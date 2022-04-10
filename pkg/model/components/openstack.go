/*
Copyright 2020 The Kubernetes Authors.

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

// OpenStackOptionsBuilder adds options for OpenStack to the model
type OpenStackOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &OpenStackOptionsBuilder{}

func (b *OpenStackOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	openstack := clusterSpec.CloudProvider.Openstack

	if openstack == nil {
		return nil
	}

	if clusterSpec.CloudConfig == nil {
		clusterSpec.CloudConfig = &kops.CloudConfiguration{}
	}

	if openstack.BlockStorage == nil {
		openstack.BlockStorage = &kops.OpenstackBlockStorageConfig{}
	}

	if openstack.BlockStorage.CreateStorageClass == nil {
		openstack.BlockStorage.CreateStorageClass = fi.Bool(true)
	}

	if openstack.Metadata == nil {
		openstack.Metadata = &kops.OpenstackMetadata{}
	}
	if openstack.Metadata.ConfigDrive == nil {
		openstack.Metadata.ConfigDrive = fi.Bool(false)
	}

	if clusterSpec.ExternalCloudControllerManager == nil {
		clusterSpec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{
			// No significant downside to always doing a leader election.
			// Also, having a replicated (HA) control plane requires leader election.
			LeaderElection: &kops.LeaderElectionConfiguration{LeaderElect: fi.Bool(true)},
		}
	}

	return nil
}
