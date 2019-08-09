package float

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/db"
	"github.com/cheggaaa/pb"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// fleetCmd .
var fleetCmd = &cobra.Command{
	Use:   "fleet [building ID]",
	Short: "Deploys to the building with the given ID",
	Args:  args.ValidRoomID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deploying to %s\n", args[0])

		dbPrompt := promptui.Select{
			Label: "Database to deploy from",
			Items: []string{"development", "stage", "production"},
		}

		_, result, err := dbPrompt.Run()
		if err != nil {
			fmt.Printf("prompt failed %v\n", err)
		}

		var dbDesignation string
		switch result {
		case "development":
			dbDesignation = "dev"
		case "stage":
			dbDesignation = "stg"
		case "production":
			dbDesignation = "prd"
		}

		finalAddr := strings.Replace(os.Getenv("DB_ADDRESS"), "dev", dbDesignation, 1)
		finalAddr = strings.Replace(finalAddr, "stg", dbDesignation, 1)
		finalAddr = strings.Replace(finalAddr, "prd", dbDesignation, 1)

		db := db.GetDBWithCustomAuth(finalAddr, dbDesignation, os.Getenv("DB_PASSWORD"))

		err = floatfleet(db, args[0], result)
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
