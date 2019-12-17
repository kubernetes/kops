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

package fi

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"k8s.io/kops/util/pkg/vfs"
)

type SecretStore interface {
	// Secret returns a secret.  Returns an error if not found
	Secret(id string) (*Secret, error)
	// DeleteSecret deletes the specified secret
	DeleteSecret(id string) error
	// FindSecret finds a secret, if exists.  Returns nil,nil if not found
	FindSecret(id string) (*Secret, error)
	// GetOrCreateSecret creates a secret
	GetOrCreateSecret(id string, secret *Secret) (current *Secret, created bool, err error)
	// ReplaceSecret will forcefully update an existing secret if it exists
	ReplaceSecret(id string, secret *Secret) (current *Secret, err error)
	// ListSecrets lists the ids of all known secrets
	ListSecrets() ([]string, error)

	// MirrorTo will copy secrets to a vfs.Path, which is often easier for a machine to read
	MirrorTo(basedir vfs.Path) error
}

type Secret struct {
	Data []byte
}

func (s *Secret) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if s == nil {
		return "", fmt.Errorf("AsString called on nil Secret")
	}

	return string(s.Data), nil
}

func CreateSecret() (*Secret, error) {
	data := make([]byte, 128)
	_, err := crypto_rand.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading crypto_rand: %v", err)
	}

	s := base64.StdEncoding.EncodeToString(data)
	r := strings.NewReplacer("+", "", "=", "", "/", "")
	s = r.Replace(s)
	s = s[:32]

	return &Secret{
		Data: []byte(s),
	}, nil
}
