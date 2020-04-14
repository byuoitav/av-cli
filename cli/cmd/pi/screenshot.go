package pi

import (
	"context"
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
	"google.golang.org/grpc"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [device ID]",
	Short: "get a screenshot of a pi",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		// TODO this url should be a constant
		conn, err := grpc.DialContext(ctx, "localhost:9999", grpc.WithInsecure())
		if err != nil {
			fmt.Printf("unable to connect to api: %s\n", err)
			os.Exit(1)
		}

		cli := avcli.NewAvCliClient(conn)

		result, err := cli.Screenshot(ctx, &avcli.ID{Id: args[0]})
		switch {
		case err != nil:
			fmt.Printf("unable to get screenshot: %s\n", err)
			os.Exit(1)
		case result == nil:
			fmt.Printf("received invalid response from server (nil result)\n")
			os.Exit(1)
		}

		// pick an address to bind to
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			fmt.Printf("unable to get network interfaces: %s\n", err)
			os.Exit(1)
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
			fmt.Printf("unable to bind listener: %s\n", err)
			os.Exit(1)
		}

		// serve
		go func() {
			srv := &http.Server{}

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(result.GetPhoto())
				switch {
				case err != nil:
					fmt.Printf("failed to send photo response: %s\n", err)
					os.Exit(1)
				}

				stopSrv <- struct{}{}
				cancel()
			})

			go func() {
				err = srv.Serve(lis)
				if err != nil {
					fmt.Printf("failed to start photo server: %s\n", err)
					os.Exit(1)
				}
			}()

			<-stopSrv
			srv.Close()
		}()

		url, err := url.Parse(fmt.Sprintf("http://%s/", lis.Addr().String()))
		if err != nil {
			fmt.Printf("unable to parse url: %s\n", err)
			os.Exit(1)
		}

		err = wso2.OpenBrowser(url.String())
		if err != nil {
			fmt.Printf("Unable to open browser: %s. Copy the below URL into your browser to see your screenshot:\n%s\n", err, color.New(color.FgBlue, color.Bold, color.Underline).Sprint(url.String()))
		}

		<-ctx.Done()
	},
}
