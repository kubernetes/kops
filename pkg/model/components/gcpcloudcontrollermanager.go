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
	"fmt"

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

	if clusterSpec.GetCloudProvider() != kops.CloudProviderGCE {
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

	ccmConfig.ConfigureCloudRoutes = fi.Bool(false)

	// TODO: we want to consolidate this with the logic from KCM
	networking := clusterSpec.Networking
	if networking == nil {
		ccmConfig.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.Kubenet != nil {
		ccmConfig.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.GCE != nil {
		ccmConfig.ConfigureCloudRoutes = fi.Bool(false)
		ccmConfig.CIDRAllocatorType = fi.String("CloudAllocator")

		if ccmConfig.ClusterCIDR == "" {
			ccmConfig.ClusterCIDR = clusterSpec.PodCIDR
		}
	} else if networking.External != nil {
		ccmConfig.ConfigureCloudRoutes = fi.Bool(false)
	} else if UsesCNI(networking) {
		ccmConfig.ConfigureCloudRoutes = fi.Bool(false)
	} else if networking.Kopeio != nil {
		// Kopeio is based on kubenet / external
		ccmConfig.ConfigureCloudRoutes = fi.Bool(false)
	} else {
		return fmt.Errorf("no networking mode set")
	}

	if ccmConfig.ClusterCIDR == "" {
		ccmConfig.ClusterCIDR = clusterSpec.PodCIDR
	}
	if ccmConfig.Image == "" {
		// TODO: Implement CCM image publishing
		switch b.KubernetesVersion.Minor {
		case 23:
			ccmConfig.Image = "k8scloudprovidergcp/cloud-controller-manager:v1.23.0"
		default:
			ccmConfig.Image = "k8scloudprovidergcp/cloud-controller-manager:latest"
		}
	}

	if b.IsKubernetesGTE("1.24") && b.IsKubernetesLT("1.25") {
		ccmConfig.EnableLeaderMigration = fi.Bool(true)
	}

	return nil
}
