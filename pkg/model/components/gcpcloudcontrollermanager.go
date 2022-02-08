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
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/loader"
)

type GCPCloudControllerManagerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = (*GCPCloudControllerManagerOptionsBuilder)(nil)

func (b *GCPCloudControllerManagerOptionsBuilder) BuildOptions(options interface{}) error {
	clusterSpec := options.(*kops.ClusterSpec)

	if kops.CloudProviderID(clusterSpec.CloudProvider) != kops.CloudProviderGCE {
		return nil
	}

	if clusterSpec.ExternalCloudControllerManager == nil && b.IsKubernetesGTE("1.24") {
		clusterSpec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{}
	}

	ccmConfig := clusterSpec.ExternalCloudControllerManager

	if ccmConfig == nil {
		return nil
	}

	// No significant downside to always doing a leader election.
	// Also, having multiple control plane nodes requires leader election.
	ccmConfig.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.Bool(true)}

	// CCM interacts directly with the GCP API, use the name safe for GCP
	ccmConfig.ClusterName = gce.SafeClusterName(b.ClusterName)
	ccmConfig.AllocateNodeCIDRs = fi.Bool(true)
	ccmConfig.CIDRAllocatorType = fi.String("CloudAllocator")
	if ccmConfig.ClusterCIDR == "" {
		ccmConfig.ClusterCIDR = clusterSpec.PodCIDR
	}
	if ccmConfig.Image == "" {
		// TODO: Implement CCM image publishing
		ccmConfig.Image = "k8scloudprovidergcp/cloud-controller-manager:v1.23.0"
	}

	if b.IsKubernetesGTE("1.24") && b.IsKubernetesLT("1.25") {
		ccmConfig.EnableLeaderMigration = fi.Bool(true)
	}

	return nil
}
