package smee

import "github.com/spf13/cobra"

var closeIssueCmd = &cobra.Command{
	Use:   "closeIssue [room ID]",
	Short: "Close a room issue in smee",
}
