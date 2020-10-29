package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/byuoitav/av-cli/cmd/slack/internal/slackcli"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var (
		port               int
		logLevel           int8
		avcliAPIAddr       string
		avcliAPIToken      string
		slackSigningSecret string
		slackToken         string
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run lazarette on")
	pflag.Int8VarP(&logLevel, "log-level", "L", 0, "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVarP(&avcliAPIAddr, "avcli-api", "a", "cli-api", "address of the avcli API")
	pflag.StringVarP(&avcliAPIToken, "avcli-token", "t", "", "token to use for request to the avcli API")
	pflag.StringVar(&slackSigningSecret, "signing-secret", "", "slack signing secret. see https://api.slack.com/authentication/verifying-requests-from-slack for more details.")
	pflag.StringVar(&slackToken, "slack-token", "", "slack token")
	pflag.Parse()

	// build the logger
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.Level(logLevel)),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "@",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err := config.Build()
	if err != nil {
		fmt.Printf("failed to build zap logger: %s\n", err)
		os.Exit(1)
	}
	defer log.Sync() // nolint:errcheck

	// build the api client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slackCli, err := slackcli.New(ctx, avcliAPIAddr, avcliAPIToken, slackToken)
	if err != nil {
		log.Fatal("failed to build slack-cli client", zap.Error(err))
	}

	slackCli.Log = log

	// build the server
	r := gin.New()
	r.Use(gin.Recovery())

	debug := r.Group("/debug")
	debug.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})
	debug.GET("/logz", func(c *gin.Context) {
		c.String(http.StatusOK, config.Level.String())
	})
	debug.GET("/logz/:level", func(c *gin.Context) {
		var level zapcore.Level
		if err := level.Set(c.Param("level")); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		fmt.Printf("***\n\tSetting log level to %s\n***\n", level.String())
		config.Level.SetLevel(level)
		c.String(http.StatusOK, config.Level.String())
	})

	api := r.Group("/", verifySlackRequest(slackSigningSecret))
	api.POST("/", handleSlackRequests(slackCli))

	/*
		e.Pre(middleware.RemoveTrailingSlash())
	*/

	// start the server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("unable to bind listener", zap.Error(err))
	}

	log.Info("Starting server", zap.String("on", lis.Addr().String()))
	err = r.RunListener(lis)
	switch {
	case errors.Is(err, http.ErrServerClosed):
	case err != nil:
		log.Fatal("failed to serve", zap.Error(err))
	}
}
