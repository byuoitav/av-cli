package avcli

//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative av-cli.proto

/*
func (s *Server) Swab(id *ID, stream AvCli_SwabServer) error {
	case 2:
		for i := range devices {
			if tmpDevice.Type.ID == "DividerSensors" || tmpDevice.Type.ID == "Pi3" {
			}
		}
	}
}
*/

/*
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
*/
