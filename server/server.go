package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/byuoitav/auth/wso2"
	avcli "github.com/byuoitav/av-cli"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ avcli.AvCliServer = &Server{}

type Server struct {
	Log        *zap.Logger
	Data       avcli.DataService
	PiPassword string
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

func (s *Server) Pis(ctx context.Context, cliID *avcli.ID) ([]avcli.Pi, error) {
	id, isRoom, err := parseID(cliID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse id: %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var pis []avcli.Pi
	if !isRoom {
		pi, err := s.Data.Device(ctx, id)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "unable to get device: %s", err)
		}

		pis = append(pis, pi)
	} else {
		pis, err = s.Data.Room(ctx, id)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "unable to get room: %s", err)
		}
	}

	if len(pis) == 0 {
		return nil, status.Errorf(codes.Unknown, "no pis found")
	}

	return pis, nil
}

func (s *Server) CloseMonitoringIssue(ctx context.Context, id *avcli.ID) (*empty.Empty, error) {
	return nil, errors.New("not implemented")
}

func (s *Server) SetLogLevel(ctx context.Context, logReq *avcli.SetLogLevelRequest) (*empty.Empty, error) {
	return nil, errors.New("not implemented")
}
