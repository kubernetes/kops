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

package configserver

import (
	"context"
	"crypto/x509"
	"fmt"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// configserverKeyStore is a KeyStore backed by the config server.
type configserverKeyStore struct{}

func NewKeyStore() fi.CAStore {
	return &configserverKeyStore{}
}

// FindPrimaryKeypair implements pki.Keystore
func (s *configserverKeyStore) FindPrimaryKeypair(name string) (*pki.Certificate, *pki.PrivateKey, error) {
	return nil, nil, fmt.Errorf("FindPrimaryKeypair %q not supported by configserverKeyStore", name)
}

// FindKeyset implements fi.Keystore
func (s *configserverKeyStore) FindKeyset(name string) (*fi.Keyset, error) {
	return nil, fmt.Errorf("FindKeyset %q not supported by configserverKeyStore", name)
}

// CreateKeypair implements fi.Keystore
func (s *configserverKeyStore) CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	return nil, fmt.Errorf("CreateKeypair not supported by configserverKeyStore")
}

// StoreKeyset implements fi.Keystore
func (s *configserverKeyStore) StoreKeyset(ctx context.Context, name string, keyset *fi.Keyset) error {
	return fmt.Errorf("StoreKeyset not supported by configserverKeyStore")
}

// MirrorTo implements fi.Keystore
func (s *configserverKeyStore) MirrorTo(ctx context.Context, basedir vfs.Path) error {
	return fmt.Errorf("MirrorTo not supported by configserverKeyStore")
}

// ListKeysets implements fi.CAStore
func (s *configserverKeyStore) ListKeysets() (map[string]*fi.Keyset, error) {
	return nil, fmt.Errorf("ListKeysets not supported by configserverKeyStore")
}
