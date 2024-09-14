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

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
	"sigs.k8s.io/yaml"
)

type keystore struct {
	keys    map[string]keystoreEntry
	keySets map[string]*fi.Keyset
}

type keystoreEntry struct {
	certificate *pki.Certificate
	key         *pki.PrivateKey
}

var _ pki.Keystore = &keystore{}
var _ fi.CAStore = &keystore{}

// FindPrimaryKeypair implements pki.Keystore
func (k *keystore) FindPrimaryKeypair(ctx context.Context, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	entry, ok := k.keys[name]
	if !ok {
		return nil, nil, fmt.Errorf("unknown CA %q", name)
	}
	return entry.certificate, entry.key, nil
}

// FindKeyset finds a Keyset.  If the keyset is not found, it returns (nil, nil).
func (k *keystore) FindKeyset(ctx context.Context, name string) (*fi.Keyset, error) {
	keySet, ok := k.keySets[name]
	if !ok {
		return nil, nil
	}
	return keySet, nil
}

// StoreKeyset writes a Keyset to the store.
func (k *keystore) StoreKeyset(ctx context.Context, name string, keyset *fi.Keyset) error {
	return fmt.Errorf("server-side client does not support StoreKeyset")
}

// MirrorTo will copy secrets to a vfs.Path, which is often easier for a machine to read
func (k *keystore) MirrorTo(ctx context.Context, basedir vfs.Path) error {
	return fmt.Errorf("server-side client does not support MirrorTo")
}

// ListKeysets will return all the KeySets.
func (k *keystore) ListKeysets() (map[string]*fi.Keyset, error) {
	return nil, fmt.Errorf("server-side client does not support ListKeysets")
}

func newKeystore(basePath string, cas []string) (*keystore, map[string]string, error) {
	keystore := &keystore{
		keys:    map[string]keystoreEntry{},
		keySets: map[string]*fi.Keyset{},
	}
	for _, name := range cas {
		certBytes, err := os.ReadFile(path.Join(basePath, name+".crt"))
		if err != nil {
			return nil, nil, fmt.Errorf("reading %q certificate: %v", name, err)
		}
		// TODO: Support multiple certificates?
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
	if err := yaml.Unmarshal(keypairIDsBytes, &keypairIDs); err != nil {
		return nil, nil, fmt.Errorf("parsing keypair-ids.yaml")
	}

	// Build keysets
	for name, keypairID := range keypairIDs {
		entry, found := keystore.keys[name]
		if !found {
			klog.Warningf("keypair %q found in keypair IDs, not found as keypair", name)
			continue
		}
		primary := &fi.KeysetItem{}
		primary.Id = keypairID
		primary.Certificate = entry.certificate
		primary.PrivateKey = entry.key

		keyset := &fi.Keyset{}
		keyset.Primary = primary
		keyset.Items = make(map[string]*fi.KeysetItem)
		keyset.Items[primary.Id] = primary

		keystore.keySets[name] = keyset
	}

	return keystore, keypairIDs, nil
}
