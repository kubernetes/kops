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
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

func TestListResourcesAzure(t *testing.T) {
	const (
		clusterName    = "cluster"
		irrelevantName = "irrelevant"
		principalID    = "pid"
		subscriptionID = "00000000-0000-0000-0000-000000000000"
		rgID           = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg"
		vnetID         = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet"
		subnetID       = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet"
		rtID           = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/routeTables/rt"
		vmssID         = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss"
		vmID           = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss/virtualmachines/0"
		diskID         = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/disks/disk"
		raID           = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Authorization/roleAssignments/ra"
		lbID           = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb"
	)

	rg, _ := arm.ParseResourceID(rgID)
	rgName := rg.Name
	vnet, _ := arm.ParseResourceID(vnetID)
	vnetName := vnet.Name
	vmss, _ := arm.ParseResourceID(vmssID)
	vmssName := vmss.Name
	vm, _ := arm.ParseResourceID(vmID)
	vmName := vm.Name
	disk, _ := arm.ParseResourceID(diskID)
	diskName := disk.Name
	subnet, _ := arm.ParseResourceID(subnetID)
	subnetName := subnet.Name
	rt, _ := arm.ParseResourceID(rtID)
	rtName := rt.Name
	ra, _ := arm.ParseResourceID(raID)
	raName := ra.Name
	lb, _ := arm.ParseResourceID(lbID)
	lbName := lb.Name

	clusterTags := map[string]*string{
		azure.TagClusterName: to.Ptr(clusterName),
	}

	cloud := azuretasks.NewMockAzureCloud("eastus")
	// Set up resources in the mock clients.
	rgs := cloud.ResourceGroupsClient.RGs
	rgs[rgName] = &armresources.ResourceGroup{
		ID:   to.Ptr(rgID),
		Name: to.Ptr(rgName),
		Tags: clusterTags,
	}
	rgs[irrelevantName] = &armresources.ResourceGroup{
		Name: to.Ptr(irrelevantName),
	}

	vnets := cloud.VirtualNetworksClient.VNets
	vnets[vnetName] = &network.VirtualNetwork{
		ID:         to.Ptr(vnetID),
		Name:       to.Ptr(vnetName),
		Tags:       clusterTags,
		Properties: &network.VirtualNetworkPropertiesFormat{},
	}
	vnets[irrelevantName] = &network.VirtualNetwork{
		Name: to.Ptr(irrelevantName),
	}

	subnets := cloud.SubnetsClient.Subnets
	subnets[rgName] = &network.Subnet{
		ID:         to.Ptr(subnetID),
		Name:       to.Ptr(subnetName),
		Properties: &network.SubnetPropertiesFormat{},
	}
	vnets[irrelevantName] = &network.VirtualNetwork{
		Name: to.Ptr(irrelevantName),
	}

	rts := cloud.RouteTablesClient.RTs
	rts[rtName] = &network.RouteTable{
		ID:   to.Ptr(rtID),
		Name: to.Ptr(rtName),
		Tags: clusterTags,
	}
	rts[irrelevantName] = &network.RouteTable{
		Name: to.Ptr(irrelevantName),
	}

	vmsses := cloud.VMScaleSetsClient.VMSSes
	networkConfig := &compute.VirtualMachineScaleSetNetworkConfiguration{
		Properties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
			IPConfigurations: []*compute.VirtualMachineScaleSetIPConfiguration{
				{
					Properties: &compute.VirtualMachineScaleSetIPConfigurationProperties{
						Subnet: &compute.APIEntityReference{
							ID: to.Ptr(subnetID),
						},
					},
				},
			},
		},
	}
	vmsses[vmssName] = &compute.VirtualMachineScaleSet{
		ID:   to.Ptr(vmssID),
		Name: to.Ptr(vmssName),
		Tags: clusterTags,
		Properties: &compute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: []*compute.VirtualMachineScaleSetNetworkConfiguration{
						networkConfig,
					},
				},
			},
		},
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type:        to.Ptr(compute.ResourceIdentityTypeSystemAssigned),
			PrincipalID: to.Ptr(principalID),
		},
	}
	vmsses[irrelevantName] = &compute.VirtualMachineScaleSet{
		Name: to.Ptr(irrelevantName),
	}

	vms := cloud.VMScaleSetVMsClient.VMs
	vms[vmName] = &compute.VirtualMachineScaleSetVM{
		Properties: &compute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: []*compute.DataDisk{
					{
						Name: to.Ptr(diskName),
					},
				},
			},
		},
	}

	disks := cloud.DisksClient.Disks
	disks[diskName] = &compute.Disk{
		ID:        to.Ptr(diskID),
		Name:      to.Ptr(diskName),
		ManagedBy: to.Ptr(vmID),
		Tags:      clusterTags,
	}
	disks[irrelevantName] = &compute.Disk{
		Name: to.Ptr(irrelevantName),
	}

	ras := cloud.RoleAssignmentsClient.RAs
	ras[raName] = &authz.RoleAssignment{
		ID:   to.Ptr(raID),
		Name: to.Ptr(raName),
		Properties: &authz.RoleAssignmentProperties{
			Scope:       to.Ptr("scope"),
			PrincipalID: to.Ptr(principalID),
		},
	}
	disks[irrelevantName] = &compute.Disk{
		Name: to.Ptr(irrelevantName),
	}

	lbs := cloud.LoadBalancersClient.LBs
	lbs[lbName] = &network.LoadBalancer{
		ID:         to.Ptr(lbID),
		Name:       to.Ptr(lbName),
		Tags:       clusterTags,
		Properties: &network.LoadBalancerPropertiesFormat{},
	}
	lbs[irrelevantName] = &network.LoadBalancer{
		Name: to.Ptr(irrelevantName),
	}

	// Call listResourcesAzure.
	g := resourceGetter{
		cloud: cloud,
		clusterInfo: resources.ClusterInfo{
			Name:                     clusterName,
			AzureSubscriptionID:      subscriptionID,
			AzureResourceGroupName:   rgName,
			AzureResourceGroupShared: true,
		},
	}
	actual, err := g.listResourcesAzure()
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	// Convert the resource map to its digest and compare with the expected.
	type resourceDigest struct {
		rtype   string
		name    string
		blocks  []string
		blocked []string
		shared  bool
	}
	toDigests := func(rs map[string]*resources.Resource) map[string]*resourceDigest {
		d := map[string]*resourceDigest{}
		for k, r := range rs {
			d[k] = &resourceDigest{
				rtype:   r.Type,
				name:    r.Name,
				blocks:  r.Blocks,
				blocked: r.Blocked,
				shared:  r.Shared,
			}
		}
		return d
	}
	a := toDigests(actual)
	e := map[string]*resourceDigest{
		toKey(typeResourceGroup, rgID): {
			rtype:  typeResourceGroup,
			name:   rgName,
			shared: true,
		},
		toKey(typeVirtualNetwork, vnetID): {
			rtype:  typeVirtualNetwork,
			name:   vnetName,
			blocks: []string{toKey(typeResourceGroup, rgID)},
		},
		toKey(typeSubnet, subnetID): {
			rtype: typeSubnet,
			name:  subnetName,
			blocks: []string{
				toKey(typeVirtualNetwork, vnetID),
				toKey(typeResourceGroup, rgID),
			},
		},
		toKey(typeRouteTable, rtID): {
			rtype:  typeRouteTable,
			name:   rtName,
			blocks: []string{toKey(typeResourceGroup, rgID)},
		},
		toKey(typeVMScaleSet, vmssID): {
			rtype: typeVMScaleSet,
			name:  vmssName,
			blocks: []string{
				toKey(typeResourceGroup, rgID),
				toKey(typeVirtualNetwork, vnetID),
				toKey(typeSubnet, subnetID),
			},
		},
		toKey(typeDisk, diskID): {
			rtype:   typeDisk,
			name:    diskName,
			blocks:  []string{toKey(typeResourceGroup, rgID)},
			blocked: []string{toKey(typeVMScaleSet, vmssID)},
		},
		toKey(typeRoleAssignment, raID): {
			rtype: typeRoleAssignment,
			name:  raName,
			blocks: []string{
				toKey(typeResourceGroup, rgID),
				toKey(typeVMScaleSet, vmssID),
			},
		},
		toKey(typeLoadBalancer, lbID): {
			rtype:  typeLoadBalancer,
			name:   lbName,
			blocks: []string{toKey(typeResourceGroup, rgID)},
		},
	}
	if !reflect.DeepEqual(a, e) {
		t.Errorf("expected %+v, but got %+v", e, a)
	}
}

func TestIsOwnedByCluster(t *testing.T) {
	clusterName := "test-cluster"

	testCases := []struct {
		tags     map[string]*string
		expected bool
	}{
		{
			tags: map[string]*string{
				azure.TagClusterName: to.Ptr(clusterName),
			},
			expected: true,
		},
		{
			tags: map[string]*string{
				azure.TagClusterName: to.Ptr(clusterName),
				"other-key":          to.Ptr("other-tag"),
			},
			expected: true,
		},
		{
			tags: map[string]*string{
				"other-key": to.Ptr("other-tag"),
			},
			expected: false,
		},
		{
			tags: map[string]*string{
				azure.TagClusterName: to.Ptr("different-cluster"),
			},
			expected: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			g := &resourceGetter{
				clusterInfo: resources.ClusterInfo{
					Name: clusterName,
				},
			}
			a := g.isOwnedByCluster(tc.tags)
			if a != tc.expected {
				t.Errorf("expected %t, but got %t", tc.expected, a)
			}
		})
	}
}
