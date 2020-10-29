package pi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [device ID]",
	Short: "get a screenshot of a pi",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()

		idToken := wso2.GetIDToken()
		auth := avcli.Auth{
			Token: idToken,
			User:  "",
		}

		client, err := avcli.NewClient(viper.GetString("api"), auth)
		if err != nil {
			fail("unable to create client: %v\n", err)
		}

		result, err := client.Screenshot(ctx, &avcli.ID{Id: args[0]})
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

			fail("unable to get screenshot: %s\n", err)
		case result == nil:
			fail("received invalid response from server (nil result)\n")
		}

		// pick an address to bind to
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			fail("unable to get network interfaces: %s\n", err)
		}

		var addr string
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				addr = ipNet.IP.String()
				break
			}
		}

		stopSrv := make(chan struct{})
		lis, err := net.Listen("tcp", addr+":0")
		if err != nil {
			fail("unable to bind listener: %s\n", err)
		}

		// serve
		go func() {
			srv := &http.Server{}

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(result.GetPhoto())
				if err != nil {
					fail("failed to send photo response: %s\n", err)
				}

				stopSrv <- struct{}{}
				cancel()
			})

			go func() {
				_ = srv.Serve(lis)
			}()

			<-stopSrv
			srv.Close()
		}()

		url, err := url.Parse(fmt.Sprintf("http://%s/", lis.Addr().String()))
		if err != nil {
			fail("unable to parse url: %s\n", err)
		}

		err = wso2.OpenBrowser(url.String())
		if err != nil {
			fmt.Printf("Unable to open browser: %s. Copy the below URL into your browser to see your screenshot:\n%s\n\n", err, color.New(color.FgBlue, color.Bold, color.Underline).Sprint(url.String()))
		}

		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Printf("Timed out waiting for you to view the screenshot.\n")
		}
	},
}
