package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	avcli "github.com/byuoitav/av-cli"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

func main() {
	var (
		port     int
		logLevel int8
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run lazarette on")
	pflag.Int8VarP(&logLevel, "log-level", "L", 0, "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
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

	// build the grpc server
	cli := &avcli.Server{
		Logger: log,
	}

	// TODO add stream interceptor if we add a streaming request
	server := grpc.NewServer(grpc.UnaryInterceptor(unaryCheckToken))
	avcli.RegisterAvCliServer(server, cli)

	// bind to a port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to bind listener: %s", err)
	}

	// start the server
	log.Infof("Starting server on %s", lis.Addr().String())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func validToken(auth []string) bool {
	if len(auth) == 0 {
		return false
	}

	token := strings.TrimPrefix(auth[0], "Bearer ")

	// TODO check opa
	return token == "secret"
}

func validUser(username []string) bool {
	if len(username) == 0 {
		return false
	}

	fmt.Printf("username: %s\n", username[0])

	// TODO check opa
	return true
}

func unaryCheckToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	// TODO these funcs should return an error
	if !validToken(md["authorization"]) {
		return nil, errInvalidToken
	}

	if !validUser(md["username"]) {
		return nil, fmt.Errorf("user %q is not authorized to call %s", md["username"][0], info.FullMethod)
	}

	return handler(ctx, req)
}
