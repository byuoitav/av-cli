package avcli

import (
	"context"
	"crypto/x509"
	"fmt"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	Auth      Auth
	cliClient AvCliClient
}

func NewClient(addr string, a Auth, opts ...grpc.DialOption) (*Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("unable to get system cert pool: %v", err)
	}

	//check to see if it matches
	add := true
	for i := range opts {
		if opts[i] == grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")) {
			add = false
		}
	}
	if add {
		opts = append(opts, getTransportSecurityDialOption(pool))
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("error making grpc connection: %v", err)
	}

	cli := NewAvCliClient(conn)

	return &Client{
		Auth:      a,
		cliClient: cli,
	}, nil
}

func (c *Client) Swab(ctx context.Context, in *ID, opts ...grpc.CallOption) (AvCli_SwabClient, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.Swab(ctx, in, opts...)
}

func (c *Client) Float(ctx context.Context, in *ID, opts ...grpc.CallOption) (AvCli_FloatClient, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.Float(ctx, in, opts...)
}

func (c *Client) Screenshot(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ScreenshotResult, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.Screenshot(ctx, in, opts...)
}

func (c *Client) DuplicateRoom(ctx context.Context, in *DuplicateRoomRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.DuplicateRoom(ctx, in, opts...)
}

func (c *Client) FixTime(ctx context.Context, in *ID, opts ...grpc.CallOption) (AvCli_FixTimeClient, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.FixTime(ctx, in, opts...)
}

func (c *Client) Sink(ctx context.Context, in *ID, opts ...grpc.CallOption) (AvCli_SinkClient, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.Sink(ctx, in, opts...)
}

func (c *Client) CloseMonitoringIssue(ctx context.Context, in *ID, opts ...grpc.CallOption) (*empty.Empty, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.CloseMonitoringIssue(ctx, in, opts...)
}

func (c *Client) SetLogLevel(ctx context.Context, in *SetLogLevelRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	opts = c.checkOpts(opts...)
	return c.cliClient.SetLogLevel(ctx, in, opts...)
}

func (c *Client) checkOpts(opts ...grpc.CallOption) []grpc.CallOption {
	add := true
	for i := range opts {
		if opts[i] == grpc.PerRPCCredentials(c.Auth) {
			add = false
		}
	}
	if add {
		opts = append(opts, grpc.PerRPCCredentials(c.Auth))
	}

	return opts
}

func GetNetID(ctx context.Context) (string, error) {
	val := ctx.Value("netID")
	if netID, ok := val.(string); ok {
		return netID, nil
	}
	return "", errors.New("unable to get netID from context")
}
