package pi

import (
	"fmt"
	"strings"

	"github.com/byuoitav/av-cli/cmd/wso2"
	"github.com/spf13/cobra"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [device ID]",
	Short: "get a screenshot of a pi",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("device ID required to get screenshot")
		}

		// validate that it is in the correct format
		split := strings.Split(args[0], "-")
		if len(split) != 3 {
			return fmt.Errorf("invalid device ID %s. must be in format BLDG-ROOM-CP1", args[0])
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("http://%s:10000/device/screenshot", args[0])
		wso2.OpenBrowser(url)
	},
}
