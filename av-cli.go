package avcli

import (
	context "context"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	empty "github.com/golang/protobuf/ptypes/empty"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

//go:generate protoc -I ./ --go_out=plugins=grpc:./ ./av-cli.proto

type Server struct {
	Logger     Logger
	DBUsername string
	DBPassword string
	DBAddress  string
}

func (s *Server) Swab(id *ID, stream AvCli_SwabServer) error {
	dbAddr := strings.Replace(s.DBAddress, "dev", id.Designation, 1)
	dbAddr = strings.Replace(s.DBAddress, "stg", id.Designation, 1)
	dbAddr = strings.Replace(s.DBAddress, "prd", id.Designation, 1)

	db := db.GetDBWithCustomAuth(dbAddr, id.Designation, s.DBPassword)

	//check if id = build, room, or device
	idChecker := strings.Split(id.Id, "-")
	switch len(idChecker) {
	case 1:
		//it's a building
		rooms, err := db.GetRoomsByBuilding(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get rooms from database: %v", err)
			return stream.Send(&SwabResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(rooms) == 0 {
			err = fmt.Errorf("no rooms found in %s", id.Id)
			return stream.Send(&SwabResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan SwabResult)
		expectedCount := 0

		for i := range rooms {
			go func() {
				tmpRoom := rooms[i]
				devices, err := db.GetDevicesByRoom(tmpRoom.ID)
				if err != nil {
					err = fmt.Errorf("unable to get devices from database: %v", err)
					c <- SwabResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				if len(devices) == 0 {
					err = fmt.Errorf("no devices found in %s", tmpRoom.ID)
					c <- SwabResult{
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
								c <- SwabResult{
									Id:    tmpDevice.ID,
									Error: err.Error(),
								}
							} else {
								c <- SwabResult{
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
		for {
			select {
			case <-c:
				res := <-c
				stream.Send(&res)
				actualCount++
				if actualCount == expectedCount {
					return nil
				}
			}
		}

	case 2:
		//it's a room
		devices, err := db.GetDevicesByRoom(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get devices from database: %v", err)
			return stream.Send(&SwabResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(devices) == 0 {
			err = fmt.Errorf("no devices found in %s", id.Id)
			return stream.Send(&SwabResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan SwabResult)
		expectedCount := 0

		for i := range devices {
			tmpDevice := devices[i]

			if tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "Pi3" {
				go func() {
					err := swabDevice(context.TODO(), tmpDevice.Address)
					if err != nil {
						c <- SwabResult{
							Id:    tmpDevice.ID,
							Error: err.Error(),
						}
					} else {
						c <- SwabResult{
							Id:    tmpDevice.ID,
							Error: "",
						}
					}
				}()
				expectedCount++
			}
		}

		actualCount := 0
		for {
			select {
			case <-c:
				res := <-c
				stream.Send(&res)
				actualCount++
				if actualCount == expectedCount {
					return nil
				}
			}
		}

	case 3:
		//it's a device
		device, err := db.GetDevice(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get device from database: %s\n", err)
			return stream.Send(&SwabResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		err = swabDevice(context.TODO(), device.Address)
		if err != nil {
			return stream.Send(&SwabResult{
				Id:    device.ID,
				Error: err.Error(),
			})
		}

		return stream.Send(&SwabResult{
			Id:    device.ID,
			Error: "",
		})
	}

	//we should never get here
	return stream.Send(&SwabResult{
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

	return nil
}

func (s *Server) Float(id *ID, stream AvCli_FloatServer) error {
	dbAddr := strings.Replace(s.DBAddress, "dev", id.Designation, 1)
	dbAddr = strings.Replace(s.DBAddress, "stg", id.Designation, 1)
	dbAddr = strings.Replace(s.DBAddress, "prd", id.Designation, 1)

	db := db.GetDBWithCustomAuth(dbAddr, id.Designation, s.DBPassword)

	//check if id = build, room, or device
	idChecker := strings.Split(id.Id, "-")
	switch len(idChecker) {
	case 1:
		//building
		rooms, err := db.GetRoomsByBuilding(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get rooms from database: %v", err)
			return stream.Send(&FloatResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(rooms) == 0 {
			err = fmt.Errorf("no rooms found in %s", id.Id)
			return stream.Send(&FloatResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan FloatResult)
		expectedCount := 0

		for i := range rooms {
			tmpRoom := rooms[i]
			go func() {
				devices, err := db.GetDevicesByRoom(tmpRoom.ID)
				if err != nil {
					err = fmt.Errorf("unable to get devices from database: %v", err)
					c <- FloatResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				if len(devices) == 0 {
					err = fmt.Errorf("no devices found in %s", tmpRoom.ID)
					c <- FloatResult{
						Id:    tmpRoom.ID,
						Error: err.Error(),
					}
					return
				}

				for x := range devices {
					tmpDevice := devices[x]
					if tmpDevice.Type.ID == "Pi3" || tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "LabAttendance" || tmpDevice.Type.ID == "Pi-STB" || tmpDevice.Type.ID == "SchedulingPanel" || tmpDevice.Type.ID == "TimeClock" {
						go func() {
							err := floatShip(tmpDevice.ID, id.Designation)
							if err != nil {
								c <- FloatResult{
									Id:    tmpDevice.ID,
									Error: err.Error(),
								}
							} else {
								c <- FloatResult{
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
		for {
			select {
			case <-c:
				res := <-c
				stream.Send(&res)
				actualCount++
				if actualCount == expectedCount {
					return nil
				}
			}
		}

	case 2:
		//room
		devices, err := db.GetDevicesByRoom(id.Id)
		if err != nil {
			err = fmt.Errorf("unable to get devices from database: %v", err)
			return stream.Send(&FloatResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		if len(devices) == 0 {
			err = fmt.Errorf("no devices found in %s", id.Id)
			return stream.Send(&FloatResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}

		c := make(chan FloatResult)
		expectedCount := 0

		for i := range devices {
			tmpDevice := devices[i]
			if tmpDevice.Type.ID == "Pi3" || tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "LabAttendance" || tmpDevice.Type.ID == "Pi-STB" || tmpDevice.Type.ID == "SchedulingPanel" || tmpDevice.Type.ID == "TimeClock" {
				go func() {
					err := floatShip(tmpDevice.ID, id.Designation)
					if err != nil {
						c <- FloatResult{
							Id:    tmpDevice.ID,
							Error: err.Error(),
						}
					} else {
						c <- FloatResult{
							Id:    tmpDevice.ID,
							Error: "",
						}
					}
				}()
				expectedCount++
			}
		}
		actualCount := 0
		for {
			select {
			case <-c:
				res := <-c
				stream.Send(&res)
				actualCount++
				if actualCount == expectedCount {
					return nil
				}
			}
		}

	case 3:
		//device
		err := floatShip(id.Id, id.Designation)
		if err != nil {
			err = fmt.Errorf("error floating ship: %v", err)
			return stream.Send(&FloatResult{
				Id:    id.Id,
				Error: err.Error(),
			})
		}
		return stream.Send(&FloatResult{
			Id:    id.Id,
			Error: "",
		})
	}

	return stream.Send(&FloatResult{
		Id:    id.Id,
		Error: "unknown id received: " + id.Id,
	})
}

func floatShip(deviceID, designation string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", designation, deviceID), nil)
	if err != nil {
		return fmt.Errorf("couldn't make request: %v", err)
	}

	// req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", wso2.GetAccessToken()))

	resp, err := http.DefaultClient.Do(req)
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

func (s *Server) DuplicateRoom(context.Context, *DuplicateRoomRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DuplicateRoom not implemented")
}
