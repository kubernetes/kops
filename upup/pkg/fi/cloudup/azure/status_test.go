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
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
)

type mockVMScaleSetsClient struct {
	vmsses []compute.VirtualMachineScaleSet
}

var _ VMScaleSetsClient = &mockVMScaleSetsClient{}

// CreateOrUpdate creates or updates a VM Scale Set.
func (c *mockVMScaleSetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName, vmScaleSetName string, parameters compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (c *mockVMScaleSetsClient) List(ctx context.Context, resourceGroupName string) ([]compute.VirtualMachineScaleSet, error) {
	return c.vmsses, nil
}

func (c *mockVMScaleSetsClient) Delete(ctx context.Context, resourceGroupName, vmssName string) error {
	return fmt.Errorf("unimplemented")
}

type mockVMScaleSetVMsClient struct {
	vms []compute.VirtualMachineScaleSetVM
}

var _ VMScaleSetVMsClient = &mockVMScaleSetVMsClient{}

func (c *mockVMScaleSetVMsClient) List(ctx context.Context, resourceGroupName, vmssName string) ([]compute.VirtualMachineScaleSetVM, error) {
	return c.vms, nil
}

func TestFindEtcdStatus(t *testing.T) {
	clusterName := "my-cluster"
	c := &azureCloudImplementation{
		tags: map[string]string{
			TagClusterName: clusterName,
		},
	}

	etcdClusterName := "main"
	disks := []compute.Disk{
		{
			Name: to.StringPtr("d0"),
			Tags: map[string]*string{
				TagClusterName:                             to.StringPtr(clusterName),
				TagNameRolePrefix + TagRoleMaster:          to.StringPtr("1"),
				TagNameEtcdClusterPrefix + etcdClusterName: to.StringPtr("a/a,b,c"),
			},
		},
		{
			Name: to.StringPtr("d1"),
			Tags: map[string]*string{
				TagClusterName:                             to.StringPtr(clusterName),
				TagNameRolePrefix + TagRoleMaster:          to.StringPtr("1"),
				TagNameEtcdClusterPrefix + etcdClusterName: to.StringPtr("b/a,b,c"),
			},
		},
		{
			Name: to.StringPtr("d2"),
			Tags: map[string]*string{
				TagClusterName:                             to.StringPtr(clusterName),
				TagNameRolePrefix + TagRoleMaster:          to.StringPtr("1"),
				TagNameEtcdClusterPrefix + etcdClusterName: to.StringPtr("c/a,b,c"),
			},
		},
		{
			// No etcd tag.
			Name: to.StringPtr("not_relevant"),
			Tags: map[string]*string{
				TagClusterName: to.StringPtr("different_cluster"),
			},
		},
		{
			// No corresponding cluster tag.
			Name: to.StringPtr("not_relevant"),
			Tags: map[string]*string{
				TagClusterName: to.StringPtr("different_cluster"),
			},
		},
	}
	etcdClusters, err := c.findEtcdStatus(disks)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if len(etcdClusters) != 1 {
		t.Fatalf("unexpected number of etcd clusters: %d", len(etcdClusters))
	}
	etcdCluster := etcdClusters[0]
	if a, e := "main", etcdCluster.Name; a != e {
		t.Errorf("expected %s, but got %s", e, a)
	}

	actual := map[string]*kops.EtcdMemberStatus{}
	for _, m := range etcdCluster.Members {
		actual[m.Name] = m
	}
	expected := map[string]*kops.EtcdMemberStatus{
		"a": {
			Name:     "a",
			VolumeID: "d0",
		},
		"b": {
			Name:     "b",
			VolumeID: "d1",
		},
		"c": {
			Name:     "c",
			VolumeID: "d2",
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, but got %+v", actual, expected)
	}
}

func TestGetCloudGroups(t *testing.T) {
	const (
		clusterName = "my-cluster"

		masterIG   = "master-eastus-1"
		masterVMSS = "master-eastus-1.masters.my-cluster"
		masterVM   = "master-eastus-1.masters.my-cluster_0"

		nodeIG   = "nodes"
		nodeVMSS = "nodes.my-cluster"
		nodeVM0  = "nodes.my-cluster_0"
		nodeVM1  = "nodes.my-cluster_1"
	)

	vmssClient := &mockVMScaleSetsClient{}
	vmssClient.vmsses = append(vmssClient.vmsses,
		compute.VirtualMachineScaleSet{
			Name: to.StringPtr(masterVMSS),
			Tags: map[string]*string{
				TagClusterName: to.StringPtr(clusterName),
			},
			Sku: &compute.Sku{
				Capacity: to.Int64Ptr(1),
			},
		},
		compute.VirtualMachineScaleSet{
			Name: to.StringPtr(nodeVMSS),
			Tags: map[string]*string{
				TagClusterName: to.StringPtr(clusterName),
			},
			Sku: &compute.Sku{
				Capacity: to.Int64Ptr(2),
			},
		},
	)

	vmClient := &mockVMScaleSetVMsClient{}
	vmClient.vms = append(vmClient.vms,
		compute.VirtualMachineScaleSetVM{
			Name: to.StringPtr(masterVM),
		},
		compute.VirtualMachineScaleSetVM{
			Name: to.StringPtr(nodeVM0),
		},
		compute.VirtualMachineScaleSetVM{
			Name: to.StringPtr(nodeVM1),
		},
	)

	c := &azureCloudImplementation{
		tags: map[string]string{
			TagClusterName: clusterName,
		},
		vmscaleSetsClient:   vmssClient,
		vmscaleSetVMsClient: vmClient,
	}

	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Azure: &kops.AzureSpec{
					ResourceGroupName: "my-rg",
				},
			},
		},
	}

	instancegroups := []*kops.InstanceGroup{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: masterIG,
			},
			Spec: kops.InstanceGroupSpec{
				Role: kops.InstanceGroupRoleMaster,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeIG,
			},
			Spec: kops.InstanceGroupSpec{
				Role: kops.InstanceGroupRoleNode,
			},
		},
	}
	nodes := []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-eastus-1000000",
			},
			Spec: v1.NodeSpec{
				ProviderID: "azure:///subscriptions/<subscription ID>/resourceGroups/my-group/providers/Microsoft.Compute/virtualMachineScaleSets/master-eastus-1.my-cluster.k8s.local/virtualMachines/0",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nodes000000",
			},
			Spec: v1.NodeSpec{
				ProviderID: "azure:///subscriptions/<subscription ID>/resourceGroups/my-group/providers/Microsoft.Compute/virtualMachineScaleSets/nodes.my-cluster.k8s.local/virtualMachines/0",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nodes000001",
			},
			Spec: v1.NodeSpec{
				ProviderID: "azure:///subscriptions/<subscription ID>/resourceGroups/my-group/providers/Microsoft.Compute/virtualMachineScaleSets/nodes.my-cluster.k8s.local/virtualMachines/1",
			},
		},
	}

	groups, err := c.GetCloudGroups(cluster, instancegroups, false /* warnUnmatched */, nodes)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := len(groups), 2; a != e {
		t.Fatalf("expected %d group(s), but found %d groups", e, a)
	}

	group := groups[masterIG]
	if a, e := group.HumanName, masterVMSS; a != e {
		t.Fatalf("expected name %s, but got %s", e, a)
	}
	if a, e := group.InstanceGroup, instancegroups[0]; a != e {
		t.Fatalf("expected instance group %+v, but got %+v", e, a)
	}
	if a, e := group.MinSize, 1; a != e {
		t.Fatalf("expected min size %d, but got %d", e, a)
	}
	if a, e := group.MaxSize, 1; a != e {
		t.Fatalf("expected min size %d, but got %d", e, a)
	}

	group = groups[nodeIG]
	if a, e := group.HumanName, nodeVMSS; a != e {
		t.Fatalf("expected name %s, but got %s", e, a)
	}
	if a, e := group.InstanceGroup, instancegroups[1]; a != e {
		t.Fatalf("expected instance group %+v, but got %+v", e, a)
	}
	if a, e := group.MinSize, 2; a != e {
		t.Fatalf("expected min size %d, but got %d", e, a)
	}
	if a, e := group.MaxSize, 2; a != e {
		t.Fatalf("expected min size %d, but got %d", e, a)
	}
}
