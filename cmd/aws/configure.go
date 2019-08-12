package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/spf13/cobra"
)

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
	fmt.Printf("Port: %s\n", port)
	files, err := getFiles()
	if err != nil {
		return fmt.Errorf("error getting files: %v", err)
	}
	_, err = findEnvvars(files)
	if err != nil {
		return fmt.Errorf("error getting env vars: %v", err)
	}
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

func findEnvvars(files []string) ([]string, error) {
	v, err := regexp.Compile(`(((.*\n){1})(.*os\.Getenv\("([a-z,A-Z,_,0-9]+)"\)))`)
	if err != nil {
		return nil, fmt.Errorf("1st regex did not compile %v", err)
	}
	commentre, err := regexp.Compile(`.*(//).*os\.Getenv\(\"([a-z,A-Z,_,0-9]+?)\"\)`)
	if err != nil {
		return nil, fmt.Errorf("2nd regex did not compile %v", err)
	}

	base := os.Getenv("GOPATH") + "/src"
	fmt.Printf("Base: %v\n", base)

	for _, i := range files {
		here, err := filepath.Glob(fmt.Sprintf("%v/%v/*", base, i))
		if err != nil {
			return nil, fmt.Errorf("error with glob: %v", err)
		}

		for _, f := range here {
			if !strings.Contains(f, ".go") {
				continue
			}
			fi, err := os.Open(f)
			if err != nil {
				return nil, fmt.Errorf("error opening file %v: %v", f, err)
			}

			byteText := make([]byte, 32<<10)
			//Read in the bytes

			_, err = fi.Read(byteText)
			if err != nil {
				fi.Close()
				return nil, fmt.Errorf("error reading file %v: %v", f, err)
			}

			res := v.FindAll(byteText, -1)
			for _, r := range res {
				if !strings.Contains(string(r[2]), "+deploy not_required") {
					res = commentre.FindAll(r[3], -1)
					if len(res) != 0 {
						//TODO add verbose option to print that I'm skipping
						continue
					}
					//TODO add potential output flag
					//TODO add golang set to hold enviroment variables
				}
			}

		}
	}

	return nil, nil
}
