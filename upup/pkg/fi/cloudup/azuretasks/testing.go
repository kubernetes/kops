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

package azuretasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	resources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

const (
	testClusterName = "test-cluster"
)

// MockAzureCloud is a mock implementation of AzureCloud.
type MockAzureCloud struct {
	Location                        string
	ResourceGroupsClient            *MockResourceGroupsClient
	VirtualNetworksClient           *MockVirtualNetworksClient
	SubnetsClient                   *MockSubnetsClient
	RouteTablesClient               *MockRouteTablesClient
	NetworkSecurityGroupsClient     *MockNetworkSecurityGroupsClient
	ApplicationSecurityGroupsClient *MockApplicationSecurityGroupsClient
	VMScaleSetsClient               *MockVMScaleSetsClient
	VMScaleSetVMsClient             *MockVMScaleSetVMsClient
	DisksClient                     *MockDisksClient
	RoleAssignmentsClient           *MockRoleAssignmentsClient
	NetworkInterfacesClient         *MockNetworkInterfacesClient
	LoadBalancersClient             *MockLoadBalancersClient
	PublicIPAddressesClient         *MockPublicIPAddressesClient
	NatGatewaysClient               *MockNatGatewaysClient
}

var _ azure.AzureCloud = &MockAzureCloud{}

// NewMockAzureCloud returns a new MockAzureCloud.
func NewMockAzureCloud(location string) *MockAzureCloud {
	return &MockAzureCloud{
		Location: location,
		ResourceGroupsClient: &MockResourceGroupsClient{
			RGs: map[string]*resources.ResourceGroup{},
		},
		VirtualNetworksClient: &MockVirtualNetworksClient{
			VNets: map[string]*network.VirtualNetwork{},
		},
		SubnetsClient: &MockSubnetsClient{
			Subnets: map[string]*network.Subnet{},
		},
		RouteTablesClient: &MockRouteTablesClient{
			RTs: map[string]*network.RouteTable{},
		},
		NetworkSecurityGroupsClient: &MockNetworkSecurityGroupsClient{
			NSGs: map[string]*network.SecurityGroup{},
		},
		ApplicationSecurityGroupsClient: &MockApplicationSecurityGroupsClient{
			ASGs: map[string]*network.ApplicationSecurityGroup{},
		},
		VMScaleSetsClient: &MockVMScaleSetsClient{
			VMSSes: map[string]*compute.VirtualMachineScaleSet{},
		},
		VMScaleSetVMsClient: &MockVMScaleSetVMsClient{
			VMs: map[string]*compute.VirtualMachineScaleSetVM{},
		},
		DisksClient: &MockDisksClient{
			Disks: map[string]*compute.Disk{},
		},
		RoleAssignmentsClient: &MockRoleAssignmentsClient{
			RAs: map[string]*authz.RoleAssignment{},
		},
		NetworkInterfacesClient: &MockNetworkInterfacesClient{
			NIs: map[string]*network.Interface{},
		},
		LoadBalancersClient: &MockLoadBalancersClient{
			LBs: map[string]*network.LoadBalancer{},
		},
		PublicIPAddressesClient: &MockPublicIPAddressesClient{
			PubIPs: map[string]*network.PublicIPAddress{},
		},
		NatGatewaysClient: &MockNatGatewaysClient{
			NGWs: map[string]*network.NatGateway{},
		},
	}
}

// Region returns the region.
func (c *MockAzureCloud) Region() string {
	return c.Location
}

// ProviderID returns the provider ID.
func (c *MockAzureCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAzure
}

// DNS returns the DNS provider.
func (c *MockAzureCloud) DNS() (dnsprovider.Interface, error) {
	return nil, errors.New("DNS not implemented on azureCloud")
}

// FindVPCInfo returns the VPCInfo.
func (c *MockAzureCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("FindVPCInfo not implemented on azureCloud")
}

