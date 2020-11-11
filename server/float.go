package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"go.uber.org/zap"
)

func (s *Server) Float(id *avcli.ID, stream avcli.AvCli_FloatServer) error {
	log := s.log(stream.Context())
	log.Info("Floating", zap.String("to", id.GetId()))

	// TODO validate designations?
	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	return s.runPerPi(ctx, id, stream, func(pi avcli.Pi) error {
		log := log.With(zap.String("pi", pi.ID))

		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", id.Designation, pi.ID), nil)
		if err != nil {
			log.Warn("unable to build request", zap.Error(err))
			return fmt.Errorf("unable to build request: %w", err)
		}

		resp, err := s.Client.Do(req)
		if err != nil {
			log.Warn("unable to do request", zap.Error(err))
			return fmt.Errorf("unable to do request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Warn("bad response from flight deck", zap.Int("statusCode", resp.StatusCode))
				return fmt.Errorf("%v response from flight deck", resp.StatusCode)
			}

			log.Warn("bad response from flight deck", zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", body))
			return fmt.Errorf("%v response from flight deck: %s", resp.StatusCode, body)
		}

		log.Info("Successfully floated")
		return nil
	})
}
