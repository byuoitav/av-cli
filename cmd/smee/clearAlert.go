package smee

import (
	"fmt"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/av-cli/cmd/wso2"
	"github.com/spf13/cobra"
)

var closeIssueCmd = &cobra.Command{
	Use:   "closeIssue [room ID]",
	Short: "Close a room issue in smee",
	Args:  args.ValidRoomID,
	Run: func(cmd *cobra.Command, args []string) {
		id, err := wso2.GetIDInfo()
		if err != nil {
			fmt.Printf("unable to get id info: %s\n", err)
		}

		fmt.Printf("id token: %+v\n", id)
	},
}
