/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog"
	pb "k8s.io/kops/pkg/proto/nodebootstrap"
)

// Options is the configuration for the NodeBootstrap server
type Options struct {
	// Server is the GRPC server we should connect to
	Server string `json:"server"`

	// CACertificate is the CA certificate for the GRPC server
	CACertificate []byte `json:"caCertificate"`
}

// PopulateDefaults sets the default configuration values
func (o *Options) PopulateDefaults() {
}

type nodeBootstrapClient struct {
	options Options

	connection *grpc.ClientConn
	client     pb.NodeBootstrapServiceClient
}

func New(ctx context.Context, options *Options) (*nodeBootstrapClient, error) {
	var grpcOpts []grpc.DialOption

	{
		tlsConfig := &tls.Config{}
		if options.CACertificate != nil {
			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(options.CACertificate) {
				return nil, fmt.Errorf("could not parse CA certificate")
			}
			tlsConfig.RootCAs = certPool
		}

		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	conn, err := grpc.DialContext(ctx, options.Server, grpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("error dialing %q: %v", options.Server, err)
	}
	client := pb.NewNodeBootstrapServiceClient(conn)
	return &nodeBootstrapClient{
		connection: conn,
		options:    *options,
		client:     client,
	}, nil
}

func (c *nodeBootstrapClient) Close() error {
	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			return fmt.Errorf("error closing GRPC connection: %v", err)
		}
		c.connection = nil
	}

	return nil
}

func (c *nodeBootstrapClient) CreateKubeletBootstrapToken(ctx context.Context) (pb.Token, error) {
	request := &pb.CreateKubeletBootstrapTokenRequest{}

	klog.V(2).Infof("sending CreateKubeletBootstrapRequest %v", request)

	response, err := c.client.CreateKubeletBootstrapToken(ctx, request)
	if err != nil {
		return pb.Token{}, fmt.Errorf("error creating bootstrap token: %v", err)
	}

	if response.Token == nil || response.Token.BearerToken == "" {
		return pb.Token{}, fmt.Errorf("created bootstrap token, but response was empty")
	}

	return *response.Token, nil
}
