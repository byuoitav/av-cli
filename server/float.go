package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
)

func (s *Server) Float(id *avcli.ID, stream avcli.AvCli_FloatServer) error {
	// TODO validate designations?
	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	return s.runPerPi(ctx, id, stream, func(pi avcli.Pi) error {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", id.Designation, pi.ID), nil)
		if err != nil {
			return fmt.Errorf("unable to build request: %w", err)
		}

		resp, err := s.Client.Do(req)
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
	})
}
