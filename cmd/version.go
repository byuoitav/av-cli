package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string

// versionCmd .
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of av",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO do a fancy ascii art thing, get sha1 sum if version not available
		fmt.Printf("av version %s\n", version)
	},
}
