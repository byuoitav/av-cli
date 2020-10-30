package main

import (
	"context"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/couch"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func dataService(ctx context.Context, config dataServiceConfig) avcli.DataService {
	var opts []couch.Option

	if len(config.Username) > 0 {
		opts = append(opts, couch.WithBasicAuth(config.Username, config.Password))
	}

	ds, err := couch.New(ctx, config.Addr, opts...)
	if err != nil {
		panic(fmt.Sprintf("unable to setup couch: %s", err))
	}

	return ds
}

func logger(logLevel string) (zap.Config, *zap.Logger) {
	var level zapcore.Level
	if err := level.Set(logLevel); err != nil {
		panic(fmt.Sprintf("invalid log level: %s", err))
	}

	config := zap.Config{
		Level: zap.NewAtomicLevelAt(level),
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
			StacktraceKey:  "trace",
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
		panic(fmt.Sprintf("unable to build logger: %s", err))
	}

	return config, log
}
