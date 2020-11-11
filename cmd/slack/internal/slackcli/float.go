package slackcli

import (
	"context"
	"fmt"
	"io"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TODO pass default designation to *Client & use that for all calls
func (c *Client) Float(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.handle(req, user, func(auth auth) []slack.MsgOption {
		c.Log.Info("Floating", zap.String("id", id), zap.String("for", user))

		args := &avcli.ID{
			Id:          id,
			Designation: "prd",
		}

		stream, err := c.cli.Float(ctx, args, grpc.PerRPCCredentials(auth))
		if err != nil {
			c.Log.Warn("unable to float", zap.Error(err))
			return []slack.MsgOption{
				slack.MsgOptionText(fmt.Sprintf("<@%s>: I was unable to float %s. :cry:. Error:\n```\n%s\n```", req.UserID, id, err), false),
			}
		}

		blocks := []slack.Block{
			slack.NewSectionBlock(&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf("*%s float result* (for <@%s>)", id, req.UserID),
			}, nil, nil),
		}

		for {
			result, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				// TODO show which ones were successful?
				c.Log.Warn("unable to recv from stream", zap.Error(err))
				return []slack.MsgOption{
					slack.MsgOptionText(fmt.Sprintf("<@%s>: I was unable to float %s. :cry:. Error:\n```\n%s\n```", req.UserID, id, err), false),
				}
			}

			if result.GetError() != "" {
				blocks = append(blocks, slack.NewSectionBlock(&slack.TextBlockObject{
					Type: slack.MarkdownType,
					Text: fmt.Sprintf(":downvote: %s `%s`", result.GetId(), result.GetError()),
				}, nil, nil))
			} else {
				blocks = append(blocks, slack.NewSectionBlock(&slack.TextBlockObject{
					Type: slack.MarkdownType,
					Text: fmt.Sprintf(":upvote: %s", result.GetId()),
				}, nil, nil))
			}

			blocks = append(blocks, slack.NewDividerBlock())
		}

		// delete the last divider
		if blocks[len(blocks)-1].BlockType() == slack.MBTDivider {
			blocks = blocks[:len(blocks)-1]
		}

		c.Log.Info("Successfully floated", zap.String("id", id))
		return []slack.MsgOption{
			slack.MsgOptionBlocks(blocks...),
		}
	})
}
