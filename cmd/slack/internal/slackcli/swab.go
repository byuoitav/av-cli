package slackcli

import (
	"context"
	"fmt"
	"io"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) Swab(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.Log.Info("Swabbing", zap.String("id", id), zap.String("for", user))

	auth := auth{
		token: c.cliToken,
		user:  user,
	}

	handle := func(err error) {
		c.Log.Warn("unable to swab", zap.Error(err))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msg := fmt.Sprintf("I was unable to swab %s. :cry:. Error:\n```\n%s\n```", id, err)
		_, _, err = c.slack.PostMessageContext(ctx, req.ChannelID, slack.MsgOptionReplaceOriginal(req.ResponseURL), slack.MsgOptionText(msg, false))
		if err != nil {
			c.Log.Warn("failed to post error to slack", zap.Error(err))
		}
	}

	stream, err := c.cli.Swab(ctx, &avcli.ID{Id: id}, grpc.PerRPCCredentials(auth))
	if err != nil {
		handle(err)
		return
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: fmt.Sprintf("%s swab result", id),
		}),
	}

	for {
		result, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			handle(fmt.Errorf("error receiving from stream: %w", err))
			return
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

	msgOpts := []slack.MsgOption{
		slack.MsgOptionReplaceOriginal(req.ResponseURL),
		slack.MsgOptionDeleteOriginal(req.ResponseURL),
		slack.MsgOptionBlocks(blocks...),
	}

	// send the message
	_, _, _, err = c.slack.SendMessageContext(ctx, req.ChannelID, msgOpts...)
	if err != nil {
		handle(fmt.Errorf("unable to send result to slack: %w", err))
		return
	}

	c.Log.Info("Successfully swabbed", zap.String("id", id))
}
