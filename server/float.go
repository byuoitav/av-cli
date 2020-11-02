package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"google.golang.org/grpc/status"
)

func (s *Server) Float(id *avcli.ID, stream avcli.AvCli_FloatServer) error {
	// TODO validate designations?
	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	pis, err := s.Pis(ctx, id)
	if err != nil {
		return err
	}

	results := make(chan *avcli.IDResult)
	defer close(results)

	for i := range pis {
		go func(pi avcli.Pi) {
			var errstr string
			if err := s.float(ctx, pi, id.GetDesignation()); err != nil {
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
				return fmt.Errorf("unable to send result: %w", err)
			}

			expectedResults--
			if expectedResults == 0 {
				return nil
			}
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		}
	}
}

func (s *Server) float(ctx context.Context, pi avcli.Pi, designation string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", designation, pi.ID), nil)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%v response from flight deck", resp.StatusCode)
		}

		return fmt.Errorf("%v response from flight deck: %s", resp.StatusCode, body)
	}

	return nil
}
