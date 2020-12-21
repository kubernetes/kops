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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	authz "github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	azureresources "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

const (
	typeResourceGroup  = "ResourceGroup"
	typeVirtualNetwork = "VirtualNetwork"
	typeSubnet         = "Subnet"
	typeRouteTable     = "RouteTable"
	typeVMScaleSet     = "VMScaleSet"
	typeDisk           = "Disk"
	typeRoleAssignment = "RoleAssignment"
)

// ListResourcesAzure lists all resources for the cluster by quering Azure.
func ListResourcesAzure(cloud azure.AzureCloud, cluster *kopsapi.Cluster) (map[string]*resources.Resource, error) {
	g := resourceGetter{
		cloud:   cloud,
		cluster: cluster,
	}
	return g.listResourcesAzure()
}

type resourceGetter struct {
	cloud   azure.AzureCloud
	cluster *kopsapi.Cluster
}

func (g *resourceGetter) resourceGroupName() string {
	return g.cluster.AzureResourceGroupName()
}

func (g *resourceGetter) listResourcesAzure() (map[string]*resources.Resource, error) {
	rs, err := g.listAll()
	if err != nil {
		return nil, err
	}

	// Convert a slice of resources to a map of resources keyed by type and ID.
	resources := make(map[string]*resources.Resource)
	for _, r := range rs {
		if r.Done {
			continue
		}
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
		g.listRouteTables,
		g.listVMScaleSetsAndRoleAssignments,
		g.listDisks,
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
	rgs, err := g.cloud.ResourceGroup().List(ctx, "" /* filter */)
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for i := range rgs {
		rg := &rgs[i]
		if !g.isOwnedByCluster(rg.Tags) {
			continue
		}
		rs = append(rs, g.toResourceGroupResource(rg))
	}
	return rs, nil
}

func (g *resourceGetter) toResourceGroupResource(rg *azureresources.Group) *resources.Resource {
	return &resources.Resource{
		Obj:     rg,
		Type:    typeResourceGroup,
		ID:      *rg.Name,
		Name:    *rg.Name,
		Deleter: g.deleteResourceGroup,
		Shared:  g.cluster.IsSharedAzureResourceGroup(),
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
	for i := range vnets {
		vnet := &vnets[i]
		if !g.isOwnedByCluster(vnet.Tags) {
			continue
		}
		rs = append(rs, g.toVirtualNetworkResource(vnet))
		// Add all subnets belonging to the virtual network.
		subnets, err := g.listSubnets(ctx, *vnet.Name)
		if err != nil {
			return nil, err
		}
		rs = append(rs, subnets...)
	}
	return rs, nil
}

func (g *resourceGetter) toVirtualNetworkResource(vnet *network.VirtualNetwork) *resources.Resource {
	return &resources.Resource{
		Obj:     vnet,
		Type:    typeVirtualNetwork,
		ID:      *vnet.Name,
		Name:    *vnet.Name,
		Deleter: g.deleteVirtualNetwork,
		Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupName())},
	}
}

func (g *resourceGetter) deleteVirtualNetwork(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.VirtualNetwork().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listSubnets(ctx context.Context, vnetName string) ([]*resources.Resource, error) {
	subnets, err := g.cloud.Subnet().List(ctx, g.resourceGroupName(), vnetName)
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for i := range subnets {
		rs = append(rs, g.toSubnetResource(&subnets[i], vnetName))
	}
	return rs, nil
}

func (g *resourceGetter) toSubnetResource(subnet *network.Subnet, vnetName string) *resources.Resource {
	return &resources.Resource{
		Obj:  subnet,
		Type: typeSubnet,
		ID:   *subnet.Name,
		Name: *subnet.Name,
		Deleter: func(_ fi.Cloud, r *resources.Resource) error {
			return g.deleteSubnet(vnetName, r)
		},
		Blocks: []string{
			toKey(typeVirtualNetwork, vnetName),
			toKey(typeResourceGroup, g.resourceGroupName()),
		},
	}
}

func (g *resourceGetter) deleteSubnet(vnetName string, r *resources.Resource) error {
	return g.cloud.Subnet().Delete(context.TODO(), g.resourceGroupName(), vnetName, r.Name)
}

func (g *resourceGetter) listRouteTables(ctx context.Context) ([]*resources.Resource, error) {
	rts, err := g.cloud.RouteTable().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for i := range rts {
		rt := &rts[i]
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
		ID:      *rt.Name,
		Name:    *rt.Name,
		Deleter: g.deleteRouteTable,
		Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupName())},
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
	for i := range vmsses {
		vmss := &vmsses[i]
		if !g.isOwnedByCluster(vmss.Tags) {
			continue
		}

		vms, err := g.cloud.VMScaleSetVM().List(ctx, g.resourceGroupName(), *vmss.Name)
		if err != nil {
			return nil, err
		}

		r, err := g.toVMScaleSetResource(vmss, vms)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)

		principalIDs[*vmss.Identity.PrincipalID] = vmss
	}

	ras, err := g.listRoleAssignments(ctx, principalIDs)
	if err != nil {
		return nil, err
	}
	rs = append(rs, ras...)

	return rs, nil
}

