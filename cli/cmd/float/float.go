package float

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

		conn, err := grpc.Dial(viper.GetString("api"), grpc.WithInsecure())
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

		stream, err := cli.Float(context.TODO(), &avcli.ID{Id: arg[0], Designation: designation})
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
				return
			}
			if in.Error != "" {
				fmt.Printf("there was an error floating to %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully swabbed %s\n", in.Id)
			}
		}
	},
}

func floatship(deviceID, designation string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/%v/webhook_device/%v", designation, deviceID), nil)
	if err != nil {
		return fmt.Errorf("couldn't make request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", wso2.GetAccessToken()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't perform request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("couldn't read the response body: %v", err)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-200 status code: %v - %s", resp.StatusCode, body)
	}

	fmt.Printf("Deployment successful\n")
	return nil
}
