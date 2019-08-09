package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

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
	fmt.Printf("Port: %s", port)
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
