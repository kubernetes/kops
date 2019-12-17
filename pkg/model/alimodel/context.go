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

package alimodel

import (
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
)

const CloudTagInstanceGroupRolePrefix = "k8s.io/role/"

type ALIModelContext struct {
	*model.KopsModelContext
}

// LinkToVPC returns the VPC object the cluster is located in
func (c *ALIModelContext) LinkToVPC() *alitasks.VPC {
	return &alitasks.VPC{Name: s(c.GetNameForVPC())}
}

func (c *ALIModelContext) GetNameForVPC() string {
	return c.ClusterName()
}

// LinkToVSwitch returns the VSwitch object the cluster is located in
func (c *ALIModelContext) LinkToVSwitch(subnetName string) *alitasks.VSwitch {
	return &alitasks.VSwitch{Name: s(c.GetNameForVSwitch(subnetName))}
}

func (c *ALIModelContext) GetNameForVSwitch(subnetName string) string {
	return subnetName + "." + c.ClusterName()
}

// LinkToNateGateway returns the NatGateway object the cluster is located in
func (c *ALIModelContext) LinkToNatGateway() *alitasks.NatGateway {
	return &alitasks.NatGateway{Name: s(c.GetNameForNatGateway())}
}

func (c *ALIModelContext) GetNameForNatGateway() string {
	return c.ClusterName()
}

// LinkToEIP returns the EIP object the NateGatway is associated to
func (c *ALIModelContext) LinkToEIP() *alitasks.EIP {
	return &alitasks.EIP{Name: s(c.GetNameForEIP())}
}

func (c *ALIModelContext) GetNameForEIP() string {
	return c.ClusterName()
}

// LinkToVSwitchSNAT returns the VSwitchSNAT object the cluster is located in
func (c *ALIModelContext) LinkToVSwitchSNAT(subnetName string) *alitasks.VSwitch {
	return &alitasks.VSwitch{Name: s(c.GetNameForVSwitch(subnetName))}
}

func (c *ALIModelContext) GetNameForVSwitchSNAT(subnetName string) string {
	return subnetName + "." + c.ClusterName()
}

func (c *ALIModelContext) GetUtilitySubnets() []*kops.ClusterSubnetSpec {
	var subnets []*kops.ClusterSubnetSpec
	for i := range c.Cluster.Spec.Subnets {
		subnet := &c.Cluster.Spec.Subnets[i]
		if subnet.Type == kops.SubnetTypeUtility {
			subnets = append(subnets, subnet)
		}
	}
	return subnets
}

// LinkLoadBalancer returns the LoadBalancer object the cluster is located in
func (c *ALIModelContext) LinkLoadBalancer() *alitasks.LoadBalancer {
	return &alitasks.LoadBalancer{Name: s(c.GetNameForLoadBalancer())}
}

func (c *ALIModelContext) GetNameForLoadBalancer() string {
	return "api." + c.ClusterName()
}

func (c *ALIModelContext) LinkToSSHKey() *alitasks.SSHKey {
	return &alitasks.SSHKey{Name: s(c.GetNameForSSHKey())}
}

func (c *ALIModelContext) GetNameForSSHKey() string {
	return "k8s.sshkey." + c.ClusterName()
}

// LinkToSecurityGroup returns the SecurityGroup with specific name
func (c *ALIModelContext) LinkToSecurityGroup(role kops.InstanceGroupRole) *alitasks.SecurityGroup {
	return &alitasks.SecurityGroup{Name: s(c.GetNameForSecurityGroup(role))}
}

func (c *ALIModelContext) GetNameForSecurityGroup(role kops.InstanceGroupRole) string {
	switch role {
	case kops.InstanceGroupRoleMaster:
		return "masters." + c.ClusterName()
	case kops.InstanceGroupRoleBastion:
		return "bastions." + c.ClusterName()
	case kops.InstanceGroupRoleNode:
		return "nodes." + c.ClusterName()

	default:
		klog.Fatalf("unknown InstanceGroup Role: %q", role)
		return ""
	}
}

func (c *ALIModelContext) LinkToRAMRole(role kops.InstanceGroupRole) *alitasks.RAMRole {
	return &alitasks.RAMRole{Name: s(c.GetNameForRAM(role))}
}

func (c *ALIModelContext) GetNameForRAM(role kops.InstanceGroupRole) string {
	name := ""
	switch role {
	case kops.InstanceGroupRoleMaster:
		name = "masters." + c.ClusterName()
	case kops.InstanceGroupRoleBastion:
		name = "bastions." + c.ClusterName()
	case kops.InstanceGroupRoleNode:
		name = "nodes." + c.ClusterName()

	default:
		klog.Fatalf("unknown InstanceGroup Role: %q", role)
		return ""
	}

	name = strings.Replace(name, ".", "-", -1)
	return name
}

func (c *ALIModelContext) LinkToScalingGroup(ig *kops.InstanceGroup) *alitasks.ScalingGroup {
	return &alitasks.ScalingGroup{Name: s(c.GetScalingGroupName(ig))}
}

func (c *ALIModelContext) GetScalingGroupName(ig *kops.InstanceGroup) string {
	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		// We need to keep this back-compatible, so we introduce the masters name,
		// though the IG name suffices for uniqueness, and with sensible naming masters
		// should be redundant...
		return ig.ObjectMeta.Name[len(ig.ObjectMeta.Name)-3:] + ".masters." + c.ClusterName()
	case kops.InstanceGroupRoleNode:
		return "nodes." + c.ClusterName()
	case kops.InstanceGroupRoleBastion:
		return "bastions." + c.ClusterName()

	default:
		klog.Fatalf("unknown InstanceGroup Role: %v", ig.Spec.Role)
		return ""
	}
}

// CloudTagsForInstanceGroup computes the tags to apply to instances in the specified InstanceGroup
// Copy from context.go, adjust parameters length to meet AliCloud requirements
func (c *ALIModelContext) CloudTagsForInstanceGroup(ig *kops.InstanceGroup) (map[string]string, error) {
	labels := make(map[string]string)

	// Apply any user-specified global labels first so they can be overridden by IG-specific labels
	for k, v := range c.Cluster.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply any user-specified labels
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply labels for cluster autoscaler node labels
	for k, v := range ig.Spec.NodeLabels {
		labels[k] = v
	}

	// Apply labels for cluster autoscaler node taints
	for _, v := range ig.Spec.Taints {
		splits := strings.SplitN(v, "=", 2)
		if len(splits) > 1 {
			labels[splits[0]] = splits[1]
		}
	}

	// The system tags take priority because the cluster likely breaks without them...

	if ig.Spec.Role == kops.InstanceGroupRoleMaster {
		labels[CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleMaster))] = "1"
	}

	if ig.Spec.Role == kops.InstanceGroupRoleNode {
		labels[CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleNode))] = "1"
	}

	if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		labels[CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleBastion))] = "1"
	}

	return labels, nil
}
