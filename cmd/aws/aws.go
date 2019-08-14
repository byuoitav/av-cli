package aws

import "github.com/spf13/cobra"

// Verbose controls if the verbosity of aws
var Verbose bool

func init() {
	configureCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	Cmd.AddCommand(configureCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "aws",
	Short: "commands for managing aws",
}
