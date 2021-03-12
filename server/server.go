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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ avcli.AvCliServer = &Server{}

type Server struct {
	Log        *zap.Logger
	Data       avcli.DataService
	PiPassword string

	MonitoringBaseURL    string
	MonitoringSecret     string
	MonitoringRedisAddr  string
	MonitoringELKBaseURL string

	Client *wso2.Client
}

func (s *Server) log(ctx context.Context) *zap.Logger {
	md, _ := metadata.FromIncomingContext(ctx)
	log := s.Log
	if len(md["x-request-id"]) > 0 {
		log = log.With(zap.String("requestID", md["x-request-id"][0]))
	}

	return log
}

type idType int

const (
	idTypeBuilding idType = iota + 1
	idTypeRoom
	idTypeDevice
)

func parseID(cliID *avcli.ID) (string, idType, error) {
	split := strings.SplitN(cliID.GetId(), "-", 3)
	switch len(split) {
	case 1:
		return strings.ToUpper(split[0]), idTypeBuilding, nil
	case 2:
		return fmt.Sprintf("%s-%s", strings.ToUpper(split[0]), strings.ToUpper(split[1])), idTypeRoom, nil
	case 3:
		return fmt.Sprintf("%s-%s-%s", strings.ToUpper(split[0]), strings.ToUpper(split[1]), strings.ToUpper(split[2])), idTypeDevice, nil
	default:
		return "", 0, errors.New("not a room or device id")
	}
}

func (s *Server) Pis(ctx context.Context, cliID *avcli.ID) ([]avcli.Pi, error) {
	id, idType, err := parseID(cliID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse id: %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var pis []avcli.Pi
	switch idType {
	case idTypeDevice:
		pi, err := s.Data.Device(ctx, id)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "unable to get device: %s", err)
		}

		pis = append(pis, pi)
	case idTypeRoom:
		pis, err = s.Data.Room(ctx, id)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "unable to get room: %s", err)
		}
	case idTypeBuilding:
		pis, err = s.Data.Building(ctx, id)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "unable to get building: %s", err)
		}
	}

	if len(pis) == 0 {
		return nil, status.Errorf(codes.Unknown, "no pis found")
	}

	return pis, nil
}

func (s *Server) SetLogLevel(ctx context.Context, logReq *avcli.SetLogLevelRequest) (*empty.Empty, error) {
	return nil, errors.New("not implemented")
}
