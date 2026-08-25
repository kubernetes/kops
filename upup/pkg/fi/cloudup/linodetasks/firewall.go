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
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type Firewall struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID    *int
	Rules []*FirewallRule
	Tags  []string
}

var _ fi.CloudupTask = &Firewall{}
var _ fi.CompareWithID = &Firewall{}

// CompareWithID returns the name of the firewall as its unique identifier.
func (v *Firewall) CompareWithID() *string {
	return v.Name
}

func (v *Firewall) Find(c *fi.CloudupContext) (*Firewall, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)
	name := fi.ValueOf(v.Name)
	if name == "" {
		return nil, fmt.Errorf("Firewall.Name is required")
	}

	listOptions, err := linode.ListOptionsForLabel(name)
	if err != nil {
		return nil, err
	}

	firewalls, err := cloud.Client().ListFirewalls(c.Context(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) firewalls: %w", err)
	}

	var matched *linodego.Firewall
	if len(firewalls) == 0 {
		return nil, nil
	}
	// Name is unique, so we should only have one match
	matched = &firewalls[0]

	actual := &Firewall{
		Name:      new(matched.Label),
		Lifecycle: v.Lifecycle,
		ID:        new(matched.ID),
		Tags:      matched.Tags,
	}
	actual.Rules = append(actual.Rules, firewallRulesFromLinode(matched.Rules)...)
	v.ID = actual.ID

	return actual, nil
}

func (v *Firewall) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (_ *Firewall) CheckChanges(actual, expected, changes *Firewall) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (*Firewall) RenderLinode(t *linode.APITarget, actual, expected, changes *Firewall) error {
	if actual == nil {
		rules, err := firewallRulesToLinode(expected.Rules)
		if err != nil {
			return fmt.Errorf("building rules for Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
		firewall, err := t.Cloud.Client().CreateFirewall(context.Background(), linodego.FirewallCreateOptions{
			Label: fi.ValueOf(expected.Name),
			Rules: rules,
			Tags:  expected.Tags,
		})
		if err != nil {
			return fmt.Errorf("error creating Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
		expected.ID = new(firewall.ID)
		return nil
	}

	expected.ID = actual.ID
	if changes == nil {
		return nil
	}

	if changes.Tags != nil {
		if _, err := t.Cloud.Client().UpdateFirewall(context.Background(), fi.ValueOf(actual.ID), linodego.FirewallUpdateOptions{Tags: expected.Tags}); err != nil {
			return fmt.Errorf("error updating Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
	}
	if changes.Rules != nil {
		rules, err := firewallRulesToUpdateOptions(expected.Rules)
		if err != nil {
			return fmt.Errorf("building rules for Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
		if _, err := t.Cloud.Client().UpdateFirewallRules(context.Background(), fi.ValueOf(actual.ID), rules); err != nil {
			return fmt.Errorf("error updating rules for Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
	}

	return nil
}

// FirewallRule represents a Firewall's rules.
type FirewallRule struct {
	Direction string
	SourceIPs []net.IPNet
	Protocol  string
	Port      *string
}

var _ fi.CloudupHasDependencies = (*FirewallRule)(nil)

func (e *FirewallRule) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

// firewallRulesToLinode converts a slice of FirewallRule objects to Linode's FirewallRulesCreateOptions format.
func firewallRulesToLinode(rules []*FirewallRule) (linodego.FirewallRulesCreateOptions, error) {
	options := linodego.FirewallRulesCreateOptions{
		InboundPolicy:  "DROP",
		OutboundPolicy: "ACCEPT",
	}
	for index, rule := range rules {
		if rule == nil {
			return linodego.FirewallRulesCreateOptions{}, fmt.Errorf("rule %d is nil", index+1)
		}
		addresses := firewallRuleAddresses(rule.SourceIPs)
		switch strings.ToLower(rule.Direction) {
		case "in":
			options.Inbound = append(options.Inbound, linodego.FirewallRuleInbound{
				Action:    "ACCEPT",
				Label:     fmt.Sprintf("rule-%d", index+1),
				Ports:     fi.ValueOf(rule.Port),
				Protocol:  linodego.NetworkProtocol(strings.ToUpper(rule.Protocol)),
				Addresses: addresses,
			})
		case "out":
			options.Outbound = append(options.Outbound, linodego.FirewallRuleOutbound{
				Action:    "ACCEPT",
				Label:     fmt.Sprintf("rule-%d", index+1),
				Ports:     fi.ValueOf(rule.Port),
				Protocol:  linodego.NetworkProtocol(strings.ToUpper(rule.Protocol)),
				Addresses: addresses,
			})
		default:
			return linodego.FirewallRulesCreateOptions{}, fmt.Errorf("rule %d has unsupported direction %q", index+1, rule.Direction)
		}
	}
	return options, nil
}

// firewallRulesToUpdateOptions converts a slice of FirewallRule objects to Linode's FirewallRulesUpdateOptions format.
func firewallRulesToUpdateOptions(rules []*FirewallRule) (linodego.FirewallRulesUpdateOptions, error) {
	createOptions, err := firewallRulesToLinode(rules)
	if err != nil {
		return linodego.FirewallRulesUpdateOptions{}, err
	}
	return linodego.FirewallRulesUpdateOptions{
		Inbound:        createOptions.Inbound,
		InboundPolicy:  createOptions.InboundPolicy,
		Outbound:       createOptions.Outbound,
		OutboundPolicy: createOptions.OutboundPolicy,
	}, nil
}

// firewallRulesFromLinode converts Linode's FirewallRules format to a slice of FirewallRule objects.
func firewallRulesFromLinode(rules linodego.FirewallRules) []*FirewallRule {
	var result []*FirewallRule
	for _, rule := range rules.Inbound {
		result = append(result, firewallRuleFromLinode("in", string(rule.Protocol), rule.Ports, rule.Addresses))
	}
	for _, rule := range rules.Outbound {
		result = append(result, firewallRuleFromLinode("out", string(rule.Protocol), rule.Ports, rule.Addresses))
	}
	return result
}

// firewallRuleFromLinode converts a single Linode firewall rule to a FirewallRule object.
func firewallRuleFromLinode(direction, protocol, ports string, addresses linodego.NetworkAddresses) *FirewallRule {
	rule := &FirewallRule{
		Direction: direction,
		Protocol:  strings.ToLower(protocol),
		Port:      new(ports),
	}
	for _, cidr := range append(addresses.IPv4, addresses.IPv6...) {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			rule.SourceIPs = append(rule.SourceIPs, *ipNet)
		}
	}
	return rule
}

// firewallRuleAddresses converts a slice of net.IPNet objects to Linode's NetworkAddresses format.
func firewallRuleAddresses(sourceIPs []net.IPNet) linodego.NetworkAddresses {
	addresses := linodego.NetworkAddresses{}
	for _, sourceIP := range sourceIPs {
		if sourceIP.IP.To4() != nil {
			addresses.IPv4 = append(addresses.IPv4, sourceIP.String())
		} else {
			addresses.IPv6 = append(addresses.IPv6, sourceIP.String())
		}
	}
	return addresses
}
