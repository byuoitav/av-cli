package slackcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Client) Screenshot(ctx context.Context, req slack.SlashCommand, user string, id string) error {
	c.infof("Getting a screenshot of %s for %s", id, user)

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
		return fmt.Errorf("unable to upload screenshot: %w", err)
	}

	return nil
}
