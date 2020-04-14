package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	arg "github.com/byuoitav/av-cli/cmd/args"
	"github.com/byuoitav/common/structs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var dupCmd = &cobra.Command{
	Use:   "dup [dst room ID] [src room ID]",
	Short: "Duplicate a room",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("must include destination and source room ID")
		}

		if err := arg.ValidRoomID(cmd, args); err != nil {
			return fmt.Errorf("invalid destination room: %v", err)
		}

		if err := arg.ValidRoomID(cmd, args[1:]); err != nil {
			return fmt.Errorf("invalid source room: %v", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Select source database:\n")
		db, _, err := arg.GetDB()
		if err != nil {
			fmt.Printf("prompt failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSelect dest database:\n")
		dbDst, _, err := arg.GetDB()
		if err != nil {
			fmt.Printf("prompt failed: %s\n", err)
			os.Exit(1)
		}

		dst := strings.ToUpper(args[0])
		src := args[1]

		// TODO make sure destination building is valid
		// TODO make sure dest room doesn't exist already

		// get docs from current room
		room, err := db.GetRoom(src)
		if err != nil {
			fmt.Printf("failed to get src room: %s\n", err)
			os.Exit(1)
		}

		devices, err := db.GetDevicesByRoom(src)
		if err != nil {
			fmt.Printf("failed to get src devices: %s\n", err)
			os.Exit(1)
		}

		uiconfig, err := db.GetUIConfig(src)
		if err != nil {
			fmt.Printf("failed to get src ui config: %s\n", err)
			os.Exit(1)
		}

		// duplicate the room
		newRoom := structs.Room{
			ID:          dst,
			Name:        strings.Replace(room.Name, room.Name, dst, 1),
			Description: strings.Replace(room.Description, room.ID, dst, -1),
			Configuration: structs.RoomConfiguration{
				ID: room.Configuration.ID,
			},
			Designation: room.Designation,
			Tags:        room.Tags,
			Attributes:  room.Attributes,
		}

		// duplicate each device
		var newDevices []structs.Device
		for _, device := range devices {
			newDevice := structs.Device{
				ID:          strings.Replace(device.ID, room.ID, dst, 1),
				Name:        device.Name,
				Address:     strings.Replace(device.Address, room.ID, dst, -1),
				Description: strings.Replace(device.Description, room.ID, dst, -1),
				DisplayName: strings.Replace(device.DisplayName, room.ID, dst, -1),
				Type: structs.DeviceType{
					ID: device.Type.ID,
				},
				Roles: device.Roles,
				Proxy: make(map[string]string),
			}

			// ports
			for _, port := range device.Ports {
				newPort := structs.Port{
					ID:                port.ID,
					FriendlyName:      port.FriendlyName,
					SourceDevice:      strings.Replace(port.SourceDevice, room.ID, dst, 1),
					DestinationDevice: strings.Replace(port.DestinationDevice, room.ID, dst, 1),
					Description:       strings.Replace(port.Description, room.ID, dst, 1),
				}

				newDevice.Ports = append(newDevice.Ports, newPort)
			}

			// proxy
			for k, v := range device.Proxy {
				newDevice.Proxy[k] = strings.Replace(v, room.ID, dst, -1)
			}

			newDevices = append(newDevices, newDevice)
		}

		// duplicate ui config
		newUIConfig := structs.UIConfig{
			ID:                  dst,
			Api:                 []string{"localhost"}, // i think this is what we always want now...
			InputConfiguration:  uiconfig.InputConfiguration,
			OutputConfiguration: uiconfig.OutputConfiguration,
			AudioConfiguration:  uiconfig.AudioConfiguration,
			PseudoInputs:        uiconfig.PseudoInputs,
		}

		// panels
		for _, panel := range uiconfig.Panels {
			newPanel := structs.Panel{
				Hostname: strings.Replace(panel.Hostname, room.ID, dst, -1),
				UIPath:   panel.UIPath,
				Preset:   panel.Preset,
				Features: panel.Features,
			}

			newUIConfig.Panels = append(newUIConfig.Panels, newPanel)
		}

		// presets
		for _, preset := range uiconfig.Presets {
			newPreset := structs.Preset{
				Name:                    preset.Name,
				Icon:                    preset.Icon,
				Displays:                preset.Displays,
				AudioDevices:            preset.AudioDevices,
				ShareablePresets:        preset.ShareablePresets,
				Inputs:                  preset.Inputs,
				VolumeMatches:           preset.VolumeMatches,
				IndependentAudioDevices: preset.IndependentAudioDevices,
				AudioGroups:             preset.AudioGroups,
			}

			// commands
			for _, cmd := range preset.Commands.PowerOn {
				newCmd := structs.ConfigCommand{
					Method:   cmd.Method,
					Port:     cmd.Port,
					Endpoint: strings.Replace(cmd.Endpoint, room.ID, dst, -1),
					// TODO BODY
				}

				newPreset.Commands.PowerOn = append(newPreset.Commands.PowerOn, newCmd)
			}

			for _, cmd := range preset.Commands.PowerOff {
				newCmd := structs.ConfigCommand{
					Method:   cmd.Method,
					Port:     cmd.Port,
					Endpoint: strings.Replace(cmd.Endpoint, room.ID, dst, -1),
					// TODO BODY
				}

				newPreset.Commands.PowerOff = append(newPreset.Commands.PowerOff, newCmd)
			}

			for _, cmd := range preset.Commands.InputSame {
				newCmd := structs.ConfigCommand{
					Method:   cmd.Method,
					Port:     cmd.Port,
					Endpoint: strings.Replace(cmd.Endpoint, room.ID, dst, -1),
					// TODO BODY
				}

				newPreset.Commands.InputSame = append(newPreset.Commands.InputSame, newCmd)
			}

			for _, cmd := range preset.Commands.InputDifferent {
				newCmd := structs.ConfigCommand{
					Method:   cmd.Method,
					Port:     cmd.Port,
					Endpoint: strings.Replace(cmd.Endpoint, room.ID, dst, -1),
					// TODO BODY
				}

				newPreset.Commands.InputDifferent = append(newPreset.Commands.InputDifferent, newCmd)
			}

			newUIConfig.Presets = append(newUIConfig.Presets, newPreset)
		}

		// write docs as tmp file
		fname := fmt.Sprintf("%s/%s->%s", os.TempDir(), src, dst)
		f, err := os.Create(fname)
		if err != nil {
			fmt.Printf("unable to create temp file: %s\n", err)
			os.Exit(1)
		}

		// write all of the docs to stdin
		_, _ = f.Write([]byte(fmt.Sprintf("******Room doc******\n")))
		buf, err := json.MarshalIndent(newRoom, "", "  ")
		if err != nil {
			_, _ = f.Write([]byte(fmt.Sprintf("unable to marshal room doc: %s\n", err)))
		} else {
			_, _ = f.Write(buf)
		}

		_, _ = f.Write([]byte(fmt.Sprintf("\n\n******Device docs******\n")))
		for _, device := range newDevices {
			buf, err = json.MarshalIndent(device, "", "  ")
			if err != nil {
				_, _ = f.Write([]byte(fmt.Sprintf("unable to marshal device doc for %q: %s\n\n", device.ID, err)))
			} else {
				_, _ = f.Write(buf)
				_, _ = f.Write([]byte("\n\n"))
			}
		}

		_, _ = f.Write([]byte(fmt.Sprintf("\n\n******UIConfig doc******\n")))
		buf, err = json.MarshalIndent(newUIConfig, "", "  ")
		if err != nil {
			_, _ = f.Write([]byte(fmt.Sprintf("unable to marshal ui config doc: %s\n", err)))
		} else {
			_, _ = f.Write(buf)
		}

		_, _ = f.Write([]byte("\n"))
		f.Close()

		// validate the docs
		less := exec.Command("less", "--prompt=Type q to exit, j/k to move down/up", fname)
		less.Stdin = os.Stdin
		less.Stdout = os.Stdout

		err = less.Run()
		if err != nil {
			fmt.Printf("failed to run less: %v\n", err)
			os.Exit(1)
		}

		// confim that the docs look good
		prompt := promptui.Prompt{
			Label:     "Would you like to save these documents?",
			IsConfirm: true,
		}

		_, err = prompt.Run()
		if err != nil {
			fmt.Printf("Documents discarded.\n")
			os.Exit(0)
		}

		// post all of the docs!
		fmt.Printf("Creating room...\n")

		_, err = dbDst.CreateRoom(newRoom)
		if err != nil {
			fmt.Printf("failed to create %s (room): %s\n", newRoom.ID, err)
			os.Exit(1)
		}
		fmt.Printf("Created %s (room)\n", newRoom.ID)

		_, err = dbDst.CreateUIConfig(newRoom.ID, newUIConfig)
		if err != nil {
			fmt.Printf("failed to create %s (uiconfig): %s\n", newUIConfig.ID, err)
			os.Exit(1)
		}
		fmt.Printf("Created %s (uiconfig)\n", newUIConfig.ID)

		for _, device := range newDevices {
			_, err = dbDst.CreateDevice(device)
			if err != nil {
				fmt.Printf("failed to create %s (device): %s\n", device.ID, err)
				os.Exit(1)
			}

			fmt.Printf("Created %s (device)\n", device.ID)
		}

		fmt.Printf("Successfully duplicated %s from %s\n", dst, src)
	},
}
