package args

import (
	"fmt"
	"os"
	"strings"

	"github.com/byuoitav/common/db"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// ValidBuildingID checks if the argument is a valid building id
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

// ValidRoomID checks if the argument is a valid room id
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

// ValidDeviceID checks if the argument is a valid device id
func ValidDeviceID(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("device ID required")
	}

	split := strings.Split(args[0], "-")
	if len(split) != 3 {
		return fmt.Errorf("invalid device ID %s. must be in format BLDG-ROOM-CP1", args[0])
	}

	return nil
}

// Valid is always valid
func Valid(cmd *cobra.Command, args []string) error {
	return nil
}

// GetDB returns the database of the user's selection
func GetDB() (db.DB, string, error) {

	if os.Getenv("DB_ADDRESS") == "" {
		return nil, "", fmt.Errorf("DB_ADDRESS not set")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		return nil, "", fmt.Errorf("DB_PASSWORD not set")
	}

	dbPrompt := promptui.Select{
		Label: "Database to deploy from",
		Items: []string{"development", "stage", "production"},
	}

	_, result, err := dbPrompt.Run()
	if err != nil {
		return nil, "", fmt.Errorf("prompt failed %v", err)
	}

	var dbDesignation string
	switch result {
	case "development":
		dbDesignation = "dev"
	case "stage":
		dbDesignation = "stg"
	case "production":
		dbDesignation = "prd"
	}

	finalAddr := strings.Replace(os.Getenv("DB_ADDRESS"), "dev", dbDesignation, 1)
	finalAddr = strings.Replace(finalAddr, "stg", dbDesignation, 1)
	finalAddr = strings.Replace(finalAddr, "prd", dbDesignation, 1)

	return db.GetDBWithCustomAuth(finalAddr, dbDesignation, os.Getenv("DB_PASSWORD")), dbDesignation, nil

}
