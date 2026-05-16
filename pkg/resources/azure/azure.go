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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	azureresources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/utils/set"
)

const (
	typeResourceGroup            = "ResourceGroup"
	typeVirtualNetwork           = "VirtualNetwork"
	typeNetworkSecurityGroup     = "NetworkSecurityGroup"
	typeApplicationSecurityGroup = "ApplicationSecurityGroup"
	typeSubnet                   = "Subnet"
	typeRouteTable               = "RouteTable"
	typeVMScaleSet               = "VMScaleSet"
	typeVMScaleSetVM             = "VMScaleSetVM"
	typeDisk                     = "Disk"
	typeRoleAssignment           = "RoleAssignment"
	typeLoadBalancer             = "LoadBalancer"
	typePublicIPAddress          = "PublicIPAddress"
	typeNatGateway               = "NatGateway"
	typeManagedIdentity          = "ManagedIdentity"
)

// ListResourcesAzure lists all resources for the cluster by quering Azure.
func ListResourcesAzure(cloud azure.AzureCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	g := resourceGetter{
		cloud:       cloud,
		clusterInfo: clusterInfo,
	}
	return g.listResourcesAzure()
}

type resourceGetter struct {
	cloud       azure.AzureCloud
	clusterInfo resources.ClusterInfo
}

func (g *resourceGetter) resourceGroupName() string {
	return g.clusterInfo.AzureResourceGroupName
}

func (g *resourceGetter) resourceGroupID() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", g.clusterInfo.AzureSubscriptionID, g.clusterInfo.AzureResourceGroupName)
}

func (g *resourceGetter) storageAccountID() string {
	return g.clusterInfo.AzureStorageAccountID
}

func (g *resourceGetter) listResourcesAzure() (map[string]*resources.Resource, error) {
	rs, err := g.listAll()
	if err != nil {
		return nil, err
	}

	// Convert a slice of resources to a map of resources keyed by type and ID.
	// Normalize IDs to lowercase since Azure resource IDs are case-insensitive
	// but different Azure APIs may return different casing for the same resource.
	resources := make(map[string]*resources.Resource)
	for _, r := range rs {
		if r.Done {
			continue
		}
		r.ID = strings.ToLower(r.ID)
		resources[toKey(r.Type, r.ID)] = r
	}
	return resources, nil
}

// listAll list all resources owned by kops for the cluster.
//
// TODO(kenji): Set the "Shared" field of each resource so that we won't delete
// shared resources.
func (g *resourceGetter) listAll() ([]*resources.Resource, error) {
	fns := []func(ctx context.Context) ([]*resources.Resource, error){
		g.listResourceGroups,
		g.listVirtualNetworksAndSubnets,
		g.listNetworkSecurityGroups,
		g.listApplicationSecurityGroups,
		g.listRouteTables,
		g.listVMScaleSetsAndRoleAssignments,
		g.listDisks,
		g.listLoadBalancers,
		g.listPublicIPAddresses,
		g.listNatGateways,
	}

	var resources []*resources.Resource
	ctx := context.TODO()
	for _, fn := range fns {
		rs, err := fn(ctx)
		if err != nil {
			return nil, err
		}
		resources = append(resources, rs...)
	}
	return resources, nil
}

func (g *resourceGetter) listResourceGroups(ctx context.Context) ([]*resources.Resource, error) {
	rgs, err := g.cloud.ResourceGroup().List(ctx)
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, rg := range rgs {
		if !g.isOwnedByCluster(rg.Tags) {
			continue
		}
		rs = append(rs, g.toResourceGroupResource(rg))
	}
	return rs, nil
}

func (g *resourceGetter) toResourceGroupResource(rg *azureresources.ResourceGroup) *resources.Resource {
	return &resources.Resource{
		Obj:     rg,
		Type:    typeResourceGroup,
		ID:      *rg.ID,
		Name:    *rg.Name,
		Deleter: g.deleteResourceGroup,
		Shared:  g.clusterInfo.AzureResourceGroupShared,
	}
}

func (g *resourceGetter) deleteResourceGroup(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.ResourceGroup().Delete(context.TODO(), r.Name)
}

