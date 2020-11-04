package server

import (
	"context"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
)

func (s *Server) CopyRoom(ctx context.Context, req *avcli.CopyRoomRequest) (*empty.Empty, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.Log.Info("Duplicating room", zap.String("from", req.GetSrc()), zap.String("to", req.GetDst()))

	// TODO validate the from/to id's ?

	if err := s.Data.CopyRoom(ctx, req.GetSrc(), req.GetDst()); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
