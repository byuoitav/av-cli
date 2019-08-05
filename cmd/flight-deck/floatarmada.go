package flightdeck

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/byuoitav/common/db"
	"github.com/spf13/cobra"
)

const (
	dev  = "development"
	stg  = "stage"
	prd  = "production"
	test = "testing"
)

// FloatarmadaCmd .
var FloatarmadaCmd = &cobra.Command{
	Use:   "floatarmada [designation ID]",
	Short: "Deploys to all rooms with the given designation",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("designation ID required to deploy")
		}

		// validate that it is in the correct format
		if args[0] != dev && args[0] != stg && args[0] != prd && args[0] != test {
			return fmt.Errorf("invalid designation")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deploying to all %s rooms\n", args[0])

		var dbDesignation string
		switch args[0] {
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

		err := floatarmada(args[0])

		os.Setenv("DB_ADDRESS", prevAddr)
		os.Setenv("DB_USERNAME", prevName)
		if err != nil {
			fmt.Printf("Error floating armada: %v", err)
			return
		}

	},
}

func floatarmada(designation string) error {
	rooms, err := db.GetDB().GetRoomsByDesignation(designation)
	if err != nil {
		return fmt.Errorf("unable to get rooms from database: %s", err)
	}

	if len(rooms) == 0 {
		return fmt.Errorf("no %s rooms found", designation)
	}

	wg := sync.WaitGroup{}

	for i := range rooms {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			fmt.Printf("Deploying to %s\n", rooms[idx].ID)
			err := floatfleet(rooms[idx].ID, designation)
			if err != nil {
				fmt.Printf("unable to deploy to %s: %s\n", rooms[idx].ID, err)
				return
			}

			fmt.Printf("Deployed to %s\n", rooms[idx].ID)
		}(i)
	}

	wg.Wait()
	return nil
}
