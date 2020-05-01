package pi

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var fixTimeCmd = &cobra.Command{
	Use:   "fixtime [device ID]",
	Short: "fix a pi who's time is off",
	Long:  "force an NTP sync of a pi to fix time drift",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Fixing time on %s\n", args[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		pool, err := x509.SystemCertPool()
		if err != nil {
			fail("unable to get system cert pool: %v", err)
		}

		idToken := wso2.GetIDToken()

		conn, err := grpc.Dial(viper.GetString("api"), avcli.getTransportSecurityDialOption(pool))
		if err != nil {
			fail("error making grpc connection: %v", err)
		}

		cli := avcli.NewAvCliClient(conn)

		auth := avcli.auth{
			token: idToken,
			user:  "",
		}

		stream, err := cli.FixTime(context.TODO(), &avcli.ID{Id: args[0]}, grpc.PerRPCCredentials(auth))
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to fix time: %v\n", err)
		}

		for {
			in, err := stream.Recv()
			switch {
			case errors.Is(err, io.EOF):
				return
			case err != nil:
				fmt.Printf("error: %s\n", err)
				return
			}

			if in.Error != "" {
				fmt.Printf("there was an error fixing time on %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully floated to %s\n", in.Id)
			}
		}
	},
}
