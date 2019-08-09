package aws

import "github.com/spf13/cobra"

func init() {
	Cmd.AddCommand(configureCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "aws",
	Short: "commands for managing aws",
}
