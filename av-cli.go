package avcli

//go:generate protoc -I ./ --go_out=plugins=grpc:./ ./av-cli.proto

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
func (s *Server) CloseMonitoringIssue(ctx context.Context, id *ID) (*empty.Empty, error) {
	if s.ShipwrightKey == "" {
		return &empty.Empty{}, fmt.Errorf("shipwright key not set")
	}
	url := fmt.Sprintf("https://smee.av.byu.edu/issues/%s/resolve", id.Id)

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
*/
