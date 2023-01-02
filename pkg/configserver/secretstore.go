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

	"k8s.io/kops/upup/pkg/fi"
)

// configserverSecretStore is a SecretStore backed by the config server.
type configserverSecretStore struct {
	nodeSecrets map[string][]byte
}

func NewSecretStore(nodeSecrets map[string][]byte) fi.SecretStoreReader {
	return &configserverSecretStore{
		nodeSecrets: nodeSecrets,
	}
}

// Secret implements fi.SecretStoreReader
func (s *configserverSecretStore) Secret(id string) (*fi.Secret, error) {
	secret, err := s.FindSecret(id)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, fmt.Errorf("secret %q not found", id)
	}
	return secret, nil
}

// FindSecret implements fi.SecretStoreReader
func (s *configserverSecretStore) FindSecret(id string) (*fi.Secret, error) {
	secretBytes, ok := s.nodeSecrets[id]
	if !ok {
		return nil, nil
	}
	secret := &fi.Secret{
		Data: secretBytes,
	}
	return secret, nil
}
