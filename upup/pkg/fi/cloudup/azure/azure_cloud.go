/*
Copyright 2024 The Kubernetes Authors.

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

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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
	TagRoleControlPlane      = "control_plane"
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
	NetworkSecurityGroup() NetworkSecurityGroupsClient
	ApplicationSecurityGroup() ApplicationSecurityGroupsClient
	VMScaleSet() VMScaleSetsClient
	VMScaleSetVM() VMScaleSetVMsClient
	Disk() DisksClient
	RoleAssignment() RoleAssignmentsClient
	NetworkInterface() NetworkInterfacesClient
	LoadBalancer() LoadBalancersClient
	PublicIPAddress() PublicIPAddressesClient
	NatGateway() NatGatewaysClient
}

type azureCloudImplementation struct {
	subscriptionID                  string
	location                        string
	tags                            map[string]string
	resourceGroupsClient            ResourceGroupsClient
	networkSecurityGroupsClient     NetworkSecurityGroupsClient
	applicationSecurityGroupsClient ApplicationSecurityGroupsClient
	vnetsClient                     VirtualNetworksClient
	subnetsClient                   SubnetsClient
	routeTablesClient               RouteTablesClient
	vmscaleSetsClient               VMScaleSetsClient
	vmscaleSetVMsClient             VMScaleSetVMsClient
	disksClient                     DisksClient
	roleAssignmentsClient           RoleAssignmentsClient
	networkInterfacesClient         NetworkInterfacesClient
	loadBalancersClient             LoadBalancersClient
	publicIPAddressesClient         PublicIPAddressesClient
	natGatewaysClient               NatGatewaysClient
}

var _ fi.Cloud = &azureCloudImplementation{}

// NewAzureCloud creates a new AzureCloud.
func NewAzureCloud(subscriptionID, location string, tags map[string]string) (AzureCloud, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating an identity: %s", err)
	}

	azureCloudImpl := &azureCloudImplementation{
		subscriptionID: subscriptionID,
		location:       location,
		tags:           tags,
	}

	if azureCloudImpl.resourceGroupsClient, err = newResourceGroupsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.vnetsClient, err = newVirtualNetworksClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.subnetsClient, err = newSubnetsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.routeTablesClient, err = newRouteTablesClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.networkSecurityGroupsClient, err = newNetworkSecurityGroupsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.applicationSecurityGroupsClient, err = newApplicationSecurityGroupsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.vmscaleSetsClient, err = newVMScaleSetsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.vmscaleSetVMsClient, err = newVMScaleSetVMsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.disksClient, err = newDisksClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.roleAssignmentsClient, err = newRoleAssignmentsClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.networkInterfacesClient, err = newNetworkInterfacesClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.loadBalancersClient, err = newLoadBalancersClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.publicIPAddressesClient, err = newPublicIPAddressesClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}
	if azureCloudImpl.natGatewaysClient, err = newNatGatewaysClientImpl(subscriptionID, cred); err != nil {
		return nil, err
	}

	return azureCloudImpl, nil
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
		if vnet.Properties == nil {
			continue
		}
		subnets := make([]*fi.SubnetInfo, 0)
		for _, subnet := range vnet.Properties.Subnets {
			if subnet.Properties == nil {
				continue
			}
			subnets = append(subnets, &fi.SubnetInfo{
				ID:   *subnet.ID,
				CIDR: *subnet.Properties.AddressPrefix,
			})
		}
		if vnet.Properties.AddressSpace == nil || len(vnet.Properties.AddressSpace.AddressPrefixes) == 0 {
			continue
		}
		return &fi.VPCInfo{
			CIDR:    *vnet.Properties.AddressSpace.AddressPrefixes[0],
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
		// Get load balancers in cluster resource group
		lbs, err := c.loadBalancersClient.List(context.TODO(), rg)
		if err != nil {
			return nil, fmt.Errorf("getting Loadbalancer for API Ingress Status: %w", err)
		}

		for _, lb := range lbs {
			val := lb.Tags[TagClusterName]
			if val == nil || *val != cluster.Name {
				continue
			}
			if lb.Properties == nil {
				continue
			}
			for _, i := range lb.Properties.FrontendIPConfigurations {
				if i.Properties == nil {
					continue
				}
				switch lbSpec.Type {
				case kops.LoadBalancerTypeInternal:
					if i.Properties.PrivateIPAddress == nil {
						continue
					}
					ingresses = append(ingresses, fi.ApiIngressStatus{
						IP: *i.Properties.PrivateIPAddress,
					})
				case kops.LoadBalancerTypePublic:
					if i.Properties.PublicIPAddress == nil || i.Properties.PublicIPAddress.ID == nil {
						continue
					}
					pips, err := c.publicIPAddressesClient.List(context.TODO(), rg)
					if err != nil {
						return nil, fmt.Errorf("error getting PublicIPAddress for API Ingress Status: %w", err)
					}
					for _, pip := range pips {
						if pip.ID == nil || pip.Properties == nil || pip.Properties.IPAddress == nil || *pip.ID != *i.Properties.PublicIPAddress.ID {
							continue
						}
						ingresses = append(ingresses, fi.ApiIngressStatus{
							IP: *pip.Properties.IPAddress,
						})
					}
				default:
					return nil, fmt.Errorf("unknown load balancer type: %q", lbSpec.Type)
				}
			}
		}
	} else {
		// Get scale sets in cluster resource group and find masters scale set
		scaleSets, err := c.vmscaleSetsClient.List(context.TODO(), rg)
		if err != nil {
			return nil, fmt.Errorf("getting cluster control plane VMSS for API ingress status: %w", err)
		}
		var vmssName string
		for _, scaleSet := range scaleSets {
			val, ok := scaleSet.Tags[TagClusterName]
			val2, ok2 := scaleSet.Tags[TagNameRolePrefix+TagRoleControlPlane]
			val3, ok3 := scaleSet.Tags[TagNameRolePrefix+TagRoleMaster]
			if ok && *val == cluster.Name && (ok2 && *val2 == "1" || ok3 && *val3 == "1") {
				vmssName = *scaleSet.Name
				break
			}
		}
		if vmssName == "" {
			return nil, fmt.Errorf("getting control plane VMSS name for API ingress status")
		}

		// Get masters scale set network interfaces and append to api ingress status
		nis, err := c.NetworkInterface().ListScaleSetsNetworkInterfaces(context.TODO(), rg, vmssName)
		if err != nil {
			return nil, fmt.Errorf("getting control plane VMSS network interfaces for API ingress status: %w", err)
		}
		for _, ni := range nis {
			if ni.Properties == nil || ni.Properties.Primary == nil || !*ni.Properties.Primary {
				continue
			}
			for _, i := range ni.Properties.IPConfigurations {
				if i.Properties == nil || i.Properties.PrivateIPAddress == nil {
					continue
				}
				ingresses = append(ingresses, fi.ApiIngressStatus{
					IP: *i.Properties.PrivateIPAddress,
				})
			}
		}
		if ingresses == nil {
			return nil, fmt.Errorf("getting API ingress status")
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

func (c *azureCloudImplementation) NetworkSecurityGroup() NetworkSecurityGroupsClient {
	return c.networkSecurityGroupsClient
}

func (c *azureCloudImplementation) ApplicationSecurityGroup() ApplicationSecurityGroupsClient {
	return c.applicationSecurityGroupsClient
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

func (c *azureCloudImplementation) NatGateway() NatGatewaysClient {
	return c.natGatewaysClient
}
