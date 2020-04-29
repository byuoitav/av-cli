package float

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

// Cmd .
var Cmd = &cobra.Command{
	Use:   "float [ID]",
	Short: "Deploys to the device/room/building with the given ID",
	Args:  args.ValidID,
	Run: func(cmd *cobra.Command, arg []string) {
		fmt.Printf("Deploying to %s\n", arg[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		pool, err := x509.SystemCertPool()
		if err != nil {
			fmt.Printf("unable to get system cert pool: %s", err)
			os.Exit(1)
		}

		idToken := wso2.GetIDToken()

		conn, err := grpc.Dial(viper.GetString("api"), getTransportSecurityDialOption(pool))
		if err != nil {
			fmt.Printf("error making grpc connection: %v", err)
			os.Exit(1)
		}

		cli := avcli.NewAvCliClient(conn)

		_, designation, err := args.GetDB()
		if err != nil {
			fmt.Printf("error getting designation: %v", err)
			os.Exit(1)
		}

		auth := auth{
			token: idToken,
			user:  "",
		}

		stream, err := cli.Float(context.TODO(), &avcli.ID{Id: arg[0], Designation: designation}, grpc.PerRPCCredentials(auth))
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable\n")
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to float: %s\n", err)
		}
		for {
			in, err := stream.Recv()
			switch {
			case errors.Is(err, io.EOF):
				return
			case err != nil:
				fmt.Printf("error: %s\nfd", err)
				return
			}

			if in.Error != "" {
				fmt.Printf("there was an error floating to %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully floated to %s\n", in.Id)
			}
		}
	},
}
