package slackcli

import (
	"context"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TODO pass default designation to *Client & use that for all calls
func (c *Client) CopyRoom(ctx context.Context, req slack.SlashCommand, user string, src, dst string) {
	c.Log.Info("Copying room", zap.String("src", src), zap.String("dst", dst))

	auth := auth{
		token: c.cliToken,
		user:  user,
	}

	handle := func(err error) {
		c.Log.Warn("unable to copy room", zap.Error(err))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msg := fmt.Sprintf("I was unable to copy %s->%s. :cry:. Error:\n```\n%s\n```", src, dst, err)
		_, _, err = c.slack.PostMessageContext(ctx, req.ChannelID, slack.MsgOptionReplaceOriginal(req.ResponseURL), slack.MsgOptionText(msg, false))
		if err != nil {
			c.Log.Warn("failed to post error to slack", zap.Error(err))
		}
	}

	args := &avcli.CopyRoomRequest{
		Src:            src,
		Dst:            dst,
		SrcDesignation: "prd",
		DstDesignation: "prd",
	}

	if _, err := c.cli.CopyRoom(ctx, args, grpc.PerRPCCredentials(auth)); err != nil {
		handle(err)
		return
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionReplaceOriginal(req.ResponseURL),
		slack.MsgOptionDeleteOriginal(req.ResponseURL),
		slack.MsgOptionText(fmt.Sprintf(":upvote: Successfully copied %s->%s", src, dst), false),
	}

	_, _, err := c.slack.PostMessageContext(ctx, req.ChannelID, msgOpts...)
	if err != nil {
		handle(fmt.Errorf("unable to send result to slack: %w", err))
		return
	}

	c.Log.Info("Successfully copied room", zap.String("src", src), zap.String("dst", dst))
}
