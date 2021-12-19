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

package pki

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"k8s.io/klog/v2"
)

func ParsePEMPublicKey(data []byte) (*PublicKey, error) {
	k, err := parsePEMPublicKey(data)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, nil
	}
	return &PublicKey{Key: k}, nil
}

type PublicKey struct {
	Key crypto.PublicKey
}

func parsePEMPublicKey(pemData []byte) (crypto.PublicKey, error) {
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			return nil, fmt.Errorf("could not parse private key")
		}

		if block.Type == "RSA PUBLIC KEY" {
			k, err := x509.ParsePKCS1PublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			return k, nil
		} else if block.Type == "PUBLIC KEY" {
			k, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			return k.(crypto.PublicKey), nil
		} else {
			klog.Infof("Ignoring unexpected PEM block: %q", block.Type)
		}

		pemData = rest
	}
}
