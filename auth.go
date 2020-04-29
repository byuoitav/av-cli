package avcli

import (
	"context"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type auth struct {
	token string
	user  string
}

func (auth) RequireTransportSecurity() bool {
	return true
}

func (a auth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + a.token,
		"x-user":        a.user,
	}, nil
}

func getTransportSecurityDialOption(pool *x509.CertPool) grpc.DialOption {
	if !(auth{}).RequireTransportSecurity() {
		return grpc.WithInsecure()
	}

	return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, ""))
}