func (g *resourceGetter) toVMScaleSetResource(vmss *compute.VirtualMachineScaleSet, vms []compute.VirtualMachineScaleSetVM) (*resources.Resource, error) {
	// Add resources whose deletion is blocked by this VMSS.
	var blocks []string
	blocks = append(blocks, toKey(typeResourceGroup, g.resourceGroupName()))

	vnets := map[string]struct{}{}
	subnets := map[string]struct{}{}
	for _, iface := range *vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations {
		for _, ip := range *iface.IPConfigurations {
			subnetID, err := azuretasks.ParseSubnetID(*ip.Subnet.ID)
			if err != nil {
				return nil, fmt.Errorf("error on parsing subnet ID: %s", err)
			}
			vnets[subnetID.VirtualNetworkName] = struct{}{}
			subnets[subnetID.SubnetName] = struct{}{}
		}
	}
	for vnet := range vnets {
		blocks = append(blocks, toKey(typeVirtualNetwork, vnet))
	}
	for subnet := range subnets {
		blocks = append(blocks, toKey(typeSubnet, subnet))
	}

	for _, vm := range vms {
		if disks := vm.StorageProfile.DataDisks; disks != nil {
			for _, d := range *disks {
				blocks = append(blocks, toKey(typeDisk, *d.Name))
			}
		}
	}

	return &resources.Resource{
		Obj:     vmss,
		Type:    typeVMScaleSet,
		ID:      *vmss.Name,
		Name:    *vmss.Name,
		Deleter: g.deleteVMScaleSet,
		Blocks:  blocks,
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
	for i := range disks {
		disk := &disks[i]
		if !g.isOwnedByCluster(disk.Tags) {
			continue
		}
		rs = append(rs, g.toDiskResource(disk))
	}
	return rs, nil
}

func (g *resourceGetter) toDiskResource(disk *compute.Disk) *resources.Resource {
	return &resources.Resource{
		Obj:     disk,
		Type:    typeDisk,
		ID:      *disk.Name,
		Name:    *disk.Name,
		Deleter: g.deleteDisk,
		Blocks:  []string{toKey(typeResourceGroup, g.resourceGroupName())},
	}
}

func (g *resourceGetter) deleteDisk(_ fi.Cloud, r *resources.Resource) error {
	return g.cloud.Disk().Delete(context.TODO(), g.resourceGroupName(), r.Name)
}

func (g *resourceGetter) listRoleAssignments(ctx context.Context, principalIDs map[string]*compute.VirtualMachineScaleSet) ([]*resources.Resource, error) {
	ras, err := g.cloud.RoleAssignment().List(ctx, g.resourceGroupName())
	if err != nil {
		return nil, err
	}

	var rs []*resources.Resource
	for i := range ras {
		// Add a Role Assignment to the slice if its principal ID is that of one of the VM Scale Sets.
		ra := &ras[i]
		if ra.PrincipalID == nil {
			continue
		}
		vmss, ok := principalIDs[*ra.PrincipalID]
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
		ID:      *ra.Name,
		Name:    *ra.Name,
		Deleter: g.deleteRoleAssignment,
		Blocks: []string{
			toKey(typeResourceGroup, g.resourceGroupName()),
			toKey(typeVMScaleSet, *vmss.Name),
		},
	}
}

func (g *resourceGetter) deleteRoleAssignment(_ fi.Cloud, r *resources.Resource) error {
	ra, ok := r.Obj.(*authz.RoleAssignment)
	if !ok {
		return fmt.Errorf("expected RoleAssignment, but got %T", r)
	}
	return g.cloud.RoleAssignment().Delete(context.TODO(), *ra.Scope, *ra.Name)
}

// isOwnedByCluster returns true if the resource is owned by the cluster.
func (g *resourceGetter) isOwnedByCluster(tags map[string]*string) bool {
	for k, v := range tags {
		if k == azure.TagClusterName && *v == g.cluster.Name {
			return true
		}
	}
	return false
}

func toKey(rtype, id string) string {
	return rtype + ":" + id
}
