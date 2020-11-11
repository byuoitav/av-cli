package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CloseMonitoringIssue(ctx context.Context, cliID *avcli.ID) (*empty.Empty, error) {
	log := s.log(ctx)
	log.Info("Closing monitoring issue", zap.String("id", cliID.GetId()))

	id, isRoom, err := parseID(cliID)
	if err != nil {
		log.Warn("unable to parse id", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse id: %s", err)
	}

	if !isRoom {
		log.Warn("id was not a room")
		return nil, status.Errorf(codes.InvalidArgument, "closing a monitoring issue requires a room id")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	netID := avcli.NetID(ctx)

	url := fmt.Sprintf("%s/issues/%s/resolve", s.MonitoringBaseURL, id)
	reqBody := []byte(fmt.Sprintf(`
	{
		"resolution-code": "Manual Removal",
		"notes": "%s manually removed room issue through av-cli"
	}`, netID))

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(reqBody))
	if err != nil {
		log.Warn("unable to build request", zap.Error(err))
		return nil, fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-av-access-key", s.MonitoringSecret)
	req.Header.Add("x-av-user", netID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn("unable to do request", zap.Error(err))
		return nil, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warn("bad response from monitoring service", zap.Int("statusCode", resp.StatusCode))
			return nil, fmt.Errorf("%v response from monitoring service", resp.StatusCode)
		}

		log.Warn("bad response from monitoring service", zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", body))
		return nil, fmt.Errorf("%v response from monitoring service: %s", resp.StatusCode, body)
	}

	return &empty.Empty{}, nil
}

func (s *Server) RemoveDeviceFromMonitoring(ctx context.Context, cliID *avcli.ID) (*empty.Empty, error) {
	log := s.log(ctx)
	log.Info("Removing device from monitoring", zap.String("id", cliID.GetId()))

	id, isRoom, err := parseID(cliID)
	if err != nil {
		log.Warn("unable to parse id", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse id: %s", err)
	}

	if isRoom {
		log.Warn("id was not a device")
		return nil, status.Errorf(codes.InvalidArgument, "removing a device from monitoring requires a device id")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// TODO double check that it's in DB 0
	// TODO should i open this some other time (main.go), and just reuse the client?
	rdb := redis.NewClient(&redis.Options{
		Addr: s.MonitoringRedisAddr,
	})
	defer rdb.Close()

	if err := rdb.Del(ctx, id).Err(); err != nil {
		log.Warn("unable to delete device from cache", zap.Error(err))
		return nil, fmt.Errorf("unable to delete device from cache: %w", err)
	}

	// delete from ELK
	url := fmt.Sprintf("%s/oit-static-av-devices-v3/_delete_by_query", s.MonitoringELKBaseURL)
	reqBody := []byte(fmt.Sprintf(`
	{
		"query": {
			"bool": {
				"filter": [
					{
						"wildcard": {
							"deviceID": "%s"
						}
					}
				]
			}
		}
	}`, id))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		log.Warn("unable to build request", zap.Error(err))
		return nil, fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Add("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn("unable to do request", zap.Error(err))
		return nil, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warn("bad response from ELK", zap.Int("statusCode", resp.StatusCode))
			return nil, fmt.Errorf("%v response from ELK", resp.StatusCode)
		}

		log.Warn("bad response from ELK", zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", body))
		return nil, fmt.Errorf("%v response from ELK: %s", resp.StatusCode, body)
	}

	log.Info("Successfully removed device from monitoring")
	return &empty.Empty{}, nil
}
