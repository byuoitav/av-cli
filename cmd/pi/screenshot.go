package pi

import (
	"fmt"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/av-cli/cmd/wso2"
	"github.com/spf13/cobra"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [device ID]",
	Short: "get a screenshot of a pi",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("http://%s:10000/device/screenshot", args[0])
		wso2.OpenBrowser(url)
	},
}
