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

package model

import (
	"fmt"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// ExternalAccessModelBuilder configures security group rules for external access
// (SSHAccess, KubernetesAPIAccess)
type ExternalAccessModelBuilder struct {
	*KopsModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &ExternalAccessModelBuilder{}

func (b *ExternalAccessModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if len(b.Cluster.Spec.KubernetesAPIAccess) == 0 {
		klog.Warningf("KubernetesAPIAccess is empty")
	}

	if len(b.Cluster.Spec.SSHAccess) == 0 {
		klog.Warningf("SSHAccess is empty")
	}

	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleMaster)
	if err != nil {
		return err
	}
	nodeGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleNode)
	if err != nil {
		return err
	}

	// SSH is open to AdminCIDR set
	if b.UsesSSHBastion() {
		// If we are using a bastion, we only access through the bastion
		// This is admittedly a little odd... adding a bastion shuts down direct access to the masters/nodes
		// But I think we can always add more permissions in this case later, but we can't easily take them away
		klog.V(2).Infof("bastion is in use; won't configure SSH access to master / node instances")
	} else {
		for _, sshAccess := range b.Cluster.Spec.SSHAccess {
			for _, masterGroup := range masterGroups {
				suffix := masterGroup.Suffix
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("ssh-external-to-master-%s%s", sshAccess, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					Protocol:      s("tcp"),
					FromPort:      i64(22),
					ToPort:        i64(22),
					CIDR:          s(sshAccess),
				}
				c.AddTask(t)
			}

			for _, nodeGroup := range nodeGroups {
				suffix := nodeGroup.Suffix
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("ssh-external-to-node-%s%s", sshAccess, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: nodeGroup.Task,
					Protocol:      s("tcp"),
					FromPort:      i64(22),
					ToPort:        i64(22),
					CIDR:          s(sshAccess),
				}
				c.AddTask(t)
			}
		}
	}

	for _, nodePortAccess := range b.Cluster.Spec.NodePortAccess {
		nodePortRange, err := b.NodePortRange()
		if err != nil {
			return err
		}

		for _, nodeGroup := range nodeGroups {
			suffix := nodeGroup.Suffix
			t1 := &awstasks.SecurityGroupRule{
				Name:          s(fmt.Sprintf("nodeport-tcp-external-to-node-%s%s", nodePortAccess, suffix)),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: nodeGroup.Task,
				Protocol:      s("tcp"),
				FromPort:      i64(int64(nodePortRange.Base)),
				ToPort:        i64(int64(nodePortRange.Base + nodePortRange.Size - 1)),
				CIDR:          s(nodePortAccess),
			}
			c.AddTask(t1)

			t2 := &awstasks.SecurityGroupRule{
				Name:          s(fmt.Sprintf("nodeport-udp-external-to-node-%s%s", nodePortAccess, suffix)),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: nodeGroup.Task,
				Protocol:      s("udp"),
				FromPort:      i64(int64(nodePortRange.Base)),
				ToPort:        i64(int64(nodePortRange.Base + nodePortRange.Size - 1)),
				CIDR:          s(nodePortAccess),
			}
			c.AddTask(t2)
		}
	}

	if !b.UseLoadBalancerForAPI() {
		// Configuration for the master, when not using a Loadbalancer (ELB)
		// We expect that either the IP address is published, or DNS is set up to point to the IPs
		// We need to open security groups directly to the master nodes (instead of via the ELB)

		// HTTPS to the master is allowed (for API access)
		for _, apiAccess := range b.Cluster.Spec.KubernetesAPIAccess {
			for _, masterGroup := range masterGroups {
				suffix := masterGroup.Suffix
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("https-external-to-master-%s%s", apiAccess, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					Protocol:      s("tcp"),
					FromPort:      i64(443),
					ToPort:        i64(443),
					CIDR:          s(apiAccess),
				}
				c.AddTask(t)
			}
		}
	}

	return nil
}
