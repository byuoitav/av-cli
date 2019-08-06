package cmd

import (
	flightdeck "github.com/byuoitav/av-cli/cmd/flight-deck"
	"github.com/spf13/cobra"
)

func init() {
	floatCmd.AddCommand(flightdeck.ShipCmd)
	floatCmd.AddCommand(flightdeck.FleetCmd)
	floatCmd.AddCommand(flightdeck.ArmadaCmd)
}

var floatCmd = &cobra.Command{
	Use:   "float",
	Short: "",
}
