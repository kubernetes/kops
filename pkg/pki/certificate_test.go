/*
Copyright 2017 The Kubernetes Authors.

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
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCertificate(t *testing.T) {
	data := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivYHVxH\nAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkEmDgA\ngvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj9WGF\nGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2QgJ/\nP2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgMwDTx\n9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABAoIBAA0ktjaTfyrAxsTI\nBezb7Zr5NBW55dvuII299cd6MJo+rI/TRYhvUv48kY8IFXp/hyUjzgeDLunxmIf9\n/Zgsoic9Ol44/g45mMduhcGYPzAAeCdcJ5OB9rR9VfDCXyjYLlN8H8iU0734tTqM\n0V13tQ9zdSqkGPZOIcq/kR/pylbOZaQMe97BTlsAnOMSMKDgnftY4122Lq3GYy+t\nvpr+bKVaQZwvkLoSU3rECCaKaghgwCyX7jft9aEkhdJv+KlwbsGY6WErvxOaLWHd\ncuMQjGapY1Fa/4UD00mvrA260NyKfzrp6+P46RrVMwEYRJMIQ8YBAk6N6Hh7dc0G\n8Z6i1m0CgYEA9HeCJR0TSwbIQ1bDXUrzpftHuidG5BnSBtax/ND9qIPhR/FBW5nj\n22nwLc48KkyirlfIULd0ae4qVXJn7wfYcuX/cJMLDmSVtlM5Dzmi/91xRiFgIzx1\nAsbBzaFjISP2HpSgL+e9FtSXaaqeZVrflitVhYKUpI/AKV31qGHf04sCgYEA6zTV\n99Sb49Wdlns5IgsfnXl6ToRttB18lfEKcVfjAM4frnkk06JpFAZeR+9GGKUXZHqs\nz2qcplw4d/moCC6p3rYPBMLXsrGNEUFZqBlgz72QA6BBq3X0Cg1Bc2ZbK5VIzwkg\nST2SSux6ccROfgULmN5ZiLOtdUKNEZpFF3i3qtsCgYADT/s7dYFlatobz3kmMnXK\nsfTu2MllHdRys0YGHu7Q8biDuQkhrJwhxPW0KS83g4JQym+0aEfzh36bWcl+u6R7\nKhKj+9oSf9pndgk345gJz35RbPJYh+EuAHNvzdgCAvK6x1jETWeKf6btj5pF1U1i\nQ4QNIw/QiwIXjWZeubTGsQKBgQCbduLu2rLnlyyAaJZM8DlHZyH2gAXbBZpxqU8T\nt9mtkJDUS/KRiEoYGFV9CqS0aXrayVMsDfXY6B/S/UuZjO5u7LtklDzqOf1aKG3Q\ndGXPKibknqqJYH+bnUNjuYYNerETV57lijMGHuSYCf8vwLn3oxBfERRX61M/DU8Z\nworz/QKBgQDCTJI2+jdXg26XuYUmM4XXfnocfzAXhXBULt1nENcogNf1fcptAVtu\nBAiz4/HipQKqoWVUYmxfgbbLRKKLK0s0lOWKbYdVjhEm/m2ZU8wtXTagNwkIGoyq\nY/C1Lox4f1ROJnCjc/hfcOjcxX5M8A8peecHWlVtUPKTJgxQ7oMKcw==\n-----END RSA PRIVATE KEY-----\n"
	publicKeyData := "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4JwpEprZ5n8RIEt6jT2l\nAh+UDgRgx/4px21gjgywQivYHVxHAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMD\nZVt+McFnWVwexnqBYFNcVjkEmDgAgvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+Cp\nOxyLhYZZNa0ZOZDHsSiJSQSj9WGFGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m\n74kjK4dsBhmjeq/7OAoTmiG2QgJ/P2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdG\nkwwZz2eF77aSPGmi/A2CSKgMwDTx9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF\n6QIDAQAB\n-----END RSA PUBLIC KEY-----\n"
	signerCertData := "-----BEGIN CERTIFICATE-----\nMIIBTDCB96ADAgECAhBjHcUz56MCdYqSYy7TYNe3MA0GCSqGSIb3DQEBCwUAMBUx\nEzARBgNVBAMTCnNlbGZzaWduZWQwHhcNMjAwNDI0MjMzNDM5WhcNMzAwNDI0MjMz\nNDM5WjAVMRMwEQYDVQQDEwpzZWxmc2lnbmVkMFwwDQYJKoZIhvcNAQEBBQADSwAw\nSAJBAL5zWUObMH5dBestQgDIa4B/rT7Cc21AK+B7gPvMcEfIWow5u6QE+EyhRTPv\n727oY+2MU9e4vq5RXBG7hneuBoECAwEAAaMjMCEwDgYDVR0PAQH/BAQDAgEGMA8G\nA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADQQBLUFz7gDKRRyjEwgRZnZzP\nOma9WIgOjX36OFllyGkspu1ZcW/EtGEGNXqtMsm1QmG38Lh7Nkehb5xoAmm6hkFA\n-----END CERTIFICATE-----"
	signerKeyData := "-----BEGIN RSA PRIVATE KEY-----\nMIIBOQIBAAJBAL5zWUObMH5dBestQgDIa4B/rT7Cc21AK+B7gPvMcEfIWow5u6QE\n+EyhRTPv727oY+2MU9e4vq5RXBG7hneuBoECAwEAAQJAZ9ZUUPwIEJ1/YJ4oYmzj\n0AfM2W8DqAlY4ufzh1YL0daGUkiuQg0p6CeqFqgnQluZ3bcXPG8iBQp1EeekULFL\nAQIhAOGbozbIEI+26Ehv41aCMWkKO1R05AVzmoNp1T2Ke6npAiEA2BtEPRSdhLek\nZR7vhk7KNTJ2XExJ+T/l2849EsojANkCIAWYD1b3ZPm7Rk0tgQyPE9yP5WK1t0Wv\nVSB3ClOJUIGpAiAfUBQbJZmNWW6gmFLsiw4RlzY/OW6ehvuvVbrTtiZMQQIgD2zY\nU2EjvR0zY5PsJYbcLHa9ieCA5ni/VW70WKn9K5s=\n-----END RSA PRIVATE KEY-----"

	key, err := ParsePEMPrivateKey([]byte(data))
	require.NoError(t, err, "ParsePEMPrivateKey")
	signerKey, err := ParsePEMPrivateKey([]byte(signerKeyData))
	require.NoError(t, err, "ParsePEMPrivateKey")

	{
		var b bytes.Buffer
		pkData, err := x509.MarshalPKIXPublicKey(key.Key.(*rsa.PrivateKey).Public())
		require.NoError(t, err, "MarshalPKIXPublicKey")

		err = pem.Encode(&b, &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkData})
		require.NoError(t, err, "serializing public key")

		require.Equal(t, b.String(), publicKeyData)
	}

	signer, err := ParsePEMCertificate([]byte(signerCertData))
	require.NoError(t, err, "ParsePEMCertificate")

	for _, tc := range []struct {
		name                string
		template            x509.Certificate
		signer              *x509.Certificate
		signerKey           *PrivateKey
		expectedExtKeyUsage []x509.ExtKeyUsage
	}{
		{
			name: "selfsigned",
			template: x509.Certificate{
				KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
				ExtKeyUsage:           []x509.ExtKeyUsage{},
				BasicConstraintsValid: true,
				IsCA:                  true,
			},
			expectedExtKeyUsage: nil,
		},
		{
			name: "client",
			template: x509.Certificate{
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
				NotAfter:    time.Now().Add(time.Hour * 24 * 365).UTC(),
			},
			signerKey:           signerKey,
			signer:              signer.Certificate,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
		{
			name: "server",
			template: x509.Certificate{
				DNSNames: []string{"a.invalid", "b.invalid"},
				NotAfter: time.Now().Add(time.Hour * 24 * 365).UTC(),
			},
			signerKey:           signerKey,
			signer:              signer.Certificate,
			expectedExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.template.Subject = pkix.Name{
				CommonName: tc.name,
			}

			cert, err := SignNewCertificate(key, &tc.template, tc.signer, tc.signerKey)
			require.NoError(t, err, "SignNewCertificate")

			{
				subject := cert.Certificate.Subject
				subject.Names = nil
				assert.Equal(t, subject, tc.template.Subject)
			}
			assert.Equal(t, cert.Subject, cert.Certificate.Subject)

			assert.Equal(t, cert.Certificate.DNSNames, tc.template.DNSNames)

			assert.Equal(t, cert.IsCA, tc.template.IsCA)
			assert.Equal(t, cert.Certificate.IsCA, tc.template.IsCA)

			{
				var b bytes.Buffer
				pkData, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
				require.NoError(t, err, "MarshalPKIXPublicKey")

				err = pem.Encode(&b, &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkData})
				require.NoError(t, err, "serializing public key")

				assert.Equal(t, b.String(), publicKeyData)
			}
			assert.Equal(t, cert.PublicKey, cert.Certificate.PublicKey)

			if tc.template.KeyUsage == 0 {
				tc.template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
			}
			assert.Equal(t, cert.Certificate.KeyUsage, tc.template.KeyUsage, "KeyUsage")

			assert.Equal(t, cert.Certificate.ExtKeyUsage, tc.expectedExtKeyUsage, "ExtKeyUsage")

			if tc.signer == nil {
				tc.signer = cert.Certificate
			}
			assert.Equal(t, cert.Certificate.Issuer, signer.Certificate.Subject, "Issuer")
			pool := x509.NewCertPool()
			pool.AddCert(tc.signer)
			_, err = cert.Certificate.Verify(x509.VerifyOptions{
				Roots:     pool,
				KeyUsages: tc.expectedExtKeyUsage,
			})
			assert.NoError(t, err, "verify certificate")

			// notbefore, notafter, serialnumber, basiccvalid
		})
	}
}

func TestCertificateRoundTrip(t *testing.T) {
	data := "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"

	cert, err := ParsePEMCertificate([]byte(data))
	if err != nil {
		t.Fatalf("error from ParsePEMCertificate: %v", err)
	}

	var b bytes.Buffer
	if _, err := cert.WriteTo(&b); err != nil {
		t.Fatalf("error from Certificate WriteTo: %v", err)
	}

	if b.String() != data {
		t.Fatalf("unexpected output from Certificate WriteTo: %q", b.String())
	}
}
