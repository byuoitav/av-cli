package aws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/spf13/cobra"
)

type jsonConfig struct {
	Name    string   `json:"name"`
	Port    string   `json:"port"`
	EnvVars []string `json:"env-vars,omitempty"`
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "creates the config.json file with the necessary enviroment variables for aws",
	Args:  args.Valid,
	Run: func(cmd *cobra.Command, args []string) {
		//TODO add command flags and stuff
		err := configure()
		if err != nil {
			fmt.Printf("Failed to generate config.json: %v", err)
		}
	},
}

func configure() error {
	port, err := getPort()
	if err != nil {
		return fmt.Errorf("couldn't get port: %v", err)
	}
	if port == "" {
		return fmt.Errorf("couldn't find port")
	}

	if Verbose {
		fmt.Printf("Port: %v\n", port)
	}

	files, err := getFiles()
	if err != nil {
		return fmt.Errorf("error getting files: %v", err)
	}
	envVars, useCases, err := findEnvvars(files)
	if err != nil {
		return fmt.Errorf("error getting env vars: %v", err)
	}
	sort.Strings(envVars)
	if Verbose {
		for _, k := range envVars {
			printy := strings.Split(useCases[k], "byuoitav/")[1]
			fmt.Printf("In %v \"%v\": \"%v\"\n", printy, k, os.Getenv(k))
		}
		fmt.Printf("\n")
	}
	err = writeJSON(port, envVars)
	if err != nil {
		return fmt.Errorf("error writing to json: %v", err)
	}
	fmt.Printf("Done\n")
	return nil
}

func getPort() (string, error) {
	//Make port regex
	portre, err := regexp.Compile(`port := ":([\d]{4,5})"`)
	if err != nil {
		return "", fmt.Errorf("error compiling port regex: %v", err)
	}

	//Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %v", err)
	}

	//Get the files in the cwd
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		return "", fmt.Errorf("could not read the files in %v: %v", cwd, err)
	}

	//Loop over the files
	for _, f := range files {
		//If one of the main file options
		if f.Name() == "server.go" || f.Name() == "main.go" {
			//Open it
			fi, err := os.Open(fmt.Sprintf("%v/%v", cwd, f.Name()))
			if err != nil {
				return "", fmt.Errorf("could not open %v: %v", f.Name(), err)
			}
			defer fi.Close()

			byteText := make([]byte, f.Size())
			//Read in the bytes
			_, err = fi.Read(byteText)
			if err != nil {
				return "", fmt.Errorf("error reading file %v: %v", f.Name(), err)
			}

			//Match the regex
			match := portre.FindSubmatch(byteText)
			if match == nil {
				fmt.Printf("no matches, moving on\n")
				continue
			} else {
				return string(match[1]), nil
			}

		}
	}
	fmt.Printf("nothing found\n")
	return "", nil
}

func getFiles() ([]string, error) {
	out, err := exec.Command("go", `list`, `-f`, `'{{ join .Deps "\n" }}'`).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing go list: %v %s", err, out)
	}
	preFilter := strings.Split(string(out[:]), "\n")
	var toReturn []string
	for _, s := range preFilter {
		if strings.Contains(s, "byuoitav") {
			toReturn = append(toReturn, s)
		}
	}
	return toReturn, nil
}

func findEnvvars(files []string) ([]string, map[string]string, error) {
	v, err := regexp.Compile(`(((.*\n){1})(.*os\.Getenv\("([a-z,A-Z,_,0-9]+)"\)))`)
	if err != nil {
		return nil, nil, fmt.Errorf("1st regex did not compile %v", err)
	}
	commentre, err := regexp.Compile(`.*(\/\/).*os\.Getenv\(\"([a-z,A-Z,_,0-9]+?)\"\)`)
	if err != nil {
		return nil, nil, fmt.Errorf("2nd regex did not compile %v", err)
	}

	base := os.Getenv("GOPATH") + "/src"
	envVars := make(map[string]bool)
	useCases := make(map[string]string)

	//Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting working directory: %v", err)
	}

	files = append(files, strings.Split(cwd, "/src/")[1])

	for _, i := range files {
		here, err := filepath.Glob(fmt.Sprintf("%v/%v/*", base, i))
		if err != nil {
			return nil, nil, fmt.Errorf("error with glob: %v", err)
		}

		for _, f := range here {
			if !strings.Contains(f, ".go") {
				continue
			}
			fi, err := os.Open(f)
			if err != nil {
				return nil, nil, fmt.Errorf("error opening file %v: %v", f, err)
			}
			defer fi.Close()

			byteText := make([]byte, 32<<10)
			//Read in the bytes

			_, err = fi.Read(byteText)
			if err != nil {
				return nil, nil, fmt.Errorf("error reading file %v: %v", f, err)
			}

			res := v.FindAllSubmatch(byteText, -1)
			for _, r := range res {
				if !strings.Contains(string(r[2]), "+deploy not_required") {
					resp := commentre.FindAllSubmatch(r[0], -1)

					if len(resp) != 0 {
						continue
					}
					//TODO add potential output flag
					envVars[string(r[5])] = true
					useCases[string(r[5])] = f

				}
			}

		}
	}
	var envList []string
	for k := range envVars {
		envList = append(envList, k)
	}
	return envList, useCases, nil
}

func writeJSON(port string, envList []string) error {

	//Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %v", err)
	}

	configPath := fmt.Sprintf("%v/config.json", cwd)

	if err := os.RemoveAll(configPath); err != nil {
		return fmt.Errorf("error deleting config.json: %v", err)
	}

	var jConf jsonConfig
	split := strings.Split(cwd, "/")
	jConf.Name = split[len(split)-1]
	jConf.Port = port
	jConf.EnvVars = envList

	configFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}
	defer configFile.Close()

	bytes, err := json.MarshalIndent(jConf, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshalling json: %v", err)
	}
	_, err = configFile.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing to config file: %v", err)
	}

	return nil
}
