/*
Copyright 2025 The Kubernetes Authors.

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

// AzureCloudControllerManagerOptionsBuilder adds options for the Azure cloud controller manager to the model.
type AzureCloudControllerManagerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.ClusterOptionsBuilder = &AzureCloudControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used for the Azure cloud controller manager manifest
func (b *AzureCloudControllerManagerOptionsBuilder) BuildOptions(cluster *kops.Cluster) error {
	if cluster.GetCloudProvider() != kops.CloudProviderAzure {
		return nil
	}

	if cluster.Spec.ExternalCloudControllerManager == nil {
		cluster.Spec.ExternalCloudControllerManager = &kops.CloudControllerManagerConfig{}
	}

	eccm := cluster.Spec.ExternalCloudControllerManager

	if eccm.Image == "" {
		eccm.Image = "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-controller-manager:v1.36.2"
	}

	if eccm.AzureNodeManagerImage == "" {
		eccm.AzureNodeManagerImage = "mcr.microsoft.com/oss/v2/kubernetes/azure-cloud-node-manager:v1.36.2"
	}

	eccm.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.PtrTo(true)}

	if eccm.LogLevel == 0 {
		eccm.LogLevel = 2
	}

	// Kubenet and kindnet rely on cloud routes to deliver pod-to-pod traffic across nodes:
	// the Azure VNet is L3 and drops packets whose destination IP is not a known VNet IP
	// (pod CIDRs are not), so the Azure CCM must populate UDRs in the route table that
	// kOps associates with the node subnet. EnableIPForwarding is already set on every NIC.
	networking := cluster.Spec.Networking
	if networking.Kubenet != nil || networking.Kindnet != nil {
		eccm.ConfigureCloudRoutes = fi.PtrTo(true)
	} else {
		eccm.ConfigureCloudRoutes = fi.PtrTo(false)
	}

	return nil
}
