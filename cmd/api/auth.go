package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authClient struct {
	Address  string
	Token    string
	Disabled bool
	Log      *zap.Logger
}

type authRequest struct {
	Input struct {
		Token  string `json:"token"`
		User   string `json:"user"`
		Method string `json:"method"`
	} `json:"input"`
}

type authResponse struct {
	DecisionID string `json:"decision_id"`
	Result     struct {
		Allow             bool   `json:"allow"`
		OriginatingClient string `json:"originating_client"`
		User              string `json:"user"`
	} `json:"result"`
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

type (
	ctxKey int
)

const (
	requestIDMetadataKey = "x-request-id"
	netIDMetadataKey     = "x-net-id"

	requestIDCtxKey ctxKey = iota + 1
	netIDCtxKey
)

// TODO support using an existing request id header?
func generateRequestID(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	id, err := ksuid.NewRandom()
	if err != nil {
		return ctx
	}

	ctx = context.WithValue(ctx, requestIDCtxKey, id.String())
	md.Set(requestIDMetadataKey, id.String())
	return metadata.NewIncomingContext(ctx, md)
}

func (client *authClient) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = generateRequestID(ctx)

		netID, err := client.authenticate(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		md.Set(netIDMetadataKey, netID)
		return handler(metadata.NewIncomingContext(ctx, md), req)
	}
}

func (client *authClient) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := generateRequestID(ss.Context())

		netID, err := client.authenticate(ctx, info.FullMethod)
		if err != nil {
			return err
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		md.Set(netIDMetadataKey, netID)
		wss := &wrappedServerStream{
			ctx:          metadata.NewIncomingContext(ctx, md),
			ServerStream: ss,
		}

		return handler(srv, wss)
	}
}

func (client *authClient) authenticate(ctx context.Context, method string) (string, error) {
	if client.Disabled {
		return "", nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}

	auth := md["authorization"]
	user := md["x-user"]

	if len(auth) == 0 {
		return "", errMissingToken
	}

	if len(user) == 0 {
		return "", errMissingUser
	}

	log := client.Log
	if len(md[requestIDMetadataKey]) > 0 {
		log = log.With(zap.String("requestID", md[requestIDMetadataKey][0]))
	}

	// build opa request
	var authReq authRequest
	authReq.Input.Token = strings.TrimPrefix(auth[0], "Bearer ")
	authReq.Input.User = user[0]
	authReq.Input.Method = method

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return "", fmt.Errorf("unable to marshal request body: %w", err)
	}

	log.Debug("Sending authentication request", zap.ByteString("body", reqBody))
	url := fmt.Sprintf("https://%s/v1/data/cli", client.Address)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		log.Warn("unable to create request", zap.Error(err))
		return "", fmt.Errorf("unable to build request: %w", err)
	}

	httpReq.Header.Add("authorization", "Bearer "+client.Token)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Warn("unable to do request", zap.Error(err))
		return "", fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("unable to read response", zap.Error(err))
		return "", fmt.Errorf("unable to read response: %w", err)
	}

	log.Debug("Authentication response", zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", respBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got a %v from auth server. response:\n%s", resp.StatusCode, respBody)
	}

	var authResp authResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		log.Warn("unable to parse body from opa", zap.Error(err))
		return "", fmt.Errorf("unable to unmarshal response: %w", err)
	}

	if !authResp.Result.Allow {
		return "", errNotAuthorized
	}

	return authResp.Result.User, nil
}
