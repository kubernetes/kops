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

package gcemodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// BastionModelBuilder adds model objects to support bastions
//
// Bastion instances live in the utility subnets created in the private topology.
// All traffic goes through an ELB, and the ELB has port 22 open to SSHAccess.
// Bastion instances have access to all internal master and node instances.

type BastionModelBuilder struct {
	*GCEModelContext
	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &BastionModelBuilder{}

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

	// Allow SSH traffic from world -> bastion
	{
		t := &gcetasks.FirewallRule{
			Name:         s(b.SafeObjectName("ssh-external-to-bastion")),
			Lifecycle:    b.Lifecycle,
			Network:      b.LinkToNetwork(),
			SourceRanges: []string{"0.0.0.0/0"},
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleBastion)},
			Allowed:      []string{"tcp:22"},
		}
		c.AddTask(t)
	}

	// Allow SSH traffic from bastion -> master
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.SafeObjectName("bastion-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleBastion)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed:    []string{"tcp:22"},
		}
		c.AddTask(t)
	}

	// Allow SSH traffic from bastion -> node
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.SafeObjectName("bastion-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleBastion)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    []string{"tcp:22"},
		}
		c.AddTask(t)
	}
	return nil
}
