package slackcli

import (
	"context"
	"errors"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Client) Screenshot(ctx context.Context, req Request, id string) error {
	auth := auth{
		token: c.cliToken,
		user:  req.UserID, // TODO should be their netID
	}

	result, err := c.cli.Screenshot(ctx, &avcli.ID{Id: req.Text}, grpc.PerRPCCredentials(auth))
	switch {
	case err != nil:
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.Unavailable:
				return s.Err()
			case codes.Unauthenticated:
				return s.Err()
			default:
				return s.Err()
			}
		}

		return err
	case result == nil:
		return errors.New("this is weird")
	}

	return slack.PostWebhookContext(ctx, req.ResponseURL, &slack.WebhookMessage{
		Text: "I have your photo!",
	})
}
