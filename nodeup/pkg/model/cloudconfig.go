/*
Copyright 2019 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

const (
	CloudConfigFilePath       = "/etc/kubernetes/cloud.config"
	InTreeCloudConfigFilePath = "/etc/kubernetes/in-tree-cloud.config"

	// VM UUID is set by cloud-init
	VM_UUID_FILE_PATH = "/etc/vmware/vm_uuid"
)

// azureCloudConfig is the configuration passed to Cloud Provider Azure.
// The specification is described in https://kubernetes-sigs.github.io/cloud-provider-azure/install/configs/.
type azureCloudConfig struct {
	// SubscriptionID is the ID of the Azure Subscription that the cluster is deployed in.
	SubscriptionID string `json:"subscriptionId,omitempty"`
	// TenantID is the ID of the tenant that the cluster is deployed in.
	TenantID string `json:"tenantId"`
	// CloudConfigType is the cloud configure type for Azure cloud provider. Supported values are file, secret and merge.
	CloudConfigType string `json:"cloudConfigType,omitempty"`
	// VMType is the type of azure nodes.
	VMType string `json:"vmType,omitempty" yaml:"vmType,omitempty"`
	// ResourceGroup is the name of the resource group that the cluster is deployed in.
	ResourceGroup string `json:"resourceGroup,omitempty"`
	// Location is the location of the resource group that the cluster is deployed in.
	Location string `json:"location,omitempty"`
	// RouteTableName is the name of the route table attached to the subnet that the cluster is deployed in.
	RouteTableName string `json:"routeTableName,omitempty"`
	// VnetName is the name of the virtual network that the cluster is deployed in.
	VnetName string `json:"vnetName"`

	// UseInstanceMetadata specifies where instance metadata service is used where possible.
	UseInstanceMetadata bool `json:"useInstanceMetadata,omitempty"`
	// UseManagedIdentityExtension specifies where managed service
	// identity is used for the virtual machine to access Azure
	// ARM APIs.
	UseManagedIdentityExtension bool `json:"useManagedIdentityExtension,omitempty"`
	// DisableAvailabilitySetNodes disables VMAS nodes support.
	DisableAvailabilitySetNodes bool `json:"disableAvailabilitySetNodes,omitempty"`
}

// CloudConfigBuilder creates the cloud configuration file
type CloudConfigBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &CloudConfigBuilder{}

func (b *CloudConfigBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.HasAPIServer && b.NodeupConfig.KubeletConfig.CloudProvider == "external" {
		return nil
	}

	if err := b.build(c, true); err != nil {
		return err
	}
	if err := b.build(c, false); err != nil {
		return err
	}
	return nil
}

func (b *CloudConfigBuilder) build(c *fi.NodeupModelBuilderContext, inTree bool) error {
	// Add cloud config file if needed
	var lines []string

	cloudProvider := b.BootConfig.CloudProvider

	var config string
	requireGlobal := true
	switch cloudProvider {
	case kops.CloudProviderGCE:
		if b.NodeupConfig.NodeTags != nil {
			lines = append(lines, "node-tags = "+*b.NodeupConfig.NodeTags)
		}
		if b.NodeupConfig.NodeInstancePrefix != nil {
			lines = append(lines, "node-instance-prefix = "+*b.NodeupConfig.NodeInstancePrefix)
		}
		if b.NodeupConfig.Multizone != nil {
			lines = append(lines, fmt.Sprintf("multizone = %t", *b.NodeupConfig.Multizone))
		}
	case kops.CloudProviderAWS:
		if b.NodeupConfig.DisableSecurityGroupIngress != nil {
			lines = append(lines, fmt.Sprintf("DisableSecurityGroupIngress = %t", *b.NodeupConfig.DisableSecurityGroupIngress))
		}
		if b.NodeupConfig.ElbSecurityGroup != nil {
			lines = append(lines, "ElbSecurityGroup = "+*b.NodeupConfig.ElbSecurityGroup)
		}
		if !inTree {
			for _, family := range b.NodeupConfig.NodeIPFamilies {
				lines = append(lines, "NodeIPFamilies = "+family)
			}
		}
	case kops.CloudProviderOpenstack:
		osc := b.Cluster.Spec.CloudProvider.Openstack
		if osc == nil {
			break
		}

		lines = append(lines, openstack.MakeCloudConfig(b.Cluster.Spec)...)

	case kops.CloudProviderAzure:
		requireGlobal = false

		var region string
		for _, subnet := range b.Cluster.Spec.Networking.Subnets {
			if subnet.Region != "" {
				region = subnet.Region
				break
			}
		}
		if region == "" {
			return fmt.Errorf("on Azure, subnets must include Regions")
		}

		vnetName := b.Cluster.Spec.Networking.NetworkID
		if vnetName == "" {
			vnetName = b.NodeupConfig.ClusterName
		}

		az := b.Cluster.Spec.CloudProvider.Azure
		c := &azureCloudConfig{
			CloudConfigType:             "file",
			SubscriptionID:              az.SubscriptionID,
			TenantID:                    az.TenantID,
			Location:                    region,
			VMType:                      "vmss",
			ResourceGroup:               b.Cluster.AzureResourceGroupName(),
			RouteTableName:              az.RouteTableName,
			VnetName:                    vnetName,
			UseInstanceMetadata:         true,
			UseManagedIdentityExtension: true,
			// Disable availability set nodes as we currently use VMSS.
			DisableAvailabilitySetNodes: true,
		}
		data, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("error marshalling azure config: %s", err)
		}
		config = string(data)
	}

	if requireGlobal {
		config = "[global]\n" + strings.Join(lines, "\n") + "\n"
	}
	path := CloudConfigFilePath
	if inTree {
		path = InTreeCloudConfigFilePath
	}
	t := &nodetasks.File{
		Path:     path,
		Contents: fi.NewStringResource(config),
		Type:     nodetasks.FileType_File,
	}
	c.AddTask(t)

	return nil
}
