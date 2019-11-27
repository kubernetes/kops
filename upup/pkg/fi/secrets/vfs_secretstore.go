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

package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/klog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSSecretStore struct {
	cluster *kops.Cluster
	basedir vfs.Path
}

var _ fi.SecretStore = &VFSSecretStore{}

func NewVFSSecretStore(cluster *kops.Cluster, basedir vfs.Path) fi.SecretStore {
	c := &VFSSecretStore{
		cluster: cluster,
		basedir: basedir,
	}
	return c
}

func (c *VFSSecretStore) VFSPath() vfs.Path {
	return c.basedir
}

func (c *VFSSecretStore) MirrorTo(basedir vfs.Path) error {
	if basedir.Path() == c.basedir.Path() {
		klog.V(2).Infof("Skipping mirror of secret store from %q to %q (same path)", c.basedir, basedir)
		return nil
	}
	klog.V(2).Infof("Mirroring secret store from %q to %q", c.basedir, basedir)

	secrets, err := c.ListSecrets()
	if err != nil {
		return fmt.Errorf("error listing secrets for mirror: %v", err)
	}

	for _, name := range secrets {
		secret, err := c.FindSecret(name)
		if err != nil {
			return fmt.Errorf("error reading secret %q for mirror: %v", name, err)
		}

		if secret == nil {
			return fmt.Errorf("unable to find secret %q for mirror", name)
		}

		p := BuildVfsSecretPath(basedir, name)

		acl, err := acls.GetACL(p, c.cluster)
		if err != nil {
			return fmt.Errorf("error building acl for secret %q for mirror: %v", name, err)
		}

		klog.Infof("mirroring secret %s -> %s", name, p)

		err = createSecret(secret, p, acl, true)
		if err != nil {
			return fmt.Errorf("error writing secret %q for mirror: %v", name, err)
		}
	}

	return nil
}

func BuildVfsSecretPath(basedir vfs.Path, name string) vfs.Path {
	return basedir.Join(name)
}

func (c *VFSSecretStore) buildSecretPath(name string) vfs.Path {
	return BuildVfsSecretPath(c.basedir, name)
}

func (c *VFSSecretStore) FindSecret(id string) (*fi.Secret, error) {
	p := c.buildSecretPath(id)
	s, err := c.loadSecret(p)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// DeleteSecret implements fi.SecretStore DeleteSecret
func (c *VFSSecretStore) DeleteSecret(name string) error {
	p := c.buildSecretPath(name)
	return p.Remove()
}

func (c *VFSSecretStore) ListSecrets() ([]string, error) {
	files, err := c.basedir.ReadDir()
	if err != nil {
		return nil, fmt.Errorf("error listing secrets directory: %v", err)
	}
	var ids []string
	for _, f := range files {
		id := f.Base()
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *VFSSecretStore) Secret(id string) (*fi.Secret, error) {
	s, err := c.FindSecret(id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, fmt.Errorf("Secret not found: %q", id)
	}
	return s, nil
}

func (c *VFSSecretStore) GetOrCreateSecret(id string, secret *fi.Secret) (*fi.Secret, bool, error) {
	p := c.buildSecretPath(id)

	for i := 0; i < 2; i++ {
		s, err := c.FindSecret(id)
		if err != nil {
			return nil, false, err
		}

		if s != nil {
			return s, false, nil
		}

		acl, err := acls.GetACL(p, c.cluster)
		if err != nil {
			return nil, false, err
		}

		err = createSecret(secret, p, acl, false)
		if err != nil {
			if os.IsExist(err) && i == 0 {
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
	s, err := c.loadSecret(p)
	if err != nil {
		klog.Fatalf("unable to load secret immediately after creation %v: %v", p, err)
		return nil, false, err
	}
	return s, true, nil
}

func (c *VFSSecretStore) ReplaceSecret(id string, secret *fi.Secret) (*fi.Secret, error) {
	p := c.buildSecretPath(id)

	acl, err := acls.GetACL(p, c.cluster)
	if err != nil {
		return nil, err
	}

	err = createSecret(secret, p, acl, true)
	if err != nil {
		return nil, fmt.Errorf("unable to write secret: %v", err)
	}

	// Confirm the secret exists
	s, err := c.loadSecret(p)
	if err != nil {
		return nil, fmt.Errorf("unable to load secret immediately after creation %v: %v", p, err)
	}
	return s, nil
}

func (c *VFSSecretStore) loadSecret(p vfs.Path) (*fi.Secret, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	s := &fi.Secret{}
	err = json.Unmarshal(data, s)
	if err != nil {
		return nil, fmt.Errorf("error parsing secret from %q: %v", p, err)
	}
	return s, nil
}

// createSecret will create the Secret, overwriting an existing secret if replace is true
func createSecret(s *fi.Secret, p vfs.Path, acl vfs.ACL, replace bool) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error serializing secret: %v", err)
	}

	rs := bytes.NewReader(data)
	if replace {
		return p.WriteFile(rs, acl)
	}
	return p.CreateFile(rs, acl)
}
