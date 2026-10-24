/*
Copyright 2026 The Kubernetes Authors.

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

package linodemodel

import (
	"fmt"
	"net"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

const (
	linodeFirewallRoleControlPlane = "control-plane"
	linodeFirewallRoleNodes        = "nodes"
	linodeFirewallRoleBastion      = "bastion"
)

type FirewallModelBuilder struct {
	*LinodeModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	sshAccess, err := parseLinodeFirewallCIDRs(b.Cluster.Spec.SSHAccess)
	if err != nil {
		return err
	}
	controlPlaneFirewall := b.newFirewall(linodeFirewallRoleControlPlane, sshAccess)
	nodesFirewall := b.newFirewall(linodeFirewallRoleNodes, sshAccess)
	bastionFirewall := b.newFirewall(linodeFirewallRoleBastion, sshAccess)

	if !b.UseLoadBalancerForAPI() && len(b.Cluster.Spec.API.Access) > 0 {
		apiAccess, err := parseLinodeFirewallCIDRs(b.Cluster.Spec.API.Access)
		if err != nil {
			return err
		}
		controlPlaneFirewall.Rules = append(controlPlaneFirewall.Rules, &linodetasks.FirewallRule{
			Direction: "in",
			Label:     "kubernetes-api",
			SourceIPs: apiAccess,
			Protocol:  "tcp",
			Port:      new("443"),
		})
	}

	if len(b.Cluster.Spec.NodePortAccess) > 0 {
		nodePortAccess, err := parseLinodeFirewallCIDRs(b.Cluster.Spec.NodePortAccess)
		if err != nil {
			return err
		}
		nodePortRange, err := b.NodePortRange()
		if err != nil {
			return err
		}
		nodePort := new(fmt.Sprintf("%d-%d", nodePortRange.Base, nodePortRange.Base+nodePortRange.Size-1))
		for _, protocol := range []string{"tcp", "udp"} {
			nodesFirewall.Rules = append(nodesFirewall.Rules, &linodetasks.FirewallRule{
				Direction: "in",
				Label:     "nodeport-" + protocol,
				SourceIPs: nodePortAccess,
				Protocol:  protocol,
				Port:      nodePort,
			})
		}
	}

	for _, firewall := range []struct {
		role     string
		firewall *linodetasks.Firewall
	}{
		{role: linodeFirewallRoleControlPlane, firewall: controlPlaneFirewall},
		{role: linodeFirewallRoleNodes, firewall: nodesFirewall},
		{role: linodeFirewallRoleBastion, firewall: bastionFirewall},
	} {
		required, err := b.publicFirewallRequired(firewall.role)
		if err != nil {
			return err
		}
		if required {
			c.AddTask(firewall.firewall)
		}
	}

	return nil
}

func (b *FirewallModelBuilder) publicFirewallRequired(role string) (bool, error) {
	for _, ig := range b.InstanceGroups {
		instanceGroupFirewallRole, err := linodeFirewallRoleForInstanceGroup(ig.Spec.Role)
		if err != nil {
			return false, err
		}
		if role != instanceGroupFirewallRole {
			continue
		}
		subnets, err := b.GatherSubnets(ig)
		if err != nil {
			return false, err
		}
		for _, subnet := range subnets {
			if fi.ValueOf(requirePublicInterface(subnet, ig)) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (b *FirewallModelBuilder) newFirewall(role string, sshAccess []net.IPNet) *linodetasks.Firewall {
	var rules []*linodetasks.FirewallRule
	if len(sshAccess) > 0 {
		rules = []*linodetasks.FirewallRule{{
			Direction: "in",
			Label:     "ssh",
			SourceIPs: sshAccess,
			Protocol:  "tcp",
			Port:      new("22"),
		}}
	}
	return &linodetasks.Firewall{
		Name:      new(linodeFirewallName(b.ClusterName(), role)),
		Lifecycle: b.Lifecycle,
		Rules:     rules,
		Tags: []string{
			fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, b.ClusterName()),
			fmt.Sprintf("%s:%s", linode.TagKubernetesFirewallRole, role),
		},
		ApplyTo: linodeFirewallTargets(b.ClusterName(), role),
	}
}

func linodeFirewallTargets(clusterName, firewallRole string) []*linodetasks.FirewallTarget {
	clusterTag := fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, clusterName)
	var instanceRoles []kops.InstanceGroupRole
	switch firewallRole {
	case linodeFirewallRoleControlPlane:
		instanceRoles = []kops.InstanceGroupRole{kops.InstanceGroupRoleControlPlane, kops.InstanceGroupRoleAPIServer}
	case linodeFirewallRoleNodes:
		instanceRoles = []kops.InstanceGroupRole{kops.InstanceGroupRoleNode}
	case linodeFirewallRoleBastion:
		instanceRoles = []kops.InstanceGroupRole{kops.InstanceGroupRoleBastion}
	default:
		return nil
	}

	targets := make([]*linodetasks.FirewallTarget, 0, len(instanceRoles))
	for _, instanceRole := range instanceRoles {
		targets = append(targets, &linodetasks.FirewallTarget{
			Tags: []string{
				clusterTag,
				fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceRole, instanceRole),
			},
		})
	}
	return targets
}

func linodeFirewallName(clusterName, role string) string {
	name := linode.NormalizeLinodeLabel(clusterName + "-" + role)
	name = truncate.TruncateString(name, truncate.TruncateStringOptions{
		MaxLength:     32,
		AlwaysAddHash: strings.Contains(name, "--"),
	})
	// Truncation can also place the hash separator next to an existing hyphen.
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return name
}

func parseLinodeFirewallCIDRs(cidrs []string) ([]net.IPNet, error) {
	result := make([]net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		result = append(result, *ipNet)
	}
	return result, nil
}

func linodeFirewallRoleForInstanceGroup(role kops.InstanceGroupRole) (string, error) {
	switch role {
	case kops.InstanceGroupRoleAPIServer, kops.InstanceGroupRoleControlPlane:
		return linodeFirewallRoleControlPlane, nil
	case kops.InstanceGroupRoleNode:
		return linodeFirewallRoleNodes, nil
	case kops.InstanceGroupRoleBastion:
		return linodeFirewallRoleBastion, nil
	default:
		return "", fmt.Errorf("Akamai (Linode) firewall is not configured for InstanceGroup role %q", role)
	}
}
