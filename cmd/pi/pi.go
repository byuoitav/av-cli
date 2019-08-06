package pi

import "github.com/spf13/cobra"

func init() {
	Cmd.AddCommand(fixTimeCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "pi",
	Short: "commands for managing pi's",
}
