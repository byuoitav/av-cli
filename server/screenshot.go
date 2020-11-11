package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Screenshot(ctx context.Context, cliID *avcli.ID) (*avcli.ScreenshotResult, error) {
	log := s.log(ctx)
	log.Info("Screenshotting", zap.String("id", cliID.GetId()))

	// TODO status errors
	id, isRoom, err := parseID(cliID)
	if err != nil {
		log.Warn("unable to parse id", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse id: %s", err)
	}

	if isRoom {
		log.Warn("id was not a device id")
		return nil, status.Errorf(codes.InvalidArgument, "screenshot requires a device id")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pi, err := s.Data.Device(ctx, id)
	if err != nil {
		log.Warn("unable to lookup device", zap.Error(err))
		return nil, fmt.Errorf("unable to lookup device: %w", err)
	}

	url := fmt.Sprintf("http://%s:10000/device/screenshot", pi.Address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Warn("unable to build request", zap.Error(err))
		return nil, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn("unable to do request", zap.Error(err))
		return nil, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("unable to read body", zap.Error(err))
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		log.Warn("bad response from ELK", zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", body))
		return nil, fmt.Errorf("failed to get screenshot: %s", body)
	}

	log.Info("Successfully got screenshot")
	return &avcli.ScreenshotResult{
		Photo: body,
	}, nil
}
