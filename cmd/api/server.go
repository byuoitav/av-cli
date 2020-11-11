package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/byuoitav/auth/wso2"
	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/server"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dataServiceConfig struct {
	Addr     string
	Username string
	Password string
}

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errMissingToken    = status.Errorf(codes.Unauthenticated, "missing token")
	errMissingUser     = status.Errorf(codes.Unauthenticated, "missing user")
	errNotAuthorized   = status.Errorf(codes.Unauthenticated, "you are not authorized to do that")
)

func main() {
	var (
		port        int
		logLevel    string
		authAddr    string
		authToken   string
		disableAuth bool

		gatewayAddr         string
		clientID            string
		clientSecret        string
		monitoringURL       string
		monitoringSecret    string
		monitoringRedisAddr string
		monitoringELKURL    string
		piPassword          string
		dataServiceConfig   dataServiceConfig
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run lazarette on")
	pflag.StringVarP(&logLevel, "log-level", "L", "", "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVar(&authAddr, "auth-addr", "", "address of the auth server")
	pflag.StringVar(&authToken, "auth-token", "", "authorization token to use when calling the auth server")
	pflag.BoolVar(&disableAuth, "disable-auth", false, "disables auth checks")
	pflag.StringVar(&gatewayAddr, "gateway-addr", "api.byu.edu", "wso2 gateway address")
	pflag.StringVar(&clientID, "client-id", "", "wso2 key")
	pflag.StringVar(&clientSecret, "client-secret", "", "wso2 secret")
	pflag.StringVar(&monitoringURL, "monitoring-url", "", "monitoring url (ie https://monitoring.com)")
	pflag.StringVar(&monitoringSecret, "monitoring-secret", "", "monitoring secret")
	pflag.StringVar(&monitoringRedisAddr, "monitoring-redis", "", "monitoring redis addr")
	pflag.StringVar(&monitoringELKURL, "monitoring-elk", "", "monitoring elk base url")
	pflag.StringVar(&piPassword, "pi-password", "", "password for the pi user of the pis")
	pflag.StringVar(&dataServiceConfig.Addr, "db-address", "", "database address")
	pflag.StringVar(&dataServiceConfig.Username, "db-username", "", "database username")
	pflag.StringVar(&dataServiceConfig.Password, "db-password", "", "database password")
	pflag.Parse()

	// ctx for setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// build the logger
	// TODO a way to change the log level
	_, log := logger(logLevel)
	defer log.Sync() // nolint:errcheck

	// build opa client
	authClient := &authClient{
		Address:  authAddr,
		Token:    authToken,
		Disabled: disableAuth,
		Log:      log,
	}

	if !authClient.Disabled && len(authClient.Address) == 0 {
		log.Fatal("auth is enabled, but opa URL is not set")
	}

	// TODO build a prd/stg/dev version of these
	ds := dataService(ctx, dataServiceConfig)

	// build the grpc server
	cli := &server.Server{
		Log:                  log,
		Data:                 ds,
		PiPassword:           piPassword,
		MonitoringBaseURL:    monitoringURL,
		MonitoringSecret:     monitoringSecret,
		MonitoringRedisAddr:  monitoringRedisAddr,
		MonitoringELKBaseURL: monitoringELKURL,
		Client: &wso2.Client{
			GatewayURL:   fmt.Sprintf("https://%s", gatewayAddr),
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(authClient.unaryServerInterceptor()), grpc.StreamInterceptor(authClient.streamServerInterceptor()))
	avcli.RegisterAvCliServer(server, cli)

	// bind to a port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("unable to bind listener", zap.Error(err))
	}

	// start the server
	log.Info("Starting server", zap.String("on", lis.Addr().String()))
	if err := server.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
