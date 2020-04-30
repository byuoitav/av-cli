package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
			return c.String(http.StatusInternalServerError, err.Error())
		}

		slackCli.Logger.Infof("Got request: %+v\n", req)

		cmd := strings.TrimSpace(req.Text)
		trim := func(s, prefix string) string {
			return strings.TrimSpace(strings.TrimPrefix(s, prefix))
		}

		switch {
		case strings.HasPrefix(cmd, "pi"):
			cmd = trim(cmd, "pi")

			switch {
			case strings.HasPrefix(cmd, "screenshot"):
				cmd = trim(cmd, "screenshot")

				// spawn a routine to post a screenshot
				cmdSplit := strings.Split(cmd, " ")
				if len(cmdSplit) != 1 {
					return c.String(http.StatusOK, "Invalid paramater to screenshot. Usage: /av pi screenshot [BLDG-ROOM-CP1]")
				}

				id := cmdSplit[0]

				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					slackCli.Screenshot(ctx, req, req.UserName, id)
				}()

				return c.String(http.StatusOK, fmt.Sprintf("Taking a screenshot of %s...", id))
			default:
				return c.String(http.StatusOK, "`pi` doesn't have that command.\n\nAvailable commands:\n\tscreenshot")
			}
		default:
			return c.String(http.StatusOK, "I don't know how to handle that command.")
		}
	}
}

func verifySlackRequest(signingSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// skip verification if no secret was given
			if len(signingSecret) == 0 {
				return next(c)
			}

			verifier, err := slack.NewSecretsVerifier(c.Request().Header, signingSecret)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			reader := io.TeeReader(c.Request().Body, &verifier)

			// read the body
			body, err := ioutil.ReadAll(reader)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			defer c.Request().Body.Close()

			if err := verifier.Ensure(); err != nil {
				return c.String(http.StatusUnauthorized, "you're not slack!")
			}

			// let the next handler read the body again
			c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))

			return next(c)
		}
	}
}
