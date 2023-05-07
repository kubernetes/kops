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
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/nodeup"
	pb "k8s.io/kops/proto/kops/bootstrap/v1"
)

type ChallengeServer struct {
	tlsConfig *tls.Config
	servingCA []byte

	mutex      sync.Mutex
	challenges map[string]*Challenge

	RequiredSubject pkix.Name

	pb.UnimplementedCallbackServiceServer
}

func NewChallengeServer(clusterName string, caBundle []byte) (*ChallengeServer, error) {
	serverCertificate, err := BuildChallengeServerCertificate(clusterName)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*serverCertificate},
	}

	var servingCA bytes.Buffer
	for _, cert := range serverCertificate.Certificate {
		if err := pem.Encode(&servingCA, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
			return nil, err
		}
	}

	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caBundle) {
		return nil, fmt.Errorf("unable to build client-cert CA pools")
	}
	tlsConfig.ClientCAs = clientCAs
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	return &ChallengeServer{
		RequiredSubject: challengeKopsControllerSubject(clusterName),
		tlsConfig:       tlsConfig,
		servingCA:       servingCA.Bytes(),
	}, nil
}

type Challenge struct {
	ChallengeID     string
	ChallengeSecret []byte
}

func (s *ChallengeServer) createChallenge() *Challenge {
	c := &Challenge{}
	c.ChallengeID = hex.EncodeToString(randomBytes(16))
	c.ChallengeSecret = randomBytes(16)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.challenges == nil {
		s.challenges = make(map[string]*Challenge)
	}
	s.challenges[c.ChallengeID] = c
	return c
}

type ChallengeListener struct {
	endpoint   string
	server     *ChallengeServer
	grpcServer *grpc.Server
}

func (s *ChallengeListener) CreateChallenge() *nodeup.ChallengeRequest {
	challenge := s.server.createChallenge()

	return &nodeup.ChallengeRequest{
		Endpoint:        s.Endpoint(),
		ChallengeID:     challenge.ChallengeID,
		ChallengeSecret: challenge.ChallengeSecret,
		ServerCA:        s.server.servingCA,
	}
}

func (s *ChallengeListener) Stop() {
	s.grpcServer.Stop()
}

func (s *ChallengeListener) Endpoint() string {
	return s.endpoint
}

func (s *ChallengeServer) NewListener(ctx context.Context, listen string) (*ChallengeListener, error) {
	var opts []grpc.ServerOption

	opts = append(opts, grpc.Creds(credentials.NewTLS(s.tlsConfig)))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterCallbackServiceServer(grpcServer, s)

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, fmt.Errorf("error listening on %q: %w", listen, err)
	}

	grpcListener := &ChallengeListener{
		server:     s,
		grpcServer: grpcServer,
		endpoint:   lis.Addr().String(),
	}

	go func() {
		klog.Infof("starting node-challenge listener on %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			lis.Close()

			klog.Warningf("error serving GRPC: %v", err)
		}
	}()

	return grpcListener, nil
}

// Answers challenges to cross-check bootstrap requests.
func (s *ChallengeServer) Challenge(ctx context.Context, req *pb.ChallengeRequest) (*pb.ChallengeResponse, error) {
	klog.Infof("got node-challenge request")
	// Explicitly authenticate the username for safety
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		klog.Warningf("no peer in context")
		return nil, status.Error(codes.Unauthenticated, "peer was nil")
	}

	tlsInfo, ok := peerInfo.AuthInfo.(credentials.TLSInfo)
	if !ok {
		klog.Warningf("peer.AuthInfo was of unexpected type %T", peerInfo.AuthInfo)
		return nil, status.Error(codes.Unauthenticated, "unexpected peer transport credentials")
	}

	if len(tlsInfo.State.VerifiedChains) == 0 || len(tlsInfo.State.VerifiedChains[0]) == 0 {
		klog.Warningf("no VerifiedChains in TLSInfo")
		return nil, status.Error(codes.Unauthenticated, "verified chains were empty")
	}

	if got, want := tlsInfo.State.VerifiedChains[0][0].Subject, s.RequiredSubject; !subjectsMatch(got, want) {
		klog.Warningf("certificate subjects did not match expected; got %q, want %q", got, want)
		return nil, status.Error(codes.Unauthenticated, "certificate subjects did not match")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := req.ChallengeId
	if key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "challenge_id is required")
	}

	challenge := s.challenges[key]
	if challenge == nil {
		return nil, status.Errorf(codes.NotFound, "challenge was not found")
	}
	// Prevent replay attacks
	delete(s.challenges, key)

	hash := buildChallengeResponse(challenge.ChallengeSecret, req.GetChallengeRandom())
	response := &pb.ChallengeResponse{
		ChallengeResponse: hash,
	}

	return response, nil
}

func buildChallengeResponse(nodeNonce []byte, kopsControllerNonde []byte) []byte {
	// Arguably this is overkill because the TLS handshake is stronger and everything is encrypted.
	hasher := sha256.New()
	hasher.Sum(nodeNonce)
	hasher.Sum(kopsControllerNonde)

	hash := hasher.Sum(nil)

	return hash
}
