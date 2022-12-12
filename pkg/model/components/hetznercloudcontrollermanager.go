/*
Copyright 2022 The Kubernetes Authors.

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
	"context"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// HetznerCloudControllerManagerOptionsBuilder adds options for the kubernetes controller manager to the model.
type HetznerCloudControllerManagerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &HetznerCloudControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used for the Hetzner cloud controller manager manifest
func (b *HetznerCloudControllerManagerOptionsBuilder) BuildOptions(ctx context.Context, o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.GetCloudProvider() != kops.CloudProviderHetzner {
		return nil
	}

	if clusterSpec.ExternalCloudControllerManager == nil {
		clusterSpec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{}
	}

	eccm := clusterSpec.ExternalCloudControllerManager
	eccm.CloudProvider = "hcloud"
	eccm.AllowUntaggedCloud = fi.PtrTo(true)
	eccm.LeaderElection = &kops.LeaderElectionConfiguration{
		LeaderElect: fi.PtrTo(false),
	}

	eccm.ClusterCIDR = clusterSpec.Networking.NonMasqueradeCIDR
	eccm.AllocateNodeCIDRs = fi.PtrTo(true)
	eccm.ConfigureCloudRoutes = fi.PtrTo(false)

	if eccm.Image == "" {
		eccm.Image = "hetznercloud/hcloud-cloud-controller-manager:v1.13.2"
	}

	return nil
}
