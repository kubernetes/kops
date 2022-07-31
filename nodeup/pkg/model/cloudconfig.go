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
	"os"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
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

var _ fi.ModelBuilder = &CloudConfigBuilder{}

func (b *CloudConfigBuilder) Build(c *fi.ModelBuilderContext) error {
	if err := b.build(c, true); err != nil {
		return err
	}
	if err := b.build(c, false); err != nil {
		return err
	}
	return nil
}

func (b *CloudConfigBuilder) build(c *fi.ModelBuilderContext, inTree bool) error {
	// Add cloud config file if needed
	var lines []string

	cloudProvider := b.CloudProvider
	cloudConfig := b.Cluster.Spec.CloudConfig

	if cloudConfig == nil {
		cloudConfig = &kops.CloudConfiguration{}
	}

	var config string
	requireGlobal := true
	switch cloudProvider {
	case "gce":
		if cloudConfig.NodeTags != nil {
			lines = append(lines, "node-tags = "+*cloudConfig.NodeTags)
		}
		if cloudConfig.NodeInstancePrefix != nil {
			lines = append(lines, "node-instance-prefix = "+*cloudConfig.NodeInstancePrefix)
		}
		if cloudConfig.Multizone != nil {
			lines = append(lines, fmt.Sprintf("multizone = %t", *cloudConfig.Multizone))
		}
	case "aws":
		if cloudConfig.DisableSecurityGroupIngress != nil {
			lines = append(lines, fmt.Sprintf("DisableSecurityGroupIngress = %t", *cloudConfig.DisableSecurityGroupIngress))
		}
		if cloudConfig.ElbSecurityGroup != nil {
			lines = append(lines, "ElbSecurityGroup = "+*cloudConfig.ElbSecurityGroup)
		}
		if !inTree {
			for _, family := range cloudConfig.NodeIPFamilies {
				lines = append(lines, "NodeIPFamilies = "+family)
			}
		}
	case "openstack":
		osc := b.Cluster.Spec.CloudProvider.Openstack
		if osc == nil {
			break
		}
		// Support mapping of older keystone API
		tenantName := os.Getenv("OS_TENANT_NAME")
		if tenantName == "" {
			tenantName = os.Getenv("OS_PROJECT_NAME")
		}
		tenantID := os.Getenv("OS_TENANT_ID")
		if tenantID == "" {
			tenantID = os.Getenv("OS_PROJECT_ID")
		}
		lines = append(lines,
			fmt.Sprintf("auth-url=\"%s\"", os.Getenv("OS_AUTH_URL")),
			fmt.Sprintf("username=\"%s\"", os.Getenv("OS_USERNAME")),
			fmt.Sprintf("password=\"%s\"", os.Getenv("OS_PASSWORD")),
			fmt.Sprintf("region=\"%s\"", os.Getenv("OS_REGION_NAME")),
			fmt.Sprintf("tenant-id=\"%s\"", tenantID),
			fmt.Sprintf("tenant-name=\"%s\"", tenantName),
			fmt.Sprintf("domain-name=\"%s\"", os.Getenv("OS_DOMAIN_NAME")),
			fmt.Sprintf("domain-id=\"%s\"", os.Getenv("OS_DOMAIN_ID")),
		)
		if b.Cluster.Spec.ExternalCloudControllerManager != nil {
			lines = append(lines,
				fmt.Sprintf("application-credential-id=\"%s\"", os.Getenv("OS_APPLICATION_CREDENTIAL_ID")),
				fmt.Sprintf("application-credential-secret=\"%s\"", os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET")),
			)
		}

		lines = append(lines,
			"",
		)

		if lb := osc.Loadbalancer; lb != nil {
			ingressHostnameSuffix := "nip.io"
			if fi.StringValue(lb.IngressHostnameSuffix) != "" {
				ingressHostnameSuffix = fi.StringValue(lb.IngressHostnameSuffix)
			}

			lines = append(lines,
				"[LoadBalancer]",
				fmt.Sprintf("floating-network-id=%s", fi.StringValue(lb.FloatingNetworkID)),
				fmt.Sprintf("lb-method=%s", fi.StringValue(lb.Method)),
				fmt.Sprintf("lb-provider=%s", fi.StringValue(lb.Provider)),
				fmt.Sprintf("use-octavia=%t", fi.BoolValue(lb.UseOctavia)),
				fmt.Sprintf("manage-security-groups=%t", fi.BoolValue(lb.ManageSecGroups)),
				fmt.Sprintf("enable-ingress-hostname=%t", fi.BoolValue(lb.EnableIngressHostname)),
				fmt.Sprintf("ingress-hostname-suffix=%s", ingressHostnameSuffix),
				"",
			)

			if monitor := osc.Monitor; monitor != nil {
				lines = append(lines,
					"create-monitor=yes",
					fmt.Sprintf("monitor-delay=%s", fi.StringValue(monitor.Delay)),
					fmt.Sprintf("monitor-timeout=%s", fi.StringValue(monitor.Timeout)),
					fmt.Sprintf("monitor-max-retries=%d", fi.IntValue(monitor.MaxRetries)),
					"",
				)
			}
		}

		if bs := osc.BlockStorage; bs != nil {
			// Block Storage Config
			lines = append(lines,
				"[BlockStorage]",
				fmt.Sprintf("bs-version=%s", fi.StringValue(bs.Version)),
				fmt.Sprintf("ignore-volume-az=%t", fi.BoolValue(bs.IgnoreAZ)),
				"")
		}

		if networking := osc.Network; networking != nil {
			// Networking Config
			// https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/openstack-cloud-controller-manager/using-openstack-cloud-controller-manager.md#networking
			var networkingLines []string

			if networking.IPv6SupportDisabled != nil {
				networkingLines = append(networkingLines, fmt.Sprintf("ipv6-support-disabled=%t", fi.BoolValue(networking.IPv6SupportDisabled)))
			}
			for _, name := range networking.PublicNetworkNames {
				networkingLines = append(networkingLines, fmt.Sprintf("public-network-name=%s", fi.StringValue(name)))
			}
			for _, name := range networking.InternalNetworkNames {
				networkingLines = append(networkingLines, fmt.Sprintf("internal-network-name=%s", fi.StringValue(name)))
			}

			if len(networkingLines) > 0 {
				lines = append(lines, "[Networking]")
				lines = append(lines, networkingLines...)
				lines = append(lines, "")
			}
		}
	case "azure":
		requireGlobal = false

		var region string
		for _, subnet := range b.Cluster.Spec.Subnets {
			if subnet.Region != "" {
				region = subnet.Region
				break
			}
		}
		if region == "" {
			return fmt.Errorf("on Azure, subnets must include Regions")
		}

		vnetName := b.Cluster.Spec.NetworkID
		if vnetName == "" {
			vnetName = b.Cluster.Name
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
