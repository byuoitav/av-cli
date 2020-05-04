package avcli

import (
	"bytes"
	context "context"
	"encoding/json"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/auth/wso2"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	empty "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/ssh"
)

//go:generate protoc -I ./ --go_out=plugins=grpc:./ ./av-cli.proto

type Server struct {
	Logger        Logger
	DBUsername    string
	DBPassword    string
	DBAddress     string
	ShipwrightKey string

	Client *wso2.Client
}

func (s *Server) Swab(id *ID, stream AvCli_SwabServer) error {
	dbAddr := strings.Replace(s.DBAddress, "dev", id.Designation, 1)
	dbAddr = strings.Replace(dbAddr, "stg", id.Designation, 1)
	dbAddr = strings.Replace(dbAddr, "prd", id.Designation, 1)

	db := db.GetDBWithCustomAuth(dbAddr, id.Designation, s.DBPassword)

	//check if id = build, room, or device
	idChecker := strings.Split(id.Id, "-")
	switch len(idChecker) {
	case 1:
		//it's a building
		rooms, err := db.GetRoomsByBuilding(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get rooms from database: %v", err)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(rooms) == 0 {
			err = fmt.Errorf("no rooms found in %s", id.Id)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan *IDResult)
		expectedCount := 0

		for i := range rooms {
			go func(tmpRoom structs.Room) {
				devices, err := db.GetDevicesByRoom(tmpRoom.ID)
				if err != nil {
					err = fmt.Errorf("unable to get devices from database: %v", err)
					c <- &IDResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				if len(devices) == 0 {
					err = fmt.Errorf("no devices found in %s", tmpRoom.ID)
					c <- &IDResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				for x := range devices {
					tmpDevice := devices[x]

					if tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "Pi3" {
						go func() {
							err := swabDevice(context.TODO(), tmpDevice.Address)
							if err != nil {
								c <- &IDResult{
									Id:    tmpDevice.ID,
									Error: err.Error(),
								}
							} else {
								c <- &IDResult{
									Id:    tmpDevice.ID,
									Error: "",
								}
							}
						}()
						expectedCount++
					}
				}
			}(rooms[i])
		}

		actualCount := 0

		for res := range c {
			if err := stream.Send(res); err != nil {
				return err
			}

			actualCount++
			if actualCount == expectedCount {
				return nil
			}
		}

	case 2:
		//it's a room
		devices, err := db.GetDevicesByRoom(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get devices from database: %v", err)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(devices) == 0 {
			err = fmt.Errorf("no devices found in %s", id.Id)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan *IDResult)
		expectedCount := 0

		for i := range devices {
			tmpDevice := devices[i]

			if tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "Pi3" {
				go func() {
					err := swabDevice(context.TODO(), tmpDevice.Address)
					if err != nil {
						c <- &IDResult{
							Id:    tmpDevice.ID,
							Error: err.Error(),
						}
					} else {
						c <- &IDResult{
							Id:    tmpDevice.ID,
							Error: "",
						}
					}
				}()
				expectedCount++
			}
		}

		actualCount := 0

		for res := range c {
			if err := stream.Send(res); err != nil {
				return err
			}

			actualCount++
			if actualCount == expectedCount {
				return nil
			}
		}

	case 3:
		//it's a device
		device, err := db.GetDevice(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get device from database: %s\n", err)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		err = swabDevice(context.TODO(), device.Address)
		if err != nil {
			return stream.Send(&IDResult{
				Id:    device.ID,
				Error: err.Error(),
			})
		}

		return stream.Send(&IDResult{
			Id:    device.ID,
			Error: "",
		})
	}

	//we should never get here
	return stream.Send(&IDResult{
		Id:    id.Id,
		Error: "unknown id received: " + id.Id,
	})
}

func swabDevice(ctx context.Context, address string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:7012/replication/start", address), nil)
	if err != nil {
		err = fmt.Errorf("unable to build replication request: %s", err)
		return err
	}

	req = req.WithContext(ctx)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("unable to start replication: %s", err)
		return err
	}

	fmt.Printf("%s\tReplication started\n", address)
	time.Sleep(3 * time.Second)

	req, err = http.NewRequest("PUT", fmt.Sprintf("http://%s:80/refresh", address), nil)
	if err != nil {
		err = fmt.Errorf("unable to build refresh request: %s", err)
		return err
	}

	req = req.WithContext(ctx)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("unable to start replication: %s", err)
		return err
	}

	fmt.Printf("%s\tUI refreshed\n", address)

	client, err := NewSSHClient(address)
	if err != nil {
		err = fmt.Errorf("unable to ssh into %s: %s", address, err)
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		err = fmt.Errorf("unable to start new session: %s", err)
		client.Close()
		return err
	}

	fmt.Printf("%s\tRestarting DMM... \n", address)

	bytes, err := session.CombinedOutput("sudo systemctl restart device-monitoring.service")
	if err != nil {
		err = fmt.Errorf("unable to reboot: %s\noutput on pi: \n%s\n", err, bytes)
		client.Close()
		return err
	}
	client.Close()

	return nil
}

func (s *Server) Float(id *ID, stream AvCli_FloatServer) error {
	dbAddr := strings.Replace(s.DBAddress, "dev", id.Designation, 1)
	dbAddr = strings.Replace(dbAddr, "stg", id.Designation, 1)
	dbAddr = strings.Replace(dbAddr, "prd", id.Designation, 1)

	db := db.GetDBWithCustomAuth(dbAddr, id.Designation, s.DBPassword)

	//check if id = build, room, or device
	idChecker := strings.Split(id.Id, "-")
	switch len(idChecker) {
	case 1:
		//building
		rooms, err := db.GetRoomsByBuilding(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get rooms from database: %v", err)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(rooms) == 0 {
			err = fmt.Errorf("no rooms found in %s", id.Id)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan *IDResult)
		expectedCount := 0

		for i := range rooms {
			tmpRoom := rooms[i]
			go func() {
				devices, err := db.GetDevicesByRoom(tmpRoom.ID)
				if err != nil {
					err = fmt.Errorf("unable to get devices from database: %v", err)
					c <- &IDResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				if len(devices) == 0 {
					err = fmt.Errorf("no devices found in %s", tmpRoom.ID)
					c <- &IDResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				for x := range devices {
					tmpDevice := devices[x]
					if tmpDevice.Type.ID == "Pi3" || tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "LabAttendance" || tmpDevice.Type.ID == "Pi-STB" || tmpDevice.Type.ID == "SchedulingPanel" || tmpDevice.Type.ID == "TimeClock" {
						go func() {
							err := s.floatShip(tmpDevice.ID, id.Designation)
							if err != nil {
								c <- &IDResult{
									Id:    tmpDevice.ID,
									Error: err.Error(),
								}
							} else {
								c <- &IDResult{
									Id:    tmpDevice.ID,
									Error: "",
								}
							}

						}()
						expectedCount++
					}
				}
			}()
		}

		actualCount := 0
		for res := range c {
			if err := stream.Send(res); err != nil {
				return err
			}

			actualCount++
			if actualCount == expectedCount {
				return nil
			}
		}

	case 2:
		//room
		devices, err := db.GetDevicesByRoom(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get devices from database: %v", err)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(devices) == 0 {
			err = fmt.Errorf("no devices found in %s", id.Id)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan *IDResult)
		expectedCount := 0

		for i := range devices {
			tmpDevice := devices[i]
			if tmpDevice.Type.ID == "Pi3" || tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "LabAttendance" || tmpDevice.Type.ID == "Pi-STB" || tmpDevice.Type.ID == "SchedulingPanel" || tmpDevice.Type.ID == "TimeClock" {
				go func() {
					err := s.floatShip(tmpDevice.ID, id.Designation)
					if err != nil {
						c <- &IDResult{
							Id:    tmpDevice.ID,
							Error: err.Error(),
						}
					} else {
						c <- &IDResult{
							Id:    tmpDevice.ID,
							Error: "",
						}
					}
				}()
				expectedCount++
			}
		}

		actualCount := 0
		for res := range c {
			if err := stream.Send(res); err != nil {
				return err
			}

			actualCount++
			if actualCount == expectedCount {
				return nil
			}
		}

	case 3:
		//device
		err := s.floatShip(id.Id, id.Designation)
		if err != nil {
			err = fmt.Errorf("error floating ship: %v", err)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: "",
		})
	}

	return stream.Send(&IDResult{
		Id:    id.Id,
		Error: "unknown id received: " + id.Id,
	})
}

func (s *Server) floatShip(deviceID, designation string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", designation, deviceID), nil)
	if err != nil {
		return fmt.Errorf("couldn't make request: %v", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't perform request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("couldn't read the response body: %v", err)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-200 status code: %v - %s", resp.StatusCode, body)
	}

	fmt.Printf("Deployment successful\n")
	return nil
}

func (s *Server) Screenshot(ctx context.Context, id *ID) (*ScreenshotResult, error) {
	// TODO validate id
	// TODO lookup id in database, use that address

	handle := func(err error) error {
		s.warnf(err.Error())
		return err
	}

	s.infof("Taking a screenshot of %q", id.GetId())
	url := fmt.Sprintf("http://%s.byu.edu:10000/device/screenshot", id.GetId())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, handle(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, handle(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, handle(err)
	}

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		return nil, handle(fmt.Errorf("failed to get screenshot: %s", body))
	}

	return &ScreenshotResult{
		Photo: body,
	}, nil
}

func (s *Server) DuplicateRoom(ctx context.Context, req *DuplicateRoomRequest) (*empty.Empty, error) {
	//replace this crap with designation
	dbAddr := strings.Replace(s.DBAddress, "dev", req.FromDesignation, 1)
	dbAddr = strings.Replace(dbAddr, "stg", req.FromDesignation, 1)
	dbAddr = strings.Replace(dbAddr, "prd", req.FromDesignation, 1)

	srcDB := db.GetDBWithCustomAuth(dbAddr, req.FromDesignation, s.DBPassword)

	dbAddr = strings.Replace(s.DBAddress, "dev", req.ToDesignation, 1)
	dbAddr = strings.Replace(dbAddr, "stg", req.ToDesignation, 1)
	dbAddr = strings.Replace(dbAddr, "prd", req.ToDesignation, 1)

	dstDB := db.GetDBWithCustomAuth(dbAddr, req.ToDesignation, s.DBPassword)

	room, err := srcDB.GetRoom(req.FromID)
	if err != nil {
		err = fmt.Errorf("failed to get src room: %s", err)
		return &empty.Empty{}, err
	}

	devices, err := srcDB.GetDevicesByRoom(req.FromID)
	if err != nil {
		err = fmt.Errorf("failed to get src devices: %s", err)
		return &empty.Empty{}, err
	}

	uiconfig, err := srcDB.GetUIConfig(req.FromID)
	if err != nil {
		err = fmt.Errorf("failed to get src ui config: %s", err)
		return &empty.Empty{}, err
	}

	// duplicate the room
	newRoom := structs.Room{
		ID:          req.ToID,
		Name:        strings.Replace(room.Name, room.Name, req.ToID, 1),
		Description: strings.Replace(room.Description, room.ID, req.ToID, -1),
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
			ID:          strings.Replace(device.ID, room.ID, req.ToID, 1),
			Name:        device.Name,
			Address:     strings.Replace(device.Address, room.ID, req.ToID, -1),
			Description: strings.Replace(device.Description, room.ID, req.ToID, -1),
			DisplayName: strings.Replace(device.DisplayName, room.ID, req.ToID, -1),
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
				SourceDevice:      strings.Replace(port.SourceDevice, room.ID, req.ToID, 1),
				DestinationDevice: strings.Replace(port.DestinationDevice, room.ID, req.ToID, 1),
				Description:       strings.Replace(port.Description, room.ID, req.ToID, 1),
			}

			newDevice.Ports = append(newDevice.Ports, newPort)
		}

		// proxy
		for k, v := range device.Proxy {
			newDevice.Proxy[k] = strings.Replace(v, room.ID, req.ToID, -1)
		}

		newDevices = append(newDevices, newDevice)
	}

	// duplicate ui config
	newUIConfig := structs.UIConfig{
		ID:                  req.ToID,
		Api:                 []string{"localhost"},
		InputConfiguration:  uiconfig.InputConfiguration,
		OutputConfiguration: uiconfig.OutputConfiguration,
		AudioConfiguration:  uiconfig.AudioConfiguration,
		PseudoInputs:        uiconfig.PseudoInputs,
	}

	// panels
	for _, panel := range uiconfig.Panels {
		newPanel := structs.Panel{
			Hostname: strings.Replace(panel.Hostname, room.ID, req.ToID, -1),
			UIPath:   panel.UIPath,
			Preset:   panel.Preset,
			Features: panel.Features,
		}

		newUIConfig.Panels = append(newUIConfig.Panels, newPanel)
	}

	// presets
	for _, preset := range uiconfig.Presets {
		newName := preset.Name
		split := strings.Split(req.FromID, "-")
		if strings.HasPrefix(preset.Name, split[0]) {
			newSplit := strings.Split(req.ToID, "-")
			newName = strings.Replace(newName, split[0], newSplit[0], 1)
			newName = strings.Replace(newName, split[1], newSplit[1], 1)
		}
		newPreset := structs.Preset{
			Name:                    newName,
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
				Endpoint: strings.Replace(cmd.Endpoint, room.ID, req.ToID, -1),
			}

			newPreset.Commands.PowerOn = append(newPreset.Commands.PowerOn, newCmd)
		}

		for _, cmd := range preset.Commands.PowerOff {
			newCmd := structs.ConfigCommand{
				Method:   cmd.Method,
				Port:     cmd.Port,
				Endpoint: strings.Replace(cmd.Endpoint, room.ID, req.ToID, -1),
			}

			newPreset.Commands.PowerOff = append(newPreset.Commands.PowerOff, newCmd)
		}

		for _, cmd := range preset.Commands.InputSame {
			newCmd := structs.ConfigCommand{
				Method:   cmd.Method,
				Port:     cmd.Port,
				Endpoint: strings.Replace(cmd.Endpoint, room.ID, req.ToID, -1),
			}

			newPreset.Commands.InputSame = append(newPreset.Commands.InputSame, newCmd)
		}

		for _, cmd := range preset.Commands.InputDifferent {
			newCmd := structs.ConfigCommand{
				Method:   cmd.Method,
				Port:     cmd.Port,
				Endpoint: strings.Replace(cmd.Endpoint, room.ID, req.ToID, -1),
			}

			newPreset.Commands.InputDifferent = append(newPreset.Commands.InputDifferent, newCmd)
		}

		newUIConfig.Presets = append(newUIConfig.Presets, newPreset)
	}

	// write docs as tmp file
	fname := fmt.Sprintf("%s/%s->%s", os.TempDir(), req.FromID, req.ToID)
	f, err := os.Create(fname)
	if err != nil {
		err = fmt.Errorf("unable to create temp file: %s", err)
		return &empty.Empty{}, err
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

	// This is
	// validate the docs
	// less := exec.Command("less", "--prompt=Type q to exit, j/k to move down/up", fname)
	// less.Stdin = os.Stdin
	// less.Stdout = os.Stdout

	// err = less.Run()
	// if err != nil {
	// 	err = fmt.Errorf("failed to run less: %v\n", err)
	// 	return &empty.Empty{}, err
	// }

	// //confirm that the docs look good
	// prompt := promptui.Prompt{
	// 	Label:     "would you like to save these documents?",
	// 	IsConfirm: true,
	// }

	// _, err = prompt.Run()
	// if err != nil {
	// 	err = fmt.Errorf("Documents discarded")
	// 	return &empty.Empty{}, err
	// }

	// BACK TO BACKEND

	// post all of the docs
	fmt.Printf("Creating room...\n")

	_, err = dstDB.CreateRoom(newRoom)
	if err != nil {
		err = fmt.Errorf("failed to create %s (room): %s", newRoom.ID, err)
		return &empty.Empty{}, err
	}
	fmt.Printf("Created %s (room)\n", newRoom.ID)

	_, err = dstDB.CreateUIConfig(newRoom.ID, newUIConfig)
	if err != nil {
		err = fmt.Errorf("failed to create %s (uiconfig): %s\n", newUIConfig.ID, err)
		return &empty.Empty{}, err
	}
	fmt.Printf("Created %s (uiconfig)\n", newUIConfig.ID)

	for _, device := range newDevices {
		_, err = dstDB.CreateDevice(device)
		if err != nil {
			err = fmt.Errorf("failed to create %s (device): %s\n", device.ID, err)
			return &empty.Empty{}, err
		}

		fmt.Printf("Created %s (device)\n", device.ID)
	}

	return &empty.Empty{}, nil
}

