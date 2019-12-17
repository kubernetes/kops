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

package gcemodel

import (
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*GCEModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {
	klog.Warningf("TODO: Harmonize gcemodel with awsmodel for firewall - GCE model is way too open")

	//// Allow all traffic from vms in our network
	//// TODO: Is this a good idea?
	//{
	//	t := &gcetasks.FirewallRule{
	//		Name:         s(b.SafeObjectName("kubernetes-internal")),
	//		Network:      b.LinkToNetwork(),
	//		SourceRanges: []string{b.Cluster.Spec.NetworkCIDR},
	//		Allowed:      []string{"tcp:1-65535", "udp:1-65535", "icmp"},
	//	}
	//	c.AddTask(t)
	//}

	// Allow all traffic from nodes -> nodes
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.SafeObjectName("node-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    []string{"tcp", "udp", "icmp", "esp", "ah", "sctp"},
		}
		c.AddTask(t)
	}

	if b.Cluster.Spec.NonMasqueradeCIDR != "" {
		// The traffic is not recognized if it's on the overlay network?
		klog.Warningf("Adding overlay network for X -> node rule - HACK")

		t := &gcetasks.FirewallRule{
			Name:         s(b.SafeObjectName("cidr-to-node")),
			Lifecycle:    b.Lifecycle,
			Network:      b.LinkToNetwork(),
			SourceRanges: []string{b.Cluster.Spec.NonMasqueradeCIDR},
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:      []string{"tcp", "udp", "icmp", "esp", "ah", "sctp"},
		}
		c.AddTask(t)
	}

	// Allow full traffic from master -> master
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.SafeObjectName("master-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed:    []string{"tcp", "udp", "icmp", "esp", "ah", "sctp"},
		}
		c.AddTask(t)
	}

	// Allow full traffic from master -> node
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.SafeObjectName("master-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    []string{"tcp", "udp", "icmp", "esp", "ah", "sctp"},
		}
		c.AddTask(t)
	}

	// Allow limited traffic from nodes -> masters
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.SafeObjectName("node-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed:    []string{"tcp:443", "tcp:4194"},
		}
		c.AddTask(t)
	}

	if b.Cluster.Spec.NonMasqueradeCIDR != "" {
		// The traffic is not recognized if it's on the overlay network?
		klog.Warningf("Adding overlay network for X -> master rule - HACK")
		t := &gcetasks.FirewallRule{
			Name:         s(b.SafeObjectName("cidr-to-master")),
			Lifecycle:    b.Lifecycle,
			Network:      b.LinkToNetwork(),
			SourceRanges: []string{b.Cluster.Spec.NonMasqueradeCIDR},
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed:      []string{"tcp:443", "tcp:4194"},
		}
		c.AddTask(t)
	}
	return nil
}
