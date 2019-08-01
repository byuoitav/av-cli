package cmd

import (
	flightdeck "github.com/byuoitav/av-cli/cmd/flight-deck"
	"github.com/spf13/cobra"
)

func init() {
	// add all the flight deck commands here
	flightDeckCmd.AddCommand(flightdeck.DeployCmd)

	rootCmd.AddCommand(flightDeckCmd)
}

var flightDeckCmd = &cobra.Command{
	Use:   "flight-deck",
	Short: "Deploy to devices",
}
