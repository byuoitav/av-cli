package float

import (
	"fmt"
	"strings"
	"sync"

	arg "github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
)

// squadronCmd .
var squadronCmd = &cobra.Command{
	Use:   "squadron [room ID]",
	Short: "Deploys to the room with the given ID",
	Args:  arg.ValidRoomID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deploying to %s\n", args[0])

		db, result, err := arg.GetDB()
		if err != nil {
			fmt.Printf("couldn't get db: %v\n", err)
			return
		}

		err = floatsquadron(db, args[0], result)
		if err != nil {
			fmt.Printf("Error floating squadron: %v\n", err)
			return
		}
	},
}

func floatsquadron(db db.DB, roomID, designation string) error {
	devices, err := db.GetDevicesByRoom(roomID)
	if err != nil {
		return fmt.Errorf("unable to get devices from database: %s", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found in %s", roomID)
	}

	var toDeploy []structs.Device
	var bars []*pb.ProgressBar
	for _, dev := range devices {
		if dev.Type.ID == "Pi3" || dev.Type.ID == "DividerSensors" || dev.Type.ID == "LabAttendance" || dev.Type.ID == "Pi-STB" || dev.Type.ID == "SchedulingPanel" || dev.Type.ID == "TimeClock" {
			toDeploy = append(toDeploy, dev)
			bar := pb.New(6).SetWidth(50).Format(fmt.Sprintf("%s [\x00=\x00>\x00-\x00]", dev.ID))
			bar.ShowCounters = false
			bars = append(bars, bar)
		}
	}

	wg := sync.WaitGroup{}
	failedCount := 0
	failedList := ""
	pool := pb.NewPool(bars...)
	pool.Start()
	for i := range toDeploy {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			err := floatshipWithBar(toDeploy[idx].ID, designation, bars[idx])
			if err != nil {
				failedList = fmt.Sprintf("%v%v: %v\n", failedList, toDeploy[idx].ID, err)
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

func floatsquadronWithBar(db db.DB, roomID, designation string, bar *pb.ProgressBar) error {

	//1
	bar.Increment()

	devices, err := db.GetDevicesByRoom(roomID)
	if err != nil {
		return fmt.Errorf("unable to get devices from database: %s", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found in %s", roomID)
	}

	//2
	bar.Increment()

	var toDeploy []structs.Device
	for _, dev := range devices {
		if idParts := strings.Split(dev.ID, "-"); strings.Contains(strings.ToUpper(idParts[2]), "CP") {
			toDeploy = append(toDeploy, dev)
		}
	}

	//3
	bar.Increment()

	wg := sync.WaitGroup{}
	failedCount := 0
	failedList := ""

	//4
	bar.Increment()

	for i := range toDeploy {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			err := floatship(toDeploy[idx].ID, designation)
			if err != nil {
				failedList = fmt.Sprintf("%v%v: %v\n", failedList, toDeploy[idx].ID, err)
				failedCount++
				return
			}

		}(i)
	}
	//5
	bar.Increment()

	wg.Wait()
	if failedCount > 0 {
		return fmt.Errorf("%v errors: %v", failedCount, failedList)
	}

	//6
	bar.Increment()
	return nil
}
