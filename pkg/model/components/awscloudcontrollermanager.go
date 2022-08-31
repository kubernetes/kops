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
	"k8s.io/kops/upup/pkg/fi/loader"
)

// AWSCloudControllerManagerOptionsBuilder adds options for the kubernetes controller manager to the model.
type AWSCloudControllerManagerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &AWSCloudControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used for the AWS cloud controller manager manifest
func (b *AWSCloudControllerManagerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.GetCloudProvider() != kops.CloudProviderAWS {
		return nil
	}

	if clusterSpec.ExternalCloudControllerManager == nil && b.IsKubernetesGTE("1.24") {
		clusterSpec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{}
	}

	eccm := clusterSpec.ExternalCloudControllerManager

	if eccm == nil {
		return nil
	}

	// No significant downside to always doing a leader election.
	// Also, having multiple control plane nodes requires leader election.
	eccm.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.Bool(true)}

	eccm.ClusterName = b.ClusterName

	eccm.ClusterCIDR = clusterSpec.NonMasqueradeCIDR

	eccm.AllocateNodeCIDRs = fi.Bool(true)
	eccm.ConfigureCloudRoutes = fi.Bool(false)

	// TODO: we want to consolidate this with the logic from KCM
	networking := clusterSpec.Networking
	if networking == nil {
		eccm.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.Kubenet != nil {
		eccm.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.GCE != nil {
		eccm.ConfigureCloudRoutes = fi.Bool(false)
		eccm.CIDRAllocatorType = fi.String("CloudAllocator")

		if eccm.ClusterCIDR == "" {
			eccm.ClusterCIDR = clusterSpec.PodCIDR
		}
	} else if networking.External != nil {
		eccm.ConfigureCloudRoutes = fi.Bool(false)
	} else if UsesCNI(networking) {
		eccm.ConfigureCloudRoutes = fi.Bool(false)
	} else if networking.Kopeio != nil {
		// Kopeio is based on kubenet / external
		eccm.ConfigureCloudRoutes = fi.Bool(false)
	} else {
		return fmt.Errorf("no networking mode set")
	}

	if eccm.Image == "" {
		// See https://us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager
		switch b.KubernetesVersion.Minor {
		case 20:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.20.1"
		case 21:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.21.3"
		case 22:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.22.2"
		case 23:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.23.2"
		case 24:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.24.1"
		case 25:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.0"
		default:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.24.1"
		}
	}

	if b.IsKubernetesGTE("1.24") && b.IsKubernetesLT("1.25") {
		eccm.EnableLeaderMigration = fi.Bool(true)
	}

	return nil
}
