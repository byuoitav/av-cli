package swab

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/pi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	Cmd.AddCommand(swabRoomCmd)
	Cmd.AddCommand(swabBuildingCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "swab [device ID]",
	Short: "Refreshes the database/ui of a pi",
	Long:  "Forces a replication of the couch database, and causes the ui to refresh shortly after",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, arg []string) {
		fmt.Printf("Swabbing %s\n", arg[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		conn, err := grpc.Dial(viper.GetString("api"), grpc.WithInsecure())
		if err != nil {
			fmt.Printf("oh no\n")
			os.Exit(1)
		}

		cli := avcli.NewAvCliClient(conn)

		_, designation, err := args.getDB()
		if err != nil {
			fmt.Printf("error getting designation: %v", err)
			os.Exit(1)
		}

		stream, err := cli.Swab(context.TODO(), &avcli.ID{Id: arg[0], Designation: designation})
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable\n")
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to swab: %s\n", err)
		}

		for {
			in, err := stream.Recv()
			if err == io.EOF {
				fmt.Printf("finished swabbing\n")
				return
			}
			if in.Error != "" {
				fmt.Printf("there was an error swabbing %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully swabbed %s\n", in.Id)
			}
		}
	},
}

func swab(ctx context.Context, address string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:7012/replication/start", address), nil)
	if err != nil {
		return fmt.Errorf("unable to build replication request: %s", err)
	}

	req = req.WithContext(ctx)

	// TODO actually validate that it worked
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to start replication: %s", err)
	}

	fmt.Printf("%s\tReplication started\n", address)
	time.Sleep(3 * time.Second) // TODO make sure this doesn't overrun ctx

	req, err = http.NewRequest("PUT", fmt.Sprintf("http://%s:80/refresh", address), nil)
	if err != nil {
		return fmt.Errorf("unable to build refresh request: %s", err)
	}

	req = req.WithContext(ctx)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to start replication: %s", err)
	}

	fmt.Printf("%s\tUI refreshed\n", address)

	client, err := pi.NewSSHClient(address)
	if err != nil {
		fmt.Printf("unable to ssh into %s: %s", address, err)
		os.Exit(1)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("unable to start new session: %s", err)
		client.Close()
		os.Exit(1)
	}

	fmt.Printf("Restarting DMM...\n")

	bytes, err := session.CombinedOutput("sudo systemctl restart device-monitoring.service")
	if err != nil {
		fmt.Printf("unable to reboot: %s\noutput on pi:\n%s\n", err, bytes)
		client.Close()
		os.Exit(1)
	}
	client.Close()
	fmt.Printf("Success\n")

	return nil
}
