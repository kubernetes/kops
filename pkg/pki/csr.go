/*
Copyright 2019 The Kubernetes Authors.

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
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"math/big"
	"time"

	"k8s.io/klog"
)

// BuildPKISerial produces a serial number for certs that is vanishingly unlikely to collide
// The timestamp should be provided as an input (time.Now().UnixNano()), and then we combine
// that with a 32 bit random crypto-rand integer.
// We also know that a bigger value was created later (modulo clock skew)
func BuildPKISerial(timestamp int64) *big.Int {
	randomLimit := new(big.Int).Lsh(big.NewInt(1), 32)
	randomComponent, err := crypto_rand.Int(crypto_rand.Reader, randomLimit)
	if err != nil {
		klog.Fatalf("error generating random number: %v", err)
	}

	serial := big.NewInt(timestamp)
	serial.Lsh(serial, 32)
	serial.Or(serial, randomComponent)

	return serial
}

func SignNewCertificate(privateKey *PrivateKey, template *x509.Certificate, signer *x509.Certificate, signerPrivateKey *PrivateKey) (*Certificate, error) {
	if template.PublicKey == nil {
		rsaPrivateKey, ok := privateKey.Key.(*rsa.PrivateKey)
		if ok {
			template.PublicKey = rsaPrivateKey.Public()
		}
	}

	if template.PublicKey == nil {
		return nil, fmt.Errorf("PublicKey not set, and cannot be determined from %T", privateKey)
	}

	now := time.Now()
	if template.NotBefore.IsZero() {
		template.NotBefore = now.Add(time.Hour * -48)
	}

	if template.NotAfter.IsZero() {
		template.NotAfter = now.Add(time.Hour * 10 * 365 * 24)
	}

	if template.SerialNumber == nil {
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := crypto_rand.Int(crypto_rand.Reader, serialNumberLimit)
		if err != nil {
			return nil, fmt.Errorf("error generating certificate serial number: %s", err)
		}
		template.SerialNumber = serialNumber
	}
	var parent *x509.Certificate
	if signer != nil {
		parent = signer
	} else {
		parent = template
		signerPrivateKey = privateKey
	}

	if template.KeyUsage == 0 {
		template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	}

	if template.ExtKeyUsage == nil && !template.IsCA {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	//c.SignatureAlgorithm  = do we want to override?

	certificateData, err := x509.CreateCertificate(crypto_rand.Reader, template, parent, template.PublicKey, signerPrivateKey.Key)
	if err != nil {
		return nil, fmt.Errorf("error creating certificate: %v", err)
	}

	c := &Certificate{}
	c.PublicKey = template.PublicKey

	cert, err := x509.ParseCertificate(certificateData)
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %v", err)
	}
	c.Certificate = cert

	return c, nil
}
