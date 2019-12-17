/*
Copyright 2017 The Kubernetes Authors.

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

package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// NamePrefix is a prefix we use to avoid collisions with other keysets
const NamePrefix = "token-"

// ClientsetSecretStore is a SecretStore backed by Keyset objects in an API server
type ClientsetSecretStore struct {
	cluster   *kops.Cluster
	namespace string
	clientset kopsinternalversion.KopsInterface
}

var _ fi.SecretStore = &ClientsetSecretStore{}

// NewClientsetSecretStore is the constructor for ClientsetSecretStore
func NewClientsetSecretStore(cluster *kops.Cluster, clientset kopsinternalversion.KopsInterface, namespace string) fi.SecretStore {
	c := &ClientsetSecretStore{
		cluster:   cluster,
		clientset: clientset,
		namespace: namespace,
	}
	return c
}

func (c *ClientsetSecretStore) MirrorTo(basedir vfs.Path) error {
	list, err := c.clientset.Keysets(c.namespace).List(v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing keysets: %v", err)
	}

	for i := range list.Items {
		keyset := &list.Items[i]

		if keyset.Spec.Type != kops.SecretTypeSecret {
			continue
		}

		primary := fi.FindPrimary(keyset)
		if primary == nil {
			return fmt.Errorf("found secret with no primary data: %s", keyset.Name)
		}

		name := strings.TrimPrefix(keyset.Name, NamePrefix)
		p := BuildVfsSecretPath(basedir, name)

		s := &fi.Secret{
			Data: primary.PrivateMaterial,
		}
		data, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("error serializing secret: %v", err)
		}

		acl, err := acls.GetACL(p, c.cluster)
		if err != nil {
			return err
		}

		if err := p.WriteFile(bytes.NewReader(data), acl); err != nil {
			return fmt.Errorf("error writing secret to %q: %v", p, err)
		}
	}

	return nil
}

// FindSecret implements fi.SecretStore::FindSecret
func (c *ClientsetSecretStore) FindSecret(name string) (*fi.Secret, error) {
	s, err := c.loadSecret(name)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListSecrets implements fi.SecretStore::ListSecrets
func (c *ClientsetSecretStore) ListSecrets() ([]string, error) {
	list, err := c.clientset.Keysets(c.namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing keysets: %v", err)
	}

	var names []string
	for i := range list.Items {
		keyset := &list.Items[i]

		switch keyset.Spec.Type {
		case kops.SecretTypeSecret:
			name := strings.TrimPrefix(keyset.Name, NamePrefix)
			names = append(names, name)
		}
	}

	return names, nil
}

// Secret implements fi.SecretStore::Secret
func (c *ClientsetSecretStore) Secret(name string) (*fi.Secret, error) {
	s, err := c.FindSecret(name)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, fmt.Errorf("Secret not found: %q", name)
	}
	return s, nil
}

// DeleteSecret implements fi.SecretStore::DeleteSecret
func (c *ClientsetSecretStore) DeleteSecret(name string) error {
	client := c.clientset.Keysets(c.namespace)

	keyset, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error reading Keyset %q: %v", name, err)
	}

	if keyset.Spec.Type != kops.SecretTypeSecret {
		return fmt.Errorf("mismatch on Keyset type on %q", name)
	}

	if err := client.Delete(name, &v1.DeleteOptions{}); err != nil {
		return fmt.Errorf("error deleting Keyset %q: %v", name, err)
	}

	return nil
}

// GetOrCreateSecret implements fi.SecretStore::GetOrCreateSecret
func (c *ClientsetSecretStore) GetOrCreateSecret(name string, secret *fi.Secret) (*fi.Secret, bool, error) {
	for i := 0; i < 2; i++ {
		s, err := c.FindSecret(name)
		if err != nil {
			return nil, false, err
		}

		if s != nil {
			return s, false, nil
		}

		_, err = c.createSecret(secret, name, false)
		if err != nil {
			if errors.IsAlreadyExists(err) && i == 0 {
				klog.Infof("Got already-exists error when writing secret; likely due to concurrent creation.  Will retry")
				continue
			} else {
				return nil, false, err
			}
		}

		if err == nil {
			break
		}
	}

	// Make double-sure it round-trips
	s, err := c.loadSecret(name)
	if err != nil {
		klog.Fatalf("unable to load secret immediately after creation %v: %v", name, err)
		return nil, false, err
	}
	return s, true, nil
}

// ReplaceSecret implements fi.SecretStore::ReplaceSecret
func (c *ClientsetSecretStore) ReplaceSecret(name string, secret *fi.Secret) (*fi.Secret, error) {
	_, err := c.createSecret(secret, name, true)
	if err != nil {
		return nil, fmt.Errorf("unable to write secret: %v", err)
	}

	// Confirm the secret exists
	s, err := c.loadSecret(name)
	if err != nil {
		return nil, fmt.Errorf("unable to load secret immediately after creation: %v", err)
	}
	return s, nil
}

// loadSecret returns the named secret, if it exists, otherwise returns nil
func (c *ClientsetSecretStore) loadSecret(name string) (*fi.Secret, error) {
	name = NamePrefix + name
	keyset, err := c.clientset.Keysets(c.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading keyset %q: %v", name, err)
	}

	return parseSecret(keyset)
}

// parseSecret attempts to parse the primary secret, otherwise returns nil
func parseSecret(keyset *kops.Keyset) (*fi.Secret, error) {
	primary := fi.FindPrimary(keyset)
	if primary == nil {
		return nil, nil
	}

	s := &fi.Secret{}
	s.Data = primary.PrivateMaterial
	return s, nil
}

// createSecret will create the Secret, overwriting an existing secret if replace is true
func (c *ClientsetSecretStore) createSecret(s *fi.Secret, name string, replace bool) (*kops.Keyset, error) {
	keyset := &kops.Keyset{}
	keyset.Name = NamePrefix + name
	keyset.Spec.Type = kops.SecretTypeSecret

	t := time.Now().UnixNano()
	id := pki.BuildPKISerial(t)

	keyset.Spec.Keys = append(keyset.Spec.Keys, kops.KeysetItem{
		Id:              id.String(),
		PrivateMaterial: s.Data,
	})

	if replace {
		return c.clientset.Keysets(c.namespace).Update(keyset)
	}
	return c.clientset.Keysets(c.namespace).Create(keyset)
}
