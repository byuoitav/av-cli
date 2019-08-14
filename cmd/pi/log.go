package pi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/db"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log [device ID] [port] [log-level]",
	Short: "change a log level",
	Long:  "change a log level on a specific port on a pi",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 3 {
			fmt.Printf("missing arguments\n")
			os.Exit(1)
		}

		dev, err := db.GetDB().GetDevice(args[0])
		if err != nil {
			fmt.Printf("unable to get device from db: %v\n", err)
			os.Exit(1)
		}

		//Make port regex
		portre, err := regexp.Compile(`[\d]{4,5}`)
		if err != nil {
			fmt.Printf("error compiling port regex: %v\n", err)
			os.Exit(1)
		}
		//Match the regex
		match := portre.FindString(args[1])
		if match == "" {
			fmt.Printf("Invalid port: %v\n", args[1])
			os.Exit(1)
		}

		req, err := http.NewRequest("PUT", fmt.Sprintf("http://%v:%v/log-level/%v", dev.Address, args[1], args[2]), nil)
		if err != nil {
			fmt.Printf("couldn't make request: %v", err)
			os.Exit(1)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("couldn't perform request: %v", err)
			os.Exit(1)
		}

		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			fmt.Printf("non-200 status code: %v", resp.StatusCode)
			os.Exit(1)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading body: %v\n", err)
		}
		fmt.Printf("Response: %s", b)

	},
}
