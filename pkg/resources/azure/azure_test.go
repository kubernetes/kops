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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	authz "github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	azureresources "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources"
	"github.com/Azure/go-autorest/autorest/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

func TestListResourcesAzure(t *testing.T) {
	const (
		clusterName    = "cluster"
		rgName         = "rg"
		vnetName       = "vnet"
		vmssName       = "vmss"
		vmName         = "vmss/0"
		diskName       = "disk"
		subnetName     = "sub"
		rtName         = "rt"
		raName         = "ra"
		irrelevantName = "irrelevant"
		principalID    = "pid"
		lbName         = "lb"
	)
	clusterTags := map[string]*string{
		azure.TagClusterName: to.StringPtr(clusterName),
	}

	cloud := azuretasks.NewMockAzureCloud("eastus")
	// Set up resources in the mock clients.
	rgs := cloud.ResourceGroupsClient.RGs
	rgs[rgName] = azureresources.Group{
		Name: to.StringPtr(rgName),
		Tags: clusterTags,
	}
	rgs[irrelevantName] = azureresources.Group{
		Name: to.StringPtr(irrelevantName),
	}

	vnets := cloud.VirtualNetworksClient.VNets
	vnets[vnetName] = network.VirtualNetwork{
		Name: to.StringPtr(vnetName),
		Tags: clusterTags,
	}
	vnets[irrelevantName] = network.VirtualNetwork{
		Name: to.StringPtr(irrelevantName),
	}

	subnets := cloud.SubnetsClient.Subnets
	subnets[rgName] = network.Subnet{
		Name: to.StringPtr(subnetName),
	}
	vnets[irrelevantName] = network.VirtualNetwork{
		Name: to.StringPtr(irrelevantName),
	}

	rts := cloud.RouteTablesClient.RTs
	rts[rtName] = network.RouteTable{
		Name: to.StringPtr(rtName),
		Tags: clusterTags,
	}
	rts[irrelevantName] = network.RouteTable{
		Name: to.StringPtr(irrelevantName),
	}

	vmsses := cloud.VMScaleSetsClient.VMSSes
	subnetID := azuretasks.SubnetID{
		SubscriptionID:     "sid",
		ResourceGroupName:  rgName,
		VirtualNetworkName: vnetName,
		SubnetName:         subnetName,
	}
	networkConfig := compute.VirtualMachineScaleSetNetworkConfiguration{
		VirtualMachineScaleSetNetworkConfigurationProperties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
			IPConfigurations: &[]compute.VirtualMachineScaleSetIPConfiguration{
				{
					VirtualMachineScaleSetIPConfigurationProperties: &compute.VirtualMachineScaleSetIPConfigurationProperties{
						Subnet: &compute.APIEntityReference{
							ID: to.StringPtr(subnetID.String()),
						},
					},
				},
			},
		},
	}
	vmsses[vmssName] = compute.VirtualMachineScaleSet{
		Name: to.StringPtr(vmssName),
		Tags: clusterTags,
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: &[]compute.VirtualMachineScaleSetNetworkConfiguration{
						networkConfig,
					},
				},
			},
		},
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type:        compute.ResourceIdentityTypeSystemAssigned,
			PrincipalID: to.StringPtr(principalID),
		},
	}
	vmsses[irrelevantName] = compute.VirtualMachineScaleSet{
		Name: to.StringPtr(irrelevantName),
	}

	vms := cloud.VMScaleSetVMsClient.VMs
	vms[vmName] = compute.VirtualMachineScaleSetVM{
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &[]compute.DataDisk{
					{
						Name: to.StringPtr(diskName),
					},
				},
			},
		},
	}

	disks := cloud.DisksClient.Disks
	disks[diskName] = compute.Disk{
		Name: to.StringPtr(diskName),
		Tags: clusterTags,
	}
	disks[irrelevantName] = compute.Disk{
		Name: to.StringPtr(irrelevantName),
	}

	ras := cloud.RoleAssignmentsClient.RAs
	ras[raName] = authz.RoleAssignment{
		Name: to.StringPtr(raName),
		RoleAssignmentPropertiesWithScope: &authz.RoleAssignmentPropertiesWithScope{
			Scope:       to.StringPtr("scope"),
			PrincipalID: to.StringPtr(principalID),
		},
	}
	disks[irrelevantName] = compute.Disk{
		Name: to.StringPtr(irrelevantName),
	}

	lbs := cloud.LoadBalancersClient.LBs
	lbs[lbName] = network.LoadBalancer{
		Name: to.StringPtr(lbName),
		Tags: clusterTags,
	}
	lbs[irrelevantName] = network.LoadBalancer{
		Name: to.StringPtr(irrelevantName),
	}

	// Call listResourcesAzure.
	g := resourceGetter{
		cloud: cloud,
		cluster: &kops.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Azure: &kops.AzureSpec{
						ResourceGroupName: rgName,
					},
				},
			},
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
		toKey(typeResourceGroup, rgName): {
			rtype:  typeResourceGroup,
			name:   rgName,
			shared: true,
		},
		toKey(typeVirtualNetwork, vnetName): {
			rtype:  typeVirtualNetwork,
			name:   vnetName,
			blocks: []string{toKey(typeResourceGroup, rgName)},
		},
		toKey(typeSubnet, subnetName): {
			rtype: typeSubnet,
			name:  subnetName,
			blocks: []string{
				toKey(typeVirtualNetwork, vnetName),
				toKey(typeResourceGroup, rgName),
			},
		},
		toKey(typeRouteTable, rtName): {
			rtype:  typeRouteTable,
			name:   rtName,
			blocks: []string{toKey(typeResourceGroup, rgName)},
		},
		toKey(typeVMScaleSet, vmssName): {
			rtype: typeVMScaleSet,
			name:  vmssName,
			blocks: []string{
				toKey(typeResourceGroup, rgName),
				toKey(typeVirtualNetwork, vnetName),
				toKey(typeSubnet, subnetName),
				toKey(typeDisk, diskName),
			},
		},
		toKey(typeDisk, diskName): {
			rtype:  typeDisk,
			name:   diskName,
			blocks: []string{toKey(typeResourceGroup, rgName)},
		},
		toKey(typeRoleAssignment, raName): {
			rtype: typeRoleAssignment,
			name:  raName,
			blocks: []string{
				toKey(typeResourceGroup, rgName),
				toKey(typeVMScaleSet, vmssName),
			},
		},
		toKey(typeLoadBalancer, lbName): {
			rtype:  typeLoadBalancer,
			name:   lbName,
			blocks: []string{toKey(typeResourceGroup, rgName)},
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
				azure.TagClusterName: to.StringPtr(clusterName),
			},
			expected: true,
		},
		{
			tags: map[string]*string{
				azure.TagClusterName: to.StringPtr(clusterName),
				"other-key":          to.StringPtr("other-tag"),
			},
			expected: true,
		},
		{
			tags: map[string]*string{
				"other-key": to.StringPtr("other-tag"),
			},
			expected: false,
		},
		{
			tags: map[string]*string{
				azure.TagClusterName: to.StringPtr("different-cluster"),
			},
			expected: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			g := &resourceGetter{
				cluster: &kops.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: clusterName,
					},
				},
			}
			a := g.isOwnedByCluster(tc.tags)
			if a != tc.expected {
				t.Errorf("expected %t, but got %t", tc.expected, a)
			}
		})
	}
}
