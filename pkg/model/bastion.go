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

package model

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

const BastionELBSecurityGroupPrefix = "bastion"

// BastionModelBuilder adds model objects to support bastions
//
// Bastion instances live in the utility subnets created in the private topology.
// All traffic goes through an ELB, and the ELB has port 22 open to SSHAccess.
// Bastion instances have access to all internal master and node instances.
type BastionModelBuilder struct {
	*KopsModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &BastionModelBuilder{}

// Build creates the security group tasks for a bastion
func (b *BastionModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var bastionGroups []*kops.InstanceGroup
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			bastionGroups = append(bastionGroups, ig)
		}
	}

	if len(bastionGroups) == 0 {
		return nil
	}

	// Create security group for bastion instances
	{
		t := &awstasks.SecurityGroup{
			Name:      s(b.SecurityGroupName(kops.InstanceGroupRoleBastion)),
			Lifecycle: b.Lifecycle,

			VPC:              b.LinkToVPC(),
			Description:      s("Security group for bastion"),
			RemoveExtraRules: []string{"port=22"},
		}
		c.AddTask(t)
	}

	// Allow traffic from bastion instances to egress freely
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("bastion-egress"),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Allow incoming SSH traffic to bastions, through the ELB
	// TODO: Could we get away without an ELB here?  Tricky to fix if dns-controller breaks though...
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("ssh-elb-to-bastion"),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
			SourceGroup:   b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Protocol:      s("tcp"),
			FromPort:      i64(22),
			ToPort:        i64(22),
		}
		c.AddTask(t)
	}

	// Allow bastion nodes to SSH to masters
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("bastion-to-master-ssh"),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
			Protocol:      s("tcp"),
			FromPort:      i64(22),
			ToPort:        i64(22),
		}
		c.AddTask(t)
	}

	// Allow bastion nodes to SSH to nodes
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("bastion-to-node-ssh"),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
			Protocol:      s("tcp"),
			FromPort:      i64(22),
			ToPort:        i64(22),
		}
		c.AddTask(t)
	}

	// Create security group for bastion ELB
	{
		t := &awstasks.SecurityGroup{
			Name:      s(b.ELBSecurityGroupName(BastionELBSecurityGroupPrefix)),
			Lifecycle: b.Lifecycle,

			VPC:              b.LinkToVPC(),
			Description:      s("Security group for bastion ELB"),
			RemoveExtraRules: []string{"port=22"},
		}
		c.AddTask(t)
	}

	// Allow traffic from ELB to egress freely
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("bastion-elb-egress"),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}

		c.AddTask(t)
	}

	// Allow external access to ELB
	for _, sshAccess := range b.Cluster.Spec.SSHAccess {
		t := &awstasks.SecurityGroupRule{
			Name:      s("ssh-external-to-bastion-elb-" + sshAccess),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Protocol:      s("tcp"),
			FromPort:      i64(22),
			ToPort:        i64(22),
			CIDR:          s(sshAccess),
		}
		c.AddTask(t)
	}

	return nil
}
