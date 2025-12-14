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

package v1alpha1

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func ComputeUniverseIDFromPEM(cert []byte) (string, error) {
	// Parse client CA certificate to find the public key info
	block, _ := pem.Decode(cert)
	if block == nil {
		// Safe to log because this is the cert, not the key
		return "", fmt.Errorf("no PEM certificate data found in client CA certificate: %q", cert)
	}
	if block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("expected CERTIFICATE PEM block in client CA certificate, got: %q", block.Type)
	}
	clientCACertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("error parsing client CA certificate: %w", err)
	}
	return ComputeUniverseIDFromCertificate(clientCACertificate), nil
}

func ComputeUniverseIDFromCertificate(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	universeID := base64.RawURLEncoding.EncodeToString(hash[:])
	return universeID
}
