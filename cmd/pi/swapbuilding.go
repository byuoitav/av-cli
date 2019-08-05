package pi

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/byuoitav/common/db"
	"github.com/spf13/cobra"
)

// SwabBuildingCmd .
var SwabBuildingCmd = &cobra.Command{
	Use:   "swabbuilding [building ID]",
	Short: "Refreshes the database/ui of all the pi's in a building",
	Long:  "Forces a replication of the couch database, and causes the ui to refresh shortly after",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("building ID required to swab building")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Swabbing %s\n", args[0])

		// TODO add a select for the database?

		err := swabBuilding(context.TODO(), args[0])
		if err != nil {
			fmt.Printf("Couldn't swab building: %v", err)
			os.Exit(1)
		}
		// look it up in the database

		fmt.Printf("Successfully swabbed the %s\n", args[0])
	},
}

func swabBuilding(ctx context.Context, buildingID string) error {
	rooms, err := db.GetDB().GetRoomsByBuilding(buildingID)
	if err != nil {
		return fmt.Errorf("unable to get rooms from database: %s", err)
	}

	if len(rooms) == 0 {
		return fmt.Errorf("no rooms found in %s", buildingID)
	}

	wg := sync.WaitGroup{}

	for i := range rooms {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			fmt.Printf("Swabbing %s\n", rooms[idx].ID)
			err := swabRoom(context.TODO(), rooms[idx].ID)
			if err != nil {
				fmt.Printf("unable to swab %s: %s\n", rooms[idx].ID, err)
				return
			}

			fmt.Printf("Swabbed %s\n", rooms[idx].ID)
		}(i)
	}

	wg.Wait()
	return nil
}
