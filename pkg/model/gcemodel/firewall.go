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

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*GCEModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {
	klog.Warningf("TODO: Harmonize gcemodel with awsmodel for firewall - GCE model is way too open")

	allProtocols := []string{"tcp", "udp", "icmp", "esp", "ah", "sctp"}

	if b.NetworkingIsCalico() {
		allProtocols = append(allProtocols, "ipip")
	}

	// Allow all traffic from nodes -> nodes
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("node-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    allProtocols,
		}
		c.AddTask(t)
	}

	// Allow full traffic from master -> master
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("master-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed:    allProtocols,
		}
		c.AddTask(t)
	}

	// Allow full traffic from master -> node
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("master-to-node")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:    allProtocols,
		}
		c.AddTask(t)
	}

	// Allow limited traffic from nodes -> masters
	{
		t := &gcetasks.FirewallRule{
			Name:       s(b.NameForFirewallRule("node-to-master")),
			Lifecycle:  b.Lifecycle,
			Network:    b.LinkToNetwork(),
			SourceTags: []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			TargetTags: []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed: []string{
				fmt.Sprintf("tcp:%d", wellknownports.KubeAPIServer),
				fmt.Sprintf("tcp:%d", wellknownports.KopsControllerPort),
			},
		}
		if b.IsGossip() {
			t.Allowed = append(t.Allowed, fmt.Sprintf("udp:%d", wellknownports.DNSControllerGossipMemberlist))
			t.Allowed = append(t.Allowed, fmt.Sprintf("udp:%d", wellknownports.ProtokubeGossipMemberlist))
		}
		if b.NetworkingIsCalico() {
			t.Allowed = append(t.Allowed, "ipip")
		}
		if b.NetworkingIsCilium() {
			t.Allowed = append(t.Allowed, fmt.Sprintf("udp:%d", wellknownports.VxlanUDP))
		}
		c.AddTask(t)
	}

	if b.NetworkingIsIPAlias() || b.NetworkingIsGCERoutes() {
		// When using IP alias or custom routes, SourceTags for identifying traffic don't work, and we must recognize by CIDR

		if b.Cluster.Spec.PodCIDR == "" {
			return fmt.Errorf("expected PodCIDR to be set for IPAlias / kubenet")
		}

		c.AddTask(&gcetasks.FirewallRule{
			Name:         s(b.NameForFirewallRule("pod-cidrs-to-node")),
			Lifecycle:    b.Lifecycle,
			Network:      b.LinkToNetwork(),
			SourceRanges: []string{b.Cluster.Spec.PodCIDR},
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleNode)},
			Allowed:      allProtocols,
		})
	}

	return nil
}

// AddFirewallRulesTasks creates and adds ipv4 and ipv6 gcetasks.FirewallRule Tasks.
// GCE does not allow us to mix ipv4 and ipv6 in the same firewall rule, so we must create separate rules.
// Furthermore, an empty SourceRange with empty SourceTags is interpreted as allow-everything,
// but we intend for it to block everything; so we can Disabled to achieve the desired blocking.
func (b *GCEModelContext) AddFirewallRulesTasks(c *fi.ModelBuilderContext, name string, rule *gcetasks.FirewallRule) {
	var ipv4SourceRanges []string
	var ipv6SourceRanges []string
	for _, sourceRange := range rule.SourceRanges {
		_, cidr, err := net.ParseCIDR(sourceRange)
		if err != nil {
			klog.Fatalf("failed to parse invalid sourceRange %q", sourceRange)
		}

		// Split into ipv4s and ipv6s, but treat IPv4-mapped IPv6 addresses as IPv6
		if cidr.IP.To4() != nil && !strings.Contains(sourceRange, ":") {
			ipv4SourceRanges = append(ipv4SourceRanges, sourceRange)
		} else {
			ipv6SourceRanges = append(ipv6SourceRanges, sourceRange)
		}
	}

	ipv4 := *rule
	ipv4.Name = s(b.NameForFirewallRule(name))
	ipv4.SourceRanges = ipv4SourceRanges
	if len(ipv4.SourceRanges) == 0 {
		// This is helpful because empty SourceRanges and SourceTags are interpreted as allow everything,
		// but the intent is usually to block everything, which can be achieved with Disabled=true.
		ipv4.Disabled = true
		ipv4.SourceRanges = []string{"0.0.0.0/0"}
	}
	c.AddTask(&ipv4)

	ipv6 := *rule
	ipv6.Name = s(b.NameForFirewallRule(name + "-ipv6"))
	ipv6.SourceRanges = ipv6SourceRanges
	if len(ipv6.SourceRanges) == 0 {
		// We specify explicitly so the rule is in IPv6 mode
		ipv6.Disabled = true
		ipv6.SourceRanges = []string{"::/0"}
	}
	c.AddTask(&ipv6)
}