func (c *MockAzureCloud) FindVNetInfo(id, resourceGroup string) (*fi.VPCInfo, error) {
	return nil, errors.New("FindVNetInfo not implemented on azureCloud")
}

// DeleteInstance deletes the instance.
func (c *MockAzureCloud) DeleteInstance(i *cloudinstances.CloudInstance) error {
	return errors.New("DeleteInstance not implemented on azureCloud")
}

func (c *MockAzureCloud) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	return nil
}

// DeleteGroup deletes the group.
func (c *MockAzureCloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return errors.New("DeleteGroup not implemented on azureCloud")
}

// DetachInstance detaches the instance.
func (c *MockAzureCloud) DetachInstance(i *cloudinstances.CloudInstance) error {
	return errors.New("DetachInstance not implemented on azureCloud")
}

// GetCloudGroups returns cloud instance groups.
func (c *MockAzureCloud) GetCloudGroups(
	cluster *kops.Cluster,
	instancegroups []*kops.InstanceGroup,
	warnUnmatched bool,
	nodes []v1.Node,
) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, errors.New("GetCloudGroups not implemented on azureCloud")
}

// AddClusterTags add the cluster tag to the given tag map.
func (c *MockAzureCloud) AddClusterTags(tags map[string]*string) {
	tags[azure.TagClusterName] = to.Ptr(testClusterName)
}

// FindClusterStatus discovers the status of the cluster, by looking for the tagged etcd volumes
func (c *MockAzureCloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return &kops.ClusterStatus{}, nil
}

// GetApiIngressStatus returns the status of API ingress.
func (c *MockAzureCloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, nil
}

// SubscriptionID returns the subscription ID.
func (c *MockAzureCloud) SubscriptionID() string {
	return ""
}

// ResourceGroup returns the resource group client.
func (c *MockAzureCloud) ResourceGroup() azure.ResourceGroupsClient {
	return c.ResourceGroupsClient
}

// VirtualNetwork returns the virtual network client.
func (c *MockAzureCloud) VirtualNetwork() azure.VirtualNetworksClient {
	return c.VirtualNetworksClient
}

// Subnet returns the subnet client.
func (c *MockAzureCloud) Subnet() azure.SubnetsClient {
	return c.SubnetsClient
}

// RouteTable returns the route table client.
func (c *MockAzureCloud) RouteTable() azure.RouteTablesClient {
	return c.RouteTablesClient
}

// NetworkSecurityGroup returns the Network Security Group client.
func (c *MockAzureCloud) NetworkSecurityGroup() azure.NetworkSecurityGroupsClient {
	return c.NetworkSecurityGroupsClient
}

// ApplicationSecurityGroup returns the Application Security Group client.
func (c *MockAzureCloud) ApplicationSecurityGroup() azure.ApplicationSecurityGroupsClient {
	return c.ApplicationSecurityGroupsClient
}

// VMScaleSet returns the VM Scale Set client.
func (c *MockAzureCloud) VMScaleSet() azure.VMScaleSetsClient {
	return c.VMScaleSetsClient
}

// VMScaleSetVM returns the VM Scale Set VM client.
func (c *MockAzureCloud) VMScaleSetVM() azure.VMScaleSetVMsClient {
	return c.VMScaleSetVMsClient
}

// Disk returns the disk client.
func (c *MockAzureCloud) Disk() azure.DisksClient {
	return c.DisksClient
}

// RoleAssignment returns the role assignment client.
func (c *MockAzureCloud) RoleAssignment() azure.RoleAssignmentsClient {
	return c.RoleAssignmentsClient
}

// NetworkInterface returns the network interface client.
func (c *MockAzureCloud) NetworkInterface() azure.NetworkInterfacesClient {
	return c.NetworkInterfacesClient
}

// LoadBalancer returns the loadbalancer client.
func (c *MockAzureCloud) LoadBalancer() azure.LoadBalancersClient {
	return c.LoadBalancersClient
}

