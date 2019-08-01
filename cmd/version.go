package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of av",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO do a fancy ascii art thing
		fmt.Printf("av version 0.0.1\n")
	},
}
