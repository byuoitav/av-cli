package slackcli

import (
	"context"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TODO pass default designation to *Client & use that for all calls
func (c *Client) CopyRoom(ctx context.Context, req slack.SlashCommand, user string, src, dst string) {
	c.handle(ctx, req, user, func(auth auth) []slack.MsgOption {
		c.Log.Info("Copying room", zap.String("src", src), zap.String("dst", dst))

		args := &avcli.CopyRoomRequest{
			Src:            src,
			Dst:            dst,
			SrcDesignation: "prd",
			DstDesignation: "prd",
		}

		if _, err := c.cli.CopyRoom(ctx, args, grpc.PerRPCCredentials(auth)); err != nil {
			c.Log.Warn("unable to copy room", zap.Error(err))
			return []slack.MsgOption{
				slack.MsgOptionText(fmt.Sprintf("I was unable to copy %s->%s. :cry:. Error:\n```\n%s\n```", src, dst, err), false),
			}
		}

		c.Log.Info("Successfully copied room", zap.String("src", src), zap.String("dst", dst))
		return []slack.MsgOption{
			slack.MsgOptionText(fmt.Sprintf(":upvote: Successfully copied %s->%s", src, dst), false),
		}
	})
}
