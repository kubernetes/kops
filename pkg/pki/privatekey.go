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
	"bytes"
	"crypto"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strconv"

	"k8s.io/klog"
)

// DefaultPrivateKeySize is the key size to use when generating private keys
// It can be overridden by the KOPS_RSA_PRIVATE_KEY_SIZE env var, or by tests
// (as generating RSA keys can be a bottleneck for testing)
var DefaultPrivateKeySize = 2048

func ParsePEMPrivateKey(data []byte) (*PrivateKey, error) {
	k, err := parsePEMPrivateKey(data)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, nil
	}
	return &PrivateKey{Key: k}, nil
}

func GeneratePrivateKey() (*PrivateKey, error) {
	var rsaKeySize = DefaultPrivateKeySize

	if os.Getenv("KOPS_RSA_PRIVATE_KEY_SIZE") != "" {
		s := os.Getenv("KOPS_RSA_PRIVATE_KEY_SIZE")
		if v, err := strconv.Atoi(s); err != nil {
			return nil, fmt.Errorf("error parsing KOPS_RSA_PRIVATE_KEY_SIZE=%s as integer", s)
		} else {
			rsaKeySize = int(v)
			klog.V(4).Infof("Generating key of size %d, set by KOPS_RSA_PRIVATE_KEY_SIZE env var", rsaKeySize)
		}
	}

	rsaKey, err := rsa.GenerateKey(crypto_rand.Reader, rsaKeySize)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	privateKey := &PrivateKey{Key: rsaKey}
	return privateKey, nil
}

type PrivateKey struct {
	Key crypto.PrivateKey
}

func (k *PrivateKey) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if k == nil {
		return "", fmt.Errorf("AsString called on nil private key")
	}

	var data bytes.Buffer
	_, err := k.WriteTo(&data)
	if err != nil {
		return "", fmt.Errorf("error writing SSL private key: %v", err)
	}
	return data.String(), nil
}

func (k *PrivateKey) AsBytes() ([]byte, error) {
	// Nicer behaviour because this is called from templates
	if k == nil {
		return nil, fmt.Errorf("AsBytes called on nil private key")
	}

	var data bytes.Buffer
	_, err := k.WriteTo(&data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL PrivateKey: %v", err)
	}
	return data.Bytes(), nil
}

func (k *PrivateKey) UnmarshalJSON(b []byte) (err error) {
	s := ""
	if err := json.Unmarshal(b, &s); err == nil {
		r, err := parsePEMPrivateKey([]byte(s))
		if err != nil {
			// Alternative form: Check if base64 encoded
			// TODO: Do we need this?  I think we need this only on nodeup, but maybe we could just not base64-it?
			d, err2 := base64.StdEncoding.DecodeString(s)
			if err2 == nil {
				r2, err2 := parsePEMPrivateKey(d)
				if err2 == nil {
					klog.Warningf("used base64 decode of PrivateKey")
					r = r2
					err = nil
				}
			}

			if err != nil {
				return fmt.Errorf("error parsing private key: %v", err)
			}
		}
		k.Key = r
		return nil
	}

	return fmt.Errorf("unknown format for private key: %q", string(b))
}

func (k *PrivateKey) MarshalJSON() ([]byte, error) {
	var data bytes.Buffer
	_, err := k.WriteTo(&data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL private key: %v", err)
	}
	return json.Marshal(data.String())
}

var _ io.WriterTo = &PrivateKey{}

func (k *PrivateKey) WriteTo(w io.Writer) (int64, error) {
	if k.Key == nil {
		// For the dry-run case
		return 0, nil
	}

	var data bytes.Buffer
	var err error

	switch pk := k.Key.(type) {
	case *rsa.PrivateKey:
		err = pem.Encode(w, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	default:
		return 0, fmt.Errorf("unknown private key type: %T", k.Key)
	}

	if err != nil {
		return 0, fmt.Errorf("error writing SSL private key: %v", err)
	}

	return data.WriteTo(w)
}

func (k *PrivateKey) WriteToFile(filename string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = k.WriteTo(f)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func parsePEMPrivateKey(pemData []byte) (crypto.PrivateKey, error) {
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			return nil, fmt.Errorf("could not parse private key")
		}

		if block.Type == "RSA PRIVATE KEY" {
			klog.V(10).Infof("Parsing pem block: %q", block.Type)
			return x509.ParsePKCS1PrivateKey(block.Bytes)
		} else if block.Type == "PRIVATE KEY" {
			klog.V(10).Infof("Parsing pem block: %q", block.Type)
			k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			return k.(crypto.PrivateKey), nil
		} else {
			klog.Infof("Ignoring unexpected PEM block: %q", block.Type)
		}

		pemData = rest
	}
}
