package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byuoitav/av-cli/slack/internal/slackcli"
	"github.com/labstack/echo"
	"github.com/slack-go/slack"
)

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "healthy")
}

func handleSlackRequests(slackCli *slackcli.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO get the user, send in metadata
		// TODO write handler logic here
		// TODO actual error handling for slack API

		req, err := slack.SlashCommandParse(c.Request())
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		// TODO validate token
		//if !req.ValidateToken() {
		//	return c.String(http.StatusForbidden, "you're not slack!")
		//}

		// TODO validate request came from slack

		req.Command = strings.TrimSpace(req.Command)

		switch {
		case strings.HasPrefix(req.Command, "pi screenshot"):
			// spawn a routine to post a screenshot
			cmd := strings.TrimPrefix(req.Command, "pi screenshot")
			cmd = strings.TrimSpace(cmd)

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				return c.String(http.StatusOK, "Invalid paramater to screenshot. Usage: /av pi screenshot [BLDG-ROOM-CP1]")
			}

			id := cmdSplit[0]

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				err := slackCli.Screenshot(ctx, req, req.UserID, id)
				if err != nil {
					slackCli.Logger.Warnf("failed to take screenshot: %s", err)
				}
			}()

			return c.String(http.StatusOK, fmt.Sprintf("Taking a screenshot of %s...", id))
		default:
			return c.String(http.StatusOK, "I don't know how to handle that command.")
		}
	}
}
