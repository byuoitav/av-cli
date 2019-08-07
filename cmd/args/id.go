package args

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ValidBuildingID .
func ValidBuildingID(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("building ID required")
	}

	split := strings.Split(args[0], "-")
	if len(split) != 1 {
		return fmt.Errorf("invalid building ID %s. must be in format BLDG", args[0])
	}

	return nil
}

// ValidRoomID .
func ValidRoomID(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("room ID required")
	}

	split := strings.Split(args[0], "-")
	if len(split) != 2 {
		return fmt.Errorf("invalid room ID %s. must be in format BLDG-ROOM", args[0])
	}

	return nil
}

// ValidDeviceID .
func ValidDeviceID(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("device ID required required")
	}

	split := strings.Split(args[0], "-")
	if len(split) != 3 {
		return fmt.Errorf("invalid device ID %s. must be in format BLDG-ROOM-CP1", args[0])
	}

	return nil
}
