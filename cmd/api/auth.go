package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authClient struct {
	Address  string
	Token    string
	Disabled bool
	Logger   *zap.Logger
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

func (client *authClient) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		netID, err := client.authenticate(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "netID", netID)

		return handler(ctx, req)
	}
}

func (client *authClient) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		netID, err := client.authenticate(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		md := metadata.Pairs("netID", netID)

		err = ss.SetHeader(md)
		if err != nil {
			return err
		}
		return handler(srv, ss)
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

	// build opa request
	var authReq authRequest
	authReq.Input.Token = strings.TrimPrefix(auth[0], "Bearer ")
	authReq.Input.User = user[0]
	authReq.Input.Method = method

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return "", fmt.Errorf("unable to marshal request body: %w", err)
	}

	client.Logger.Debug("Authenticating", zap.String("user", authReq.Input.User), zap.String("for", authReq.Input.Method))
	url := fmt.Sprintf("https://%s/v1/data/cli", client.Address)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("unable to build request: %w", err)
	}

	httpReq.Header.Add("authorization", "Bearer "+client.Token)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got a %v from auth server. response:\n%s", resp.StatusCode, respBody)
	}

	var authResp authResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return "", fmt.Errorf("unable to unmarshal response: %w", err)
	}

	if !authResp.Result.Allow {
		client.Logger.Debug("Not authorized", zap.String("user", authResp.Result.User), zap.String("for", authReq.Input.Method))
		return "", errNotAuthorized
	}

	client.Logger.Debug("Authorized", zap.String("user", authResp.Result.User), zap.String("for", authReq.Input.Method), zap.String("from", authResp.Result.OriginatingClient))
	return authResp.Result.User, nil
}
