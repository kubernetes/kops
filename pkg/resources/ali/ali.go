/*
Copyright 2018 The Kubernetes Authors.

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

package ali

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"

	common "github.com/denverdino/aliyungo/common"
	ecs "github.com/denverdino/aliyungo/ecs"
	ess "github.com/denverdino/aliyungo/ess"
	ram "github.com/denverdino/aliyungo/ram"
	slb "github.com/denverdino/aliyungo/slb"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
)

type aliListFn func() ([]*resources.Resource, error)

const (
	typeScalingGroup      = "ScalingGroup"
	typeLoadBalancer      = "LoadBalancer"
	typeSecurityGroup     = "SecurityGroup"
	typeSecurityGroupRole = "SecurityGroupRole"
	typeVswitch           = "Vswitch"
	typeRamRole           = "RamRole"
	typeVolume            = "Volume"
	typeSSHKey            = "SSHKey"
	typeVPC               = "VPC"
)

type clusterDiscoveryALI struct {
	aliCloud     aliup.ALICloud
	clusterName  string
	scalingGroup map[string]string
}

func ListResourcesALI(aliCloud aliup.ALICloud, clusterName string, region string) (map[string]*resources.Resource, error) {
	if region == "" {
		region = aliCloud.Region()
	}

	resources := make(map[string]*resources.Resource)

	d := clusterDiscoveryALI{
		aliCloud:    aliCloud,
		clusterName: clusterName,
	}
	d.scalingGroup = make(map[string]string)

	// SecurityGroup and VPC should be deleted after ScalingGroup.
	// We list ScalingGroups first to configure dependencies' blocked parameter.
	resourceTrackers, err := d.ListScalingGroups()
	if err != nil {
		return nil, err
	}

	for _, t := range resourceTrackers {
		d.scalingGroup[t.Name] = t.ID
		resources[t.Type+":"+t.ID] = t
	}

	listFunctions := []aliListFn{
		d.ListLoadBalancer,
		d.ListRam,
		d.ListSecurityGroup,
		d.ListSSHKey,
		d.ListVPC,
		d.ListVolume,
	}

	for _, fn := range listFunctions {
		resourceTrackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range resourceTrackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	for k, t := range resources {
		if t.Done {
			delete(resources, k)
		}
	}

	return resources, nil
}

func (d *clusterDiscoveryALI) ListScalingGroups() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	// ScalingGroup Name defined in /pkg/model/alimodel/context.go
	// The should be changed at the same time.
	remove := []string{
		"masters." + d.clusterName,
		"nodes." + d.clusterName,
		"bastions." + d.clusterName,
	}
	pageNumber := 1
	pageSize := 50

	for {
		describeScalingGroupsArgs := &ess.DescribeScalingGroupsArgs{
			RegionId: common.Region(d.aliCloud.Region()),
			Pagination: common.Pagination{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			},
		}
		groups, _, err := d.aliCloud.EssClient().DescribeScalingGroups(describeScalingGroupsArgs)

		if err != nil {
			return nil, fmt.Errorf("error listing ScalingGroup: %v", err)
		}

		scalingGroups := []ess.ScalingGroupItemType{}

		for _, group := range groups {
			for _, r := range remove {
				if strings.Contains(group.ScalingGroupName, r) {
					scalingGroups = append(scalingGroups, group)
				}
			}
		}

		for _, scalingGroup := range scalingGroups {

			resourceTracker := &resources.Resource{
				Name:    scalingGroup.ScalingGroupName,
				ID:      scalingGroup.ScalingGroupId,
				Type:    typeScalingGroup,
				Deleter: DeleteScalingGroup,
			}

			d.scalingGroup[scalingGroup.ScalingGroupName] = scalingGroup.ScalingGroupId
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}

		if len(groups) < pageSize {
			break
		} else {
			pageNumber++
		}
	}

	return resourceTrackers, nil
}

func DeleteScalingGroup(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	glog.V(2).Infof("Removing ScalingGroup with Id %s", r.ID)

	// Force to delete the ScalingGroup
	// TODO: Should we delete the group softly? Like set ForceDelete to false.
	// That will be safer.

	// All the resource of the ScalingGroup will be deleted.
	deleteScalingGroupArgs := &ess.DeleteScalingGroupArgs{
		ScalingGroupId: r.ID,
		ForceDelete:    true,
	}

	_, err := c.EssClient().DeleteScalingGroup(deleteScalingGroupArgs)
	if err != nil {
		return fmt.Errorf("error deleting ScalingGroup: %v", err)
	}

	return nil
}

func (d *clusterDiscoveryALI) ListLoadBalancer() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	// LoadBalancer has the cluster tag: KubernetesCluster:$ClusterName
	// Find the cluster loadBalancer by this key value pair.
	// TODO: Should we check the tags of loadbalancer?

	// ScalingGroup Name defined in /pkg/model/alimodel/context.go
	// The should be changed at the same time.
	// ScalingGroup can not be renamed in alicloud, so we recognize by name
	loadBalancerName := "api." + d.clusterName

	describeLoadBalancersArgs := &slb.DescribeLoadBalancersArgs{
		RegionId:         common.Region(d.aliCloud.Region()),
		LoadBalancerName: loadBalancerName,
	}
	loadBalancers, err := d.aliCloud.SlbClient().DescribeLoadBalancers(describeLoadBalancersArgs)
	if err != nil {
		return nil, fmt.Errorf("err listing LoadBalancers")
	}

	if len(loadBalancers) > 1 {
		return nil, fmt.Errorf("found multiple LoadBalancers with the same name: %v", loadBalancerName)
	}

	if len(loadBalancers) == 1 {
		resourceTracker := &resources.Resource{
			Name:    loadBalancers[0].LoadBalancerName,
			ID:      loadBalancers[0].LoadBalancerId,
			Type:    typeLoadBalancer,
			Deleter: DeleteLoadBalancer,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DeleteLoadBalancer(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	glog.V(2).Infof("Removing LoadBalancer with Id %s", r.ID)

	err := c.SlbClient().DeleteLoadBalancer(r.ID)
	if err != nil {
		return fmt.Errorf("err deleting LoadBalancer:%v", err)
	}
	return nil
}

func (d *clusterDiscoveryALI) ListSecurityGroup() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	// SecurityGroup's name can be renamed in alicloud.
	// We recognize SecurityGroup with its name and clusterTags
	tags := make(map[string]string)
	d.aliCloud.AddClusterTags(tags)

	securityGroupNames := []string{
		"masters." + d.clusterName,
		"bastions." + d.clusterName,
		"nodes." + d.clusterName,
	}

	describeResourceByTagsArgs := &ecs.DescribeResourceByTagsArgs{
		ResourceType: ecs.TagResourceType(alitasks.SecurityResource),
		RegionId:     common.Region(d.aliCloud.Region()),
		Tag:          tags,
	}

	resourceList, _, err := d.aliCloud.EcsClient().DescribeResourceByTags(describeResourceByTagsArgs)
	if err != nil {
		return nil, fmt.Errorf("err listing securityGroup:%v", err)
	}

	blocked := []string{}
	groupTrackers := []*resources.Resource{}

	if len(resourceList) != 0 {
		for _, resource := range resourceList {
			find := false

			describeSecurityGroupAttributeArgs := &ecs.DescribeSecurityGroupAttributeArgs{
				SecurityGroupId: resource.ResourceId,
				RegionId:        resource.RegionId,
			}
			respGroup, err := d.aliCloud.EcsClient().DescribeSecurityGroupAttribute(describeSecurityGroupAttributeArgs)
			if err != nil {
				return nil, fmt.Errorf("err listing securityGroup:%v", err)
			}

			for _, value := range securityGroupNames {
				if respGroup.SecurityGroupName == value {
					find = true
				}
			}

			if find {
				groupTracker := &resources.Resource{
					Name:    respGroup.SecurityGroupName,
					ID:      respGroup.SecurityGroupId,
					Type:    typeSecurityGroup,
					Deleter: DeleteSecurityGroup,
				}
				groupTrackers = append(groupTrackers, groupTracker)

				if len(respGroup.Permissions.Permission) != 0 {
					for i, securityGroupRole := range respGroup.Permissions.Permission {
						roleTracker := &resources.Resource{
							Name:    respGroup.SecurityGroupId,
							ID:      respGroup.SecurityGroupName + strconv.Itoa(i),
							Type:    typeSecurityGroupRole,
							Deleter: DeleteSecurityGroupRole,
							Obj:     securityGroupRole,
						}
						resourceTrackers = append(resourceTrackers, roleTracker)

						// Any SecurityGroupRole which depends on the current SecurityGroup should be deleted before delete the SecurityGroup.
						// So before deleting any SecurityGroup, we should delete all SecurityGroupRoles of all SecurityGroups.
						blocked = append(blocked, roleTracker.Type+":"+roleTracker.ID)
					}
				}

			}
		}
	}

	if len(groupTrackers) != 0 {
		for _, groupTracker := range groupTrackers {

			for scalingGroupName, scalingGroupId := range d.scalingGroup {
				if strings.Contains(scalingGroupName, groupTracker.Name) {
					groupTracker.Blocked = append(groupTracker.Blocked, typeScalingGroup+":"+scalingGroupId)
				}
			}

			for _, block := range blocked {
				groupTracker.Blocked = append(groupTracker.Blocked, block)
			}
			resourceTrackers = append(resourceTrackers, groupTracker)
		}
	}
	return resourceTrackers, nil
}

func DeleteSecurityGroup(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	region := common.Region(c.Region())
	glog.V(2).Infof("Removing SecurityGroup with Id %s", r.ID)

	err := c.EcsClient().DeleteSecurityGroup(region, r.ID)
	if err != nil {
		return fmt.Errorf("err deleting securityGroup:%v", err)
	}
	return nil
}

func DeleteSecurityGroupRole(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	permission := r.Obj.(ecs.PermissionType)

	glog.V(2).Infof("Removing SecurityGroupRole of SecurityGroup %s", r.Name)

	if permission.Direction == "ingress" {
		authorizeSecurityGroupArgs := ecs.AuthorizeSecurityGroupArgs{
			SecurityGroupId:         r.Name,
			RegionId:                common.Region(c.Region()),
			IpProtocol:              permission.IpProtocol,
			PortRange:               permission.PortRange,
			SourceGroupId:           permission.SourceGroupId,
			SourceGroupOwnerAccount: permission.SourceGroupOwnerAccount,
			SourceCidrIp:            permission.SourceCidrIp,
			Policy:                  permission.Policy,
			Priority:                permission.Priority,
			NicType:                 permission.NicType,
		}

		revokeSecurityGroupArgs := &ecs.RevokeSecurityGroupArgs{
			AuthorizeSecurityGroupArgs: authorizeSecurityGroupArgs,
		}

		err := c.EcsClient().RevokeSecurityGroup(revokeSecurityGroupArgs)
		if err != nil {
			return fmt.Errorf("err deleting securityGroup Role:%v", err)
		}
	}

	if permission.Direction == "egress" {
		authorizeSecurityGroupEgressArgs := ecs.AuthorizeSecurityGroupEgressArgs{
			SecurityGroupId:       r.Name,
			RegionId:              common.Region(c.Region()),
			IpProtocol:            permission.IpProtocol,
			PortRange:             permission.PortRange,
			DestGroupId:           permission.DestGroupId,
			DestGroupOwnerAccount: permission.DestGroupOwnerAccount,
			DestCidrIp:            permission.DestCidrIp,
			Policy:                permission.Policy,
			Priority:              permission.Priority,
			NicType:               permission.NicType,
		}

		revokeSecurityGroupEgressArgs := &ecs.RevokeSecurityGroupEgressArgs{
			AuthorizeSecurityGroupEgressArgs: authorizeSecurityGroupEgressArgs,
		}

		err := c.EcsClient().RevokeSecurityGroupEgress(revokeSecurityGroupEgressArgs)
		if err != nil {
			return fmt.Errorf("err deleting securityGroup Role:%v", err)
		}
	}

	return nil
}

func (d *clusterDiscoveryALI) ListRam() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	clusterName := strings.Replace(d.clusterName, ".", "-", -1)
	names := []string{
		"masters-" + clusterName,
		"nodes-" + clusterName,
		"bastions-" + clusterName,
	}

	roleToDelete := []string{}

	response, err := d.aliCloud.RamClient().ListRoles()
	if err != nil {
		return nil, fmt.Errorf("err listing RamRole:%v", err)

	}

	if len(response.Roles.Role) != 0 {
		for _, role := range response.Roles.Role {
			for _, roleName := range names {
				if role.RoleName == roleName {
					roleToDelete = append(roleToDelete, role.RoleId)
					resourceTracker := &resources.Resource{
						Name:    role.RoleName,
						ID:      role.RoleId,
						Type:    typeRamRole,
						Deleter: DeleteRoleRam,
					}
					resourceTrackers = append(resourceTrackers, resourceTracker)

				}
			}
		}
	}

	return resourceTrackers, nil
}

func DeleteRoleRam(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	policies := []string{}

	glog.V(2).Infof("Removing RamRole  %s", r.Name)

	roleQueryRequest := ram.RoleQueryRequest{
		RoleName: r.Name,
	}
	response, err := c.RamClient().ListPoliciesForRole(roleQueryRequest)
	if err != nil {
		return fmt.Errorf("err listing Policices for role:%v", err)
	} else {
		if len(response.Policies.Policy) != 0 {
			for _, policy := range response.Policies.Policy {
				policies = append(policies, policy.PolicyName)
			}
		}
	}

	for _, policy := range policies {
		glog.V(2).Infof("Removing RolePolicy %s of RamRole %s", policy, r.Name)

		policyRequest := ram.PolicyRequest{
			PolicyName: policy,
			PolicyType: ram.Custom,
		}

		attachPolicyToRoleRequest := ram.AttachPolicyToRoleRequest{
			PolicyRequest: policyRequest,
			RoleName:      r.Name,
		}
		_, err := c.RamClient().DetachPolicyFromRole(attachPolicyToRoleRequest)
		if err != nil {
			return fmt.Errorf("err detaching policy from role:%v", err)
		}

		_, err = c.RamClient().DeletePolicy(policyRequest)
		if err != nil {
			return fmt.Errorf("err deleting policy:%v", err)
		}
	}

	_, err = c.RamClient().DeleteRole(roleQueryRequest)
	if err != nil {
		return fmt.Errorf("err deleting ram role:%v", err)
	}

	return nil
}

func (d *clusterDiscoveryALI) ListSSHKey() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	name := "k8s.sshkey." + d.clusterName

	resourceTracker := &resources.Resource{
		Name:    name,
		ID:      name,
		Type:    typeSSHKey,
		Deleter: DeleteSSHKey,
	}

	resourceTrackers = append(resourceTrackers, resourceTracker)
	return resourceTrackers, nil
}

func DeleteSSHKey(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	region := common.Region(c.Region())
	glog.V(2).Infof("Removing SSHKsy %s", r.Name)

	deleteKeyPairsArgs := &ecs.DeleteKeyPairsArgs{
		RegionId: region,
	}
	keyPairs := []string{r.Name}
	keyPairsJson, _ := json.Marshal(keyPairs)
	deleteKeyPairsArgs.KeyPairNames = string(keyPairsJson)

	err := c.EcsClient().DeleteKeyPairs(deleteKeyPairsArgs)
	if err != nil {
		return fmt.Errorf("err deleting sshkey pairs:%v", err)
	}

	return nil
}

func (d *clusterDiscoveryALI) ListVPC() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	// Delete VPC with specified name. All of the Switches will be deleted.
	// We think the VPC which owns the designated name is owned.
	name := d.clusterName
	vpcsToDelete := []string{}
	vswitchsToDelete := []string{}

	pageNumber := 1
	pageSize := 50
	for {
		describeVpcsArgs := &ecs.DescribeVpcsArgs{
			RegionId: common.Region(d.aliCloud.Region()),
			Pagination: common.Pagination{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			},
		}

		vpcs, _, err := d.aliCloud.EcsClient().DescribeVpcs(describeVpcsArgs)
		if err != nil {
			return nil, fmt.Errorf("err listing VPC:%v", err)
		}

		if len(vpcs) != 0 {
			for _, vpc := range vpcs {
				if name == vpc.VpcName {
					vpcsToDelete = append(vpcsToDelete, vpc.VpcId)
					for _, vswitch := range vpc.VSwitchIds.VSwitchId {
						vswitchsToDelete = append(vswitchsToDelete, vswitch)
					}
				}
			}
		}
		if len(vpcs) < pageSize {
			break
		} else {
			pageNumber++
		}
	}

	if len(vpcsToDelete) > 1 {
		glog.V(8).Infof("Found multiple vpcs with name %q", name)
	} else if len(vpcsToDelete) == 1 {
		vpcTracker := &resources.Resource{
			Name:    name,
			ID:      vpcsToDelete[0],
			Type:    typeVPC,
			Deleter: DeleteVPC,
		}
		resourceTrackers = append(resourceTrackers, vpcTracker)

		for _, vswitchId := range vswitchsToDelete {

			resourceTracker := &resources.Resource{
				Name:    name,
				ID:      vswitchId,
				Type:    typeVswitch,
				Deleter: DeleteVswitch,
			}
			resourceTracker.Blocks = append(resourceTracker.Blocks, typeVPC+":"+vpcTracker.ID)

			//Waiting for all autoScalingGroups to be deleted.
			for _, scalingGroupId := range d.scalingGroup {
				resourceTracker.Blocked = append(resourceTracker.Blocked, typeScalingGroup+":"+scalingGroupId)
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}

func DeleteVswitch(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	glog.V(2).Infof("Removing Vswitch with Id %s", r.ID)

	err := c.EcsClient().DeleteVSwitch(r.ID)
	if err != nil {
		return fmt.Errorf("err deleting Vswitch:%v", err)
	}
	return nil
}

func DeleteVPC(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	glog.V(2).Infof("Removing VPC with Id %s", r.ID)

	err := c.EcsClient().DeleteVpc(r.ID)
	if err != nil {
		return fmt.Errorf("err deleting VPC:%v", err)
	}
	return nil
}

func (d *clusterDiscoveryALI) ListVolume() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	// Volume has cluster tags
	tags := make(map[string]string)
	d.aliCloud.AddClusterTags(tags)
	disksToDelete := []string{}

	pageNumber := 1
	pageSize := 50
	for {
		describeDisksArgs := &ecs.DescribeDisksArgs{
			RegionId: common.Region(d.aliCloud.Region()),
			Tag:      tags,
			DiskType: ecs.DiskTypeAllData,
			Pagination: common.Pagination{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			},
		}

		disks, _, err := d.aliCloud.EcsClient().DescribeDisks(describeDisksArgs)
		if err != nil {
			return nil, fmt.Errorf("err listing disks:%v", err)
		}

		if len(disks) != 0 {
			for _, disk := range disks {
				disksToDelete = append(disksToDelete, disk.DiskId)
			}
		}
		if len(disks) < pageSize {
			break
		} else {
			pageNumber++
		}
	}

	if len(disksToDelete) == 0 {
		glog.V(8).Infof("Found no disks to delete")
	} else {
		for _, disk := range disksToDelete {

			resourceTracker := &resources.Resource{
				Name:    disk,
				ID:      disk,
				Type:    typeVolume,
				Deleter: DeleteVolume,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil

}

func DeleteVolume(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(aliup.ALICloud)
	glog.V(2).Infof("Removing Disk with Id %s", r.ID)

	err := c.EcsClient().DeleteDisk(r.ID)
	if err != nil {
		return fmt.Errorf("err deleting volume:%v", err)
	}
	return nil
}
