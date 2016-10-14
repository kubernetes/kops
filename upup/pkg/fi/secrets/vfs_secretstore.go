/*
Copyright 2016 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
	"os"
)

type VFSSecretStore struct {
	basedir vfs.Path
}

var _ fi.SecretStore = &VFSSecretStore{}

func NewVFSSecretStore(basedir vfs.Path) fi.SecretStore {
	c := &VFSSecretStore{
		basedir: basedir,
	}
	return c
}

func (s *VFSSecretStore) VFSPath() vfs.Path {
	return s.basedir
}

func (c *VFSSecretStore) buildSecretPath(id string) vfs.Path {
	return c.basedir.Join(id)
}

func (c *VFSSecretStore) FindSecret(id string) (*fi.Secret, error) {
	p := c.buildSecretPath(id)
	s, err := c.loadSecret(p)
	if err != nil {
		return nil, err
	}
	return s, nil
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

		err = c.createSecret(secret, p)
		if err != nil {
			if os.IsExist(err) && i == 0 {
				glog.Infof("Got already-exists error when writing secret; likely due to concurrent creation.  Will retry")
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
		glog.Fatalf("unable to load secret immmediately after creation %v: %v", p, err)
		return nil, false, err
	}
	return s, true, nil
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

// createSecret writes the secret, but only if it does not exists
func (c *VFSSecretStore) createSecret(s *fi.Secret, p vfs.Path) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error serializing secret: %v", err)
	}
	return p.CreateFile(data)
}
