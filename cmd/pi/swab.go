package pi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/av-cli/cmd/wso2"
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
		/*
		token, err := wso2.GetWSO2Token()
		if err != nil {
			fmt.Printf("unable to get wso2 token: %s\n", err)
			os.Exit(1)
		}
		*/

		fmt.Printf("token: %s\n", token)
			fmt.Printf("Swabbing %s\n", args[0])

			// TODO add a select for the database?

			// look it up in the database
			device, err := db.GetDB().GetDevice(args[0])
			if err != nil {
				fmt.Printf("unable to get device from database: %s\n", err)
				os.Exit(1)
			}

			err = swab(context.TODO(), device.Address)
			if err != nil {
				fmt.Printf("unable to swab %s: %s\n", device.ID, err)
				os.Exit(1)
			}

			fmt.Printf("Successfully swabbed %s\n", device.ID)
	},
}

func swab(ctx context.Context, address string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:7012/replication/start", address), nil)
	if err != nil {
		return fmt.Errorf("unable to build replication request: %s", err)
	}

	req = req.WithContext(ctx)

	// TODO actually validate that it worked
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to start replication: %s", err)
	}

	fmt.Printf("%s\tReplication started\n", address)
	time.Sleep(3 * time.Second) // TODO make sure this doesn't overrun ctx

	req, err = http.NewRequest("PUT", fmt.Sprintf("http://%s:8888/refresh", address), nil)
	if err != nil {
		return fmt.Errorf("unable to build refresh request: %s", err)
	}

	req = req.WithContext(ctx)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to start replication: %s", err)
	}

	fmt.Printf("%s\tUI refreshed\n", address)
	return nil
}
