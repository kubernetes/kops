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
	"context"
	"fmt"
	"os"
	"sync"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSKeystoreReader struct {
	basedir vfs.Path

	mutex    sync.Mutex
	cachedCA *Keyset
}

var _ KeystoreReader = &VFSKeystoreReader{}

func NewVFSKeystoreReader(basedir vfs.Path) *VFSKeystoreReader {
	k := &VFSKeystoreReader{
		basedir: basedir,
	}

	return k
}

func (c *VFSKeystoreReader) VFSPath() vfs.Path {
	return c.basedir
}

func (c *VFSKeystoreReader) buildPrivateKeyPoolPath(name string) vfs.Path {
	return c.basedir.Join("private", name)
}

func (c *VFSKeystoreReader) parseKeysetYaml(data []byte) (*kops.Keyset, bool, error) {
	defaultReadVersion := v1alpha2.SchemeGroupVersion.WithKind("Keyset")

	object, gvk, err := kopscodecs.Decode(data, &defaultReadVersion)
	if err != nil {
		return nil, false, fmt.Errorf("error parsing keyset: %v", err)
	}

	keyset, ok := object.(*kops.Keyset)
	if !ok {
		return nil, false, fmt.Errorf("object was not a keyset, was a %T", object)
	}

	if gvk == nil {
		return nil, false, fmt.Errorf("object did not have GroupVersionKind: %q", keyset.Name)
	}

	return keyset, gvk.Version != keysetFormatLatest, nil
}

// loadKeyset loads a Keyset from the path.
// Returns (nil, nil) if the file is not found
// Bundles avoid the need for a list-files permission, which can be tricky on e.g. GCE
func (c *VFSKeystoreReader) loadKeyset(ctx context.Context, p vfs.Path) (*Keyset, error) {
	bundlePath := p.Join("keyset.yaml")
	data, err := bundlePath.ReadFile(ctx)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read bundle %q: %v", p, err)
	}

	o, legacyFormat, err := c.parseKeysetYaml(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing bundle %q: %v", p, err)
	}

	keyset, err := parseKeyset(o)
	if err != nil {
		return nil, fmt.Errorf("error mapping bundle %q: %v", p, err)
	}

	keyset.LegacyFormat = legacyFormat
	return keyset, nil
}

var legacyKeysetMappings = map[string]string{
	// The strange name is because kOps prior to 1.19 used the api-server TLS key for this.
	"service-account": "master",
	// Renamed in kOps 1.22
	"kubernetes-ca": "ca",
}

// FindPrimaryKeypair implements pki.Keystore
func (c *VFSKeystoreReader) FindPrimaryKeypair(ctx context.Context, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	keyset, err := c.FindKeyset(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	if keyset == nil {
		return nil, nil, nil
	}
	if keyset.Primary == nil {
		return nil, nil, nil
	}
	if keyset.Primary.Certificate == nil {
		return nil, nil, nil
	}
	return keyset.Primary.Certificate, keyset.Primary.PrivateKey, nil
}

func (c *VFSKeystoreReader) FindKeyset(ctx context.Context, id string) (*Keyset, error) {
	keys, err := c.findPrivateKeyset(ctx, id)
	if keys == nil || os.IsNotExist(err) {
		if legacyId := legacyKeysetMappings[id]; legacyId != "" {
			keys, err = c.findPrivateKeyset(ctx, legacyId)
			if keys != nil {
				keys.LegacyFormat = true
			}
		}
	}

	return keys, err
}

func (c *VFSKeystoreReader) findPrivateKeyset(ctx context.Context, id string) (*Keyset, error) {
	var keys *Keyset
	var err error
	if id == CertificateIDCA {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		cached := c.cachedCA
		if cached != nil {
			return cached, nil
		}

		keys, err = c.loadKeyset(ctx, c.buildPrivateKeyPoolPath(id))
		if err != nil {
			return nil, err
		}

		if keys == nil {
			klog.Warningf("CA private key was not found")
			// We no longer generate CA certificates automatically - too race-prone
		} else {
			c.cachedCA = keys
		}
	} else {
		p := c.buildPrivateKeyPoolPath(id)
		keys, err = c.loadKeyset(ctx, p)
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}
