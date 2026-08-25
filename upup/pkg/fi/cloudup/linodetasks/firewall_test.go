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

package linodetasks

import (
	"net"
	"reflect"
	"testing"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func TestFirewallFindByLabel(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListFirewallsResponse: []linodego.Firewall{{
			ID:    101,
			Label: "example-k8s-local-nodes",
			Tags:  []string{"kops.k8s.io/cluster:example.k8s.local"},
			Rules: linodego.FirewallRules{
				Inbound: []linodego.FirewallRuleInbound{{
					Action:   "ACCEPT",
					Label:    "rule-1",
					Ports:    "22",
					Protocol: linodego.TCP,
					Addresses: linodego.NetworkAddresses{
						IPv4: []string{"192.0.2.0/24"},
					},
				}},
				Outbound: []linodego.FirewallRuleOutbound{{
					Action:   "ACCEPT",
					Label:    "rule-2",
					Protocol: linodego.UDP,
					Addresses: linodego.NetworkAddresses{
						IPv6: []string{"2001:db8::/64"},
					},
				}},
			},
		}},
	}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
	task := &Firewall{Name: new("example-k8s-local-nodes")}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find firewall")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected firewall ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(task.ID), 101; got != want {
		t.Fatalf("expected task ID to be propagated after Find: got %d, want %d", got, want)
	}
	if got, want := actual.Rules, []*FirewallRule{
		{Direction: "in", Protocol: "tcp", Port: new("22"), SourceIPs: []net.IPNet{mustParseCIDR(t, "192.0.2.0/24")}},
		{Direction: "out", Protocol: "udp", Port: new(""), SourceIPs: []net.IPNet{mustParseCIDR(t, "2001:db8::/64")}},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected rules: got %#v, want %#v", got, want)
	}

	expectedListOptions, err := linode.ListOptionsForLabel("example-k8s-local-nodes")
	if err != nil {
		t.Fatalf("ListOptionsForLabel returned error: %v", err)
	}
	if client.LastListFirewallsOpts == nil {
		t.Fatalf("expected firewall list options to be recorded")
	}
	if got, want := client.LastListFirewallsOpts.Filter, expectedListOptions.Filter; got != want {
		t.Fatalf("unexpected firewall list filter: got %q, want %q", got, want)
	}
}

func TestFirewallRenderLinodeCreate(t *testing.T) {
	client := &linode.MockLinodeClient{CreateFirewallResponse: &linodego.Firewall{ID: 42}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	expected := &Firewall{
		Name: new("example-k8s-local-nodes"),
		Tags: []string{"kops.k8s.io/cluster:example.k8s.local"},
		Rules: []*FirewallRule{
			{Direction: "in", Protocol: "tcp", Port: new("22"), SourceIPs: []net.IPNet{mustParseCIDR(t, "192.0.2.0/24")}},
			{Direction: "out", Protocol: "udp", Port: new("53"), SourceIPs: []net.IPNet{mustParseCIDR(t, "2001:db8::/64")}},
		},
	}

	if err := (&Firewall{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateFirewallCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateFirewallOpts.Label, "example-k8s-local-nodes"; got != want {
		t.Fatalf("unexpected firewall label: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateFirewallOpts.Rules.InboundPolicy, "DROP"; got != want {
		t.Fatalf("unexpected inbound policy: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateFirewallOpts.Rules.OutboundPolicy, "ACCEPT"; got != want {
		t.Fatalf("unexpected outbound policy: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateFirewallOpts.Rules.Inbound[0].Addresses.IPv4, []string{"192.0.2.0/24"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected inbound IPv4 addresses: got %v, want %v", got, want)
	}
	if got, want := client.LastCreateFirewallOpts.Rules.Outbound[0].Addresses.IPv6, []string{"2001:db8::/64"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected outbound IPv6 addresses: got %v, want %v", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to be populated from create response: got %d, want %d", got, want)
	}
}

func TestFirewallRenderLinodeUpdatesRules(t *testing.T) {
	client := &linode.MockLinodeClient{}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	actual := &Firewall{ID: new(42), Name: new("example-k8s-local-nodes")}
	expected := &Firewall{
		Name:  new("example-k8s-local-nodes"),
		Rules: []*FirewallRule{{Direction: "in", Protocol: "tcp", Port: new("443"), SourceIPs: []net.IPNet{mustParseCIDR(t, "198.51.100.0/24")}}},
	}
	changes := &Firewall{Rules: expected.Rules}

	if err := (&Firewall{}).RenderLinode(target, actual, expected, changes); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.UpdateFirewallRulesCalls, 1; got != want {
		t.Fatalf("unexpected rule update calls: got %d, want %d", got, want)
	}
	if got, want := client.UpdatedFirewallRulesIDs, []int{42}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected updated firewall IDs: got %v, want %v", got, want)
	}
	if got, want := client.LastUpdateFirewallRulesOpts.Inbound[0].Ports, "443"; got != want {
		t.Fatalf("unexpected updated port: got %q, want %q", got, want)
	}
}

func TestFirewallRenderLinodeRejectsInvalidRuleDirection(t *testing.T) {
	client := &linode.MockLinodeClient{}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	expected := &Firewall{
		Name:  new("example-k8s-local-nodes"),
		Rules: []*FirewallRule{{Direction: "sideways", Protocol: "tcp"}},
	}

	if err := (&Firewall{}).RenderLinode(target, nil, expected, nil); err == nil {
		t.Fatalf("expected invalid direction to be rejected")
	}
	if got, want := client.CreateFirewallCalls, 0; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
}

func mustParseCIDR(t *testing.T, cidr string) net.IPNet {
	t.Helper()
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR(%q): %v", cidr, err)
	}
	return *ipNet
}
