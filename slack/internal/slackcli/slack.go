package slackcli

import (
	"context"
	"crypto/x509"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Request struct {
	Token          string `form:"token"`
	Command        string `form:"command"`
	Text           string `form:"text"`
	ResponseURL    string `form:"response_url"`
	TriggerID      string `form:"trigger_id"`
	UserID         string `form:"user_id"`
	UserName       string `form:"user_name"`
	TeamID         string `form:"team_id"`
	TeamName       string `form:"team_name"`
	EnterpriseID   string `form:"enterprise_id"`
	EnterpriseName string `form:"enterprise_name"`
	ChannelID      string `form:"channel_id"`
	ChannelName    string `form:"channel_name"`
}

type Client struct {
	cli      avcli.AvCliClient
	cliToken string

	Logger *zap.SugaredLogger
}

func NewClient(ctx context.Context, cliAddr string, cliToken string) (*Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("unable to get system cert pool: %s", err)
	}

	conn, err := grpc.DialContext(ctx, cliAddr, getTransportSecurityDialOption(pool))
	if err != nil {
		return nil, fmt.Errorf("unable to connec to avcli API: %s", err)
	}

	return &Client{
		cli:      avcli.NewAvCliClient(conn),
		cliToken: cliToken,
	}, nil
}
