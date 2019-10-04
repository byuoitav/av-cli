package float

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	arg "github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/av-cli/cmd/wso2"
	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
)

// shipCmd .
var shipCmd = &cobra.Command{
	Use:   "ship [device ID]",
	Short: "Deploys to the device with the given ID",
	Args:  arg.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deploying to %s\n", args[0])

		_, result, err := arg.GetDB()
		if err != nil {
			fmt.Printf("prompt failed %v\n", err)
			os.Exit(1)
		}

		bar := pb.New(6).SetWidth(50).Format(fmt.Sprintf("%s [\x00=\x00>\x00-\x00]", args[0]))
		bar.ShowCounters = false
		bar.Start()
		err = floatshipWithBar(args[0], result, bar)
		if err != nil {
			fmt.Printf("Error floating ship: %v\n", err)
			return
		}

	},
}

func floatship(deviceID, designation string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", designation, deviceID), nil)
	if err != nil {
		return fmt.Errorf("couldn't make request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", wso2.GetAccessToken()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't perform request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("couldn't read the response body: %v", err)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-200 status code: %v - %s", resp.StatusCode, body)
	}

	fmt.Printf("Deployment successful\n")
	return nil
}

func floatshipWithBar(deviceID, designation string, bar *pb.ProgressBar) error {
	//1
	bar.Increment()

	//2
	bar.Increment()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", designation, deviceID), nil)
	if err != nil {
		return fmt.Errorf("couldn't make request: %v", err)
	}

	//3
	bar.Increment()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", wso2.GetAccessToken()))

	//4
	bar.Increment()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't perform request: %v", err)
	}
	//5
	bar.Increment()

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("couldn't read the response body: %v", err)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-200 status code: %v - %s", resp.StatusCode, body)
	}
	//6
	bar.Increment()
	bar.Finish()
	return nil
}
