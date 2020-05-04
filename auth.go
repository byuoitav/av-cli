package avcli

import (
	"context"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Auth struct {
	Token string
	User  string
}

func (Auth) RequireTransportSecurity() bool {
	return false
}

func (a Auth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + a.Token,
		"x-user":        a.User,
	}, nil
}

func getTransportSecurityDialOption(pool *x509.CertPool) grpc.DialOption {
	if !(Auth{}).RequireTransportSecurity() {
		return grpc.WithInsecure()
	}

	return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, ""))
}
