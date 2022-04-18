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

package components

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

const DefaultBackupImage = "k8s.gcr.io/etcdadm/etcd-backup:3.0.20210707"

// EtcdOptionsBuilder adds options for etcd to the model
type EtcdOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &EtcdOptionsBuilder{}

const (
	DefaultEtcd3Version_1_17 = "3.4.3"
	DefaultEtcd3Version_1_19 = "3.4.13"
	DefaultEtcd3Version_1_22 = "3.5.3"
)

// BuildOptions is responsible for filling in the defaults for the etcd cluster model
func (b *EtcdOptionsBuilder) BuildOptions(o interface{}) error {
	spec := o.(*kops.ClusterSpec)

	for i := range spec.EtcdClusters {
		c := &spec.EtcdClusters[i]
		if c.Provider == "" {
			c.Provider = kops.EtcdProviderTypeManager
		}

		// Ensure the version is set
		if c.Version == "" {
			// We run the k8s-recommended versions of etcd
			if b.IsKubernetesGTE("1.22") {
				c.Version = DefaultEtcd3Version_1_22
			} else if b.IsKubernetesGTE("1.19") {
				c.Version = DefaultEtcd3Version_1_19
			} else {
				c.Version = DefaultEtcd3Version_1_17
			}
		}

		// We make sure that etcd v3 is used
		version := strings.TrimPrefix(c.Version, "v")
		if !strings.HasPrefix(version, "3.") {
			return fmt.Errorf("unexpected etcd version %q", c.Version)
		}

		// We enable TLS if we're running EtcdManager
		if c.Provider == kops.EtcdProviderTypeManager {
			c.EnableEtcdTLS = true
			c.EnableTLSAuth = true
		}

		// We remap the etcd manager image when we build the manifest,
		// but we need to map the standalone images here because protokube launches them
		if c.Provider == kops.EtcdProviderTypeLegacy {

			// remap etcd image
			{
				image := c.Image
				if image == "" {
					image = fmt.Sprintf("k8s.gcr.io/etcd:%s", c.Version)
				}

				if image != "" {
					image, err := b.AssetBuilder.RemapImage(image)
					if err != nil {
						return fmt.Errorf("unable to remap container %q: %v", image, err)
					}
					c.Image = image
				}
			}

			// remap backup manager image
			if c.Backups != nil {
				image := c.Backups.Image
				if image == "" {
					image = DefaultBackupImage
				}

				if image != "" {
					image, err := b.AssetBuilder.RemapImage(image)
					if err != nil {
						return fmt.Errorf("unable to remap container %q: %v", image, err)
					}
					c.Backups.Image = image
				}
			}
		}
	}

	return nil
}
