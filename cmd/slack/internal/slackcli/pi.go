package slackcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
	c.Log.Info("Sinking", zap.String("id", id), zap.String("for", user))

	auth := auth{
		token: c.cliToken,
		user:  user,
	}

	handle := func(err error) {
		c.Log.Warn("unable to sink", zap.Error(err))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msg := fmt.Sprintf("I was unable to sink %s. :cry:. Error:\n```\n%s\n```", id, err)
		_, _, err = c.slack.PostMessageContext(ctx, req.ChannelID, slack.MsgOptionReplaceOriginal(req.ResponseURL), slack.MsgOptionText(msg, false))
		if err != nil {
			c.Log.Warn("failed to post error to slack", zap.Error(err))
		}
	}

	stream, err := c.cli.Float(ctx, &avcli.ID{Id: id}, grpc.PerRPCCredentials(auth))
	if err != nil {
		handle(err)
		return
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: fmt.Sprintf("%s sink result", id),
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

	c.Log.Info("Successfully sunk", zap.String("id", id))
}

func (c *Client) FixTime(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.Log.Info("Fixing time", zap.String("on", id), zap.String("for", user))

	auth := auth{
		token: c.cliToken,
		user:  user,
	}

	handle := func(err error) {
		c.Log.Warn("unable to fix time", zap.Error(err))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msg := fmt.Sprintf("I was unable to fix time on %s. :cry:. Error:\n```\n%s\n```", id, err)
		_, _, err = c.slack.PostMessageContext(ctx, req.ChannelID, slack.MsgOptionReplaceOriginal(req.ResponseURL), slack.MsgOptionText(msg, false))
		if err != nil {
			c.Log.Warn("failed to post error to slack", zap.Error(err))
		}
	}

	stream, err := c.cli.Float(ctx, &avcli.ID{Id: id}, grpc.PerRPCCredentials(auth))
	if err != nil {
		handle(err)
		return
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: fmt.Sprintf("%s fix time result", id),
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

	c.Log.Info("Successfully fixed time", zap.String("on", id))
}