func (g *resourceGetter) listVirtualNetworksAndSubnets(ctx context.Context) ([]*resources.Resource, error) {
	vnets, err := g.cloud.VirtualNetwork().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, vnet := range vnets {
		if !g.isOwnedByCluster(vnet.Tags) {
			continue
		}
		r, err := g.toVirtualNetworkResource(vnet)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
		// Add all subnets belonging to the virtual network.
		subnets, err := g.listSubnets(ctx, *vnet.ID)
		if err != nil {
			return nil, err
		}
		rs = append(rs, subnets...)
	}
	return rs, nil
}

func (g *resourceGetter) toVirtualNetworkResource(vnet *network.VirtualNetwork) (*resources.Resource, error) {
	var blocks []string
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupID()))

	nsgs := set.New[string]()
	if vnet.Properties != nil && vnet.Properties.Subnets != nil {
		for _, sn := range vnet.Properties.Subnets {
			if sn.Properties == nil || sn.Properties.NetworkSecurityGroup == nil || sn.Properties.NetworkSecurityGroup.ID == nil {
				continue
			}
			nsgs.Insert(*sn.Properties.NetworkSecurityGroup.ID)
		}
	}
	for nsg := range nsgs {
		blocks = append(blocks, toKey(typeNetworkSecurityGroup, nsg))
	}

	return &resources.Resource{
		Obj:     vnet,
		Type:    typeVirtualNetwork,
		ID:      *vnet.ID,
		Name:    *vnet.Name,
		Deleter: g.deleteVirtualNetwork,
		Blocks:  blocks,
		Shared:  g.clusterInfo.AzureNetworkShared,
	}, nil
}

func (g *resourceGetter) deleteVirtualNetwork(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.VirtualNetwork().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listSubnets(ctx context.Context, vnetID string) ([]*resources.Resource, error) {
	vnet, err := arm.ParseResourceID(vnetID)
	if err != nil {
		return nil, err
	}
	subnets, err := g.cloud.Subnet().List(ctx, g.resourceGroupName(), vnet.Name)
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, sn := range subnets {
		rs = append(rs, g.toSubnetResource(sn, vnetID))
	}
	return rs, nil
}

func (g *resourceGetter) toSubnetResource(subnet *network.Subnet, vnetID string) *resources.Resource {
	var blocks []string
	blocks = append(blocks, toKey(typeVirtualNetwork, vnetID))
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupID()))

	if subnet.Properties != nil {
		if subnet.Properties.NatGateway != nil && subnet.Properties.NatGateway.ID != nil {
			blocks = append(blocks, toKey(typeNatGateway, *subnet.Properties.NatGateway.ID))
		}
		if subnet.Properties.RouteTable != nil && subnet.Properties.RouteTable.ID != nil {
			blocks = append(blocks, toKey(typeRouteTable, *subnet.Properties.RouteTable.ID))
		}
		if subnet.Properties.NetworkSecurityGroup != nil && subnet.Properties.NetworkSecurityGroup.ID != nil {
			blocks = append(blocks, toKey(typeNetworkSecurityGroup, *subnet.Properties.NetworkSecurityGroup.ID))
		}
	}

	vnet, err := arm.ParseResourceID(vnetID)
	if err != nil {
		return nil
	}

	return &resources.Resource{
		Obj:  subnet,
		Type: typeSubnet,
		ID:   *subnet.ID,
		Name: *subnet.Name,
		Deleter: func(_ fi.Cloud, r *resources.Resource) error {
			return g.deleteSubnet(vnet.Name, r)
		},
		Blocks: blocks,
		Shared: g.clusterInfo.AzureNetworkShared,
	}
}

func (g *resourceGetter) deleteSubnet(vnetName string, r *resources.Resource) error {
	return g.cloud.Subnet().Delete(context.TODO(), g.resourceGroupName(), vnetName, r.Name)
}

