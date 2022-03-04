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

// OpenStackOptionsBulder adds options for OpenStack to the model
type OpenStackOptionsBulder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &OpenStackOptionsBulder{}

func (b *OpenStackOptionsBulder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.CloudProvider.Openstack == nil {
		return nil
	}

	if clusterSpec.CloudConfig == nil {
		clusterSpec.CloudConfig = &kops.CloudConfiguration{}
	}

	if clusterSpec.CloudProvider.Openstack.BlockStorage == nil {
		clusterSpec.CloudProvider.Openstack.BlockStorage = &kops.OpenstackBlockStorageConfig{}
	}

	if clusterSpec.CloudProvider.Openstack.BlockStorage.CreateStorageClass == nil {
		clusterSpec.CloudProvider.Openstack.BlockStorage.CreateStorageClass = fi.Bool(true)
	}

	if clusterSpec.CloudProvider.Openstack.Metadata == nil {
		clusterSpec.CloudProvider.Openstack.Metadata = &kops.OpenstackMetadata{}
	}
	if clusterSpec.CloudProvider.Openstack.Metadata.ConfigDrive == nil {
		clusterSpec.CloudProvider.Openstack.Metadata.ConfigDrive = fi.Bool(false)
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
