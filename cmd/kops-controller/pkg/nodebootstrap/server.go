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

package nodebootstrap

import (
	"context"
	"crypto"
	crypto_rand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"

	"google.golang.org/grpc/peer"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog"
	"k8s.io/kops/node-authorizer/pkg/server"
	"k8s.io/kops/pkg/pki"
	pb "k8s.io/kops/pkg/proto/nodebootstrap"
	"k8s.io/kops/pkg/rbac"
)

const (
	// the namespace to place the secrets
	tokenNamespace = "kube-system"
)

type nodeBootstrapService struct {
	authorizer server.Authorizer

	signerCert       *x509.Certificate
	signerPrivateKey crypto.Signer

	options Options
}

// Options is the configuration for the NodeBootstrap server
type Options struct {
	CertificateTTL time.Duration `json:"certificateTTL,omitempty"`

	SignerCertificatePath string `json:"signerCertificatePath,omitempty"`
	SignerKeyPath         string `json:"signerKeyPath,omitempty"`
}

// PopulateDefaults sets the default configuration values
func (o *Options) PopulateDefaults() {
	o.CertificateTTL = 15 * time.Minute
}

func NewNodeBootstrapService(signerCert *x509.Certificate, signerPrivateKey crypto.Signer, authorizer server.Authorizer, options *Options) (*nodeBootstrapService, error) {
	s := &nodeBootstrapService{}

	s.signerCert = signerCert
	s.signerPrivateKey = signerPrivateKey

	s.authorizer = authorizer

	s.options = *options

	return s, nil
}

var _ pb.NodeBootstrapServiceServer = &nodeBootstrapService{}

func (s *nodeBootstrapService) CreateKubeletBootstrapToken(ctx context.Context, request *pb.CreateKubeletBootstrapTokenRequest) (*pb.CreateKubeletBootstrapTokenResponse, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer context")
	}

	if request.NodeName == "" {
		return nil, fmt.Errorf("node_name not supplied")
	}

	var publicKey interface{}
	if request.PublicKey != nil && request.PublicKey.PemData != nil {
		publicKeys, err := keyutil.ParsePublicKeysPEM(request.PublicKey.PemData)
		if err != nil {
			return nil, fmt.Errorf("error parsing public key: %v", err)
		}

		if len(publicKeys) == 0 {
			return nil, fmt.Errorf("no public key was parsed")
		}
		if len(publicKeys) != 1 {
			return nil, fmt.Errorf("multiple public keys were parsed")
		}

		publicKey = publicKeys[0]
	} else {
		return nil, fmt.Errorf("no public key data supplied")
	}

	{
		nodeRegistration := &server.NodeRegistration{}
		nodeRegistration.Spec.RemoteAddr = peer.Addr.String()

		if err := s.authorizer.Authorize(ctx, nodeRegistration); err != nil {
			// In general we prefer to log error details here, and only return minimal information to the client, because it may be an attacker
			klog.Warningf("internal error during authorization: %v", err)
			return nil, fmt.Errorf("internal error during authorization")
		}

		if !nodeRegistration.Status.Allowed {
			// TODO: Use grpc Status errors?
			return nil, fmt.Errorf("node registration not allowed")
		}
	}

	pemData, err := s.createBootstrapToken(ctx, request.NodeName, publicKey)
	if err != nil {
		klog.Warningf("error creating bootstrap token: %v", err)
		return nil, fmt.Errorf("error creating bootstrap token")
	}

	response := &pb.CreateKubeletBootstrapTokenResponse{}
	if pemData != nil {
		response.Certificate = &pb.Certificate{
			PemData: pemData,
		}
	}

	return response, nil
}

// createBootstrapToken generates a bootstrap token for the node, inserting it into k8s
func (s *nodeBootstrapService) createBootstrapToken(ctx context.Context, nodeName string, publicKey interface{}) ([]byte, error) {

	// Build a Certificate template; note that for security reasons we don't allow the client to build it
	now := time.Now()

	template := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Allow for a small amount of clock-skew
	template.NotBefore = now.Add(time.Hour * -1)

	// Rotate the cert fairly quickly
	template.NotAfter = now.Add(s.options.CertificateTTL)

	template.Subject = pkix.Name{
		CommonName:   fmt.Sprintf("system:node:%s", nodeName),
		Organization: []string{rbac.NodesGroup},
	}

	// https://tools.ietf.org/html/rfc5280#section-4.2.1.3
	//
	// Digital signature allows the certificate to be used to verify
	// digital signatures used during TLS negotiation.
	template.KeyUsage = template.KeyUsage | x509.KeyUsageDigitalSignature
	// KeyEncipherment allows the cert/key pair to be used to encrypt
	// keys, including the symmetric keys negotiated during TLS setup
	// and used for data transfer.
	template.KeyUsage = template.KeyUsage | x509.KeyUsageKeyEncipherment
	// ClientAuth allows the cert to be used by a TLS client to
	// authenticate itself to the TLS server.
	template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)

	template.SerialNumber = pki.BuildPKISerial(now.UnixNano())

	certificateData, err := x509.CreateCertificate(crypto_rand.Reader, template, s.signerCert, publicKey, s.signerPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating certificate: %v", err)
	}

	b := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateData})

	return b, nil
}