// PublicIPAddress returns the public ip address client.
func (c *MockAzureCloud) PublicIPAddress() azure.PublicIPAddressesClient {
	return c.PublicIPAddressesClient
}

// NatGateway returns the nat gateway client.
func (c *MockAzureCloud) NatGateway() azure.NatGatewaysClient {
	return c.NatGatewaysClient
}

// MockResourceGroupsClient is a mock implementation of resource group client.
type MockResourceGroupsClient struct {
	RGs map[string]*resources.ResourceGroup
}

var _ azure.ResourceGroupsClient = &MockResourceGroupsClient{}

// CreateOrUpdate creates or updates a resource group.
func (c *MockResourceGroupsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, parameters resources.ResourceGroup) error {
	parameters.Name = &resourceGroupName
	parameters.ID = &resourceGroupName
	c.RGs[resourceGroupName] = &parameters
	return nil
}

// List returns a slice of resource groups.
func (c *MockResourceGroupsClient) List(ctx context.Context) ([]*resources.ResourceGroup, error) {
	var l []*resources.ResourceGroup
	for _, rg := range c.RGs {
		l = append(l, rg)
	}
	return l, nil
}

// Delete deletes a specified resource group.
func (c *MockResourceGroupsClient) Delete(ctx context.Context, name string) error {
	if _, ok := c.RGs[name]; !ok {
		return fmt.Errorf("%s does not exist", name)
	}
	delete(c.RGs, name)
	return nil
}

// MockVirtualNetworksClient is a mock implementation of virtual network client.
type MockVirtualNetworksClient struct {
	VNets map[string]*network.VirtualNetwork
}

var _ azure.VirtualNetworksClient = &MockVirtualNetworksClient{}

