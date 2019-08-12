package float

import (
	"fmt"
	"sync"

	arg "github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/pb"
	"github.com/spf13/cobra"
)

// fleetCmd .
var fleetCmd = &cobra.Command{
	Use:   "fleet [building ID]",
	Short: "Deploys to the building with the given ID",
	Args:  arg.ValidRoomID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deploying to %s\n", args[0])

		db, designation, err := arg.GetDB()
		if err != nil {
			fmt.Printf("Error getting db: %v\n", err)
			return
		}

		err = floatfleet(db, args[0], designation)
		if err != nil {
			fmt.Printf("Error floating fleet: %v\n", err)
			return
		}
	},
}

func floatfleet(db db.DB, buildingID, designation string) error {
	rooms, err := db.GetRoomsByBuilding(buildingID)
	if err != nil {
		return fmt.Errorf("unable to get rooms from database: %s", err)
	}

	if len(rooms) == 0 {
		return fmt.Errorf("no rooms found in %s", buildingID)
	}

	var bars []*pb.ProgressBar
	for _, room := range rooms {
		devs, err := db.GetDevicesByRoom(room.ID)
		if err != nil {
			return fmt.Errorf("couldn't get devices for room %v: %v", room.ID, err)
		}
		bar := pb.New(len(devs) + 6).SetWidth(50).Format(fmt.Sprintf("%s [\x00=\x00>\x00-\x00]", room.ID))
		bar.ShowCounters = false
		bars = append(bars, bar)

	}

	wg := sync.WaitGroup{}

	failedCount := 0
	failedList := ""
	pool := pb.NewPool(bars...)
	pool.Start()

	for i := range rooms {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			err := floatsquadronWithBar(db, rooms[idx].ID, designation, bars[idx])
			if err != nil {
				failedList = fmt.Sprintf("%v%v: %v\n", failedList, rooms[idx].ID, err)
				failedCount++
				bars[idx].Finish()
				return
			}

		}(i)
	}
	wg.Wait()
	pool.Stop()

	fmt.Printf("%v failures:\n%v\n", failedCount, failedList)

	return nil
}