func (g *resourceGetter) listNetworkSecurityGroups(ctx context.Context) ([]*resources.Resource, error) {
	NetworkSecurityGroups, err := g.cloud.NetworkSecurityGroup().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for i := range NetworkSecurityGroups {
		r, err := g.toNetworkSecurityGroupResource(NetworkSecurityGroups[i])
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func (g *resourceGetter) toNetworkSecurityGroupResource(NetworkSecurityGroup *network.SecurityGroup) (*resources.Resource, error) {
	var blocks []string
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupID()))

	asgs := set.New[string]()
	if NetworkSecurityGroup.Properties.SecurityRules != nil {
		for _, nsr := range NetworkSecurityGroup.Properties.SecurityRules {
			if nsr.Properties.SourceApplicationSecurityGroups != nil {
				for _, sasg := range nsr.Properties.SourceApplicationSecurityGroups {
					asgs.Insert(*sasg.ID)
				}
			}
			if nsr.Properties.DestinationApplicationSecurityGroups != nil {
				for _, dasg := range nsr.Properties.DestinationApplicationSecurityGroups {
					asgs.Insert(*dasg.ID)
				}
			}
		}
	}
	for asg := range asgs {
		blocks = append(blocks, toKey(typeApplicationSecurityGroup, asg))
	}

	return &resources.Resource{
		Obj:  NetworkSecurityGroup,
		Type: typeNetworkSecurityGroup,
		ID:   *NetworkSecurityGroup.ID,
		Name: *NetworkSecurityGroup.Name,
		Deleter: func(_ fi.Cloud, r *resources.Resource) error {
			return g.deleteNetworkSecurityGroup(r)
		},
		Blocks: blocks,
	}, nil
}

func (g *resourceGetter) deleteNetworkSecurityGroup(r *resources.Resource) error {
	return g.cloud.NetworkSecurityGroup().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listApplicationSecurityGroups(ctx context.Context) ([]*resources.Resource, error) {
	ApplicationSecurityGroups, err := g.cloud.ApplicationSecurityGroup().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, asg := range ApplicationSecurityGroups {
		rs = append(rs, g.toApplicationSecurityGroupResource(asg))
	}
	return rs, nil
}

func (g *resourceGetter) toApplicationSecurityGroupResource(ApplicationSecurityGroup *network.ApplicationSecurityGroup) *resources.Resource {
	return &resources.Resource{
		Obj:  ApplicationSecurityGroup,
		Type: typeApplicationSecurityGroup,
		ID:   *ApplicationSecurityGroup.ID,
		Name: *ApplicationSecurityGroup.Name,
		Deleter: func(_ fi.Cloud, r *resources.Resource) error {
			return g.deleteApplicationSecurityGroup(r)
		},
		Blocks: []string{
			toKey(typeResourceGroup, g.resourceGroupID()),
		},
	}
}

func (g *resourceGetter) deleteApplicationSecurityGroup(r *resources.Resource) error {
	return g.cloud.ApplicationSecurityGroup().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listRouteTables(ctx context.Context) ([]*resources.Resource, error) {
	rts, err := g.cloud.RouteTable().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, rt := range rts {
		if !g.isOwnedByCluster(rt.Tags) {
			continue
		}
		rs = append(rs, g.toRouteTableResource(rt))
	}
	return rs, nil
}

func (g *resourceGetter) toRouteTableResource(rt *network.RouteTable) *resources.Resource {
	return &resources.Resource{
		Obj:     rt,
		Type:    typeRouteTable,
		ID:      *rt.ID,
		Name:    *rt.Name,
		Deleter: g.deleteRouteTable,
		Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupID())},
		Shared:  g.clusterInfo.AzureRouteTableShared,
	}
}

func (g *resourceGetter) deleteRouteTable(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.RouteTable().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listVMScaleSetsAndRoleAssignments(ctx context.Context) ([]*resources.Resource, error) {
	vmsses, err := g.cloud.VMScaleSet().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	principalIDs := map[string]*compute.VirtualMachineScaleSet{}
	for _, vmss := range vmsses {
		if !g.isOwnedByCluster(vmss.Tags) {
			continue
		}

		vms, err := g.cloud.VMScaleSetVM().List(ctx, g.resourceGroupName(), *vmss.Name)
		if err != nil {
			return nil, err
		}

		for _, vm := range vms {
			vmr, err := g.toVMScaleSetVMResource(vmss, vm)
			if err != nil {
				return nil, err
			}
			rs = append(rs, vmr)
		}

		r, err := g.toVMScaleSetResource(vmss, vms)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)

		if vmss.Identity != nil {
			// Collect principal IDs from both system-assigned and user-assigned identities.
			if vmss.Identity.PrincipalID != nil {
				principalIDs[*vmss.Identity.PrincipalID] = vmss
			}
			for _, uai := range vmss.Identity.UserAssignedIdentities {
				if uai != nil && uai.PrincipalID != nil {
					principalIDs[*uai.PrincipalID] = vmss
				}
			}
		}
	}

	// Collect VMSS IDs so that managed identities are not deleted before all VMSS are gone.
	var blocked []string
	for _, r := range rs {
		if r.Type == typeVMScaleSet {
			blocked = append(blocked, toKey(typeVMScaleSet, r.ID))
		}
	}

	// Also list and delete managed identities owned by the cluster.
	miResources, err := g.listManagedIdentities(ctx, blocked)
	if err != nil {
		return nil, err
	}
	rs = append(rs, miResources...)

	resourceGroupRAs, err := g.listRoleAssignments(ctx, principalIDs, g.resourceGroupID())
	if err != nil {
		return nil, err
	}
	rs = append(rs, resourceGroupRAs...)

	storageAccountRAs, err := g.listRoleAssignments(ctx, principalIDs, g.storageAccountID())
	if err != nil {
		return nil, err
	}
	rs = append(rs, storageAccountRAs...)

	return rs, nil
}

