package v1alpha1

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
)

func ComputeUniverseIDFromCertificate(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	universeID := base64.RawURLEncoding.EncodeToString(hash[:])
	return universeID
}
