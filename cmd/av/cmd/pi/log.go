package pi

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var logCmd = &cobra.Command{
	Use:   "log [device ID] [port] [log-level]",
	Short: "change a log level",
	Long:  "change a log level on a specific port on a pi",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 3 {
			fmt.Printf("missing arguments\n")
			os.Exit(1)
		}

		level := ""
		switch args[2] {
		case "-1":
			level = "debug"
		case "0":
			level = "info"
		case "1":
			level = "warn"
		case "2":
			level = "error"
		case "3":
			level = "dpanic"
		case "4":
			level = "panic"
		case "5":
			level = "fatal"
		}

		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}
		fmt.Printf("Setting log level on port %s of %s to %s\n", args[1], args[0], level)

		idToken := wso2.GetIDToken()

		auth := avcli.Auth{
			Token: idToken,
			User:  "",
		}

		port, err := strconv.Atoi(args[1])
		if err != nil {
			fail("error converting port to int: %v\n", err)
		}

		logLevel, err := strconv.Atoi(args[2])
		if err != nil {
			fail("error converting log level to int: %v\n", err)
		}

		req := avcli.SetLogLevelRequest{
			Id:    args[0],
			Port:  int32(port),
			Level: int32(logLevel),
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()

		client, err := avcli.NewClient(viper.GetString("api"), auth)
		if err != nil {
			fail("unable to create client: %v\n", err)
		}

		_, err = client.SetLogLevel(ctx, &req)
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("bad %s\n", s.Err())
				}
			}

			fail("unable to set log level: %s\n", err)
		}

		fmt.Printf("Log level on port %s of %s set to %s\n", args[1], args[0], level)

	},
}