func (g *resourceGetter) toVMScaleSetResource(vmss *compute.VirtualMachineScaleSet, vms []*compute.VirtualMachineScaleSetVM) (*resources.Resource, error) {
	// Add resources whose deletion is blocked by this VMSS.
	var blocks []string
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupID()))

	// The VM Scale Set deletion is blocked by its instances.
	var blocked []string
	for _, vm := range vms {
		if vm == nil || vm.ID == nil {
			continue
		}
		blocked = append(blocked, toKey(typeVMScaleSetVM, *vm.ID))
	}

	vnets := set.New[string]()
	subnets := set.New[string]()
	asgs := set.New[string]()
	lbs := set.New[string]()
	if vmss.Properties == nil || vmss.Properties.VirtualMachineProfile == nil || vmss.Properties.VirtualMachineProfile.NetworkProfile == nil {
		return nil, fmt.Errorf("VMSS %s has no network profile", fi.ValueOf(vmss.Name))
	}
	for _, iface := range vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations {
		if iface.Properties == nil {
			continue
		}
		for _, ip := range iface.Properties.IPConfigurations {
			if ip.Properties == nil || ip.Properties.Subnet == nil {
				continue
			}
			subnet, err := arm.ParseResourceID(*ip.Properties.Subnet.ID)
			if err != nil {
				return nil, err
			}
			vnets.Insert(subnet.Parent.String())
			subnets.Insert(subnet.String())
			if ip.Properties.ApplicationSecurityGroups != nil {
				for _, asg := range ip.Properties.ApplicationSecurityGroups {
					asgs.Insert(*asg.ID)
				}
			}
			if ip.Properties.LoadBalancerBackendAddressPools != nil {
				for _, lbbap := range ip.Properties.LoadBalancerBackendAddressPools {
					pool, err := arm.ParseResourceID(*lbbap.ID)
					if err != nil {
						return nil, err
					}
					lbs.Insert(pool.Parent.String())
				}
			}
		}
	}
	for vnet := range vnets {
		blocks = append(blocks, toKey(typeVirtualNetwork, vnet))
	}
	for subnet := range subnets {
		blocks = append(blocks, toKey(typeSubnet, subnet))
	}
	for asg := range asgs {
		blocks = append(blocks, toKey(typeApplicationSecurityGroup, asg))
	}
	for lb := range lbs {
		blocks = append(blocks, toKey(typeLoadBalancer, lb))
	}

	return &resources.Resource{
		Obj:     vmss,
		Type:    typeVMScaleSet,
		ID:      *vmss.ID,
		Name:    *vmss.Name,
		Deleter: g.deleteVMScaleSet,
		Blocks:  blocks,
		Blocked: blocked,
		Dumper:  DumpVMScaleSet,
	}, nil
}

