package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/byuoitav/av-cli/slack/internal/slackcli"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run lazarette on")
	pflag.Int8VarP(&logLevel, "log-level", "L", 0, "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVarP(&avcliAPIAddr, "avcli-api", "a", "cli.av.byu.edu:443", "address of the avcli API")
	pflag.StringVarP(&avcliAPIToken, "avcli-token", "t", "", "token to use for request to the avcli API")
	pflag.StringVar(&slackSigningSecret, "signing-secret", "", "slack signing secret. see https://api.slack.com/authentication/verifying-requests-from-slack for more details.")
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

	lPlain, err := config.Build()
	if err != nil {
		fmt.Printf("failed to build zap logger: %s\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = lPlain.Sync()
	}()

	log := lPlain.Sugar()

	// build the api client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slackCli, err := slackcli.NewClient(ctx, avcliAPIAddr, avcliAPIToken)
	if err != nil {
		log.Fatalf("failed to build slack-cli client: %s", err)
	}

	slackCli.Logger = log

	// build the server
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	e.GET("/healthz", healthHandler)
	e.POST("/", handleSlackRequests(slackCli), verifySlackRequest(slackSigningSecret))

	// start the server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to bind listener: %s", err)
	}

	log.Infof("Starting server on %s", lis.Addr().String())
	if err := e.Server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
