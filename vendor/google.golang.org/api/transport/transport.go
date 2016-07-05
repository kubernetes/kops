// Package transport supports network connections to HTTP and GRPC servers.
// This package is not intended for use by end developers. Use the
// google.golang.org/api/option package to configure API clients.
package transport

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"

	"google.golang.org/api/internal"
	"google.golang.org/api/option"
)

// NewHTTPClient returns an HTTP client for use communicating with a Google cloud
// service, configured with the given ClientOptions. It also returns the endpoint
// for the service as specified in the options.
func NewHTTPClient(ctx context.Context, opts ...option.ClientOption) (*http.Client, string, error) {
	var o internal.DialSettings
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if o.GRPCConn != nil {
		return nil, "", errors.New("unsupported gRPC connection specified")
	}
	// TODO(djd): Set UserAgent on all outgoing requests.
	if o.HTTPClient != nil {
		return o.HTTPClient, o.Endpoint, nil
	}
	if o.TokenSource == nil {
		var err error
		o.TokenSource, err = google.DefaultTokenSource(ctx, o.Scopes...)
		if err != nil {
			return nil, "", fmt.Errorf("google.DefaultTokenSource: %v", err)
		}
	}
	return oauth2.NewClient(ctx, o.TokenSource), o.Endpoint, nil
}

// DialGRPC returns a GRPC connection for use communicating with a Google cloud
// service, configured with the given ClientOptions.
func DialGRPC(ctx context.Context, opts ...option.ClientOption) (*grpc.ClientConn, error) {
	var o internal.DialSettings
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if o.HTTPClient != nil {
		return nil, errors.New("unsupported HTTP client specified")
	}
	if o.GRPCConn != nil {
		return o.GRPCConn, nil
	}
	if o.TokenSource == nil {
		var err error
		o.TokenSource, err = google.DefaultTokenSource(ctx, o.Scopes...)
		if err != nil {
			return nil, fmt.Errorf("google.DefaultTokenSource: %v", err)
		}
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(oauth.TokenSource{o.TokenSource}),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
	}
	grpcOpts = append(grpcOpts, o.GRPCDialOpts...)
	if o.UserAgent != "" {
		grpcOpts = append(grpcOpts, grpc.WithUserAgent(o.UserAgent))
	}
	return grpc.Dial(o.Endpoint, grpcOpts...)
}
