package server

import (
	"context"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
	"google.golang.org/grpc/status"
)

type idResultStream interface {
	Send(*avcli.IDResult) error
}

func (s *Server) runPerPi(ctx context.Context, id *avcli.ID, stream idResultStream, f func(avcli.Pi) error) error {
	pis, err := s.Pis(ctx, id)
	if err != nil {
		return err
	}

	// buffer results so that if we exit early (from ctx or stream error)
	// the goroutines in the for loop below aren't blocked forever
	results := make(chan *avcli.IDResult, len(pis))

	for i := range pis {
		go func(pi avcli.Pi) {
			var errstr string
			if err := f(pi); err != nil {
				errstr = err.Error()
			}

			results <- &avcli.IDResult{
				Id:    pi.ID,
				Error: errstr,
			}
		}(pis[i])
	}

	for i := len(pis); i > 0; i-- {
		select {
		case result := <-results:
			if err := stream.Send(result); err != nil {
				return fmt.Errorf("unable to send result: %w", err)
			}
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		}
	}

	return nil
}
