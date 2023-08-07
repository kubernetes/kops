/*
Copyright 2023 The Kubernetes Authors.

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

package bootstrap

import (
	cryptorand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/pki"
)

func randomBytes(length int) []byte {
	b := make([]byte, length)
	if _, err := cryptorand.Read(b); err != nil {
		klog.Fatalf("failed to read from crypto/rand: %v", err)
	}
	return b
}

func challengeKopsControllerSubject(clusterName string) pkix.Name {
	// Note: keep in sync with subjectsMatch if you add (additional) fields here
	return pkix.Name{
		CommonName: "kops-controller." + clusterName,
	}
}

func subjectsMatch(l, r pkix.Name) bool {
	// We need to check all the fields in challengeKopsControllerSubject
	return l.CommonName == r.CommonName
}

func challengeServerHostName(clusterName string) string {
	return "challenge-server." + clusterName
}

func BuildChallengeServerCertificate(clusterName string) (*tls.Certificate, error) {
	serverName := challengeServerHostName(clusterName)

	privateKey, err := pki.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("generating ecdsa key: %w", err)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment

	now := time.Now()
	notBefore := now.Add(-15 * time.Minute)
	notAfter := notBefore.Add(time.Hour)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: serverName,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.DNSNames = append(template.DNSNames, serverName)

	der, err := x509.CreateCertificate(cryptorand.Reader, &template, &template, privateKey.Key.Public(), privateKey.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	tlsCertificate := &tls.Certificate{
		PrivateKey:  privateKey.Key,
		Certificate: [][]byte{parsed.Raw},
		Leaf:        parsed,
	}

	return tlsCertificate, nil
}
