package pi

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/byuoitav/common/db"
	"github.com/spf13/cobra"
)

// SwabRoomCmd .
var SwabRoomCmd = &cobra.Command{
	Use:   "swabroom [room ID]",
	Short: "Refreshes the database/ui of all the pi's in a room",
	Long:  "Forces a replication of the couch database, and causes the ui to refresh shortly after",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("room ID required to swab room")
		}

		// validate that it is in the correct format
		split := strings.Split(args[0], "-")
		if len(split) != 2 {
			return fmt.Errorf("invalid room ID %s. must be in format BLDG-ROOM", args[0])
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Swabbing %s\n", args[0])

		// TODO add a select for the database?

		// look it up in the database
		devices, err := db.GetDB().GetDevicesByRoom(args[0])
		if err != nil {
			fmt.Printf("unable to get devices from database: %s\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Printf("No devices found in %s\n", args[0])
			os.Exit(1)
		}

		wg := sync.WaitGroup{}

		for i := range devices {
			if devices[i].Type.ID == "DividerSensors" || devices[i].Type.ID == "Pi3" {
				wg.Add(1)

				go func(idx int) {
					defer wg.Done()
					fmt.Printf("Swabbing %s\n", devices[idx].ID)
					err := swab(context.TODO(), devices[idx].Address)
					if err != nil {
						fmt.Printf("unable to swab %s: %s\n", devices[idx].ID, err)
						return
					}

					fmt.Printf("Swabbed %s\n", devices[idx].ID)
				}(i)
			}
		}

		wg.Wait()
		fmt.Printf("Successfully swabbed %s\n", args[0])
	},
}
