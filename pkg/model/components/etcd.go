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

const DefaultBackupImage = "kopeio/etcd-backup:3.0.20200307"

// EtcdOptionsBuilder adds options for etcd to the model
type EtcdOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &EtcdOptionsBuilder{}

const (
	DefaultEtcd2Version = "2.2.1"

	//  1.11 originally recommended 3.2.18, but there was an advisory to update to 3.2.24
	DefaultEtcd3Version_1_11 = "3.2.24"

	DefaultEtcd3Version_1_13 = "3.2.24"

	DefaultEtcd3Version_1_14 = "3.3.10"

	DefaultEtcd3Version_1_17 = "3.4.3"
)

// BuildOptions is responsible for filling in the defaults for the etcd cluster model
func (b *EtcdOptionsBuilder) BuildOptions(o interface{}) error {
	spec := o.(*kops.ClusterSpec)

	for _, c := range spec.EtcdClusters {
		if c.Provider == "" {
			if b.IsKubernetesGTE("1.12") {
				c.Provider = kops.EtcdProviderTypeManager
			} else if c.Manager != nil {
				c.Provider = kops.EtcdProviderTypeManager
			} else {
				c.Provider = kops.EtcdProviderTypeLegacy
			}
		}

		// Ensure the version is set
		if c.Version == "" && c.Provider == kops.EtcdProviderTypeLegacy {
			// Even if in legacy mode, etcd version 2 is unsupported as of k8s 1.13
			if b.IsKubernetesGTE("1.17") {
				c.Version = DefaultEtcd3Version_1_17
			} else if b.IsKubernetesGTE("1.14") {
				c.Version = DefaultEtcd3Version_1_14
			} else if b.IsKubernetesGTE("1.13") {
				c.Version = DefaultEtcd3Version_1_13
			} else {
				c.Version = DefaultEtcd2Version
			}
		}

		if c.Version == "" && c.Provider == kops.EtcdProviderTypeManager {
			// From 1.11, we run the k8s-recommended versions of etcd when using the manager
			if b.IsKubernetesGTE("1.17") {
				c.Version = DefaultEtcd3Version_1_17
			} else if b.IsKubernetesGTE("1.14") {
				c.Version = DefaultEtcd3Version_1_14
			} else if b.IsKubernetesGTE("1.13") {
				c.Version = DefaultEtcd3Version_1_13
			} else if b.IsKubernetesGTE("1.11") {
				c.Version = DefaultEtcd3Version_1_11
			} else {
				c.Version = DefaultEtcd2Version
			}
		}

		// From 1.12, we enable TLS if we're running EtcdManager & etcd3
		//
		// (Moving to etcd3 is a disruptive upgrade, so we
		// force TLS at the same time as we enable
		// etcd-manager by default).
		if c.Provider == kops.EtcdProviderTypeManager {
			etcdV3 := true
			version := c.Version
			version = strings.TrimPrefix(version, "v")
			if strings.HasPrefix(version, "2.") {
				etcdV3 = false
			} else if strings.HasPrefix(version, "3.") {
				etcdV3 = true
			} else {
				return fmt.Errorf("unexpected etcd version %q", c.Version)
			}

			if b.IsKubernetesGTE("1.12.0") && etcdV3 {
				c.EnableEtcdTLS = true
				c.EnableTLSAuth = true
			}
		}
	}

	// Remap the well known images
	for _, c := range spec.EtcdClusters {

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
