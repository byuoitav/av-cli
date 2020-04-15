package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	var (
		port          int
		logLevel      int8
		avcliApiUrl   string
		avcliApiToken string
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run lazarette on")
	pflag.Int8VarP(&logLevel, "log-level", "L", 0, "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVarP(&avcliApiUrl, "avcli-api", "a", "cli.av.byu.edu:443", "host/port of the avcli API")
	pflag.StringVarP(&avcliApiToken, "avcli-token", "t", "", "token to use for request to the avcli API")
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
	defer lPlain.Sync()

	log := lPlain.Sugar()

	// build the api client
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("unable to get system cert pool: %s", err)
	}

	opts := []grpc.DialOption{
		getTransportSecurityDialOption(pool),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, avcliApiUrl, opts...)
	if err != nil {
		log.Fatalf("unable to connect to avcli API: %s", err)
	}

	slack := &slack{
		cli:   avcli.NewAvCliClient(conn),
		token: avcliApiToken,
	}

	// build the server
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	e.POST("/", slack.handleRequest)

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
