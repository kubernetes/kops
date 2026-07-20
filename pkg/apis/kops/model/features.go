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

package model

import (
	"k8s.io/kops/pkg/apis/kops"
)

// UseChallengeCallback is true if we should use a callback challenge during node provisioning with kops-controller.
func UseChallengeCallback(cloudProvider kops.CloudProviderID) bool {
	switch cloudProvider {
	case kops.CloudProviderHetzner:
		return true
	case kops.CloudProviderDO:
		return true
	case kops.CloudProviderScaleway:
		return true
	case kops.CloudProviderAzure:
		return true
	case kops.CloudProviderLinode:
		return true
	default:
		return false
	}
}

// UseKopsControllerForNodeConfig checks if nodeup should use kops-controller to get nodeup.Config.
// Gossip and None-DNS clusters need a fixed kops-controller endpoint baked into worker boot config,
// which is exactly what UsesLoadBalancerForKopsController offers.
func UseKopsControllerForNodeConfig(cluster *kops.Cluster) bool {
	return !cluster.UsesLegacyGossip() || cluster.UsesLoadBalancerForKopsController()
}

// UseKopsControllerForInstanceGroupNodeConfig checks if nodeup should retrieve its configuration from kops-controller.
func UseKopsControllerForInstanceGroupNodeConfig(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) bool {
	if cluster.GetCloudProvider() == kops.CloudProviderGCE && instanceGroup.Spec.DisableTPM {
		return false
	}
	return UseKopsControllerForNodeConfig(cluster)
}

// UseCiliumEtcd is true if we are using the Cilium etcd cluster.
func UseCiliumEtcd(cluster *kops.Cluster) bool {
	if cluster.Spec.Networking.Cilium == nil {
		return false
	}

	for _, cluster := range cluster.Spec.EtcdClusters {
		if cluster.Name == "cilium" {
			return true
		}
	}

	return false
}
