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
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCertificate(t *testing.T) {
	data := "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCthOYz2rjTI2x7hyWDiqxUS4LPBXctThTRCm8mGANvCaHESVhMMAVkeIINn90zhVHvr48CnMR2qSWwxR8AZJpFvR7hKk8gHh494pLW/eSUwP3pJbSOxt0OsxXEflaiHR5x1VJ9DsTl9Vy3qSkhsYBFPMMoLk7kpb3Y0F+t6ewIFrFw0IHXvbSYm07dN6amJXU/L78X/kAcsD9IDfE4EqOf5P4WBpXZ4YnCt+kpQR0sFmCst/2ktlgU+TCsP+PNUuzHsAovmLNdOdpplC1VmIUaT5/CuAITz0YASs8L2H9MlgxtwJ3CkNr/H4Y1UhTXxgKPMskdvrMGLLkh+Y7z4k/vAgMBAAECggEABLEsI759h/a/F2SzVCga+oPji3iQ3UXtadetOMD70B6uY/yTFCr/l6oV44ttSQLeZoD/TilHZhQIg7auMhhbqZjFwyid5MxGXL20jcwQyHWi3Waacs/tfgPEChZb07ITjC13YmVUrV6MqB9oEClYwsaJWnFMyyxbgrI5HB6lYMnnjzU+otkqwVtMbe1mO1nX7paBxJ+pNbtasihKCWRJZAMZ9bcRw01+QFRca2CVJjPWlkD9N2S/0Q7+4YqUTNh/dx3t2uBuxg1ji7TQzBKT5ftF9lLlDwlEyXYqqOYUFgU9541sGVUVisXp/mykJVN7z3osoz+oU62kXqQm0YLlQQKBgQDdZqDKwDVmL23Tguwhrs6quj8VWCpD78YC8zzICFBDDWWY5Jdl+07Bw9gwgHyb8h3JfRMXXFg62tFald10hqi5raLva6QKYwkLSe372syrJXaEej/x8NE1bQf8wzl6yS0fv/UwZ1aZTiKUSj7LKNS720IDAGdEn/1NdL6zUBNkQQKBgQDIorZGO/zkLyWRU1Zzc2pxwtL/iYwFmsndKEEojvs0NIe6ySIG0HOw+wFqczjoiyIORten8Vypz4CEUe4fVJwkeJa6GrVFNWTlNi4mX2p33L9OJEFa82Uisf0amTedC2RQEfnMu5PKDedVl5WZ3HOUbQPNJ+qYy+9SyhHi73joLwKBgBpbb2TzwOerWc3GVkokP2I/zebCmjWAQ/hx8Jh3tOZmn+O1wvhXFKcoo4ISqcL+7eDgzPcI/U/0YNwB311R8qA4NZ9/FwZNh/QaFwTWpWryiMt4qkgpPR65HixPKXaeoIqZFZ1vj/WsQZ2ZwSP6dmjuz0sAL0sSKNuhvFoofEaBAoGAGzppvjJZ6aW0VXqX2uco5PNpqyBBjmkpSAg0f4qX8MfIO8McCQy1Bqmp0YZ9jKGFJ6bZkYMh7jGo4Uw1Iq9a2WA8JFmHjDLo1Gp77N06F7YviC1HaU5qxUCedsOgVoG7RVqLKguyzNMCOA1wUgcm8FezEl5+aeoTOosNzlxtbiUCgYB0OWavac9ZP/BDU2gfTkeks/Z+HtIssToBQ1AiByQxdTPNHZ5GCDvvy/9g8CKETkHU1DoG78lAsMCMDUc1mPFVpTJllxaO9SOIgfYxgkRt9fyenQmdhiXbvJ4vv503sAQdU1knw2UgIcwPjXAaBR3Rf2gyMBkdZ2icQvILKz2OOQ==\n-----END PRIVATE KEY-----"
	publicKeyData := "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArYTmM9q40yNse4clg4qs\nVEuCzwV3LU4U0QpvJhgDbwmhxElYTDAFZHiCDZ/dM4VR76+PApzEdqklsMUfAGSa\nRb0e4SpPIB4ePeKS1v3klMD96SW0jsbdDrMVxH5Woh0ecdVSfQ7E5fVct6kpIbGA\nRTzDKC5O5KW92NBfrensCBaxcNCB1720mJtO3TempiV1Py+/F/5AHLA/SA3xOBKj\nn+T+FgaV2eGJwrfpKUEdLBZgrLf9pLZYFPkwrD/jzVLsx7AKL5izXTnaaZQtVZiF\nGk+fwrgCE89GAErPC9h/TJYMbcCdwpDa/x+GNVIU18YCjzLJHb6zBiy5IfmO8+JP\n7wIDAQAB\n-----END RSA PUBLIC KEY-----\n"
	signerCertData := "-----BEGIN CERTIFICATE-----\nMIIDAzCCAeugAwIBAgIUX5OneoJSyzLbD5/cUsacl0+kbLcwDQYJKoZIhvcNAQELBQAwETEPMA0GA1UEAwwGc2lnbmVyMB4XDTI1MDMxMDEzMzkwNFoXDTM1MDMwODEzMzkwNFowETEPMA0GA1UEAwwGc2lnbmVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2Ym6Hcv4f7OU8rkrZr6ur9VC9HnJe/Gi/lmHvwLe+RqnpqDWxbQDR5Q3dwVnBgbMS3CVVTwcERqK5AkHSehG68OtUu9ygMFZDoqMU+7UJglfsY550SBk5YZuvhTRFgtlGqiKjRoD+p8/thrx0l4BV54FJvE9g7k31X+ynHP75hAsPTIobsAk3DYAtV651NLwpmyEQqCiyImJKi7jcvR/YpZobzFSQ1fPp9yBxxKTJZrBf54ZATPOLc0Pxv0rA5MscT0ujFYe7lC1EKUAlZ6mSIWFXsuJaXQmPmGctX5m50DxhKXfVEfr77xIa1+Xqh9bq4/sFxryuPyYiNJNnRMY4QIDAQABo1MwUTAdBgNVHQ4EFgQU0pax2FviEHNrcKbXtjmQsB8LBe8wHwYDVR0jBBgwFoAU0pax2FviEHNrcKbXtjmQsB8LBe8wDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAxtC4lB/HqvfFh/7qcgUInfZ7f0p4EzthTSDOf83zuOV4YFzD1K003YloTC720PNjURc3JMCSl2gT4vzM3zxwfp0FXuzTkW9mU8IPoWX7iArXp80wl1zLBc9HJ2eVN4zK8abTgnq2j7U3Fd/pCQjNeToyHfYk7VW3d4Xco38NIh9GWTCDwjfYAr0jNLBATuYZreCSatysIITQcE7LNUhD1QywNtF74SS4J3rHOgYHK+/jFW4KqCMA5VeXBenutpqYKIcP8lg4/qQiAWrqiY/dTyhKSDdb8vczaq31H6XMv5czcmUHRx0yI2XwoKuD4uTBIPGisJWqoMaW/GyyKdvugw==\n-----END CERTIFICATE-----"
	signerKeyData := "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDZibody/h/s5TyuStmvq6v1UL0ecl78aL+WYe/At75GqemoNbFtANHlDd3BWcGBsxLcJVVPBwRGorkCQdJ6Ebrw61S73KAwVkOioxT7tQmCV+xjnnRIGTlhm6+FNEWC2UaqIqNGgP6nz+2GvHSXgFXngUm8T2DuTfVf7Kcc/vmECw9MihuwCTcNgC1XrnU0vCmbIRCoKLIiYkqLuNy9H9ilmhvMVJDV8+n3IHHEpMlmsF/nhkBM84tzQ/G/SsDkyxxPS6MVh7uULUQpQCVnqZIhYVey4lpdCY+YZy1fmbnQPGEpd9UR+vvvEhrX5eqH1urj+wXGvK4/JiI0k2dExjhAgMBAAECggEAAXDKDLx3DtFvoRPc17dXjM6KvPe5f9qfy7NoFLm+JEQq7A2QnoqMowK2Q1GD1yRgYfeC5aeaP/q/BLeSlsi0/4ayNSRky7l8D36XY07nlMDnI1PgNqRSRrrXLOcSY2T77GtFT53mfNhlIZ2YEF6S/7OKMTHTyHWHiyBnXGXgOyvJHufKZ6+VECL4PlYLIAom1PlClNVt/TafwbEIhpbfpH/XYJCJ6GcpbWU2EpkZPYxxlCF+kd7PVI7kPzQQ9DTWDBmBI6mX9t0HPFmzlo4TfaUdgqIZfOwuBEVwi5+Pb4kppVK1QGKAl4DxqBgQnmqeWfEelwhsEEzIyUaocU0PYQKBgQDw1e+5tjL4bqNvuKGJwZcMplVMmEv6W+sxV4qMfFkV/dzEwsYQgqjDz2B9lipTadlXiuQ6B6iNVkAGwfZhTHjq0ruf5UB7yZ5CJSA8sXARR1Y4aEefes4np5Pw+N/jOY30PWrgtClT4PJxON3w9LnSWHDg62whvMUzqDZuz2JVdQKBgQDnPEB607FNoEgFAnfm4hy0JcmePvgndIPRJe+2vlWr0+xwtLoFCAQWIm4sagXWWzRco981eCPovElQ7T9SZwFVH1xBSEY54N4KSoX7L+HffeDB52i1a+UlQQEY0vi6ep48smBoJ+7UOFpDHp2LSn3bkyNy5XtxAG9hTXilBW5MPQKBgQCZXFJ0my5n/uQ6b4MGWu2aE4174ft36PKjEBDdFw4PsAHWlgVUXC+lyTezoV1AksXhNkPRJDFUF1lcNEV1fiH9vsXVs0HV0fTiQAwAOimYBypDbzw0tRn0LIVLzN+dLXhU0Itvnao3jKY2LTU/jEeMR99RivjnnvKgy3wmIg+HRQKBgFmzTtQW8M3LIoUG+xpOlpHvorHHfZ5YnZXxoHcEiNlaIXtrMEopXOR1QMXr7w3DXaGeVEU6sLtk5xAEqK6/lI2/15rffZaQO7JETIsvfPCktR6jNURDcaWs/M7zcFdun5muHKXq78PVhHZLFxRktkQKZRL6IJOqdoqJcgaZ/7qFAoGAWb2dubn2yi6EoRh24T9+qQ1FZLMOigHEnv0+X93A8KS5pqKJ7Fc9NBnXng1f3R5wBjinXjenVGWDJYLe3LQA2w85OmdEL/cqkLIuCARyc/r5MFB0jzGTvbXZUvg2TCE2MdwZMX91dN4RS5B1o06g4ISWO6cdLl9y0EUHLtUTiZc=\n-----END PRIVATE KEY-----"

	key, err := ParsePEMPrivateKey([]byte(data))
	require.NoError(t, err, "ParsePEMPrivateKey")
	signerKey, err := ParsePEMPrivateKey([]byte(signerKeyData))
	require.NoError(t, err, "ParsePEMPrivateKey")

	{
		var b bytes.Buffer
		pkData, err := x509.MarshalPKIXPublicKey(key.Key.Public())
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
				Issuer:                pkix.Name{CommonName: "selfsigned"},
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

			cert, err := signNewCertificate(key, &tc.template, tc.signer, tc.signerKey)
			require.NoError(t, err, "signNewCertificate")

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
			if tc.name == "selfsigned" {
				assert.Equal(t, cert.Certificate.Issuer, cert.Certificate.Subject, "Issuer")
			} else {
				assert.Equal(t, cert.Certificate.Issuer, signer.Certificate.Subject, "Issuer")
			}
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
