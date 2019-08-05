package flightdeck

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/byuoitav/av-cli/cmd/wso2"
	"github.com/cheggaaa/pb/v3"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// FloatshipCmd .
var FloatshipCmd = &cobra.Command{
	Use:   "floatship [device ID]",
	Short: "Deploys to the device with the given ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("device ID required to deploy")
		}

		// validate that it is in the correct format
		split := strings.Split(args[0], "-")
		if len(split) != 3 {
			return fmt.Errorf("invalid device ID %s. must be in format BLDG-ROOM-CP#", args[0])
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

		fmt.Printf("result: %s\n", result)

		err = floatship(args[0], result)
		if err != nil {
			fmt.Printf("Error floating ship: %v", err)
			return
		}

		// use result to build flight-deck addr
		// hit webhook_deploy/
	},
}

func floatship(deviceID, designation string) error {
	count := 7
	bar := pb.StartNew(count)

	var dbDesignation string
	switch designation {
	case "development":
		dbDesignation = "dev"
	case "stage":
		dbDesignation = "stg"
	case "production":
		dbDesignation = "prd"
	}

	//1
	bar.Increment()

	//2
	bar.Increment()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", dbDesignation, deviceID), nil)
	if err != nil {
		return fmt.Errorf("Couldn't make request: %v", err)
	}

	//3
	bar.Increment()

	//4
	bar.Increment()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", wso2.GetToken()))

	//5
	bar.Increment()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Couldn't perform request: %v", err)
	}
	defer resp.Body.Close()

	//6
	bar.Increment()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Non-200 status code: %v", resp.StatusCode)
	}

	//7
	bar.Increment()
	bar.Finish()

	fmt.Printf("Deployment successful\n")
	return nil
}
