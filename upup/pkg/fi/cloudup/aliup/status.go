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

package aliup

import (
	"errors"
	"fmt"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/ess"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi"
)

// FindClusterStatus discovers the status of the cluster, by looking for the tagged etcd volumes
func (c *aliCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	etcdStatus, err := findEtcdStatus(c, cluster)
	if err != nil {
		return nil, err
	}
	status := &kops.ClusterStatus{
		EtcdClusters: etcdStatus,
	}
	klog.V(2).Infof("Cluster status (from cloud): %v", fi.DebugAsJsonString(status))
	return status, nil
}

// findEtcdStatus discovers the status of etcd, by looking for the tagged etcd volumes
func findEtcdStatus(c ALICloud, cluster *kops.Cluster) ([]kops.EtcdClusterStatus, error) {
	klog.V(2).Infof("Querying ALI for etcd volumes")
	statusMap := make(map[string]*kops.EtcdClusterStatus)

	maxPageSize := 50
	tags := c.GetClusterTags()
	var disks []ecs.DiskItemType

	describeDisksArgs := &ecs.DescribeDisksArgs{
		RegionId: common.Region(c.Region()),
		Tag:      tags,
		Pagination: common.Pagination{
			PageNumber: 1,
			PageSize:   maxPageSize,
		},
	}

	for {
		resp, page, err := c.EcsClient().DescribeDisks(describeDisksArgs)
		if err != nil {
			return nil, fmt.Errorf("error describing disks: %v", err)
		}
		disks = append(disks, resp...)

		if page.NextPage() == nil {
			break
		}
		describeDisksArgs.Pagination = *(page.NextPage())
	}

	// Don't exist disk with specified ClusterTags.
	if len(disks) == 0 {
		return nil, nil
	}

	for _, disk := range disks {

		etcdClusterName := ""
		var etcdClusterSpec *etcd.EtcdClusterSpec
		master := false

		describeTagsArgs := &ecs.DescribeTagsArgs{
			RegionId:     common.Region(c.Region()),
			ResourceType: ecs.TagResourceDisk,
			ResourceId:   disk.DiskId,
		}
		tags, _, err := c.EcsClient().DescribeTags(describeTagsArgs)
		if err != nil {
			return nil, fmt.Errorf("error querying Aliyun disk tags: %v", err)
		}

		for _, tag := range tags {

			k := tag.TagKey
			v := tag.TagValue

			if strings.HasPrefix(k, TagNameEtcdClusterPrefix) {
				etcdClusterName := strings.TrimPrefix(k, TagNameEtcdClusterPrefix)
				etcdClusterSpec, err = etcd.ParseEtcdClusterSpec(etcdClusterName, v)
				if err != nil {
					return nil, fmt.Errorf("error parsing etcd cluster tag %q on volume %q: %v", v, disk.DiskId, err)
				}
			} else if k == TagNameRolePrefix+TagRoleMaster {
				master = true
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

		memberName := etcdClusterSpec.NodeName
		status.Members = append(status.Members, &kops.EtcdMemberStatus{
			Name:     memberName,
			VolumeId: disk.DiskId,
		})
	}

	var status []kops.EtcdClusterStatus
	for _, v := range statusMap {
		status = append(status, *v)
	}
	return status, nil
}

func (c *aliCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

func getCloudGroups(c ALICloud, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	asgs, err := FindAutoscalingGroups(c)
	if err != nil {
		return nil, fmt.Errorf("unable to find autoscale groups: %v", err)
	}

	for _, asg := range asgs {
		name := asg.ScalingGroupName

		instancegroup, err := matchInstanceGroup(name, cluster.ObjectMeta.Name, instancegroups)
		if err != nil {
			return nil, fmt.Errorf("error getting instance group for ASG %q", name)
		}
		if instancegroup == nil {
			if warnUnmatched {
				klog.Warningf("Found ASG with no corresponding instance group %q", name)
			}
			continue
		}

		groups[instancegroup.ObjectMeta.Name], err = buildCloudInstanceGroup(c, instancegroup, asg, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", instancegroup.ObjectMeta.Name, err)
		}
	}

	return groups, nil

}

// FindAutoscalingGroups finds autoscaling groups matching the specified tags
// This isn't entirely trivial because autoscaling doesn't let us filter with as much precision as we would like
func FindAutoscalingGroups(c ALICloud) ([]ess.ScalingGroupItemType, error) {
	var sgs []ess.ScalingGroupItemType
	var relsult []ess.ScalingGroupItemType
	var clusterName string

	clusterName, ok := c.GetClusterTags()[TagClusterName]
	if !ok {
		return nil, errors.New("error describing ScalingGroups:can not get clusterName")
	}

	klog.V(2).Infof("Listing all Autoscaling groups matching clusterName")

	request := &ess.DescribeScalingGroupsArgs{
		RegionId: common.Region(c.Region()),
	}
	for {
		resp, page, err := c.EssClient().DescribeScalingGroups(request)
		if err != nil {
			return nil, fmt.Errorf("error describing ScalingGroups: %v", err)
		}
		sgs = append(sgs, resp...)

		if page.NextPage() == nil {
			break
		}
		request.Pagination = *(page.NextPage())
	}
	for _, sg := range sgs {
		if strings.HasSuffix(sg.ScalingGroupName, clusterName) {
			relsult = append(relsult, sg)
		}
	}

	return relsult, nil
}

func matchInstanceGroup(name string, clusterName string, instancegroups []*kops.InstanceGroup) (*kops.InstanceGroup, error) {
	var instancegroup *kops.InstanceGroup
	for _, g := range instancegroups {
		var groupName string
		switch g.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			groupName = g.ObjectMeta.Name[len(g.ObjectMeta.Name)-3:] + ".masters." + clusterName
		case kops.InstanceGroupRoleNode:
			groupName = g.ObjectMeta.Name + "." + clusterName
		case kops.InstanceGroupRoleBastion:
			groupName = g.ObjectMeta.Name + "." + clusterName
		default:
			klog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
			continue
		}

		if name == groupName {
			if instancegroup != nil {
				return nil, fmt.Errorf("found multiple instance groups matching ASG %q", groupName)
			}
			instancegroup = g
		}
	}

	return instancegroup, nil
}

func buildCloudInstanceGroup(c ALICloud, ig *kops.InstanceGroup, g ess.ScalingGroupItemType, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	newLaunchConfigName := g.ActiveScalingConfigurationId
	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     g.ScalingGroupName,
		InstanceGroup: ig,
		MinSize:       g.MinSize,
		MaxSize:       g.MaxSize,
		Raw:           g,
	}

	var instances []ess.ScalingInstanceItemType
	request := &ess.DescribeScalingInstancesArgs{
		RegionId:       common.Region(c.Region()),
		ScalingGroupId: g.ScalingGroupId,
	}
	for {
		resp, page, err := c.EssClient().DescribeScalingInstances(request)
		if err != nil {
			return nil, fmt.Errorf("error describing ScalingGroups: %v", err)
		}
		instances = append(instances, resp...)

		if page.NextPage() == nil {
			break
		}
		request.Pagination = *(page.NextPage())
	}

	for _, i := range instances {
		instanceId := i.InstanceId
		if instanceId == "" {
			klog.Warningf("ignoring instance with no instance id: %s", i)
			continue
		}
		err := cg.NewCloudInstanceGroupMember(instanceId, newLaunchConfigName, i.ScalingConfigurationId, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}
	}

	return cg, nil
}