func (s *Server) FixTime(id *ID, stream AvCli_FixTimeServer) error {
	client, err := NewSSHClient(id.Id + ".byu.edu")
	if err != nil {
		err = fmt.Errorf("unable to ssh into %s: %s", id.Id, err)
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: err.Error(),
		})
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		err = fmt.Errorf("unable to start new session: %s", err)
		client.Close()
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: err.Error(),
		})
	}

	fmt.Printf("Fixing time on pi...\n")

	bytes, err := session.CombinedOutput("date; sudo ntpdate tick.byu.edu && date")
	if err != nil {
		err = fmt.Errorf("unable to run fix time command: %s\noutput on pi:\n%s", err, bytes)
		client.Close()
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: err.Error(),
		})
	}

	f := func(c rune) bool {
		return c == 0x0a
	}

	split := strings.FieldsFunc(string(bytes), f)
	if len(split) != 3 {
		er := fmt.Sprintf("weird response while update time:\n%s\n", bytes)
		client.Close()
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: er,
		})
	}

	return stream.Send(&IDResult{
		Id:    id.Id,
		Error: "",
	})
}

func (s *Server) Sink(id *ID, stream AvCli_SinkServer) error {
	device, err := db.GetDB().GetDevice(id.Id)
	if err != nil {
		err = fmt.Errorf("unable to get device from db: %v", err)
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: err.Error(),
		})
	}

	client, err := NewSSHClient(device.Address)
	if err != nil {
		err = fmt.Errorf("unable to ssh into %s: %v", id.Id, err)
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: err.Error(),
		})
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		err = fmt.Errorf("unable to start new session: %v", err)
		client.Close()
		return stream.Send(&IDResult{
			Id:    id.Id,
			Error: err.Error(),
		})
	}

	fmt.Printf("Rebooting...\n")

	bytes, err := session.CombinedOutput("sudo reboot")
	if err != nil {
		switch err.(type) {
		case *ssh.ExitMissingError:
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: "",
			})
		default:
			err = fmt.Errorf("unable to reboot: %v\noutput on pi:\n%s", err, bytes)
			return stream.Send(&IDResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}
	}

	er := fmt.Sprintf("unable to reboot:\n%s\n", bytes)
	return stream.Send(&IDResult{
		Id:    id.Id,
		Error: er,
	})
}

