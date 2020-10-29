package swab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Cmd .
var Cmd = &cobra.Command{
	Use:   "swab [ID]",
	Short: "Refreshes the database/ui of a pi/room/building",
	Long:  "Forces a replication of the couch database, and causes the ui to refresh shortly after",
	Args:  args.ValidID,
	Run: func(cmd *cobra.Command, arg []string) {
		fmt.Printf("Swabbing %s\n", arg[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		_, designation, err := args.GetDB()
		if err != nil {
			fmt.Printf("error getting designation: %v", err)
			os.Exit(1)
		}

		idToken := wso2.GetIDToken()
		auth := avcli.Auth{
			Token: idToken,
			User:  "",
		}

		client, err := avcli.NewClient(viper.GetString("api"), auth)
		if err != nil {
			fail("unable to create client: %v\n", err)
		}

		stream, err := client.Swab(context.TODO(), &avcli.ID{Id: arg[0], Designation: designation})
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to swab: %s\n", err)
		}

		for {
			in, err := stream.Recv()
			switch {
			case errors.Is(err, io.EOF):
				return
			case err != nil:
				fmt.Printf("error: %s\nfd", err)
			}

			if in.Error != "" {
				fmt.Printf("there was an error swabbing %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully swabbed %s\n", in.Id)
			}
		}
	},
}
