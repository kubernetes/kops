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

package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagClusterName = "KubernetesCluster"
	// Use dash (_) as a splitter. Other CSPs use slash (/), but slash is not
	// allowed as a tag key in Azure.
	TagNameRolePrefix        = "k8s.io_role_"
	TagRoleMaster            = "master"
	TagNameEtcdClusterPrefix = "k8s.io_etcd_"
)

// AzureCloud provides clients to make API calls to Azure.
type AzureCloud interface {
	fi.Cloud
	AddClusterTags(tags map[string]*string)
	FindVNetInfo(id, resourceGroup string) (*fi.VPCInfo, error)
	SubscriptionID() string
	ResourceGroup() ResourceGroupsClient
	VirtualNetwork() VirtualNetworksClient
	Subnet() SubnetsClient
	RouteTable() RouteTablesClient
	VMScaleSet() VMScaleSetsClient
	VMScaleSetVM() VMScaleSetVMsClient
	Disk() DisksClient
	RoleAssignment() RoleAssignmentsClient
	NetworkInterface() NetworkInterfacesClient
	LoadBalancer() LoadBalancersClient
	PublicIPAddress() PublicIPAddressesClient
}

type azureCloudImplementation struct {
	subscriptionID          string
	location                string
	tags                    map[string]string
	resourceGroupsClient    ResourceGroupsClient
	vnetsClient             VirtualNetworksClient
	subnetsClient           SubnetsClient
	routeTablesClient       RouteTablesClient
	vmscaleSetsClient       VMScaleSetsClient
	vmscaleSetVMsClient     VMScaleSetVMsClient
	disksClient             DisksClient
	roleAssignmentsClient   RoleAssignmentsClient
	networkInterfacesClient NetworkInterfacesClient
	loadBalancersClient     LoadBalancersClient
	publicIPAddressesClient PublicIPAddressesClient
}

var _ fi.Cloud = &azureCloudImplementation{}

// NewAzureCloud creates a new AzureCloud.
func NewAzureCloud(subscriptionID, location string, tags map[string]string) (AzureCloud, error) {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &azureCloudImplementation{
		subscriptionID:          subscriptionID,
		location:                location,
		tags:                    tags,
		resourceGroupsClient:    newResourceGroupsClientImpl(subscriptionID, authorizer),
		vnetsClient:             newVirtualNetworksClientImpl(subscriptionID, authorizer),
		subnetsClient:           newSubnetsClientImpl(subscriptionID, authorizer),
		routeTablesClient:       newRouteTablesClientImpl(subscriptionID, authorizer),
		vmscaleSetsClient:       newVMScaleSetsClientImpl(subscriptionID, authorizer),
		vmscaleSetVMsClient:     newVMScaleSetVMsClientImpl(subscriptionID, authorizer),
		disksClient:             newDisksClientImpl(subscriptionID, authorizer),
		roleAssignmentsClient:   newRoleAssignmentsClientImpl(subscriptionID, authorizer),
		networkInterfacesClient: newNetworkInterfacesClientImpl(subscriptionID, authorizer),
		loadBalancersClient:     newLoadBalancersClientImpl(subscriptionID, authorizer),
		publicIPAddressesClient: newPublicIPAddressesClientImpl(subscriptionID, authorizer),
	}, nil
}

func (c *azureCloudImplementation) Region() string {
	return c.location
}

func (c *azureCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAzure
}

func (c *azureCloudImplementation) DNS() (dnsprovider.Interface, error) {
	return nil, errors.New("DNS not implemented on azureCloud")
}

func (c *azureCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("FindVPCInfo not implemented on azureCloud, use FindVNETInfo instead")
}

func (c *azureCloudImplementation) FindVNetInfo(id, resourceGroup string) (*fi.VPCInfo, error) {
	vnets, err := c.vnetsClient.List(context.TODO(), resourceGroup)
	if err != nil {
		return nil, err
	}
	for _, vnet := range vnets {
		if *vnet.ID != id {
			continue
		}
		subnets := make([]*fi.SubnetInfo, 0)
		for _, subnet := range *vnet.Subnets {
			subnets = append(subnets, &fi.SubnetInfo{
				ID:   *subnet.ID,
				CIDR: *subnet.AddressPrefix,
			})
		}
		return &fi.VPCInfo{
			CIDR:    (*vnet.AddressSpace.AddressPrefixes)[0],
			Subnets: subnets,
		}, nil
	}
	return nil, nil
}

func (c *azureCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	return errors.New("DeleteInstance not implemented on azureCloud")
}

// DeregisterInstance drains a cloud instance and loadbalancers.
func (c *azureCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Info("Azure DeregisterInstance not implemented")
	return nil
}

func (c *azureCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return errors.New("DeleteGroup not implemented on azureCloud")
}

func (c *azureCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	return errors.New("DetachInstance not implemented on azureCloud")
}

// AddClusterTags adds cluster tags to the resource.
func (c *azureCloudImplementation) AddClusterTags(tags map[string]*string) {
	for k, v := range c.tags {
		tags[k] = &v
	}
}

