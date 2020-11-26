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
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueCert(t *testing.T) {
	privateKey, err := ParsePEMPrivateKey([]byte("-----BEGIN RSA PRIVATE KEY-----\nMIIBOQIBAAJBANgL5cR2cLOB7oZZTiuiUmMwQRBaia8yLULt+XtBtDHf0lPOrn78\nvLPh7P7zRBgHczbTddcsg68g9vAfb9TC5M8CAwEAAQJAJytxCv+WS1VhU4ZZf9u8\nKDOVeEuR7uuf/SR8OPaenvPqONpYbZSVjnWnRBRHvg3HaHchQqH32UljZUojs9z4\nEQIhAO/yoqCFckfqswOGwWyYX1oNOtU8w9ulXlZqAtZieavVAiEA5n/tKHoZyx3U\nbZcks/wns1WqhAoSmDJpMyVXOVrUlBMCIDGnalQBiYasYOMn7bsFRSYjertJ2dYI\nQJ9tTK0Er90JAiAmpVQx8SbZ80pmhWzV8HUHkFligf3UHr+cn6ocJ6p0mQIgB728\npdvrS5zRPoUN8BHfWOZcPrElKTuJjP2kH6eNPvI=\n-----END RSA PRIVATE KEY-----"))
	require.NoError(t, err)

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
				PrivateKey:     privateKey,
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
				PrivateKey:     privateKey,
			},
			expectedKeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			expectedSubject:     pkix.Name{CommonName: "Test server"},
			expectedDNSNames:    []string{"*.internal.test.cluster.local", "localhost"},
			expectedIPAddresses: []net.IP{net.ParseIP("127.0.0.1").To4()},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var minExpectedValidity int64
			if tc.req.Validity == 0 {
				minExpectedValidity = time.Now().Add(time.Hour * 10 * 365 * 24).Unix()
			} else {
				minExpectedValidity = time.Now().Add(tc.req.Validity).Unix()
			}

			keystore, _ := NewMockKeystore()
			if tc.req.Type != "ca" {
				tc.req.Signer = tc.name + "-signer"
				keystore.Signer = tc.req.Signer
			}
			certificate, key, caCert, err := IssueCert(&tc.req, keystore)
			require.NoError(t, err)

			cert := certificate.Certificate

			caCertificate := keystore.cert

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
