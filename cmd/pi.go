package cmd

import (
	"github.com/byuoitav/av-cli/cmd/pi"
	"github.com/spf13/cobra"
)

func init() {
	// add all the flight deck commands here
	piCmd.AddCommand(pi.SwabCmd)

	rootCmd.AddCommand(piCmd)
}

var piCmd = &cobra.Command{
	Use:   "pi",
	Short: "For management of raspberry pi's",
}
