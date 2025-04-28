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
	"fmt"
	"net"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*GCEModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	klog.Warningf("TODO: Harmonize gcemodel with awsmodel for firewall - GCE model is way too open")

	allProtocols := []string{"tcp", "udp", "icmp", "esp", "ah", "sctp"}

	if b.NetworkingIsCalico() {
		allProtocols = append(allProtocols, "ipip")
	}

	// Allow all TCP traffic from load balancer health checks
	if b.Cluster.Spec.API.LoadBalancer != nil {
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		c.AddTask(&gcetasks.FirewallRule{
			Name:      s(b.NameForFirewallRule("lb-health-checks")),
			Lifecycle: b.Lifecycle,
			Network:   network,
			Family:    gcetasks.AddressFamilyIPv4,
			SourceRanges: sets.New(
				// IP ranges for load balancer health checks
				// https://cloud.google.com/load-balancing/docs/health-checks
				"35.191.0.0/16",
				"130.211.0.0/22",
				"209.85.204.0/22",
				"209.85.152.0/22",
			),
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
			Allowed:    sets.New("tcp"),
		})
	}

	// Allow all traffic from nodes -> nodes
	{
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("node-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    network,
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    sets.New(allProtocols...),
		}
		c.AddTask(t)
	}

	// Allow full traffic from master -> master
	{
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("master-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    network,
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			Allowed:    sets.New(allProtocols...),
		}
		c.AddTask(t)
	}

	// Allow full traffic from master -> node
	{
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("master-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    network,
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    sets.New(allProtocols...),
		}
		c.AddTask(t)
	}

	// Allow limited traffic from nodes -> masters
	{
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("node-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    network,
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane), b.GCETagForRole("Master")},
			Allowed: sets.New(
				fmt.Sprintf("tcp:%d", wellknownports.KubeAPIServer),
				fmt.Sprintf("tcp:%d", wellknownports.KubeletAPI),
				fmt.Sprintf("tcp:%d", wellknownports.KopsControllerPort),
			),
		}
		if b.Cluster.UsesLegacyGossip() {
			t.Allowed.Insert(fmt.Sprintf("udp:%d", wellknownports.DNSControllerGossipMemberlist))
			t.Allowed.Insert(fmt.Sprintf("tcp:%d", wellknownports.DNSControllerGossipMemberlist))
			t.Allowed.Insert(fmt.Sprintf("udp:%d", wellknownports.ProtokubeGossipMemberlist))
			t.Allowed.Insert(fmt.Sprintf("tcp:%d", wellknownports.ProtokubeGossipMemberlist))
		}
		if b.NetworkingIsCalico() {
			t.Allowed.Insert("ipip")
		}
		if b.NetworkingIsCilium() {
			t.Allowed.Insert(fmt.Sprintf("udp:%d", wellknownports.VxlanUDP))
			if model.UseCiliumEtcd(b.Cluster) {
				t.Allowed.Insert(fmt.Sprintf("tcp:%d", wellknownports.EtcdCiliumClientPort))
			}
		}
		c.AddTask(t)
	}

	if b.NetworkingIsIPAlias() || b.NetworkingIsGCERoutes() {
		if b.IsIPv6Only() {
			// We can use tags for IPv6, and this is covered by prior rules
		} else {
			// When using IP alias or custom routes, SourceTags for identifying traffic don't work, and we must recognize by CIDR

			if b.Cluster.Spec.Networking.PodCIDR == "" {
				return fmt.Errorf("expected PodCIDR to be set for IPAlias / kubenet")
			}

			network, err := b.LinkToNetwork()
			if err != nil {
				return err
			}
			b.AddFirewallRulesTasks(c, "pod-cidrs-to-node", &gcetasks.FirewallRule{
				Lifecycle:    b.Lifecycle,
				Network:      network,
				SourceRanges: sets.New(b.Cluster.Spec.Networking.PodCIDR),
				TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
				Allowed:      sets.New(allProtocols...),
			})
		}
	}

	return nil
}

// AddFirewallRulesTasks creates and adds ipv4 and ipv6 gcetasks.FirewallRule Tasks.
// GCE does not allow us to mix ipv4 and ipv6 in the same firewall rule, so we must create separate rules.
// Furthermore, an empty SourceRange with empty SourceTags is interpreted as allow-everything,
// but we intend for it to block everything; so we can Disabled to achieve the desired blocking.
func (b *GCEModelContext) AddFirewallRulesTasks(c *fi.CloudupModelBuilderContext, name string, rule *gcetasks.FirewallRule) {
	ipv4SourceRanges := sets.New[string]()
	ipv6SourceRanges := sets.New[string]()
	for sourceRange := range rule.SourceRanges {
		_, cidr, err := net.ParseCIDR(sourceRange)
		if err != nil {
			klog.Fatalf("failed to parse invalid sourceRange %q", sourceRange)
		}

		// Split into ipv4s and ipv6s, but treat IPv4-mapped IPv6 addresses as IPv6
		if cidr.IP.To4() != nil && !strings.Contains(sourceRange, ":") {
			ipv4SourceRanges.Insert(sourceRange)
		} else {
			ipv6SourceRanges.Insert(sourceRange)
		}
	}

	ipv4 := *rule
	ipv4.Name = s(b.NameForFirewallRule(name))
	ipv4.Family = gcetasks.AddressFamilyIPv4
	if len(ipv4.SourceTags) == 0 {
		ipv4.SourceRanges = ipv4SourceRanges
		if len(ipv4.SourceRanges) == 0 {
			// This is helpful because empty SourceRanges and SourceTags are interpreted as allow everything,
			// but the intent is usually to block everything, which can be achieved with Disabled=true.
			ipv4.Disabled = true
			ipv4.SourceRanges = sets.New("0.0.0.0/0")
		}
	}
	c.AddTask(&ipv4)

	ipv6 := *rule
	ipv6.Name = s(b.NameForFirewallRule(name + "-ipv6"))
	ipv6.Family = gcetasks.AddressFamilyIPv6
	if len(ipv6.SourceTags) == 0 {
		ipv6.SourceRanges = ipv6SourceRanges
		if len(ipv6.SourceRanges) == 0 {
			// We specify explicitly so the rule is in IPv6 mode
			ipv6.Disabled = true
			ipv6.SourceRanges = sets.New("::/0")
		}
	}
	ipv6Allowed := sets.New[string]()
	for allowed := range ipv6.Allowed {
		// Map icmp to icmpv6; easier than maintaining separate lists
		if allowed == "icmp" {
			allowed = "58" // 58 == the IANA protocol number for ICMPv6
		}
		ipv6Allowed.Insert(allowed)
	}
	ipv6.Allowed = ipv6Allowed
	c.AddTask(&ipv6)
}