func (g *resourceGetter) toVMScaleSetVMResource(vmss *compute.VirtualMachineScaleSet, vm *compute.VirtualMachineScaleSetVM) (*resources.Resource, error) {
	if vm == nil || vm.ID == nil {
		return nil, fmt.Errorf("VMScaleSetVM is missing ID")
	}
	if vmss == nil || vmss.Name == nil || vmss.ID == nil {
		return nil, fmt.Errorf("VMScaleSet is missing ID or Name")
	}

	rid, err := arm.ParseResourceID(*vm.ID)
	if err != nil {
		return nil, err
	}
	instanceID := rid.Name

	name := ""
	if vm.Properties != nil && vm.Properties.OSProfile != nil && vm.Properties.OSProfile.ComputerName != nil {
		name = *vm.Properties.OSProfile.ComputerName
	} else if vm.Name != nil {
		name = *vm.Name
	} else {
		name = instanceID
	}

	return &resources.Resource{
		Obj:    vm,
		Type:   typeVMScaleSetVM,
		ID:     *vm.ID,
		Name:   name,
		Dumper: DumpVMScaleSetVM,
		Deleter: func(_ fi.Cloud, r *resources.Resource) error {
			return g.cloud.VMScaleSetVM().Delete(context.TODO(), g.resourceGroupName(), *vmss.Name, instanceID)
		},
		Blocks: []string{
			toKey(typeResourceGroup, g.resourceGroupID()),
			toKey(typeVMScaleSet, *vmss.ID),
		},
	}, nil
}

func (g *resourceGetter) deleteVMScaleSet(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.VMScaleSet().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listDisks(ctx context.Context) ([]*resources.Resource, error) {
	disks, err := g.cloud.Disk().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, disk := range disks {
		if !g.isOwnedByCluster(disk.Tags) {
			continue
		}
		rs = append(rs, g.toDiskResource(disk))
	}
	return rs, nil
}

func (g *resourceGetter) toDiskResource(disk *compute.Disk) *resources.Resource {
	var blocked []string
	if disk.ManagedBy != nil {
		// Block on the parent VMScaleSet, not the individual VM instance.
		// The raw ManagedBy path may not match the listed VM's resource ID,
		// but parsing it to extract the parent VMSS gives a reliable match.
		vmID, err := arm.ParseResourceID(*disk.ManagedBy)
		if err == nil && vmID.Parent != nil {
			blocked = append(blocked, toKey(typeVMScaleSet, vmID.Parent.String()))
		}
	}

	return &resources.Resource{
		Obj:     disk,
		Type:    typeDisk,
		ID:      *disk.ID,
		Name:    *disk.Name,
		Deleter: g.deleteDisk,
		Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupID())},
		Blocked: blocked,
	}
}

func (g *resourceGetter) deleteDisk(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.Disk().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listRoleAssignments(ctx context.Context, principalIDs map[string]*compute.VirtualMachineScaleSet, scope string) ([]*resources.Resource, error) {
	ras, err := g.cloud.RoleAssignment().List(ctx, scope)
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, ra := range ras {
		// Add a Role Assignment to the slice if its principal ID is that of one of the VM Scale Sets.
		if ra.Properties == nil || ra.Properties.PrincipalID == nil {
			continue
		}
		vmss, ok := principalIDs[*ra.Properties.PrincipalID]
		if !ok {
			continue
		}
		rs = append(rs, g.toRoleAssignmentResource(ra, vmss))
	}
	return rs, nil
}

func (g *resourceGetter) toRoleAssignmentResource(ra *authz.RoleAssignment, vmss *compute.VirtualMachineScaleSet) *resources.Resource {
	return &resources.Resource{
		Obj:     ra,
		Type:    typeRoleAssignment,
		ID:      *ra.ID,
		Name:    *ra.Name,
		Deleter: g.deleteRoleAssignment,
		Blocks: []string{
			toKey(typeResourceGroup, g.resourceGroupID()),
		},
		// Wait for the VMSS to be deleted before removing role assignments,
		// to avoid permission issues during VMSS teardown.
		Blocked: []string{toKey(typeVMScaleSet, *vmss.ID)},
	}
}

func (g *resourceGetter) deleteRoleAssignment(_ fi.Cloud, r *resources.Resource) error {
	ra, ok := r.Obj.(*authz.RoleAssignment)
	if !ok {
		return fmt.Errorf("expected RoleAssignment, but got %T", r)
	}
	return g.cloud.RoleAssignment().Delete(context.TODO(), *ra.Properties.Scope, *ra.Name)
}

func (g *resourceGetter) listLoadBalancers(ctx context.Context) ([]*resources.Resource, error) {
	loadBalancers, err := g.cloud.LoadBalancer().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, lb := range loadBalancers {
		if !g.isOwnedByCluster(lb.Tags) {
			continue
		}
		r, err := g.toLoadBalancerResource(lb)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func (g *resourceGetter) toLoadBalancerResource(loadBalancer *network.LoadBalancer) (*resources.Resource, error) {
	var blocks []string
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupID()))

	pips := set.New[string]()
	if loadBalancer.Properties != nil {
		for _, fip := range loadBalancer.Properties.FrontendIPConfigurations {
			if fip.Properties == nil || fip.Properties.PublicIPAddress == nil {
				continue
			}
			pips.Insert(*fip.Properties.PublicIPAddress.ID)
		}
	}
	for pip := range pips {
		blocks = append(blocks, toKey(typePublicIPAddress, pip))
	}

	return &resources.Resource{
		Obj:     loadBalancer,
		Type:    typeLoadBalancer,
		ID:      *loadBalancer.ID,
		Name:    *loadBalancer.Name,
		Deleter: g.deleteLoadBalancer,
		Blocks:  blocks,
		Dumper:  DumpLoadBalancer,
	}, nil
}

func (g *resourceGetter) deleteLoadBalancer(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.LoadBalancer().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listPublicIPAddresses(ctx context.Context) ([]*resources.Resource, error) {
	publicIPAddresses, err := g.cloud.PublicIPAddress().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, pip := range publicIPAddresses {
		if !g.isOwnedByCluster(pip.Tags) {
			continue
		}
		rs = append(rs, g.toPublicIPAddressResource(pip))
	}
	return rs, nil
}

func (g *resourceGetter) toPublicIPAddressResource(publicIPAddress *network.PublicIPAddress) *resources.Resource {
	return &resources.Resource{
		Obj:     publicIPAddress,
		Type:    typePublicIPAddress,
		ID:      *publicIPAddress.ID,
		Name:    *publicIPAddress.Name,
		Deleter: g.deletePublicIPAddress,
		Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupID())},
	}
}

