/*
Copyright 2025 The Kubernetes Authors.

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

package discovery

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net/http"
)

type UserInfo struct {
	UniverseID string
	ClientID   string
}

// AuthenticateClientToUniverse extracts the Universe ID and Client ID from the mTLS connection.
// The Universe ID is defined as the SHA256 hash of the root CA certificate (DER bytes)
// presented in the client's certificate chain.
// The Client ID is taken from the Common Name (CN) of the leaf certificate.
func AuthenticateClientToUniverse(r *http.Request, universeID string) (*UserInfo, error) {
	if r.TLS == nil {
		return nil, fmt.Errorf("no TLS connection")
	}
	if len(r.TLS.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no client certificate presented")
	}

	// Verify the chain is valid, though we don't validate that the CA certificate is trusted.
	var verifiedChains [][]*x509.Certificate
	{
		peerCertificates := r.TLS.PeerCertificates

		opts := x509.VerifyOptions{
			Roots:         x509.NewCertPool(),
			Intermediates: x509.NewCertPool(),
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}

		for i := 1; i < len(peerCertificates); i++ {
			if i == len(peerCertificates)-1 {
				// Last cert is the root
				opts.Roots.AddCert(peerCertificates[i])
			} else {
				opts.Intermediates.AddCert(peerCertificates[i])
			}
		}

		chains, err := peerCertificates[0].Verify(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to verify client certificate chain: %w", err)
		}
		verifiedChains = chains
	}

	// The universe ID must match at least one of the certificates in the chain (typically the root CA).
	var matchingChain []*x509.Certificate
	for _, verifiedChain := range verifiedChains {
		for _, cert := range verifiedChain {
			hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
			calculatedUniverseID := hex.EncodeToString(hash[:])
			if calculatedUniverseID == universeID {
				matchingChain = verifiedChain
				break
			}
		}
	}

	if matchingChain == nil {
		return nil, fmt.Errorf("client certificate chain does not match universe ID")
	}

	clientID := matchingChain[0].Subject.CommonName
	if clientID == "" {
		return nil, fmt.Errorf("client certificate missing Common Name")
	}

	return &UserInfo{
		UniverseID: universeID,
		ClientID:   clientID,
	}, nil
}
