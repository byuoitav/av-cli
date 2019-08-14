package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates av-cli to the newest version",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func update() {
	os.Getenv("")
}
