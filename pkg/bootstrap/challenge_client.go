/*
Copyright 2023 The Kubernetes Authors.

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

package bootstrap

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/pki"
	pb "k8s.io/kops/proto/kops/bootstrap/v1"
	"k8s.io/kops/upup/pkg/fi"
)

type ChallengeClient struct {
	keystore pki.Keystore
}

func NewChallengeClient(keystore pki.Keystore) (*ChallengeClient, error) {
	return &ChallengeClient{
		keystore: keystore,
	}, nil
}

func (c *ChallengeClient) getClientCertificate(ctx context.Context, clusterName string) (*tls.Certificate, error) {
	subject := challengeKopsControllerSubject(clusterName)

	certificate, privateKey, _, err := pki.IssueCert(ctx, &pki.IssueCertRequest{
		Validity: 1 * time.Hour,
		Signer:   fi.CertificateIDCA,
		Type:     "client",
		Subject:  subject,
	}, c.keystore)
	if err != nil {
		return nil, fmt.Errorf("error creating certificate: %w", err)
	}

	// TODO: Caching and rotation
	clientCertificate := &tls.Certificate{
		PrivateKey:  privateKey.Key,
		Certificate: [][]byte{certificate.Certificate.Raw},
		Leaf:        certificate.Certificate,
	}
	return clientCertificate, nil
}

func (c *ChallengeClient) DoCallbackChallenge(ctx context.Context, clusterName string, targetEndpoint string, bootstrapRequest *nodeup.BootstrapRequest) error {
	challenge := bootstrapRequest.Challenge

	if challenge == nil {
		return fmt.Errorf("challenge not set")
	}
	if challenge.ChallengeID == "" {
		return fmt.Errorf("challenge.id not set")
	}
	if len(challenge.ChallengeSecret) == 0 {
		return fmt.Errorf("challenge.secret not set")
	}
	if challenge.Endpoint == "" {
		return fmt.Errorf("challenge.endpoint not set")
	}
	if len(challenge.ServerCA) == 0 {
		return fmt.Errorf("challenge.ca not set")
	}

	clientCertificate, err := c.getClientCertificate(ctx, clusterName)
	if err != nil {
		return err
	}

	serverCAs := x509.NewCertPool()
	if !serverCAs.AppendCertsFromPEM(challenge.ServerCA) {
		return fmt.Errorf("error loading certificate pool")
	}

	serverName := challengeServerHostName(clusterName)
	tlsConfig := &tls.Config{
		RootCAs:      serverCAs,
		Certificates: []tls.Certificate{*clientCertificate},
		ServerName:   serverName,
	}

	kospControllerNonce := randomBytes(16)
	req := &pb.ChallengeRequest{
		ChallengeId:     challenge.ChallengeID,
		ChallengeRandom: kospControllerNonce,
	}

	expectedChallengeResponse := buildChallengeResponse(challenge.ChallengeSecret, kospControllerNonce)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	conn, err := grpc.DialContext(ctx, targetEndpoint, opts...)
	if err != nil {
		return fmt.Errorf("error dialing target %q: %w", targetEndpoint, err)
	}
	client := pb.NewCallbackServiceClient(conn)

	response, err := client.Challenge(ctx, req)
	if err != nil {
		return fmt.Errorf("error from callback challenge: %w", err)
	}

	if subtle.ConstantTimeCompare(response.GetChallengeResponse(), expectedChallengeResponse) != 1 {
		return fmt.Errorf("callback challenge returned wrong result")
	}
	return nil
}
