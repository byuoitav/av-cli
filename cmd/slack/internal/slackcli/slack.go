package slackcli

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Client struct {
	cli      avcli.AvCliClient
	cliToken string

	slack *slack.Client
	Log   *zap.Logger
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
		slack:    slack.New(slackToken, slack.OptionDebug(true)),
	}, nil
}

func (c *Client) handle(ctx context.Context, req slack.SlashCommand, user string, f func(auth auth) []slack.MsgOption) {
	auth := auth{
		token: c.cliToken,
		user:  user,
	}

	opts := []slack.MsgOption{
		slack.MsgOptionResponseURL(req.ResponseURL, slack.ResponseTypeInChannel),
	}
	opts = append(opts, f(auth)...)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, _, err := c.slack.PostMessageContext(ctx, req.ChannelID, opts...)
	if err != nil {
		c.Log.Warn("unable to post message to slack", zap.Error(err), zap.String("cmd", req.Command))
		return
	}
}
