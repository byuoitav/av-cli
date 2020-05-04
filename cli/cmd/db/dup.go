package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	avcli "github.com/byuoitav/av-cli"
	arg "github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var dupCmd = &cobra.Command{
	Use:   "dup [dst room ID] [src room ID]",
	Short: "Duplicate a room",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("must include destination and source room ID")
		}

		if err := arg.ValidRoomID(cmd, args); err != nil {
			return fmt.Errorf("invalid destination room: %v", err)
		}

		if err := arg.ValidRoomID(cmd, args[1:]); err != nil {
			return fmt.Errorf("invalid source room: %v", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Duplicating %s to %s\n", args[1], args[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		fmt.Printf("Select source database:\n")
		_, srcDesignation, err := arg.GetDB()
		if err != nil {
			fmt.Printf("prompt failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSelect dest database:\n")
		_, dstDesignation, err := arg.GetDB()
		if err != nil {
			fmt.Printf("prompt failed: %s\n", err)
			os.Exit(1)
		}

		dst := strings.ToUpper(args[0])
		src := strings.ToUpper(args[1])

		idToken := wso2.GetIDToken()
		auth := avcli.Auth{
			Token: idToken,
			User:  "",
		}

		client, err := avcli.NewClient(viper.GetString("api"), auth)
		if err != nil {
			fail("unable to create client: %v\n", err)
		}

		req := avcli.DuplicateRoomRequest{
			FromID:          src,
			FromDesignation: srcDesignation,
			ToID:            dst,
			ToDesignation:   dstDesignation,
		}
		result, err := client.DuplicateRoom(context.TODO(), &req)
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

			fail("unable to duplicate room: %s\n", err)
		case result == nil:
			fail("received invalid response from server (nil result)\n")
		}
		fmt.Printf("Successfully duplicated %s from %s\n", dst, src)
	},
}
