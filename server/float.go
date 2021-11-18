package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"go.uber.org/zap"
)

func (s *Server) Float(id *avcli.ID, stream avcli.AvCli_FloatServer) error {
	log := s.log(stream.Context())
	log.Info("Floating", zap.String("to", id.GetId()))

	// TODO validate designations?
	ctx, cancel := context.WithTimeout(stream.Context(), 90*time.Second)
	defer cancel()

	return s.runPerPi(ctx, id, stream, func(pi avcli.Pi) error {
		// Buildings that have been converted over to use Ansible
		var buildings = []string{"ITB", "JKB"}
		var buildingMatch = false

		log := log.With(zap.String("pi", pi.ID))

		// no individual pi can take longer than 30 seconds
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Loop will interate over list of buildings and if it finds a match, it will change variable to true and break the loop
		for _, building := range buildings {
			if strings.Contains(pi.ID, building) {
				buildingMatch = true
				break
			}
		}

		// if the building matches with the list of buildings above, run that ansible endpoint verses running the old flightdeck endpoint
		if buildingMatch {
			Req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/dev/refloat/%v", pi.ID), nil)
			if err != nil {
				log.Warn("Unable to build request:", zap.Error(err))
				return fmt.Errorf("Unable to build request: %w", err)
			}
		} else {
			Req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", id.Designation, pi.ID), nil)
			if err != nil {
				log.Warn("unable to build request", zap.Error(err))
				return fmt.Errorf("unable to build request: %w", err)
			}
		}

		resp, err := s.Client.Do(Req)
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
