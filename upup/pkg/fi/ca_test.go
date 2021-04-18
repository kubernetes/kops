/*
Copyright 2021 The Kubernetes Authors.

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

package fi_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

const certData = "-----BEGIN CERTIFICATE-----\nMIIC+DCCAeCgAwIBAgIMFnbTyckBfEgW0qf0MA0GCSqGSIb3DQEBCwUAMBgxFjAU\nBgNVBAMTDWNuPWt1YmVybmV0ZXMwHhcNMjEwNDE2MDI0NjE5WhcNMzEwNDE2MDI0\nNjE5WjAYMRYwFAYDVQQDEw1jbj1rdWJlcm5ldGVzMIIBIjANBgkqhkiG9w0BAQEF\nAAOCAQ8AMIIBCgKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivY\nHVxHAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkE\nmDgAgvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj\n9WGFGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2\nQgJ/P2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgM\nwDTx9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABo0IwQDAOBgNVHQ8B\nAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUVOJ/Z21wnbF5Kjsi\nJFdZL/rKHJ4wDQYJKoZIhvcNAQELBQADggEBABWm3zNEiAO+HxfT+WYUhS/ot1Rf\nwzQUPmFR+1GGdXxOam4dMQ0X9uJsE5PtNw7kg9ja+jkUuharSXraueu6IbcRCXQ8\nycnjmbhcPdvH8dhQAB+2Y1xyf/olWan7NU98fwwWk3ubX1aUaoBbXJW9IqkmRG6K\nZVTnz0yGdQfqZyZbl48WfIQ3W7qsAnsSYSrhUWNNOHgy6NQRxpgzGohYFAUprFqc\n8Dm02rzpNwe+ZDOInm05UUOblsdeHYrRetfvhnYJl/CEPGAdJGnOjLMjr7V85y4a\n41XKFsjMo2ztvNdLmYiw0dfar4WSdK/AKFzUXEPRwjCe5xMtsMOIkyJtvxw=\n-----END CERTIFICATE-----\n"
const tooBigSerialCertData = "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"
const privatekeyData = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivYHVxH\nAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkEmDgA\ngvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj9WGF\nGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2QgJ/\nP2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgMwDTx\n9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABAoIBAA0ktjaTfyrAxsTI\nBezb7Zr5NBW55dvuII299cd6MJo+rI/TRYhvUv48kY8IFXp/hyUjzgeDLunxmIf9\n/Zgsoic9Ol44/g45mMduhcGYPzAAeCdcJ5OB9rR9VfDCXyjYLlN8H8iU0734tTqM\n0V13tQ9zdSqkGPZOIcq/kR/pylbOZaQMe97BTlsAnOMSMKDgnftY4122Lq3GYy+t\nvpr+bKVaQZwvkLoSU3rECCaKaghgwCyX7jft9aEkhdJv+KlwbsGY6WErvxOaLWHd\ncuMQjGapY1Fa/4UD00mvrA260NyKfzrp6+P46RrVMwEYRJMIQ8YBAk6N6Hh7dc0G\n8Z6i1m0CgYEA9HeCJR0TSwbIQ1bDXUrzpftHuidG5BnSBtax/ND9qIPhR/FBW5nj\n22nwLc48KkyirlfIULd0ae4qVXJn7wfYcuX/cJMLDmSVtlM5Dzmi/91xRiFgIzx1\nAsbBzaFjISP2HpSgL+e9FtSXaaqeZVrflitVhYKUpI/AKV31qGHf04sCgYEA6zTV\n99Sb49Wdlns5IgsfnXl6ToRttB18lfEKcVfjAM4frnkk06JpFAZeR+9GGKUXZHqs\nz2qcplw4d/moCC6p3rYPBMLXsrGNEUFZqBlgz72QA6BBq3X0Cg1Bc2ZbK5VIzwkg\nST2SSux6ccROfgULmN5ZiLOtdUKNEZpFF3i3qtsCgYADT/s7dYFlatobz3kmMnXK\nsfTu2MllHdRys0YGHu7Q8biDuQkhrJwhxPW0KS83g4JQym+0aEfzh36bWcl+u6R7\nKhKj+9oSf9pndgk345gJz35RbPJYh+EuAHNvzdgCAvK6x1jETWeKf6btj5pF1U1i\nQ4QNIw/QiwIXjWZeubTGsQKBgQCbduLu2rLnlyyAaJZM8DlHZyH2gAXbBZpxqU8T\nt9mtkJDUS/KRiEoYGFV9CqS0aXrayVMsDfXY6B/S/UuZjO5u7LtklDzqOf1aKG3Q\ndGXPKibknqqJYH+bnUNjuYYNerETV57lijMGHuSYCf8vwLn3oxBfERRX61M/DU8Z\nworz/QKBgQDCTJI2+jdXg26XuYUmM4XXfnocfzAXhXBULt1nENcogNf1fcptAVtu\nBAiz4/HipQKqoWVUYmxfgbbLRKKLK0s0lOWKbYdVjhEm/m2ZU8wtXTagNwkIGoyq\nY/C1Lox4f1ROJnCjc/hfcOjcxX5M8A8peecHWlVtUPKTJgxQ7oMKcw==\n-----END RSA PRIVATE KEY-----\n"

func TestAddItem(t *testing.T) {
	cert, _ := pki.ParsePEMCertificate([]byte(certData))
	tooBigSerialCert, _ := pki.ParsePEMCertificate([]byte(tooBigSerialCertData))
	privateKey, _ := pki.ParsePEMPrivateKey([]byte(privatekeyData))

	type expectedItems []struct {
		validateId func(t *testing.T, id string)
		cert       *pki.Certificate
		privateKey *pki.PrivateKey
	}
	tests := []struct {
		name            string
		keyset          fi.Keyset
		cert            *pki.Certificate
		privateKey      *pki.PrivateKey
		expectedItems   expectedItems
		expectedPrimary *pki.PrivateKey
		expectedError   string
	}{
		{
			name: "first",
			keyset: fi.Keyset{
				Items: map[string]*fi.KeysetItem{},
			},
			cert:       cert,
			privateKey: privateKey,
			expectedItems: expectedItems{
				{
					cert:       cert,
					privateKey: privateKey,
				},
			},
			expectedPrimary: privateKey,
		},
		{
			name: "BigSerial",
			keyset: fi.Keyset{
				Items: map[string]*fi.KeysetItem{},
			},
			cert:       tooBigSerialCert,
			privateKey: privateKey,
			expectedItems: expectedItems{
				{
					validateId: func(t *testing.T, id string) {
						version, ok := big.NewInt(0).SetString(id, 10)
						require.True(t, ok, "parses as integer")
						if version.Cmp(pki.BuildPKISerial(time.Now().UnixNano())) > 0 {
							t.Errorf("id %q larger than serial for current time", id)
						}
					},
					cert:       tooBigSerialCert,
					privateKey: privateKey,
				},
			},
			expectedPrimary: privateKey,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.keyset.AddItem(tc.cert, tc.privateKey)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedPrimary, tc.keyset.Primary.PrivateKey, "primary key")
		expected:
			for _, expected := range tc.expectedItems {
				for id, item := range tc.keyset.Items {
					if (expected.cert != nil && expected.cert == item.Certificate) ||
						expected.privateKey != nil && expected.privateKey == item.PrivateKey {
						assert.Same(t, expected.cert, item.Certificate, "item %s", id)
						assert.Same(t, expected.privateKey, item.PrivateKey, "item %s", id)
						if expected.validateId == nil {
							assert.Equal(t, expected.cert.Certificate.SerialNumber.String(), id, "id")
						} else {
							expected.validateId(t, id)
						}
						delete(tc.keyset.Items, id)
						continue expected
					}
				}
				if expected.cert != nil {
					t.Errorf("did not find expected item %q", expected.cert.Certificate.SerialNumber)
				} else {
					t.Errorf("did not find expected key %q", expected.privateKey)
				}
			}
			for id := range tc.keyset.Items {
				t.Errorf("unexpected item %q", id)
			}
		})
	}

}
