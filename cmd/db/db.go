package db

import "github.com/spf13/cobra"

func init() {
	Cmd.AddCommand(dupCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "db",
	Short: "",
}
