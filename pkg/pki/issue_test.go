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

package pki

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKeystore struct {
	t      *testing.T
	signer string
	cert   *Certificate
	key    *PrivateKey

	invoked bool
}

// FindPrimaryKeypair implements Keystore
func (m *mockKeystore) FindPrimaryKeypair(ctx context.Context, name string) (*Certificate, *PrivateKey, error) {
	assert.False(m.t, m.invoked, "invoked already")
	m.invoked = true
	assert.Equal(m.t, m.signer, name, "name argument")
	return m.cert, m.key, nil
}

func TestIssueCert(t *testing.T) {
	origSize := os.Getenv("KOPS_RSA_PRIVATE_KEY_SIZE")
	os.Unsetenv("KOPS_RSA_PRIVATE_KEY_SIZE")
	defer func() {
		os.Setenv("KOPS_RSA_PRIVATE_KEY_SIZE", origSize)
	}()

	// Generate a new RSA key pair using rsa.GenerateKey
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create pki.PrivateKey wrapper for CA key
	caPrivateKey := &PrivateKey{Key: caKey}

	// Create the CA
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	caCert, err := x509.ParseCertificate(caCertDER)
	require.NoError(t, err)
	caCertificate := &Certificate{Certificate: caCert}

	for _, tc := range []struct {
		name                string
		req                 IssueCertRequest
		expectedKeyUsage    x509.KeyUsage
		expectedExtKeyUsage []x509.ExtKeyUsage
		expectedSubject     pkix.Name
		expectedDNSNames    []string
		expectedIPAddresses []net.IP
	}{
		{
			name: "ca",
			req: IssueCertRequest{
				Type: "ca",
				Subject: pkix.Name{
					CommonName: "Test CA",
				},
			},
			expectedKeyUsage: x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
			expectedSubject:  pkix.Name{CommonName: "Test CA"},
		},
		{
			name: "client",
			req: IssueCertRequest{
				Type: "client",
				Subject: pkix.Name{
					CommonName:   "Test client",
					Organization: []string{"system:masters"},
				},
				Serial: BuildPKISerial(123456),
			},
			expectedKeyUsage:    x509.KeyUsageDigitalSignature,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			expectedSubject:     pkix.Name{CommonName: "Test client", Organization: []string{"system:masters"}},
		},
		{
			name: "clientOneYear",
			req: IssueCertRequest{
				Type: "client",
				Subject: pkix.Name{
					CommonName: "Test client",
				},
				Validity: time.Hour * 24 * 365,
			},
			expectedKeyUsage:    x509.KeyUsageDigitalSignature,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			expectedSubject:     pkix.Name{CommonName: "Test client"},
		},
		{
			req: IssueCertRequest{
				Type: "clientServer",
				Subject: pkix.Name{
					CommonName: "Test client/server",
				},
				AlternateNames: []string{"*.internal.test.cluster.local", "localhost", "127.0.0.1"},
				PrivateKey:     caPrivateKey,
			},
			expectedKeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			expectedSubject:     pkix.Name{CommonName: "Test client/server"},
			expectedDNSNames:    []string{"*.internal.test.cluster.local", "localhost"},
			expectedIPAddresses: []net.IP{net.ParseIP("127.0.0.1").To4()},
		},
		{
			name: "server",
			req: IssueCertRequest{
				Type: "server",
				Subject: pkix.Name{
					CommonName: "Test server",
				},
				AlternateNames: []string{"*.internal.test.cluster.local", "localhost", "127.0.0.1"},
				PrivateKey:     caPrivateKey,
			},
			expectedKeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			expectedSubject:     pkix.Name{CommonName: "Test server"},
			expectedDNSNames:    []string{"*.internal.test.cluster.local", "localhost"},
			expectedIPAddresses: []net.IP{net.ParseIP("127.0.0.1").To4()},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()

			var minExpectedValidity int64
			if tc.req.Validity == 0 {
				minExpectedValidity = time.Now().Add(time.Hour * 10 * 365 * 24).Unix()
			} else {
				minExpectedValidity = time.Now().Add(tc.req.Validity).Unix()
			}

			var keystore Keystore
			if tc.req.Type != "ca" {
				tc.req.Signer = tc.name + "-signer"
				keystore = &mockKeystore{
					t:      t,
					signer: tc.req.Signer,
					cert:   caCertificate,
					key:    caPrivateKey,
				}
			}
			certificate, key, caCert, err := IssueCert(ctx, &tc.req, keystore)
			require.NoError(t, err)

			cert := certificate.Certificate
			if tc.req.Signer == "" {
				assert.Equal(t, cert.Issuer, cert.Subject, "self-signed")
				assert.NoError(t, cert.CheckSignatureFrom(cert), "check signature")
				assert.Equal(t, certificate, caCert, "returned CA cert")
			} else {
				assert.Equal(t, cert.Issuer, caCertificate.Certificate.Subject, "cert signer")
				assert.NoError(t, cert.CheckSignatureFrom(caCertificate.Certificate), "check signature")
				assert.Equal(t, caCertificate, caCert, "returned CA cert")
			}

			// type
			assert.Equal(t, certificate.IsCA, cert.IsCA, "IsCA matches")
			assert.Equal(t, tc.req.Type == "ca", cert.IsCA, "IsCA")
			assert.True(t, cert.BasicConstraintsValid, "BasicConstraintsValid")
			assert.Equal(t, tc.expectedKeyUsage, cert.KeyUsage, "KeyUsage")
			assert.ElementsMatch(t, tc.expectedExtKeyUsage, cert.ExtKeyUsage)
			assert.Nil(t, cert.ExtraExtensions, "ExtraExtensions")

			// subject
			assert.Equal(t, certificate.Subject, cert.Subject, "Subject matches")
			actualName := cert.Subject
			actualName.Names = nil
			assert.Equal(t, tc.expectedSubject, actualName, "Subject")

			// alternateNames
			assert.Equal(t, tc.expectedDNSNames, cert.DNSNames, "DNSNames")
			assert.Equal(t, tc.expectedIPAddresses, cert.IPAddresses, "IPAddresses")
			assert.Empty(t, cert.EmailAddresses, "EmailAddresses")

			// privateKey
			rsaPrivateKey, ok := key.Key.(*rsa.PrivateKey)
			require.True(t, ok, "private key is RSA")
			if tc.req.PrivateKey == nil {
				assert.Equal(t, 2048, rsaPrivateKey.N.BitLen(), "Private key length")
			} else {
				assert.Equal(t, tc.req.PrivateKey, key, "Private key")
			}
			assert.Equal(t, &rsaPrivateKey.PublicKey, cert.PublicKey, "certificate public key matches private key")
			assert.Equal(t, certificate.PublicKey, cert.PublicKey, "PublicKey")

			// serial
			if tc.req.Serial != nil {
				assert.Equal(t, cert.SerialNumber, tc.req.Serial, "SerialNumber")
			} else {
				assert.Greater(t, cert.SerialNumber.BitLen(), 110, "SerialNumber bit length")
			}

			// validity
			var maxExpectedValidity int64
			if tc.req.Validity == 0 {
				maxExpectedValidity = time.Now().Add(time.Hour * 10 * 365 * 24).Unix()
			} else {
				maxExpectedValidity = time.Now().Add(tc.req.Validity).Unix()
			}
			assert.Less(t, cert.NotBefore.Unix(), time.Now().Add(time.Hour*-47).Unix(), "NotBefore")
			assert.GreaterOrEqual(t, cert.NotAfter.Unix(), minExpectedValidity, "NotAfter")
			assert.LessOrEqual(t, cert.NotAfter.Unix(), maxExpectedValidity, "NotAfter")
		})
	}
}
