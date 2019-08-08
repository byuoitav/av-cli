package float

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/pb/v3"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// fleetCmd .
var fleetCmd = &cobra.Command{
	Use:   "fleet [room ID]",
	Short: "Deploys to the room with the given ID",
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

		prevAddr := os.Getenv("DB_ADDRESS")
		prevName := os.Getenv("DB_USERNAME")
		finalAddr := strings.Replace(prevAddr, "dev", dbDesignation, 1)
		finalAddr = strings.Replace(finalAddr, "stg", dbDesignation, 1)
		finalAddr = strings.Replace(finalAddr, "prd", dbDesignation, 1)

		os.Setenv("DB_USERNAME", dbDesignation)
		os.Setenv("DB_ADDRESS", finalAddr)

		err = floatfleet(args[0], result)

		os.Setenv("DB_ADDRESS", prevAddr)
		os.Setenv("DB_USERNAME", prevName)
		if err != nil {
			fmt.Printf("Error floating fleet: %v\n", err)
			return
		}

	},
}

func floatfleet(roomID, designation string) error {
	devices, err := db.GetDB().GetDevicesByRoom(roomID)
	if err != nil {
		return fmt.Errorf("unable to get devices from database: %s\n", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found in %s\n", roomID)
	}

	var toDeploy []structs.Device
	var bars []*pb.ProgressBar
	for _, dev := range devices {
		if idParts := strings.Split(dev.ID, "-"); strings.Contains(strings.ToUpper(idParts[2]), "CP") {
			toDeploy = append(toDeploy, dev)
			tmpl := fmt.Sprintf(`{{ white "%v" }} {{ bar . (green "[") (green "#") (cycle . "\\" "|" "/" "-") (red "-") (green "]") }} {{ percent . | white}}`, dev.ID)
			bars = append(bars, pb.ProgressBarTemplate(tmpl).New(6).SetWidth(50))
			//bars = append(bars, pb.New(6).SetWidth(50))
		}
	}

	//tmpl := fmt.Sprintf(`{{ green "%v" }} {{ bar . "[" (blue "#") (rndcolor ">") (red "-") (white "]") }} {{speed . | yellow }} {{percent .}}`, roomID)
	//tmpl := `{{ red "With funcs:" }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{speed . | rndcolor }} {{percent .}}`

	wg := sync.WaitGroup{}
	failedCount := 0
	failedList := ""
	pool := pb.NewPool(bars...)
	pool.Start()
	for i := range toDeploy {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			fmt.Printf("Deploying to %s\n", toDeploy[idx].ID)
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
	fmt.Printf("%v failures: %v\n", failedCount, failedList)

	return nil
}
