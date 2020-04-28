package slackcli

import (
	"context"
	"crypto/x509"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Client struct {
	cli      avcli.AvCliClient
	cliToken string

	slack *slack.Client

	Logger *zap.SugaredLogger
}

func New(ctx context.Context, cliAddr string, cliToken string, slackToken string) (*Client, error) {
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
		slack:    slack.New(slackToken),
	}, nil
}

//func (c *Client) debugf(format string, a ...interface{}) {
//	if c.Logger != nil {
//		c.Logger.Debugf(format, a...)
//	}
//}

func (c *Client) infof(format string, a ...interface{}) {
	if c.Logger != nil {
		c.Logger.Infof(format, a...)
	}
}

func (c *Client) warnf(format string, a ...interface{}) {
	if c.Logger != nil {
		c.Logger.Warnf(format, a...)
	}
}

//func (c *Client) errorf(format string, a ...interface{}) {
//	if c.Logger != nil {
//		c.Logger.Errorf(format, a...)
//	}
//}
