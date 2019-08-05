package flightdeck

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/byuoitav/common/db"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// FloatfleetCmd .
var FloatfleetCmd = &cobra.Command{
	Use:   "floatfleet [room ID]",
	Short: "Deploys to the room with the given ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("room ID required to deploy")
		}

		// validate that it is in the correct format
		split := strings.Split(args[0], "-")
		if len(split) != 2 {
			return fmt.Errorf("invalid room ID %s. must be in format BLDG-ROOM", args[0])
		}

		return nil
	},
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
			fmt.Printf("Error floating fleet: %v", err)
			return
		}

	},
}

func floatfleet(roomID, designation string) error {
	devices, err := db.GetDB().GetDevicesByRoom(roomID)
	if err != nil {
		return fmt.Errorf("unable to get devices from database: %s", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found in %s", roomID)
	}

	wg := sync.WaitGroup{}

	for i := range devices {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			fmt.Printf("Deploying to %s\n", devices[idx].ID)
			err := floatship(devices[idx].ID, designation)
			if err != nil {
				fmt.Printf("unable to deploy to %s: %s\n", devices[idx].ID, err)
				return
			}

			fmt.Printf("Deployed to %s\n", devices[idx].ID)
		}(i)
	}

	wg.Wait()
	return nil
}