func (s *Server) CloseMonitoringIssue(ctx context.Context, id *ID) (*empty.Empty, error) {
	if s.ShipwrightKey == "" {
		return &empty.Empty{}, fmt.Errorf("shipwright key not set")
	}
	url := fmt.Sprintf("https://smee.avs.byu.edu/issues/%s/resolve", id.Id)

	netID, err := GetNetID(ctx)
	if err != nil {
		return &empty.Empty{}, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"resolution-code": "Manual Removal",
		"notes":           fmt.Sprintf("%s manually removed room issue through av-cli", netID), //add in net id later if possible
	})
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("unable to build marshal request body: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("unable to build request: %v", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-av-access-key", s.ShipwrightKey)
	req.Header.Add("x-av-user", netID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("unable to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &empty.Empty{}, fmt.Errorf("unable to close issue; response code %v. unable to read response body: %s", resp.StatusCode, err)
		}

		return &empty.Empty{}, fmt.Errorf("unable to close issue: %s", body)
	}

	return &empty.Empty{}, nil
}

//change loglevelrequest level to string
func (s *Server) SetLogLevel(ctx context.Context, logReq *SetLogLevelRequest) (*empty.Empty, error) {
	device, err := db.GetDB().GetDevice(logReq.Id)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("unable to get device from db: %v", err)
	}

	level := ""
	switch logReq.Level {
	case -1:
		level = "debug"
	case 0:
		level = "info"
	case 1:
		level = "warn"
	case 2:
		level = "error"
	case 3:
		level = "dpanic"
	case 4:
		level = "panic"
	case 5:
		level = "fatal"
	}

	//Make port regex
	portre, err := regexp.Compile(`[\d]{4,5}`)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("error compiling port regex: %v", err)
	}

	//Match the regex
	match := portre.FindString(strconv.Itoa(int(logReq.Port)))
	if match == "" {
		return &empty.Empty{}, fmt.Errorf("Invalid port: %v", logReq.Port)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%v:%v/log-level/%s", device.Address, logReq.Port, level), nil)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("couldn't make request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("couldn't perform request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return &empty.Empty{}, fmt.Errorf("non-200 status code: %v", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("error reading body: %v", err)
	}
	fmt.Printf("Response: %s\n", b)

	return &empty.Empty{}, nil
}
