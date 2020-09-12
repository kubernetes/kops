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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi"
)

// FindClusterStatus discovers the status of the cluster by looking for the tagged etcd volume.
func (c *azureCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	klog.V(2).Infof("Listing Azure managed disks.")
	disks, err := c.Disk().List(context.TODO(), cluster.AzureResourceGroupName())
	if err != nil {
		return nil, fmt.Errorf("error listing disks: %s", err)
	}

	etcdStatus, err := c.findEtcdStatus(disks)
	if err != nil {
		return nil, err
	}
	status := &kops.ClusterStatus{
		EtcdClusters: etcdStatus,
	}
	klog.V(2).Infof("Cluster status (from cloud): %v", fi.DebugAsJsonString(status))
	return status, nil
}

func (c *azureCloudImplementation) findEtcdStatus(disks []compute.Disk) ([]kops.EtcdClusterStatus, error) {
	statusMap := make(map[string]*kops.EtcdClusterStatus)
	for _, disk := range disks {
		if !c.isDiskForCluster(&disk) {
			continue
		}

		var (
			etcdClusterName string
			etcdClusterSpec *etcd.EtcdClusterSpec
			master          bool
		)
		for k, v := range disk.Tags {
			if k == TagNameRolePrefix+TagRoleMaster {
				master = true
				continue
			}

			if strings.HasPrefix(k, TagNameEtcdClusterPrefix) {
				etcdClusterName = strings.TrimPrefix(k, TagNameEtcdClusterPrefix)
				var err error
				etcdClusterSpec, err = etcd.ParseEtcdClusterSpec(etcdClusterName, *v)
				if err != nil {
					return nil, fmt.Errorf("error parsing etcd cluster tag %q on volume %q: %s", *v, *disk.Name, err)
				}
			}
		}

		if etcdClusterName == "" || etcdClusterSpec == nil || !master {
			continue
		}

		status := statusMap[etcdClusterName]
		if status == nil {
			status = &kops.EtcdClusterStatus{
				Name: etcdClusterName,
			}
			statusMap[etcdClusterName] = status
		}
		status.Members = append(status.Members, &kops.EtcdMemberStatus{
			Name:     etcdClusterSpec.NodeName,
			VolumeId: *disk.Name,
		})
	}

	var status []kops.EtcdClusterStatus
	for _, v := range statusMap {
		status = append(status, *v)
	}
	return status, nil
}

// isDiskForCluster returns true if the managed disk is for the cluster.
func (c *azureCloudImplementation) isDiskForCluster(disk *compute.Disk) bool {
	found := 0
	for k, v := range disk.Tags {
		if c.tags[k] == *v {
			found++
		}
	}
	return found == len(c.tags)
}

// GetCloudGroups returns Cloud Instance Groups for the cluster
// by querying Azure.
func (c *azureCloudImplementation) GetCloudGroups(
	cluster *kops.Cluster,
	instancegroups []*kops.InstanceGroup,
	warnUnmatched bool,
	nodes []v1.Node,
) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	igsByName, err := keyedByName(instancegroups, cluster.Name)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	vmsses, err := c.vmscaleSetsClient.List(ctx, cluster.AzureResourceGroupName())
	if err != nil {
		return nil, fmt.Errorf("unable to find VM Scale Sets: %s", err)
	}

	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	for _, vmss := range vmsses {
		if !isOwnedByCluster(&vmss, cluster.Name) {
			continue
		}

		ig, ok := igsByName[*vmss.Name]
		if !ok {
			if warnUnmatched {
				klog.Warningf("Found VM Scale Set with no corresponding instance group %q", *vmss.Name)
			}
			continue
		}

		cig, err := c.buildCloudInstanceGroup(ctx, cluster, ig, &vmss, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", ig.Name, err)
		}
		groups[ig.Name] = cig
	}
	return groups, nil
}

func (c *azureCloudImplementation) buildCloudInstanceGroup(
	ctx context.Context,
	cluster *kops.Cluster,
	ig *kops.InstanceGroup,
	vmss *compute.VirtualMachineScaleSet,
	nodeMap map[string]*v1.Node,
) (*cloudinstances.CloudInstanceGroup, error) {
	cap := int(*vmss.Sku.Capacity)
	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     *vmss.Name,
		InstanceGroup: ig,
		MinSize:       cap,
		MaxSize:       cap,
		Raw:           vmss,
	}

	// Add members (VMs) to the Cloud Instance Group.
	vms, err := c.vmscaleSetVMsClient.List(ctx, cluster.AzureResourceGroupName(), *vmss.Name)
	if err != nil {
		return nil, fmt.Errorf("error querying VM ScaleSet VMs: %s", err)
	}
	for _, vm := range vms {
		// TODO(kenji): Ignore an instance that is being terminated.

		// TODO(kenji): Set the status properly so that kops can
		// tell whether a VM is up-to-date or not.
		status := cloudinstances.CloudInstanceStatusUpToDate
		_, err := cg.NewCloudInstance(*vm.Name, status, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %s", err)
		}
		// TODO(kenji): Set addCloudInstanceData.
	}

	return cg, nil
}

func isOwnedByCluster(vmss *compute.VirtualMachineScaleSet, clusterName string) bool {
	for k, v := range vmss.Tags {
		if k == TagClusterName && *v == clusterName {
			return true
		}
	}
	return false
}

// keyedByName creates a map of instance groups keyed by VM Scale Set names.
func keyedByName(instancegroups []*kops.InstanceGroup, clusterName string) (map[string]*kops.InstanceGroup, error) {
	m := map[string]*kops.InstanceGroup{}
	for _, ig := range instancegroups {
		var name string
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			name = ig.Name + ".masters." + clusterName
		case kops.InstanceGroupRoleNode, kops.InstanceGroupRoleBastion:
			name = ig.Name + "." + clusterName
		default:
			klog.Warningf("Ignoring InstanceGroup of unknown role %q", ig.Spec.Role)
			continue
		}
		if _, ok := m[name]; ok {
			return nil, fmt.Errorf("found multiple instance groups matching %q", name)
		}
		m[name] = ig
	}

	return m, nil
}
