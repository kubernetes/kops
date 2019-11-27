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

package k8sapi

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/pki"
)

// KeypairSecret is a wrapper around a k8s Secret object that holds a TLS keypair
type KeypairSecret struct {
	Namespace string
	Name      string

	Certificate *pki.Certificate
	PrivateKey  *pki.PrivateKey
}

// ParseKeypairSecret parses the secret object, decoding the certificate & private-key, if present
func ParseKeypairSecret(secret *v1.Secret) (*KeypairSecret, error) {
	k := &KeypairSecret{}
	k.Namespace = secret.Namespace
	k.Name = secret.Name

	certData := secret.Data[v1.TLSCertKey]
	if certData != nil {
		cert, err := pki.ParsePEMCertificate(certData)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate in %s/%s: %q", k.Namespace, k.Name, err)
		}
		k.Certificate = cert
	}
	keyData := secret.Data[v1.TLSPrivateKeyKey]
	if keyData != nil {
		key, err := pki.ParsePEMPrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("error parsing key in %s/%s: %q", k.Namespace, k.Name, err)
		}
		k.PrivateKey = key
	}

	return k, nil
}

// Encode maps a KeypairSecret into a k8s Secret
func (k *KeypairSecret) Encode() (*v1.Secret, error) {
	secret := &v1.Secret{}
	secret.Namespace = k.Namespace
	secret.Name = k.Name
	secret.Type = v1.SecretTypeTLS

	secret.Data = make(map[string][]byte)

	if k.Certificate != nil {
		data, err := k.Certificate.AsBytes()
		if err != nil {
			return nil, err
		}
		secret.Data[v1.TLSCertKey] = data
	}

	if k.PrivateKey != nil {
		data, err := k.PrivateKey.AsBytes()
		if err != nil {
			return nil, err
		}
		secret.Data[v1.TLSPrivateKeyKey] = data
	}

	return secret, nil
}
