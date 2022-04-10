/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// CloudConfigurationOptionsBuilder prepares settings related to the backing cloud provider.
type CloudConfigurationOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &CloudConfigurationOptionsBuilder{}

func (b *CloudConfigurationOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	c := clusterSpec.CloudConfig
	if c == nil {
		c = &kops.CloudConfiguration{}
		clusterSpec.CloudConfig = c
	}

	// NB: See file openstack.go for establishing default values for the CloudConfig.Openstack
	// field.

	if c.ManageStorageClasses == nil {
		var manage *bool
		if clusterSpec.CloudProvider.Openstack != nil && clusterSpec.CloudProvider.Openstack.BlockStorage != nil && clusterSpec.CloudProvider.Openstack.BlockStorage.CreateStorageClass != nil {
			// Avoid a spurious conflict with a user-specified configuration for OpenStack by
			// adopting that more particular setting generally.
			manage = clusterSpec.CloudProvider.Openstack.BlockStorage.CreateStorageClass
		} else {
			manage = fi.Bool(true)
		}
		c.ManageStorageClasses = manage
	}

	if clusterSpec.IsIPv6Only() && len(c.NodeIPFamilies) == 0 {
		c.NodeIPFamilies = []string{"ipv6", "ipv4"}
	}

	return nil
}
