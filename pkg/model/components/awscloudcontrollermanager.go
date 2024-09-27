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

var _ loader.ClusterOptionsBuilder = &AWSCloudControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used for the AWS cloud controller manager manifest
func (b *AWSCloudControllerManagerOptionsBuilder) BuildOptions(cluster *kops.Cluster) error {
	clusterSpec := &cluster.Spec

	if cluster.GetCloudProvider() != kops.CloudProviderAWS {
		return nil
	}

	if clusterSpec.ExternalCloudControllerManager == nil {
		clusterSpec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{}
	}

	eccm := clusterSpec.ExternalCloudControllerManager

	// No significant downside to always doing a leader election.
	// Also, having multiple control plane nodes requires leader election.
	eccm.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.PtrTo(true)}

	eccm.ClusterName = b.ClusterName

	eccm.AllocateNodeCIDRs = fi.PtrTo(!clusterSpec.IsKopsControllerIPAM())

	if eccm.ClusterCIDR == "" && !clusterSpec.IsKopsControllerIPAM() {
		eccm.ClusterCIDR = clusterSpec.Networking.PodCIDR
	}

	// TODO: we want to consolidate this with the logic from KCM
	networking := &clusterSpec.Networking
	if networking.Kubenet != nil {
		eccm.ConfigureCloudRoutes = fi.PtrTo(true)
	} else if networking.External != nil {
		eccm.ConfigureCloudRoutes = fi.PtrTo(false)
	} else if UsesCNI(networking) {
		eccm.ConfigureCloudRoutes = fi.PtrTo(false)
	} else if networking.Kopeio != nil {
		// Kopeio is based on kubenet / external
		eccm.ConfigureCloudRoutes = fi.PtrTo(false)
	} else {
		return fmt.Errorf("no networking mode set")
	}

	if eccm.Image == "" {
		// See https://us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager
		switch b.KubernetesVersion.Minor {
		case 25:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.15"
		case 26:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.12"
		case 27:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.9"
		case 28:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.28.9"
		case 29:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.29.6"
		case 30:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.30.3"
		case 31:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.31.0"
		default:
			eccm.Image = "registry.k8s.io/provider-aws/cloud-controller-manager:v1.31.0"
		}
	}

	if b.IsKubernetesLT("1.25") {
		eccm.EnableLeaderMigration = fi.PtrTo(true)
	}

	return nil
}
