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
	"strings"
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
				InboundPolicy:  "DROP",
				OutboundPolicy: "ACCEPT",
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
	if got, want := fi.ValueOf(actual.InboundPolicy), "DROP"; got != want {
		t.Fatalf("unexpected inbound policy: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(actual.OutboundPolicy), "ACCEPT"; got != want {
		t.Fatalf("unexpected outbound policy: got %q, want %q", got, want)
	}
	if got, want := actual.Rules, []*FirewallRule{
		{Direction: "in", Label: "rule-1", Protocol: "tcp", Action: new("ACCEPT"), Port: new("22"), SourceIPs: []net.IPNet{mustParseCIDR(t, "192.0.2.0/24")}},
		{Direction: "out", Label: "rule-2", Protocol: "udp", Action: new("ACCEPT"), SourceIPs: []net.IPNet{mustParseCIDR(t, "2001:db8::/64")}},
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
			{Direction: "in", Label: "ssh", Protocol: "tcp", Port: new("22"), SourceIPs: []net.IPNet{mustParseCIDR(t, "192.0.2.0/24")}},
			{Direction: "out", Label: "dns", Protocol: "udp", Port: new("53"), SourceIPs: []net.IPNet{mustParseCIDR(t, "2001:db8::/64")}},
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
	if got, want := client.LastCreateFirewallOpts.Rules.Inbound[0].Label, "ssh"; got != want {
		t.Fatalf("unexpected inbound rule label: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateFirewallOpts.Rules.Outbound[0].Label, "dns"; got != want {
		t.Fatalf("unexpected outbound rule label: got %q, want %q", got, want)
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

func TestFirewallCheckChangesRejectsInvalidRules(t *testing.T) {
	expected := &Firewall{
		Name:  new("example-k8s-local-nodes"),
		Rules: []*FirewallRule{{Direction: "sideways", Protocol: "tcp"}},
	}

	err := (&Firewall{}).CheckChanges(nil, expected, &Firewall{})
	if err == nil || !strings.Contains(err.Error(), "unsupported direction") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFirewallRenderLinodeRemovesAllRules(t *testing.T) {
	client := &linode.MockLinodeClient{}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	actual := &Firewall{
		ID:    new(42),
		Name:  new("example-k8s-local-nodes"),
		Rules: []*FirewallRule{{Direction: "in", Protocol: "tcp", Action: new("ACCEPT"), Port: new("443")}},
	}
	expected := &Firewall{Name: new("example-k8s-local-nodes")}

	if err := (&Firewall{}).RenderLinode(target, actual, expected, &Firewall{}); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.UpdateFirewallRulesCalls, 1; got != want {
		t.Fatalf("unexpected rule update calls: got %d, want %d", got, want)
	}
	if got, want := client.LastUpdateFirewallRulesOpts.Inbound, []linodego.FirewallRuleInbound(nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected inbound rules: got %v, want %v", got, want)
	}
}

func TestFirewallFindApplyTo(t *testing.T) {
	target := &FirewallTarget{
		Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:Node"},
	}

	for _, testCase := range []struct {
		name               string
		interfaceFirewalls map[int][]linodego.Firewall
		wantApplied        bool
		wantError          string
	}{
		{name: "missing firewall", interfaceFirewalls: map[int][]linodego.Firewall{}},
		{
			name: "expected firewall",
			interfaceFirewalls: map[int][]linodego.Firewall{
				201: {{ID: 101, Label: "example-k8s-local-nodes"}},
			},
			wantApplied: true,
		},
		{
			name: "unexpected firewall",
			interfaceFirewalls: map[int][]linodego.Firewall{
				201: {{ID: 102, Label: "user-managed"}},
			},
			wantError: "has firewall \"user-managed\", expected firewall ID 101",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client := &linode.MockLinodeClient{
				ListFirewallsResponse: []linodego.Firewall{{ID: 101, Label: "example-k8s-local-nodes", Rules: linodego.FirewallRules{InboundPolicy: "DROP", OutboundPolicy: "ACCEPT"}}},
				ListInstancesResponse: []linodego.Instance{{ID: 1, Label: "nodes-us-east-1"}},
				ListInterfacesResponses: map[int][]linodego.LinodeInterface{
					1: {{ID: 201, Public: &linodego.PublicInterface{}}},
				},
				ListInterfaceFirewallsResponses: testCase.interfaceFirewalls,
			}
			ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
			task := &Firewall{Name: new("example-k8s-local-nodes"), ApplyTo: []*FirewallTarget{target}}

			actual, err := task.Find(ctx)
			if testCase.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.wantError) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Find returned error: %v", err)
			}
			if got := len(actual.ApplyTo) == 1; got != testCase.wantApplied {
				t.Fatalf("unexpected ApplyTo state: got %v, want %v", actual.ApplyTo, testCase.wantApplied)
			}
		})
	}
}

func TestFirewallRunAppliesMissingTargets(t *testing.T) {
	target := &FirewallTarget{
		Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:Node"},
	}
	client := &linode.MockLinodeClient{
		ListFirewallsResponse: []linodego.Firewall{{ID: 101, Label: "example-k8s-local-nodes", Rules: linodego.FirewallRules{InboundPolicy: "DROP", OutboundPolicy: "ACCEPT"}}},
		ListInstancesResponse: []linodego.Instance{{ID: 1, Label: "nodes-us-east-1"}},
		ListInterfacesResponses: map[int][]linodego.LinodeInterface{
			1: {{ID: 201, Public: &linodego.PublicInterface{}}, {ID: 202, VPC: &linodego.VPCInterface{}}},
		},
		ListInterfaceFirewallsResponses: map[int][]linodego.Firewall{},
	}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
	task := &Firewall{
		Name:      new("example-k8s-local-nodes"),
		Lifecycle: fi.LifecycleSync,
		ApplyTo:   []*FirewallTarget{target},
	}

	if err := task.Run(ctx); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got, want := client.CreatedFirewallDeviceIDs, []int{101}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected attached firewall IDs: got %v, want %v", got, want)
	}
	if got, want := client.CreatedFirewallDeviceOpts, []linodego.FirewallDeviceCreateOptions{{ID: 201, Type: linodego.FirewallDeviceLinodeInterface}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected firewall attachment options: got %v, want %v", got, want)
	}
}

func TestFirewallFindDetectsPolicyAndActionChanges(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListFirewallsResponse: []linodego.Firewall{{
			ID:    101,
			Label: "example-k8s-local-nodes",
			Rules: linodego.FirewallRules{
				InboundPolicy:  "ACCEPT",
				OutboundPolicy: "DROP",
				Inbound: []linodego.FirewallRuleInbound{{
					Action:   "DROP",
					Label:    "rule-1",
					Protocol: linodego.TCP,
				}},
			},
		}},
	}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
	task := &Firewall{
		Name:  new("example-k8s-local-nodes"),
		Rules: []*FirewallRule{{Direction: "in", Label: "rule-1", Protocol: "tcp"}},
	}
	task.setDefaults()

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	changes := &Firewall{}
	if !fi.BuildChanges(actual, task, changes) {
		t.Fatalf("expected policy and action changes to be detected")
	}
	if got, want := fi.ValueOf(changes.InboundPolicy), "DROP"; got != want {
		t.Fatalf("unexpected inbound policy change: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(changes.OutboundPolicy), "ACCEPT"; got != want {
		t.Fatalf("unexpected outbound policy change: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(changes.Rules[0].Action), "ACCEPT"; got != want {
		t.Fatalf("unexpected rule action change: got %q, want %q", got, want)
	}
}

func TestFirewallFindNormalizesEmptyPort(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListFirewallsResponse: []linodego.Firewall{{
			ID:    101,
			Label: "example-k8s-local-nodes",
			Rules: linodego.FirewallRules{
				InboundPolicy:  "DROP",
				OutboundPolicy: "ACCEPT",
				Inbound: []linodego.FirewallRuleInbound{{
					Action:   "ACCEPT",
					Label:    "rule-1",
					Protocol: linodego.ICMP,
				}},
			},
		}},
	}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
	task := &Firewall{
		Name:  new("example-k8s-local-nodes"),
		Rules: []*FirewallRule{{Direction: "in", Label: "rule-1", Protocol: "icmp"}},
	}
	task.setDefaults()

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual.Rules[0].Port != nil {
		t.Fatalf("expected empty discovered port to be normalized to nil, got %q", fi.ValueOf(actual.Rules[0].Port))
	}
	if fi.BuildChanges(actual, task, &Firewall{}) {
		t.Fatalf("expected equivalent omitted ports to produce no changes")
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
