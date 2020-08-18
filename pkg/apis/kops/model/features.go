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

// UseKopsControllerForNodeBootstrap is true if nodeup should use kops-controller for bootstrapping.
func UseKopsControllerForNodeBootstrap(cluster *kops.Cluster) bool {
	return kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderAWS && cluster.IsKubernetesGTE("1.19")
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
