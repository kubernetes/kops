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
	"fmt"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

// configserverKeyStore is a KeyStore backed by the config server.
type configserverKeyStore struct{}

func NewKeyStore() fi.KeystoreReader {
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
