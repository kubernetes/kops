/*
Copyright 2026 The Kubernetes Authors.

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
	"k8s.io/kops/upup/pkg/fi/loader"
)

// LinodeCloudControllerManagerOptionsBuilder adds options for the Linode (Akamai) cloud controller manager to the model.
type LinodeCloudControllerManagerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.ClusterOptionsBuilder = &LinodeCloudControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used for the Linode (Akamai) cloud controller manager manifest
func (b *LinodeCloudControllerManagerOptionsBuilder) BuildOptions(cluster *kops.Cluster) error {
	clusterSpec := &cluster.Spec

	if cluster.GetCloudProvider() != kops.CloudProviderLinode {
		return nil
	}

	if clusterSpec.ExternalCloudControllerManager == nil {
		clusterSpec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{}
	}

	eccm := clusterSpec.ExternalCloudControllerManager
	eccm.CloudProvider = "linode"
	eccm.LeaderElection = &kops.LeaderElectionConfiguration{
		LeaderElect: new(true),
	}

	if eccm.ClusterCIDR == "" {
		eccm.ClusterCIDR = clusterSpec.Networking.PodCIDR
	}
	eccm.AllocateNodeCIDRs = new(true)
	eccm.ConfigureCloudRoutes = new(false)

	if eccm.Image == "" {
		// Using the official Linode (Akamai) CCM image
		// https://github.com/linode/linode-cloud-controller-manager
		eccm.Image = "linode/linode-cloud-controller-manager:v0.9.5"
	}

	return nil
}
