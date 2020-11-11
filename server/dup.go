package server

import (
	"context"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
)

func (s *Server) CopyRoom(ctx context.Context, req *avcli.CopyRoomRequest) (*empty.Empty, error) {
	log := s.log(ctx)
	log.Info("Duplicating room", zap.String("from", req.GetSrc()), zap.String("to", req.GetDst()))

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// TODO validate the from/to id's ?

	if err := s.Data.CopyRoom(ctx, req.GetSrc(), req.GetDst()); err != nil {
		log.Warn("unable to duplicate room", zap.Error(err))
		return nil, err
	}

	log.Info("Successfully duplicated room")
	return &empty.Empty{}, nil
}
