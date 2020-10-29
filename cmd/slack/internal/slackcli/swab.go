package slackcli

import (
	"context"
	"io"

	avcli "github.com/byuoitav/av-cli"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) Swab(ctx context.Context, req slack.SlashCommand, user string, id string) {
	c.Log.Info("Swabbing", zap.String("id", id), zap.String("for", user))

	auth := auth{
		token: c.cliToken,
		user:  user, // TODO should be their netID
	}

	stream, err := c.cli.Swab(ctx, &avcli.ID{Id: id}, grpc.PerRPCCredentials(auth))
	if err != nil {
		// TODO handle error
	}

	// TODO figure out what we want the response message to look like

	for {
		result, err := stream.Recv()
		switch {
		case err == io.EOF:
			break
		case err != nil:
			// TODO handle error
		}
	}
}
