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
	"strconv"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
)

const (
	IpProtocolTCP = "tcp"
	IpProtocolUDP = "udp"
)

// ExternalAccessModelBuilder configures security group rules for external access
// (SSHAccess, KubernetesAPIAccess)
type ExternalAccessModelBuilder struct {
	*ALIModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *ExternalAccessModelBuilder) Build(c *fi.ModelBuilderContext) error {

	if len(b.Cluster.Spec.SSHAccess) == 0 {
		klog.Warningf("SSHAccess is empty")
	}

	// SSH is open to AdminCIDR set
	if b.UsesSSHBastion() {
		// If we are using a bastion, we only access through the bastion
		// This is admittedly a little odd... adding a bastion shuts down direct access to the masters/nodes
		// But I think we can always add more permissions in this case later, but we can't easily take them away
		klog.V(2).Infof("bastion is in use; won't configure SSH access to master / node instances")
	} else {
		for _, sshAccess := range b.Cluster.Spec.SSHAccess {
			ipProtocolTCP := IpProtocolTCP
			t := &alitasks.SecurityGroupRule{
				Name:          s("ssh-external-to-master-" + sshAccess),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				IpProtocol:    s(ipProtocolTCP),
				PortRange:     s("22/22"),
				SourceCidrIp:  s(sshAccess),
				In:            fi.Bool(true),
			}
			c.AddTask(t)

			t = &alitasks.SecurityGroupRule{
				Name:          s("ssh-external-to-node-" + sshAccess),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
				IpProtocol:    s(ipProtocolTCP),
				PortRange:     s("22/22"),
				SourceCidrIp:  s(sshAccess),
				In:            fi.Bool(true),
			}
			c.AddTask(t)
		}
	}

	for _, nodePortAccess := range b.Cluster.Spec.NodePortAccess {

		nodePortRange, err := b.NodePortRange()
		if err != nil {
			return err
		}
		ipProtocolTCP := IpProtocolTCP
		ipProtocolUDP := IpProtocolUDP
		fromPort := strconv.Itoa(nodePortRange.Base)
		toPort := strconv.Itoa(nodePortRange.Base + nodePortRange.Size - 1)

		t := &alitasks.SecurityGroupRule{
			Name:          s("nodeport-tcp-external-to-node-" + nodePortAccess),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			IpProtocol:    s(ipProtocolTCP),
			PortRange:     s(fromPort + "/" + toPort),
			SourceCidrIp:  s(nodePortAccess),
			In:            fi.Bool(true),
		}
		c.AddTask(t)

		t = &alitasks.SecurityGroupRule{
			Name:          s("nodeport-udp-external-to-node-" + nodePortAccess),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			IpProtocol:    s(ipProtocolUDP),
			PortRange:     s(fromPort + "/" + toPort),
			SourceCidrIp:  s(nodePortAccess),
			In:            fi.Bool(true),
		}
		c.AddTask(t)
	}

	return nil

}
