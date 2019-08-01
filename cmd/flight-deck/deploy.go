package flightdeck

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// DeployCmd .
var DeployCmd = &cobra.Command{
	Use:   "deploy [device ID]",
	Short: "Deploys to the device with the given ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("device ID required to deploy")
		}

		// validate that it is in the correct format
		split := strings.Split(args[0], "-")
		if len(split) != 3 {
			return fmt.Errorf("invalid device ID %s. must be in format BLDG-ROOM-CP1", args[0])
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

		fmt.Printf("result: %s", result)

		// use result to build flight-deck addr
		// hit webhook_deploy/
	},
}