func (g *resourceGetter) deletePublicIPAddress(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.PublicIPAddress().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listNatGateways(ctx context.Context) ([]*resources.Resource, error) {
	natGateways, err := g.cloud.NatGateway().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, ngw := range natGateways {
		if !g.isOwnedByCluster(ngw.Tags) {
			continue
		}
		r, err := g.toNatGatewayResource(ngw)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func (g *resourceGetter) toNatGatewayResource(natGateway *network.NatGateway) (*resources.Resource, error) {
	var blocks []string
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupID()))

	pips := set.New[string]()
	if natGateway.Properties != nil && natGateway.Properties.PublicIPAddresses != nil {
		for _, pip := range natGateway.Properties.PublicIPAddresses {
			pips.Insert(*pip.ID)
		}
	}
	for pip := range pips {
		blocks = append(blocks, toKey(typePublicIPAddress, pip))
	}

	return &resources.Resource{
		Obj:     natGateway,
		Type:    typeNatGateway,
		ID:      *natGateway.ID,
		Name:    *natGateway.Name,
		Deleter: g.deleteNatGateway,
		Blocks:  blocks,
	}, nil
}

func (g *resourceGetter) deleteNatGateway(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.NatGateway().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listManagedIdentities(ctx context.Context, blocked []string) ([]*resources.Resource, error) {
	mis, err := g.cloud.ManagedIdentity().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for _, mi := range mis {
		if !g.isOwnedByCluster(mi.Tags) {
			continue
		}
		rs = append(rs, &resources.Resource{
			Obj:  mi,
			Type: typeManagedIdentity,
			ID:   *mi.ID,
			Name: *mi.Name,
			Deleter: func(_ fi.Cloud, r *resources.Resource) error {
				return g.cloud.ManagedIdentity().Delete(context.TODO(), g.resourceGroupName(), r.Name)
			},
			Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupID())},
			Blocked: blocked,
		})
	}
	return rs, nil
}

// isOwnedByCluster returns true if the resource is owned by the cluster.
func (g *resourceGetter) isOwnedByCluster(tags map[string]*string) bool {
	for k, v := range tags {
		if k == azure.TagClusterName && *v == g.clusterInfo.Name {
			return true
		}
	}
	return false
}

func toKey(rtype, id string) string {
	return rtype + ":" + strings.ToLower(id)
}
