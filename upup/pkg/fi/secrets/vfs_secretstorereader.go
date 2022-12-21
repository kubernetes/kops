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
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSSecretStoreReader struct {
	basedir vfs.Path
}

var _ fi.SecretStoreReader = &VFSSecretStoreReader{}

func NewVFSSecretStoreReader(basedir vfs.Path) fi.SecretStoreReader {
	c := &VFSSecretStoreReader{
		basedir: basedir,
	}
	return c
}

func (c *VFSSecretStoreReader) VFSPath() vfs.Path {
	return c.basedir
}

func BuildVfsSecretPath(basedir vfs.Path, name string) vfs.Path {
	return basedir.Join(name)
}

func (c *VFSSecretStoreReader) buildSecretPath(name string) vfs.Path {
	return BuildVfsSecretPath(c.basedir, name)
}

func (c *VFSSecretStoreReader) FindSecret(id string) (*fi.Secret, error) {
	p := c.buildSecretPath(id)
	s, err := c.loadSecret(p)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (c *VFSSecretStoreReader) Secret(id string) (*fi.Secret, error) {
	s, err := c.FindSecret(id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, fmt.Errorf("secret %q not found", id)
	}
	return s, nil
}

func (c *VFSSecretStoreReader) loadSecret(p vfs.Path) (*fi.Secret, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	s := &fi.Secret{}
	err = json.Unmarshal(data, s)
	if err != nil {
		return nil, fmt.Errorf("parsing secret from %q: %v", p, err)
	}
	return s, nil
}
