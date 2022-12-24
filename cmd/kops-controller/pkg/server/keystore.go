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

package server

import (
	"context"
	"fmt"
	"os"
	"path"

	"k8s.io/kops/pkg/pki"
	"sigs.k8s.io/yaml"
)

type keystore struct {
	keys map[string]keystoreEntry
}

type keystoreEntry struct {
	certificate *pki.Certificate
	key         *pki.PrivateKey
}

var _ pki.Keystore = keystore{}

// FindPrimaryKeypair implements pki.Keystore
func (k keystore) FindPrimaryKeypair(ctx context.Context, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	entry, ok := k.keys[name]
	if !ok {
		return nil, nil, fmt.Errorf("unknown CA %q", name)
	}
	return entry.certificate, entry.key, nil
}

func newKeystore(basePath string, cas []string) (pki.Keystore, map[string]string, error) {
	keystore := &keystore{
		keys: map[string]keystoreEntry{},
	}
	for _, name := range cas {
		certBytes, err := os.ReadFile(path.Join(basePath, name+".crt"))
		if err != nil {
			return nil, nil, fmt.Errorf("reading %q certificate: %v", name, err)
		}
		certificate, err := pki.ParsePEMCertificate(certBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %q certificate: %v", name, err)
		}

		keyBytes, err := os.ReadFile(path.Join(basePath, name+".key"))
		if err != nil {
			return nil, nil, fmt.Errorf("reading %q key: %v", name, err)
		}
		key, err := pki.ParsePEMPrivateKey(keyBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %q key: %v", name, err)
		}

		keystore.keys[name] = keystoreEntry{
			certificate: certificate,
			key:         key,
		}
	}

	var keypairIDs map[string]string
	keypairIDsBytes, err := os.ReadFile(path.Join(basePath, "keypair-ids.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("reading keypair-ids.yaml")
	}
	err = yaml.Unmarshal(keypairIDsBytes, &keypairIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing keypair-ids.yaml")
	}

	return keystore, keypairIDs, nil
}
