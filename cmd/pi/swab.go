package pi

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/spf13/cobra"
)

// SwabCmd .
var SwabCmd = &cobra.Command{
	Use:   "swab [device ID]",
	Short: "Refreshes the database/ui of a pi",
	Long:  "Forces a replication of the couch database, and causes the ui to refresh shortly after",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("device ID required to swab")
		}

		// validate that it is in the correct format
		split := strings.Split(args[0], "-")
		if len(split) != 3 {
			return fmt.Errorf("invalid device ID %s. must be in format BLDG-ROOM-CP1", args[0])
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Swabbing %s\n", args[0])

		// TODO add a select for the database?

		// look it up in the database
		device, err := db.GetDB().GetDevice(args[0])
		if err != nil {
			fmt.Printf("unable to get device from database: %s\n", err)
			os.Exit(1)
		}

		client := http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:7012/replication/start", device.Address), nil)
		if err != nil {
			fmt.Printf("unable to build replication request: %s\n", err)
			os.Exit(1)
		}

		// TODO actually validate that it worked
		_, err = client.Do(req)
		if err != nil {
			fmt.Printf("unable to start replication: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Replication started\n")
		time.Sleep(3 * time.Second)

		req, err = http.NewRequest("PUT", fmt.Sprintf("http://%s:8888/refresh", device.Address), nil)
		if err != nil {
			fmt.Printf("unable to build refresh request: %s\n", err)
			os.Exit(1)
		}

		_, err = client.Do(req)
		if err != nil {
			fmt.Printf("unable to refresh pi: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("UI refreshed\n")
		fmt.Printf("Done!\n")
	},
}
