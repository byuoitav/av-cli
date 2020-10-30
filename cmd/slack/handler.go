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

	"github.com/byuoitav/av-cli/cmd/slack/internal/slackcli"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

func handleSlackRequests(slackCli *slackcli.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO get the user, send in metadata
		// TODO req.UserName should be their real netID (is there some mapping somewhere?)
		req, err := slack.SlashCommandParse(c.Request)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		slackCli.Log.Info("Got request", zap.Any("req", req))

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

				cmdSplit := strings.Split(cmd, " ")
				if len(cmdSplit) != 1 {
					c.String(http.StatusOK, "Invalid paramater to screenshot. Usage: /av pi screenshot [BLDG-ROOM-CP1]")
					return
				}

				id := cmdSplit[0]

				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					slackCli.Screenshot(ctx, req, req.UserName, id)
				}()

				c.String(http.StatusOK, fmt.Sprintf("Taking a screenshot of %s...", id))
				return
			default:
				c.String(http.StatusOK, "`pi` doesn't have that command.\n\nAvailable commands:\n\tscreenshot")
				return
			}
		case strings.HasPrefix(cmd, "swab"):
			cmd = trim(cmd, "swab")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				c.String(http.StatusOK, "Invalid paramater to swab. Usage: /av pi swab [(BLDG-ROOM)|(BLDG-ROOM-CP1)]")
				return
			}

			id := cmdSplit[0]

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				slackCli.Swab(ctx, req, req.UserName, id)
			}()

			c.String(http.StatusOK, fmt.Sprintf("Swabbing %s...", id))
			return
		default:
			c.String(http.StatusOK, "I don't know how to handle that command.")
			return
		}
	}
}

func verifySlackRequest(signingSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// skip verification if no secret was given
		if len(signingSecret) == 0 {
			return
		}

		verifier, err := slack.NewSecretsVerifier(c.Request.Header, signingSecret)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Abort()
			return
		}

		reader := io.TeeReader(c.Request.Body, &verifier)

		// read the body
		body, err := ioutil.ReadAll(reader)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Abort()
			return
		}
		defer c.Request.Body.Close()

		if err := verifier.Ensure(); err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// let the next handler read the body again
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
}
