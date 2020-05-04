package smee

import (
	"context"
	"fmt"
	"os"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var closeIssueCmd = &cobra.Command{
	Use:   "closeIssue [room ID]",
	Short: "Close a room issue in smee",
	Args:  args.ValidRoomID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Closing issue for %s\n", args[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
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

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		_, err = client.CloseMonitoringIssue(ctx, &avcli.ID{Id: args[0]})
		switch {
		case err != nil:
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to close issue: %s\n", err)
		}

		fmt.Printf("Room issue closed for %s.\n", args[0])
	},
}
