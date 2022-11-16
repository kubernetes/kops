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
	"k8s.io/kops/util/pkg/vfs"
)

// configserverSecretStore is a SecretStore backed by the config server.
type configserverSecretStore struct {
	nodeSecrets map[string][]byte
}

func NewSecretStore(nodeSecrets map[string][]byte) fi.SecretStore {
	return &configserverSecretStore{
		nodeSecrets: nodeSecrets,
	}
}

// Secret implements fi.SecretStore
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

// DeleteSecret implements fi.SecretStore
func (s *configserverSecretStore) DeleteSecret(id string) error {
	return fmt.Errorf("DeleteSecret not supported by configserverSecretStore")
}

// FindSecret implements fi.SecretStore
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

// GetOrCreateSecret implements fi.SecretStore
func (s *configserverSecretStore) GetOrCreateSecret(id string, secret *fi.Secret) (current *fi.Secret, created bool, err error) {
	return nil, false, fmt.Errorf("GetOrCreateSecret not supported by configserverSecretStore")
}

// ReplaceSecret implements fi.SecretStore
func (s *configserverSecretStore) ReplaceSecret(id string, secret *fi.Secret) (current *fi.Secret, err error) {
	return nil, fmt.Errorf("ReplaceSecret not supported by configserverSecretStore")
}

// ListSecrets implements fi.SecretStore
func (s *configserverSecretStore) ListSecrets() ([]string, error) {
	return nil, fmt.Errorf("ListSecrets not supported by configserverSecretStore")
}

// MirrorTo implements fi.SecretStore
func (s *configserverSecretStore) MirrorTo(basedir vfs.Path) error {
	return fmt.Errorf("MirrorTo not supported by configserverSecretStore")
}