// CreateOrUpdate creates or updates a virtual network.
func (c *MockVirtualNetworksClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, virtualNetworkName string, parameters network.VirtualNetwork) (*network.VirtualNetwork, error) {
	if _, ok := c.VNets[virtualNetworkName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &virtualNetworkName
	parameters.ID = &virtualNetworkName
	c.VNets[virtualNetworkName] = &parameters
	return &parameters, nil
}

// List returns a slice of virtual networks.
func (c *MockVirtualNetworksClient) List(ctx context.Context, resourceGroupName string) ([]*network.VirtualNetwork, error) {
	var l []*network.VirtualNetwork
	for _, vnet := range c.VNets {
		l = append(l, vnet)
	}
	return l, nil
}

// Delete deletes a specified virtual network.
func (c *MockVirtualNetworksClient) Delete(ctx context.Context, resourceGroupName, vnetName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.VNets[vnetName]; !ok {
		return fmt.Errorf("%s does not exist", vnetName)
	}
	delete(c.VNets, vnetName)
	return nil
}

// MockSubnetsClient is a mock implementation of a subnet client.
type MockSubnetsClient struct {
	Subnets map[string]*network.Subnet
}

var _ azure.SubnetsClient = &MockSubnetsClient{}

// CreateOrUpdate creates or updates a subnet.
func (c *MockSubnetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string, parameters network.Subnet) (*network.Subnet, error) {
	// Ignore resourceGroupName and virtualNetworkName for simplicity.
	if _, ok := c.Subnets[subnetName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &subnetName
	parameters.ID = &subnetName
	c.Subnets[subnetName] = &parameters
	return &parameters, nil
}

// List returns a slice of subnets.
func (c *MockSubnetsClient) List(ctx context.Context, resourceGroupName, virtualNetworkName string) ([]*network.Subnet, error) {
	var l []*network.Subnet
	for _, subnet := range c.Subnets {
		l = append(l, subnet)
	}
	return l, nil
}

// Delete deletes a specified subnet.
func (c *MockSubnetsClient) Delete(ctx context.Context, resourceGroupName, vnetName, subnetName string) error {
	// Ignore resourceGroupName and virtualNetworkName for simplicity.
	if _, ok := c.Subnets[subnetName]; !ok {
		return fmt.Errorf("%s does not exist", subnetName)
	}
	delete(c.Subnets, subnetName)
	return nil
}

// MockRouteTablesClient is a mock implementation of a route table client.
type MockRouteTablesClient struct {
	RTs map[string]*network.RouteTable
}

var _ azure.RouteTablesClient = &MockRouteTablesClient{}

// CreateOrUpdate creates or updates a route table.
func (c *MockRouteTablesClient) CreateOrUpdate(ctx context.Context, resourceGroupName, routeTableName string, parameters network.RouteTable) (*network.RouteTable, error) {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.RTs[routeTableName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &routeTableName
	parameters.ID = &routeTableName
	c.RTs[routeTableName] = &parameters
	return &parameters, nil
}

// List returns a slice of route tables.
func (c *MockRouteTablesClient) List(ctx context.Context, resourceGroupName string) ([]*network.RouteTable, error) {
	var l []*network.RouteTable
	for _, rt := range c.RTs {
		l = append(l, rt)
	}
	return l, nil
}

// Delete deletes a specified routeTable.
func (c *MockRouteTablesClient) Delete(ctx context.Context, resourceGroupName, routeTableName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.RTs[routeTableName]; !ok {
		return fmt.Errorf("%s does not exist", routeTableName)
	}
	delete(c.RTs, routeTableName)
	return nil
}

// MockVMScaleSetsClient is a mock implementation of VM Scale Set client.
type MockVMScaleSetsClient struct {
	VMSSes map[string]*compute.VirtualMachineScaleSet
}

var _ azure.VMScaleSetsClient = &MockVMScaleSetsClient{}

// CreateOrUpdate creates or updates a VM Scale Set.
func (c *MockVMScaleSetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName, vmScaleSetName string, parameters compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error) {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.VMSSes[vmScaleSetName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &vmScaleSetName
	parameters.ID = &vmScaleSetName
	parameters.Identity.PrincipalID = to.Ptr(uuid.New().String())
	c.VMSSes[vmScaleSetName] = &parameters
	return &parameters, nil
}

// List returns a slice of VM Scale Sets.
func (c *MockVMScaleSetsClient) List(ctx context.Context, resourceGroupName string) ([]*compute.VirtualMachineScaleSet, error) {
	var l []*compute.VirtualMachineScaleSet
	for _, vmss := range c.VMSSes {
		l = append(l, vmss)
	}
	return l, nil
}

// Get Returns a specified VM Scale Set.
func (c *MockVMScaleSetsClient) Get(ctx context.Context, resourceGroupName string, vmssName string) (*compute.VirtualMachineScaleSet, error) {
	vmss, ok := c.VMSSes[vmssName]
	if !ok {
		return nil, nil
	}
	return vmss, nil
}

// Delete deletes a specified VM Scale Set.
func (c *MockVMScaleSetsClient) Delete(ctx context.Context, resourceGroupName, vmssName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.VMSSes[vmssName]; !ok {
		return fmt.Errorf("%s does not exist", vmssName)
	}
	delete(c.VMSSes, vmssName)
	return nil
}

// MockVMScaleSetVMsClient is a mock implementation of VM Scale Set VM client.
type MockVMScaleSetVMsClient struct {
	VMs map[string]*compute.VirtualMachineScaleSetVM
}

var _ azure.VMScaleSetVMsClient = &MockVMScaleSetVMsClient{}

// List returns a slice of VM Scale Set VMs.
func (c *MockVMScaleSetVMsClient) List(ctx context.Context, resourceGroupName, vmssName string) ([]*compute.VirtualMachineScaleSetVM, error) {
	// Ignore resourceGroupName and vmssName for simplicity.
	var l []*compute.VirtualMachineScaleSetVM
	for _, vm := range c.VMs {
		l = append(l, vm)
	}
	return l, nil
}

// MockDisksClient is a mock implementation of disk client.
type MockDisksClient struct {
	Disks map[string]*compute.Disk
}

var _ azure.DisksClient = &MockDisksClient{}

// CreateOrUpdate creates or updates a disk.
func (c *MockDisksClient) CreateOrUpdate(ctx context.Context, resourceGroupName, diskName string, parameters compute.Disk) (*compute.Disk, error) {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.Disks[diskName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &diskName
	parameters.ID = &diskName
	c.Disks[diskName] = &parameters
	return &parameters, nil
}

// List returns a slice of disks.
func (c *MockDisksClient) List(ctx context.Context, resourceGroupName string) ([]*compute.Disk, error) {
	var l []*compute.Disk
	for _, disk := range c.Disks {
		l = append(l, disk)
	}
	return l, nil
}

// Delete deletes a specified disk.
func (c *MockDisksClient) Delete(ctx context.Context, resourceGroupName, diskName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.Disks[diskName]; !ok {
		return fmt.Errorf("%s does not exist", diskName)
	}
	delete(c.Disks, diskName)
	return nil
}

// MockRoleAssignmentsClient is a mock implementation of role assignment client.
type MockRoleAssignmentsClient struct {
	RAs map[string]*authz.RoleAssignment
}

var _ azure.RoleAssignmentsClient = &MockRoleAssignmentsClient{}

// Create creates a new role assignment.
func (c *MockRoleAssignmentsClient) Create(
	ctx context.Context,
	scope string,
	roleAssignmentName string,
	parameters authz.RoleAssignmentCreateParameters,
) (*authz.RoleAssignment, error) {
	if _, ok := c.RAs[roleAssignmentName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	ra := &authz.RoleAssignment{
		ID:   to.Ptr(roleAssignmentName),
		Name: to.Ptr(roleAssignmentName),
		Properties: &authz.RoleAssignmentProperties{
			Scope:            to.Ptr(scope),
			RoleDefinitionID: parameters.Properties.RoleDefinitionID,
			PrincipalID:      parameters.Properties.PrincipalID,
		},
	}
	c.RAs[roleAssignmentName] = ra
	return ra, nil
}

// List returns a slice of role assignments.
func (c *MockRoleAssignmentsClient) List(ctx context.Context, resourceGroupName string) ([]*authz.RoleAssignment, error) {
	var l []*authz.RoleAssignment
	for _, ra := range c.RAs {
		l = append(l, ra)
	}
	return l, nil
}

// Delete deletes a specified role assignment.
func (c *MockRoleAssignmentsClient) Delete(ctx context.Context, scope, raName string) error {
	// Ignore scope for simplicity.
	if _, ok := c.RAs[raName]; !ok {
		return fmt.Errorf("%s does not exist", raName)
	}
	delete(c.RAs, raName)
	return nil
}

// MockNetworkInterfacesClient is a mock implementation of network interfaces client.
type MockNetworkInterfacesClient struct {
	NIs map[string]*network.Interface
}

var _ azure.NetworkInterfacesClient = &MockNetworkInterfacesClient{}

// List returns a slice of VM Scale Set Network Interfaces.
func (c *MockNetworkInterfacesClient) ListScaleSetsNetworkInterfaces(ctx context.Context, resourceGroupName, vmssName string) ([]*network.Interface, error) {
	// Ignore resourceGroupName and vmssName for simplicity.
	var l []*network.Interface
	for _, ni := range c.NIs {
		l = append(l, ni)
	}
	return l, nil
}

// MockLoadBalancersClient is a mock implementation of role assignment client.
type MockLoadBalancersClient struct {
	LBs map[string]*network.LoadBalancer
}

var _ azure.LoadBalancersClient = &MockLoadBalancersClient{}

// CreateOrUpdate creates a new loadbalancer.
func (c *MockLoadBalancersClient) CreateOrUpdate(ctx context.Context, resourceGroupName, loadBalancerName string, parameters network.LoadBalancer) (*network.LoadBalancer, error) {
	if _, ok := c.LBs[loadBalancerName]; ok {
		return nil, nil
	}
	parameters.Name = &loadBalancerName
	parameters.ID = &loadBalancerName
	c.LBs[loadBalancerName] = &parameters
	return &parameters, nil
}

// List returns a slice of loadbalancer.
func (c *MockLoadBalancersClient) List(ctx context.Context, resourceGroupName string) ([]*network.LoadBalancer, error) {
	var l []*network.LoadBalancer
	for _, lb := range c.LBs {
		l = append(l, lb)
	}
	return l, nil
}

// Get returns a loadbalancer.
func (c *MockLoadBalancersClient) Get(ctx context.Context, resourceGroupName string, loadBalancerName string) (*network.LoadBalancer, error) {
	for _, lb := range c.LBs {
		if *lb.Name == loadBalancerName {
			return nil, nil
		}
	}
	return nil, nil
}

// Delete deletes a specified loadbalancer.
func (c *MockLoadBalancersClient) Delete(ctx context.Context, scope, lbName string) error {
	// Ignore scope for simplicity.
	if _, ok := c.LBs[lbName]; !ok {
		return fmt.Errorf("%s does not exist", lbName)
	}
	delete(c.LBs, lbName)
	return nil
}

// MockPublicIPAddressesClient is a mock implementation of role assignment client.
type MockPublicIPAddressesClient struct {
	PubIPs map[string]*network.PublicIPAddress
}

var _ azure.PublicIPAddressesClient = &MockPublicIPAddressesClient{}

// CreateOrUpdate creates a new public ip address.
func (c *MockPublicIPAddressesClient) CreateOrUpdate(ctx context.Context, resourceGroupName, publicIPAddressName string, parameters network.PublicIPAddress) (*network.PublicIPAddress, error) {
	if _, ok := c.PubIPs[publicIPAddressName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &publicIPAddressName
	parameters.ID = &publicIPAddressName
	c.PubIPs[publicIPAddressName] = &parameters
	return &parameters, nil
}

// List returns a slice of public ip address.
func (c *MockPublicIPAddressesClient) List(ctx context.Context, resourceGroupName string) ([]*network.PublicIPAddress, error) {
	var l []*network.PublicIPAddress
	for _, lb := range c.PubIPs {
		l = append(l, lb)
	}
	return l, nil
}

// Delete deletes a specified public ip address.
func (c *MockPublicIPAddressesClient) Delete(ctx context.Context, scope, publicIPAddressName string) error {
	// Ignore scope for simplicity.
	if _, ok := c.PubIPs[publicIPAddressName]; !ok {
		return fmt.Errorf("%s does not exist", publicIPAddressName)
	}
	delete(c.PubIPs, publicIPAddressName)
	return nil
}

// MockNetworkSecurityGroupsClient is a mock implementation of Network Security Group client.
type MockNetworkSecurityGroupsClient struct {
	NSGs map[string]*network.SecurityGroup
}

var _ azure.NetworkSecurityGroupsClient = &MockNetworkSecurityGroupsClient{}

// CreateOrUpdate creates or updates a Network Security Group.
func (c *MockNetworkSecurityGroupsClient) CreateOrUpdate(ctx context.Context, resourceGroupName, nsgName string, parameters network.SecurityGroup) (*network.SecurityGroup, error) {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.NSGs[nsgName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &nsgName
	parameters.ID = &nsgName
	c.NSGs[nsgName] = &parameters
	return &parameters, nil
}

// List returns a slice of Network Security Groups.
func (c *MockNetworkSecurityGroupsClient) List(ctx context.Context, resourceGroupName string) ([]*network.SecurityGroup, error) {
	var l []*network.SecurityGroup
	for _, nsg := range c.NSGs {
		l = append(l, nsg)
	}
	return l, nil
}

// Get Returns a specified Network Security Group.
func (c *MockNetworkSecurityGroupsClient) Get(ctx context.Context, resourceGroupName string, nsgName string) (*network.SecurityGroup, error) {
	nsg, ok := c.NSGs[nsgName]
	if !ok {
		return nil, nil
	}
	return nsg, nil
}

// Delete deletes a specified Network Security Group.
func (c *MockNetworkSecurityGroupsClient) Delete(ctx context.Context, resourceGroupName, nsgName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.NSGs[nsgName]; !ok {
		return fmt.Errorf("%s does not exist", nsgName)
	}
	delete(c.NSGs, nsgName)
	return nil
}

// MockApplicationSecurityGroupsClient is a mock implementation of Application Security Group client.
type MockApplicationSecurityGroupsClient struct {
	ASGs map[string]*network.ApplicationSecurityGroup
}

var _ azure.ApplicationSecurityGroupsClient = &MockApplicationSecurityGroupsClient{}

// CreateOrUpdate creates or updates a Application Security Group.
func (c *MockApplicationSecurityGroupsClient) CreateOrUpdate(ctx context.Context, resourceGroupName, asgName string, parameters network.ApplicationSecurityGroup) (*network.ApplicationSecurityGroup, error) {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.ASGs[asgName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &asgName
	parameters.ID = &asgName
	c.ASGs[asgName] = &parameters
	return &parameters, nil
}

// List returns a slice of Application Security Groups.
func (c *MockApplicationSecurityGroupsClient) List(ctx context.Context, resourceGroupName string) ([]*network.ApplicationSecurityGroup, error) {
	var l []*network.ApplicationSecurityGroup
	for _, nsg := range c.ASGs {
		l = append(l, nsg)
	}
	return l, nil
}

// Get Returns a specified Application Security Group.
func (c *MockApplicationSecurityGroupsClient) Get(ctx context.Context, resourceGroupName string, asgName string) (*network.ApplicationSecurityGroup, error) {
	asg, ok := c.ASGs[asgName]
	if !ok {
		return nil, nil
	}
	return asg, nil
}

// Delete deletes a specified Application Security Group.
func (c *MockApplicationSecurityGroupsClient) Delete(ctx context.Context, resourceGroupName, asgName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.ASGs[asgName]; !ok {
		return fmt.Errorf("%s does not exist", asgName)
	}
	delete(c.ASGs, asgName)
	return nil
}

// MockNatGatewaysClient is a mock implementation of Nat Gateway client.
type MockNatGatewaysClient struct {
	NGWs map[string]*network.NatGateway
}

var _ azure.NatGatewaysClient = &MockNatGatewaysClient{}

// CreateOrUpdate creates or updates a Nat Gateway.
func (c *MockNatGatewaysClient) CreateOrUpdate(ctx context.Context, resourceGroupName, ngwName string, parameters network.NatGateway) (*network.NatGateway, error) {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.NGWs[ngwName]; ok {
		return nil, fmt.Errorf("update not supported")
	}
	parameters.Name = &ngwName
	parameters.ID = &ngwName
	c.NGWs[ngwName] = &parameters
	return &parameters, nil
}

// List returns a slice of Nat Gateways.
func (c *MockNatGatewaysClient) List(ctx context.Context, resourceGroupName string) ([]*network.NatGateway, error) {
	var l []*network.NatGateway
	for _, ngw := range c.NGWs {
		l = append(l, ngw)
	}
	return l, nil
}

// Get Returns a specified Nat Gateway.
func (c *MockNatGatewaysClient) Get(ctx context.Context, resourceGroupName string, ngwName string) (*network.NatGateway, error) {
	ngw, ok := c.NGWs[ngwName]
	if !ok {
		return nil, nil
	}
	return ngw, nil
}

// Delete deletes a specified Nat Gateway.
func (c *MockNatGatewaysClient) Delete(ctx context.Context, resourceGroupName, ngwName string) error {
	// Ignore resourceGroupName for simplicity.
	if _, ok := c.NGWs[ngwName]; !ok {
		return fmt.Errorf("%s does not exist", ngwName)
	}
	delete(c.NGWs, ngwName)
	return nil
}
