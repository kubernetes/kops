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

package components

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

const DefaultBackupImage = "kopeio/etcd-backup:1.0.20180220"

// EtcdOptionsBuilder adds options for etcd to the model
type EtcdOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &EtcdOptionsBuilder{}

const DefaultEtcdVersion = "2.2.1"

// BuildOptions is responsible for filling in the defaults for the etcd cluster model
func (b *EtcdOptionsBuilder) BuildOptions(o interface{}) error {
	spec := o.(*kops.ClusterSpec)

	for _, c := range spec.EtcdClusters {
		// Ensure the version is set
		// @TODO once we have a way of detecting a 'new' cluster we can default all clusters to v3
		if c.Version == "" {
			c.Version = DefaultEtcdVersion
		}

		useEtcdManager := false
		if c.Manager != nil {
			useEtcdManager = true
		}

		// remap etcd image
		{
			image := c.Image
			if image == "" {
				if !useEtcdManager {
					image = fmt.Sprintf("k8s.gcr.io/etcd:%s", c.Version)
				}
			}

			if image != "" {
				image, err := b.Context.AssetBuilder.RemapImage(image)
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
				if !useEtcdManager {
					image = fmt.Sprintf(DefaultBackupImage)
				}
			}

			if image != "" {
				image, err := b.Context.AssetBuilder.RemapImage(image)
				if err != nil {
					return fmt.Errorf("unable to remap container %q: %v", image, err)
				}
				c.Backups.Image = image
			}
		}

		// remap etcd manager image
		if useEtcdManager {
			image := c.Manager.Image
			if image == "" {
				// We can make this easier later - maybe put it into the channel?  Or an addon?
				//image = fmt.Sprintf(DefaultManagerImage)
				return fmt.Errorf("EtcdManager.Image must be specified (currently)")
			}

			if image != "" {
				image, err := b.Context.AssetBuilder.RemapImage(image)
				if err != nil {
					return fmt.Errorf("unable to remap container %q: %v", image, err)
				}
				c.Manager.Image = image
			}
		}
	}

	return nil
}
