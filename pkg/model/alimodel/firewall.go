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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
)

const IpProtocolAll = "all"

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*ALIModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {

	// Create nodeInstances security group
	var nodeSecurityGroup *alitasks.SecurityGroup
	{
		groupName := b.GetNameForSecurityGroup(kops.InstanceGroupRoleNode)
		nodeSecurityGroup = &alitasks.SecurityGroup{
			Name:      s(groupName),
			Lifecycle: b.Lifecycle,
			VPC:       b.LinkToVPC(),
		}
		c.AddTask(nodeSecurityGroup)
	}

	// Create masterInstances security group
	var masterSecurityGroup *alitasks.SecurityGroup
	{
		groupName := b.GetNameForSecurityGroup(kops.InstanceGroupRoleMaster)
		masterSecurityGroup = &alitasks.SecurityGroup{
			Name:      s(groupName),
			Lifecycle: b.Lifecycle,
			VPC:       b.LinkToVPC(),
		}
		c.AddTask(masterSecurityGroup)
	}

	// Allow traffic from masters to nodes
	ipProtocolAll := IpProtocolAll
	{
		nodeSecurityGroupRules := &alitasks.SecurityGroupRule{
			Name:          s("node-to-master"),
			Lifecycle:     b.Lifecycle,
			IpProtocol:    s(ipProtocolAll),
			SecurityGroup: nodeSecurityGroup,
			SourceGroup:   masterSecurityGroup,
			PortRange:     s("-1/-1"),
			In:            fi.Bool(true),
		}
		c.AddTask(nodeSecurityGroupRules)
	}

	// Allow traffic from nodes to masters
	{
		masterSecurityGroupRules := &alitasks.SecurityGroupRule{
			Name:          s("node-master"),
			Lifecycle:     b.Lifecycle,
			IpProtocol:    s(ipProtocolAll),
			SecurityGroup: masterSecurityGroup,
			SourceGroup:   nodeSecurityGroup,
			PortRange:     s("-1/-1"),
			In:            fi.Bool(true),
		}
		c.AddTask(masterSecurityGroupRules)
	}

	return nil

}
