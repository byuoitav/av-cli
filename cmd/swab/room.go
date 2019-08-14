package swab

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/db"
	"github.com/spf13/cobra"
)

// swabRoomCmd .
var swabRoomCmd = &cobra.Command{
	Use:   "room [room ID]",
	Short: "Refreshes the database/ui of all the pi's in a room",
	Long:  "Forces a replication of the couch database, and causes the ui to refresh shortly after",
	Args:  args.ValidRoomID,
	Run: func(cmd *cobra.Command, arg []string) {
		fmt.Printf("Swabbing %s\n", arg[0])

		db, _, err := args.GetDB()
		if err != nil {
			fmt.Printf("unable to get database: %v", err)
			os.Exit(1)
		}

		err = swabRoom(context.TODO(), db, arg[0])
		if err != nil {
			fmt.Printf("Couldn't swab room: %v", err)
			os.Exit(1)
		}
		// look it up in the database

		fmt.Printf("Successfully swabbed %s\n", arg[0])
	},
}

func swabRoom(ctx context.Context, db db.DB, roomID string) error {
	devices, err := db.GetDevicesByRoom(roomID)
	if err != nil {
		return fmt.Errorf("unable to get devices from database: %s", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found in %s", roomID)
	}

	wg := sync.WaitGroup{}

	for i := range devices {
		if devices[i].Type.ID == "DividerSensors" || devices[i].Type.ID == "Pi3" {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()
				fmt.Printf("Swabbing %s\n", devices[idx].ID)
				err := swab(ctx, devices[idx].Address)
				if err != nil {
					fmt.Printf("unable to swab %s: %s\n", devices[idx].ID, err)
					return
				}

				fmt.Printf("Swabbed %s\n", devices[idx].ID)
			}(i)
		}
	}

	wg.Wait()
	return nil
}
