package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/auth/wso2"
	avcli "github.com/byuoitav/av-cli"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
)

var _ avcli.AvCliServer = &Server{}

type Server struct {
	Log  *zap.Logger
	Data avcli.DataService
	// ShipwrightKey string

	Client *wso2.Client
}

func parseID(cliID *avcli.ID) (id string, isRoom bool, err error) {
	split := strings.Split(cliID.GetId(), "-")
	switch len(split) {
	case 2:
		// it's a room
		id = fmt.Sprintf("%s-%s", split[0], split[1])
		isRoom = true
	case 3:
		// it's a device
		id = fmt.Sprintf("%s-%s-%s", split[0], split[1], split[2])
		isRoom = false
	default:
		err = errors.New("not a room or device id")
	}

	return
}

func (s *Server) DuplicateRoom(ctx context.Context, req *avcli.DuplicateRoomRequest) (*empty.Empty, error) {
	return nil, errors.New("not implemented")
}

func (s *Server) FixTime(id *avcli.ID, stream avcli.AvCli_FixTimeServer) error {
	return errors.New("not implemented")
}

func (s *Server) Sink(id *avcli.ID, stream avcli.AvCli_SinkServer) error {
	return errors.New("not implemented")
}

func (s *Server) CloseMonitoringIssue(ctx context.Context, id *avcli.ID) (*empty.Empty, error) {
	return nil, errors.New("not implemented")
}

func (s *Server) SetLogLevel(ctx context.Context, logReq *avcli.SetLogLevelRequest) (*empty.Empty, error) {
	return nil, errors.New("not implemented")
}
