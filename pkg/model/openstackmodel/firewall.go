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

package openstackmodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

const DirectionEgress = "egress"
const DirectionIngress = "ingress"
const IpProtocolTCP = "tcp"
const IpProtocolUDP = "udp"
const IPV4 = "IPv4"

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*OpenstackModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {

	for _, role := range []kops.InstanceGroupRole{kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleNode, kops.InstanceGroupRoleBastion} {

		// Create Security Group for Role
		groupName := b.SecurityGroupName(role)
		sg := &openstacktasks.SecurityGroup{
			Name:      s(groupName),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(sg)

		//Allow local traffic
		localTCP := &openstacktasks.SecurityGroupRule{
			Lifecycle: b.Lifecycle,
			SecGroup:  sg,
			Direction: s(DirectionIngress),
			Protocol:  s(IpProtocolTCP),
			EtherType: s(IPV4),

			PortRangeMin: i(1),
			PortRangeMax: i(65535),

			RemoteIPPrefix: s(b.Cluster.Spec.NetworkCIDR),
		}
		c.AddTask(localTCP)

		// Add SSH Rules
		if b.UsesSSHBastion() {
			if role == kops.InstanceGroupRoleBastion {
				for _, sshAccess := range b.Cluster.Spec.SSHAccess {
					t := &openstacktasks.SecurityGroupRule{
						Lifecycle: b.Lifecycle,
						SecGroup:  sg,
						Direction: s(DirectionIngress),
						Protocol:  s(IpProtocolTCP),
						EtherType: s(IPV4),

						PortRangeMin: i(22),
						PortRangeMax: i(22),

						RemoteIPPrefix: s(sshAccess),
					}
					c.AddTask(t)
				}
			}
		} else {
			for _, sshAccess := range b.Cluster.Spec.SSHAccess {
				t := &openstacktasks.SecurityGroupRule{
					Lifecycle: b.Lifecycle,
					SecGroup:  sg,
					Direction: s(DirectionIngress),
					Protocol:  s(IpProtocolTCP),
					EtherType: s(IPV4),

					PortRangeMin: i(22),
					PortRangeMax: i(22),

					RemoteIPPrefix: s(sshAccess),
				}
				c.AddTask(t)
			}
		}

		// Add NodePort Rules:
		if role == kops.InstanceGroupRoleNode {
			for _, nodePortAccess := range b.Cluster.Spec.NodePortAccess {

				nodePortRange, err := b.NodePortRange()
				if err != nil {
					return err
				}

				for _, protocol := range []string{IpProtocolTCP, IpProtocolUDP} {
					t := &openstacktasks.SecurityGroupRule{
						Lifecycle: b.Lifecycle,
						SecGroup:  sg,
						Direction: s(DirectionIngress),
						Protocol:  s(protocol),
						EtherType: s(IPV4),

						PortRangeMin: i(nodePortRange.Base),
						PortRangeMax: i(nodePortRange.Base + nodePortRange.Size - 1),

						RemoteIPPrefix: s(nodePortAccess),
					}
					c.AddTask(t)
				}
			}
		} else if role == kops.InstanceGroupRoleMaster {
			for _, apiAccess := range b.Cluster.Spec.NodePortAccess {
				t := &openstacktasks.SecurityGroupRule{
					Lifecycle: b.Lifecycle,
					SecGroup:  sg,
					Direction: s(DirectionIngress),
					Protocol:  s(IpProtocolTCP),
					EtherType: s(IPV4),

					PortRangeMin: i(443),
					PortRangeMax: i(443),

					RemoteIPPrefix: s(apiAccess),
				}
				c.AddTask(t)
			}
		}

	}

	return nil

}
