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
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// ExternalAccessModelBuilder configures security group rules for external access
// (SSHAccess, KubernetesAPIAccess)
type ExternalAccessModelBuilder struct {
	*GCEModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &ExternalAccessModelBuilder{}

func (b *ExternalAccessModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	klog.Warningf("TODO: Harmonize gcemodel ExternalAccessModelBuilder with awsmodel")
	if len(b.Cluster.Spec.API.Access) == 0 {
		klog.Warningf("KubernetesAPIAccess is empty")
	}

	if len(b.Cluster.Spec.SSHAccess) == 0 {
		klog.Warningf("SSHAccess is empty")
	}

	// SSH is open to AdminCIDR set
	if b.UsesSSHBastion() {
		// If we are using a bastion, we only access through the bastion
		// This is admittedly a little odd... adding a bastion shuts down direct access to the masters/nodes
		// But I think we can always add more permissions in this case later, but we can't easily take them away
		klog.V(2).Infof("bastion is in use; won't configure SSH access to control-plane / worker node instances")
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		b.AddFirewallRulesTasks(c, "ssh-external-to-bastion", &gcetasks.FirewallRule{
			Lifecycle:    b.Lifecycle,
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleBastion)},
			Allowed:      sets.New("tcp:22"),
			SourceRanges: sets.New(b.Cluster.Spec.SSHAccess...),
			Network:      network,
		})
		b.AddFirewallRulesTasks(c, "bastion-to-master-ssh", &gcetasks.FirewallRule{
			Lifecycle:  b.Lifecycle,
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			Allowed:    sets.New("tcp:22"),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleBastion)},
			Network:    network,
		})
		b.AddFirewallRulesTasks(c, "bastion-to-node-ssh", &gcetasks.FirewallRule{
			Lifecycle:  b.Lifecycle,
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    sets.New("tcp:22"),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleBastion)},
			Network:    network,
		})
	} else {
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		b.AddFirewallRulesTasks(c, "ssh-external-to-master", &gcetasks.FirewallRule{
			Lifecycle:    b.Lifecycle,
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			Allowed:      sets.New("tcp:22"),
			SourceRanges: sets.New(b.Cluster.Spec.SSHAccess...),
			Network:      network,
		})

		b.AddFirewallRulesTasks(c, "ssh-external-to-node", &gcetasks.FirewallRule{
			Lifecycle:    b.Lifecycle,
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:      sets.New("tcp:22"),
			SourceRanges: sets.New(b.Cluster.Spec.SSHAccess...),
			Network:      network,
		})
	}

	// NodePort access
	{
		nodePortRange, err := b.NodePortRange()
		if err != nil {
			return err
		}

		nodePortRangeString := nodePortRange.String()
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		b.AddFirewallRulesTasks(c, "nodeport-external-to-node", &gcetasks.FirewallRule{
			Lifecycle:  b.Lifecycle,
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed: sets.New(
				"tcp:"+nodePortRangeString,
				"udp:"+nodePortRangeString,
			),
			SourceRanges: sets.New(b.Cluster.Spec.NodePortAccess...),
			Network:      network,
		})
	}

	if !b.UseLoadBalancerForAPI() {
		// Configuration for the master, when not using a Loadbalancer (ELB)
		// We expect that either the IP address is published, or DNS is set up to point to the IPs
		// We need to open security groups directly to the master nodes (instead of via the ELB)

		// HTTPS to the master is allowed (for API access)

		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		b.AddFirewallRulesTasks(c, "kubernetes-master-https", &gcetasks.FirewallRule{
			Lifecycle:    b.Lifecycle,
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			Allowed:      sets.New("tcp:443"),
			SourceRanges: sets.New(b.Cluster.Spec.API.Access...),
			Network:      network,
		})

		if b.NetworkingIsIPAlias() {
			c.AddTask(&gcetasks.FirewallRule{
				Name:         s(b.NameForFirewallRule("pod-cidrs-to-https-api")),
				Lifecycle:    b.Lifecycle,
				Network:      network,
				Family:       gcetasks.AddressFamilyIPv4, // ip alias is always ipv4
				SourceRanges: sets.New(b.Cluster.Spec.Networking.PodCIDR),
				TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
				Allowed:      sets.New("tcp:" + strconv.Itoa(wellknownports.KubeAPIServer)),
			})
		}
	}

	return nil
}
