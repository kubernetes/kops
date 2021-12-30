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

package azuremodel

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	nodeidentityazure "k8s.io/kops/pkg/nodeidentity/azure"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// AzureModelContext is
type AzureModelContext struct {
	*model.KopsModelContext
}

// LinkToVirtualNetwork returns the Azure Virtual Network object the cluster is located in.
func (c *AzureModelContext) LinkToVirtualNetwork() *azuretasks.VirtualNetwork {
	return &azuretasks.VirtualNetwork{Name: fi.String(c.NameForVirtualNetwork())}
}

// NameForVirtualNetwork returns the name of the Azure Virtual Network object the cluster is located in.
func (c *AzureModelContext) NameForVirtualNetwork() string {
	networkName := c.Cluster.Spec.NetworkID
	if networkName == "" {
		networkName = c.ClusterName()
	}
	return networkName
}

// LinkToResourceGroup returns the Resource Group object the cluster is located in.
func (c *AzureModelContext) LinkToResourceGroup() *azuretasks.ResourceGroup {
	return &azuretasks.ResourceGroup{Name: fi.String(c.NameForResourceGroup())}
}

// NameForResourceGroup returns the name of the Resource Group object the cluster is located in.
func (c *AzureModelContext) NameForResourceGroup() string {
	return c.Cluster.AzureResourceGroupName()
}

// LinkToAzureSubnet returns the Azure Subnet object the cluster is located in.
func (c *AzureModelContext) LinkToAzureSubnet(spec *kops.ClusterSubnetSpec) *azuretasks.Subnet {
	return &azuretasks.Subnet{Name: fi.String(spec.Name)}
}

// NameForRouteTable returns the name of the Route Table object for the cluster.
func (c *AzureModelContext) NameForRouteTable() string {
	return c.Cluster.Spec.CloudProvider.Azure.RouteTableName
}

// LinkToLoadBalancer returns the Load Balancer object for the cluster.
func (c *AzureModelContext) LinkToLoadBalancer() *azuretasks.LoadBalancer {
	return &azuretasks.LoadBalancer{Name: fi.String(c.NameForLoadBalancer())}
}

// NameForLoadBalancer returns the name of the Load Balancer object for the cluster.
func (c *AzureModelContext) NameForLoadBalancer() string {
	return "api-" + c.ClusterName()
}

// CloudTagsForInstanceGroup computes the tags to apply to instances in the specified InstanceGroup
// Mostly copied from pkg/model/context.go, but "/" in tag keys are replaced with "_" as Azure
// doesn't allow "/" in tag keys.
func (c *AzureModelContext) CloudTagsForInstanceGroup(ig *kops.InstanceGroup) map[string]*string {
	const (
		clusterNodeTemplateLabel = "k8s.io_cluster_node-template_label_"
		clusterNodeTemplateTaint = "k8s.io_cluster_node-template_taint_"
	)

	labels := make(map[string]string)
	// Apply any user-specified global labels first so they can be overridden by IG-specific labels.
	for k, v := range c.Cluster.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply any user-specified labels.
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply labels for cluster node labels.
	i := 0
	for k, v := range ig.Spec.NodeLabels {
		// Store the label key in the tag value
		// so that we don't need to espace "/" in the label key.
		labels[fmt.Sprintf("%s%d", clusterNodeTemplateLabel, i)] = fmt.Sprintf("%s=%s", k, v)
		i++
	}

	// Apply labels for cluster node taints.
	for _, v := range ig.Spec.Taints {
		splits := strings.SplitN(v, "=", 2)
		if len(splits) > 1 {
			labels[clusterNodeTemplateTaint+splits[0]] = splits[1]
		}
	}

	// The system tags take priority because the cluster likely breaks without them...
	labels[azure.TagNameRolePrefix+strings.ToLower(string(ig.Spec.Role))] = "1"

	// Set the tag used by kops-controller to identify the instance group to which the VM ScaleSet belongs.
	labels[nodeidentityazure.InstanceGroupNameTag] = ig.Name

	// Replace all "/" with "_" as "/" is not an allowed key character in Azure.
	m := make(map[string]*string)
	for k, v := range labels {
		m[strings.ReplaceAll(k, "/", "_")] = fi.String(v)
	}
	return m
}
