package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Swab(cliID *avcli.ID, stream avcli.AvCli_SwabServer) error {
	// TODO handle different types differently
	// TODO status errors
	id, isRoom, err := parseID(cliID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "unable to parse id: %s", err)
	}

	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	var pis []avcli.Pi
	if !isRoom {
		pi, err := s.Data.Device(ctx, id)
		if err != nil {
			// TODO
		}

		pis = append(pis, pi)
	} else {
		pis, err = s.Data.Room(ctx, id)
		if err != nil {
			// TODO
		}
	}

	results := make(chan *avcli.IDResult)
	defer close(results)

	for i := range pis {
		go func(pi avcli.Pi) {
			var errstr string
			if err := swab(ctx, pi); err != nil {
				errstr = err.Error()
			}

			results <- &avcli.IDResult{
				Id:    pi.ID,
				Error: errstr,
			}
		}(pis[i])
	}

	expectedResults := len(pis)

	for {
		select {
		case result := <-results:
			if err := stream.Send(result); err != nil {
				return fmt.Errorf("unable to send message: %w", err)
			}

			expectedResults--
			if expectedResults == 0 {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("unable to send all results: %w", err)
		}
	}
}

func swab(ctx context.Context, pi avcli.Pi) error {
	// start the replication
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:7012/replication/start", pi.Address), nil)
	if err != nil {
		return fmt.Errorf("unable to build replication request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do replication request: %w", err)
	}
	resp.Body.Close()

	// let it replicate for a bit
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		return fmt.Errorf("unable to wait for replication: %w", ctx.Err())
	}

	// refresh the ui (TODO check the url)
	req, err = http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("http://%s:8888/refresh", pi.Address), nil)
	if err != nil {
		return fmt.Errorf("unable to build refresh request: %w", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do refresh request: %w", err)
	}
	resp.Body.Close()

	// TODO restart the dmm

	return nil
}
