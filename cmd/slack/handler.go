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
		// TODO request id in ctx
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
		case strings.HasPrefix(cmd, "screenshot"):
			cmd = trim(cmd, "screenshot")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				c.String(http.StatusOK, "Invalid paramater to screenshot. Usage: /av screenshot [BLDG-ROOM-CP1]")
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
		case strings.HasPrefix(cmd, "sink"):
			cmd = trim(cmd, "sink")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				c.String(http.StatusOK, "Invalid paramater to sink. Usage: /av sink [(BLDG-ROOM)|(BLDG-ROOM-CP1)]")
				return
			}

			id := cmdSplit[0]

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				slackCli.Sink(ctx, req, req.UserName, id)
			}()

			c.String(http.StatusOK, fmt.Sprintf("Sinking %s...", id))
			return
		case strings.HasPrefix(cmd, "fixtime"):
			cmd = trim(cmd, "fixtime")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				c.String(http.StatusOK, "Invalid paramater to fixtime. Usage: /av fixtime [(BLDG-ROOM)|(BLDG-ROOM-CP1)]")
				return
			}

			id := cmdSplit[0]

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				slackCli.FixTime(ctx, req, req.UserName, id)
			}()

			c.String(http.StatusOK, fmt.Sprintf("Fixing time on %s...", id))
			return
		case strings.HasPrefix(cmd, "swab"):
			cmd = trim(cmd, "swab")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				c.String(http.StatusOK, "Invalid paramater to swab. Usage: /av swab [(BLDG-ROOM)|(BLDG-ROOM-CP1)]")
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
		case strings.HasPrefix(cmd, "float"):
			cmd = trim(cmd, "float")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 1 {
				c.String(http.StatusOK, "Invalid paramater to float. Usage: /av float [(BLDG-ROOM)|(BLDG-ROOM-CP1)]")
				return
			}

			id := cmdSplit[0]

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				slackCli.Float(ctx, req, req.UserName, id)
			}()

			c.String(http.StatusOK, fmt.Sprintf("Swabbing %s...", id))
			return
		case strings.HasPrefix(cmd, "db dup"):
			cmd = trim(cmd, "db dup")

			cmdSplit := strings.Split(cmd, " ")
			if len(cmdSplit) != 2 {
				c.String(http.StatusOK, "Invalid paramater to db dup. Usage: /av db dup [SRC-ID] [DST-ID]")
				return
			}

			src := cmdSplit[0]
			dst := cmdSplit[1]

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				slackCli.CopyRoom(ctx, req, req.UserName, src, dst)
			}()

			c.String(http.StatusOK, fmt.Sprintf("Copying %s -> %s...", src, dst))
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
