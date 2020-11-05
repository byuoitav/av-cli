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

func (c *Client) CloseMonitoringIssue(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.Log.Info("Closing monitoring issue", zap.String("id", id))

	auth := auth{
		token: c.cliToken,
		user:  user,
	}

	handle := func(err error) {
		c.Log.Warn("unable to close monitoring issue", zap.Error(err))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msg := fmt.Sprintf("I was unable to close monitoring issue %s. :cry:. Error:\n```\n%s\n```", id, err)
		_, _, err = c.slack.PostMessageContext(ctx, req.ChannelID, slack.MsgOptionReplaceOriginal(req.ResponseURL), slack.MsgOptionText(msg, false))
		if err != nil {
			c.Log.Warn("failed to post error to slack", zap.Error(err))
		}
	}

	args := &avcli.ID{
		Id: id,
	}

	if _, err := c.cli.CloseMonitoringIssue(ctx, args, grpc.PerRPCCredentials(auth)); err != nil {
		handle(err)
		return
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionReplaceOriginal(req.ResponseURL),
		slack.MsgOptionDeleteOriginal(req.ResponseURL),
		slack.MsgOptionText(fmt.Sprintf(":upvote: Closed issue %s", id), false),
	}

	_, _, err := c.slack.PostMessageContext(ctx, req.ChannelID, msgOpts...)
	if err != nil {
		handle(fmt.Errorf("unable to send result to slack: %w", err))
		return
	}

	c.Log.Info("Successfully closed monitoring issue", zap.String("id", id))
}
