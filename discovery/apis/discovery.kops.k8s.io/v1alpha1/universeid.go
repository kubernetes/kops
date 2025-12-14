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
