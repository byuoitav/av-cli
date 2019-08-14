package smee

import "github.com/spf13/cobra"

func init() {
	Cmd.AddCommand(closeIssueCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "smee",
	Short: "commands for managing smee",
}
