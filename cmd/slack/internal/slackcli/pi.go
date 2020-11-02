package slackcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Client) Screenshot(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.Log.Info("Getting a screenshot", zap.String("id", id), zap.String("for", user))

	handle := func(err error) {
		c.Log.Warn("unable to take screenshot", zap.Error(err))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msg := fmt.Sprintf("I was unable to get a screenshot of %s. :cry:. Error:\n```\n%s\n```", id, err)

		_, _, err = c.slack.PostMessageContext(ctx, req.ChannelID, slack.MsgOptionReplaceOriginal(req.ResponseURL), slack.MsgOptionText(msg, false))
		if err != nil {
			c.Log.Warn("failed to post error to slack", zap.Error(err))
		}
	}

	auth := auth{
		token: c.cliToken,
		user:  user, // TODO should be their netID
	}

	result, err := c.cli.Screenshot(ctx, &avcli.ID{Id: id}, grpc.PerRPCCredentials(auth))
	switch {
	case err != nil:
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.Unavailable:
				handle(s.Err())
				return
			case codes.Unauthenticated:
				handle(s.Err())
				return
			default:
				handle(s.Err())
				return
			}
		}

		handle(err)
		return
	case result == nil:
		handle(errors.New("result from api was nil"))
	}

	now := time.Now()

	// upload the screenshot
	_, err = c.slack.UploadFileContext(ctx, slack.FileUploadParameters{
		Filetype:       "jpg",
		Filename:       fmt.Sprintf("%s_%s.jpg", id, now.Format(time.RFC3339)),
		Title:          fmt.Sprintf("%s Screenshot @ %s", id, now.Format(time.RFC3339)),
		InitialComment: fmt.Sprintf("<@%s> - here is your screenshot of %s!", req.UserID, id),
		Reader:         bytes.NewBuffer(result.GetPhoto()),
		Channels:       []string{req.ChannelID},
	})
	if err != nil {
		handle(fmt.Errorf("unable to upload screenshot to slack: %w", err))
	}

	c.Log.Info("Successfully took screenshot", zap.String("of", id))
}

func (c *Client) Sink(ctx context.Context, req slack.SlashCommand, user string, id string) {
}

func (c *Client) FixTime(ctx context.Context, req slack.SlashCommand, user string, id string) {
}