func (c *azureCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus
	rg := cluster.AzureResourceGroupName()

	lbSpec := cluster.Spec.API.LoadBalancer
	if lbSpec != nil {
		// Get loadbalancers in cluster resource group
		lbs, err := c.loadBalancersClient.List(context.TODO(), rg)
		if err != nil {
			return nil, fmt.Errorf("error getting Loadbalancer for API Ingress Status: %s", err)
		}

		for _, lb := range lbs {
			val := lb.Tags[TagClusterName]
			if val == nil || *val != cluster.Name {
				continue
			}
			if lb.LoadBalancerPropertiesFormat == nil {
				continue
			}
			for _, i := range *lb.LoadBalancerPropertiesFormat.FrontendIPConfigurations {
				if i.FrontendIPConfigurationPropertiesFormat == nil {
					continue
				}
				switch lbSpec.Type {
				case kops.LoadBalancerTypeInternal:
					if i.FrontendIPConfigurationPropertiesFormat.PrivateIPAddress == nil {
						continue
					}
					ingresses = append(ingresses, fi.ApiIngressStatus{
						IP: *i.FrontendIPConfigurationPropertiesFormat.PrivateIPAddress,
					})
				case kops.LoadBalancerTypePublic:
					if i.FrontendIPConfigurationPropertiesFormat.PublicIPAddress == nil ||
						i.FrontendIPConfigurationPropertiesFormat.PublicIPAddress.PublicIPAddressPropertiesFormat == nil ||
						i.FrontendIPConfigurationPropertiesFormat.PublicIPAddress.PublicIPAddressPropertiesFormat.IPAddress == nil {
						continue
					}
					ingresses = append(ingresses, fi.ApiIngressStatus{
						IP: *i.FrontendIPConfigurationPropertiesFormat.PublicIPAddress.PublicIPAddressPropertiesFormat.IPAddress,
					})
				default:
					return nil, fmt.Errorf("unknown load balancer Type: %q", lbSpec.Type)
				}
			}
		}
	} else {
		// Get scale sets in cluster resource group and find masters scale set
		scaleSets, err := c.vmscaleSetsClient.List(context.TODO(), rg)
		if err != nil {
			return nil, fmt.Errorf("error getting Cluster Master Scale Set for API Ingress Status: %s", err)
		}
		var vmssName string
		for _, scaleSet := range scaleSets {
			val, ok := scaleSet.Tags[TagClusterName]
			val2, ok2 := scaleSet.Tags[TagNameRolePrefix+TagRoleMaster]
			if ok && *val == cluster.Name && ok2 && *val2 == "1" {
				vmssName = *scaleSet.Name
				break
			}
		}
		if vmssName == "" {
			return nil, fmt.Errorf("error getting Master Scale Set Name for API Ingress Status")
		}

		// Get masters scale set network interfaces and append to api ingress status
		nis, err := c.NetworkInterface().ListScaleSetsNetworkInterfaces(context.TODO(), rg, vmssName)
		if err != nil {
			return nil, fmt.Errorf("error getting Master Scale Set Network Interfaces for API Ingress Status: %s", err)
		}
		for _, ni := range nis {
			if ni.Primary == nil || !*ni.Primary {
				continue
			}
			for _, i := range *ni.IPConfigurations {
				ingresses = append(ingresses, fi.ApiIngressStatus{
					IP: *i.PrivateIPAddress,
				})
			}
		}
		if ingresses == nil {
			return nil, fmt.Errorf("error getting API Ingress Status so make sure to update your kubecfg accordingly")
		}
	}

	return ingresses, nil
}

func (c *azureCloudImplementation) SubscriptionID() string {
	return c.subscriptionID
}

func (c *azureCloudImplementation) ResourceGroup() ResourceGroupsClient {
	return c.resourceGroupsClient
}

func (c *azureCloudImplementation) VirtualNetwork() VirtualNetworksClient {
	return c.vnetsClient
}

func (c *azureCloudImplementation) Subnet() SubnetsClient {
	return c.subnetsClient
}

func (c *azureCloudImplementation) RouteTable() RouteTablesClient {
	return c.routeTablesClient
}

func (c *azureCloudImplementation) VMScaleSet() VMScaleSetsClient {
	return c.vmscaleSetsClient
}

func (c *azureCloudImplementation) VMScaleSetVM() VMScaleSetVMsClient {
	return c.vmscaleSetVMsClient
}

func (c *azureCloudImplementation) Disk() DisksClient {
	return c.disksClient
}

func (c *azureCloudImplementation) RoleAssignment() RoleAssignmentsClient {
	return c.roleAssignmentsClient
}

func (c *azureCloudImplementation) NetworkInterface() NetworkInterfacesClient {
	return c.networkInterfacesClient
}

func (c *azureCloudImplementation) LoadBalancer() LoadBalancersClient {
	return c.loadBalancersClient
}

func (c *azureCloudImplementation) PublicIPAddress() PublicIPAddressesClient {
	return c.publicIPAddressesClient
}
