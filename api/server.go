package main

import (
	"fmt"
	"net"
	"os"

	avcli "github.com/byuoitav/av-cli"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
)

func main() {
	var (
		port int
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run lazarette on")
	pflag.Parse()

	addr := fmt.Sprintf(":%v", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("failed to bind listener: %s\n", err)
		os.Exit(1)
	}

	cli := &avcli.Server{}
	server := grpc.NewServer()
	avcli.RegisterAvCliServer(server, cli)

	fmt.Printf("Starting server on %s\n", lis.Addr().String())

	if err := server.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %s\n", err)
		os.Exit(1)
	}
}
