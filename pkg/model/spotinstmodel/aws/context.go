/*
Copyright 2016 The Kubernetes Authors.

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

package aws

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	awstasks "k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks/aws"
)

type ModelContext struct {
	*model.KopsModelContext
}

func (b *ModelContext) LinkToSecurityGroup(role kops.InstanceGroupRole) *awstasks.SecurityGroup {
	name := b.SecurityGroupName(role)
	return &awstasks.SecurityGroup{Name: &name}
}

func (b *ModelContext) LinkToVPC() *awstasks.VPC {
	name := b.ClusterName()
	return &awstasks.VPC{Name: &name}
}

func (b *ModelContext) LinkToSubnet(z *kops.ClusterSubnetSpec) *awstasks.Subnet {
	name := z.Name + "." + b.ClusterName()
	return &awstasks.Subnet{Name: &name}
}

func (b *ModelContext) LinkToPublicSubnetInZone(zoneName string) (*awstasks.Subnet, error) {
	var matches []*kops.ClusterSubnetSpec
	for i := range b.Cluster.Spec.Subnets {
		z := &b.Cluster.Spec.Subnets[i]
		if z.Zone != zoneName {
			continue
		}
		if z.Type != kops.SubnetTypePublic {
			continue
		}
		matches = append(matches, z)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find public subnet in zone: %q", zoneName)
	}
	if len(matches) > 1 {
		// TODO: Support this (arbitrary choice I think, for ELBs)
		return nil, fmt.Errorf("found multiple public subnets in zone: %q", zoneName)
	}

	return b.LinkToSubnet(matches[0]), nil
}

func (b *ModelContext) LinkToUtilitySubnetInZone(zoneName string) (*awstasks.Subnet, error) {
	var matches []*kops.ClusterSubnetSpec
	for i := range b.Cluster.Spec.Subnets {
		s := &b.Cluster.Spec.Subnets[i]
		if s.Zone != zoneName {
			continue
		}
		if s.Type != kops.SubnetTypeUtility {
			continue
		}
		matches = append(matches, s)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find utility subnet in zone: %q", zoneName)
	}
	if len(matches) > 1 {
		// TODO: Support this
		return nil, fmt.Errorf("found multiple utility subnets in zone: %q", zoneName)
	}

	return b.LinkToSubnet(matches[0]), nil
}

func (b *ModelContext) LinkToELB(prefix string) *awstasks.LoadBalancer {
	name := b.ELBName(prefix)
	return &awstasks.LoadBalancer{Name: &name}
}

func (b *ModelContext) LinkToPrivateRouteTableInZone(zoneName string) *awstasks.RouteTable {
	return &awstasks.RouteTable{Name: fi.String(b.NamePrivateRouteTableInZone(zoneName))}
}

func (b *ModelContext) LinkToDNSZone() *awstasks.DNSZone {
	name := b.NameForDNSZone()
	return &awstasks.DNSZone{Name: &name}
}

func (b *ModelContext) LinkToELBSecurityGroup(prefix string) *awstasks.SecurityGroup {
	name := b.ELBSecurityGroupName(prefix)
	return &awstasks.SecurityGroup{Name: &name}
}

func (b *ModelContext) LinkToIAMInstanceProfile(ig *kops.InstanceGroup) *awstasks.IAMInstanceProfile {
	name := b.IAMName(ig.Spec.Role)
	return &awstasks.IAMInstanceProfile{Name: &name}
}

func (b *ModelContext) LinkToSSHKey() (*awstasks.SSHKey, error) {
	sshKeyName, err := b.SSHKeyName()
	if err != nil {
		return nil, err
	}
	return &awstasks.SSHKey{Name: &sshKeyName}, nil
}

func (b *ModelContext) LinkToAutoscalingGroup(ig *kops.InstanceGroup) *awstasks.AutoscalingGroup {
	name := b.AutoscalingGroupName(ig)
	return &awstasks.AutoscalingGroup{Name: &name}
}

func (b *ModelContext) CloudTags(name string, shared bool) map[string]string {
	tags := make(map[string]string)

	tags[awsup.TagClusterName] = b.ClusterName()
	if name != "" {
		tags["Name"] = name
	}

	if shared {
		tags["kubernetes.io/cluster/"+b.ClusterName()] = "shared"
	} else {
		tags["kubernetes.io/cluster/"+b.ClusterName()] = "owned"
	}

	return tags
}
