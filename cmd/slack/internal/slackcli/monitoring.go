package slackcli

import (
	"context"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) CloseMonitoringIssue(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.handle(ctx, req, user, func(auth auth) []slack.MsgOption {
		c.Log.Info("Closing monitoring issue", zap.String("id", id))

		args := &avcli.ID{
			Id:          id,
			Designation: "prd",
		}

		if _, err := c.cli.CloseMonitoringIssue(ctx, args, grpc.PerRPCCredentials(auth)); err != nil {
			c.Log.Warn("unable to close monitoring issue", zap.Error(err))
			return []slack.MsgOption{
				slack.MsgOptionText(fmt.Sprintf("I was unable to close monitoring issue %s. :cry:. Error:\n```\n%s\n```", id, err), false),
			}
		}

		c.Log.Info("Successfully closed monitoring issue", zap.String("id", id))
		return []slack.MsgOption{
			slack.MsgOptionText(fmt.Sprintf(":upvote: Closed issue %s", id), false),
		}
	})
}

func (c *Client) RemoveDeviceFromMonitoring(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.handle(ctx, req, user, func(auth auth) []slack.MsgOption {
		c.Log.Info("Removing device from monitoring", zap.String("id", id))

		args := &avcli.ID{
			Id:          id,
			Designation: "prd",
		}

		if _, err := c.cli.RemoveDeviceFromMonitoring(ctx, args, grpc.PerRPCCredentials(auth)); err != nil {
			c.Log.Warn("unable to remove device from monitoring issue", zap.Error(err))
			return []slack.MsgOption{
				slack.MsgOptionText(fmt.Sprintf("I was unable to remove %s from monitoring. :cry:. Error:\n```\n%s\n```", id, err), false),
			}
		}

		c.Log.Info("Successfully removed device from monitoring", zap.String("id", id))
		return []slack.MsgOption{
			slack.MsgOptionText(fmt.Sprintf(":upvote: Removed %s from monitoring", id), false),
		}
	})
}
