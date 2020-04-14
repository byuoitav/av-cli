package pi

import (
	"fmt"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/av-cli/cmd/wso2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [device ID]",
	Short: "get a screenshot of a pi",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("http://%s:10000/device/screenshot", args[0])
		err := wso2.OpenBrowser(url)
		if err != nil {
			fmt.Printf("Unable to open browser: %s. Copy the below URL into your browser to see your screenshot:\n%s\n", err, color.New(color.FgBlue, color.Bold, color.Underline).Sprint(url))
		}
	},
}
