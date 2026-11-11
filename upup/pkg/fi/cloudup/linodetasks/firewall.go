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

	ID             *int
	InboundPolicy  *string
	OutboundPolicy *string
	Rules          []*FirewallRule
	Tags           []string
	ApplyTo        []*FirewallTarget
}

type FirewallTarget struct {
	fi.CloudupNotADependency
	Tags []string
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
		Name:           new(matched.Label),
		Lifecycle:      v.Lifecycle,
		ID:             new(matched.ID),
		InboundPolicy:  new(matched.Rules.InboundPolicy),
		OutboundPolicy: new(matched.Rules.OutboundPolicy),
		Tags:           matched.Tags,
	}
	actual.Rules = append(actual.Rules, firewallRulesFromLinode(matched.Rules)...)
	v.ID = actual.ID
	if v.ApplyTo != nil {
		applied, err := firewallTargetsAreApplied(c.Context(), cloud.Client(), matched.ID, v.ApplyTo)
		if err != nil {
			return nil, err
		}
		if applied {
			actual.ApplyTo = v.ApplyTo
		}
	}

	return actual, nil
}

func (v *Firewall) Run(c *fi.CloudupContext) error {
	v.setDefaults()
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (v *Firewall) setDefaults() {
	if v.InboundPolicy == nil {
		v.InboundPolicy = new("DROP")
	}
	if v.OutboundPolicy == nil {
		v.OutboundPolicy = new("ACCEPT")
	}
	for index, rule := range v.Rules {
		if rule != nil && rule.Action == nil {
			rule.Action = new("ACCEPT")
		}
		if rule != nil && rule.Label == "" {
			rule.Label = fmt.Sprintf("rule-%d", index+1)
		}
	}
}

func (_ *Firewall) CheckChanges(actual, expected, changes *Firewall) error {
	if err := validateFirewallTargets(expected.ApplyTo); err != nil {
		return err
	}
	if _, err := firewallRulesToLinode(expected.InboundPolicy, expected.OutboundPolicy, expected.Rules); err != nil {
		return fmt.Errorf("invalid firewall rules: %w", err)
	}
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
	expected.setDefaults()
	client := t.Cloud.Client()
	var firewallID int
	if actual == nil {
		rules, err := firewallRulesToLinode(expected.InboundPolicy, expected.OutboundPolicy, expected.Rules)
		if err != nil {
			return fmt.Errorf("building rules for Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
		firewall, err := client.CreateFirewall(context.Background(), linodego.FirewallCreateOptions{
			Label: fi.ValueOf(expected.Name),
			Rules: rules,
			Tags:  expected.Tags,
		})
		if err != nil {
			return fmt.Errorf("error creating Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
		}
		expected.ID = new(firewall.ID)
		firewallID = firewall.ID
	} else {
		expected.ID = actual.ID
		firewallID = fi.ValueOf(actual.ID)
		if changes == nil {
			return nil
		}

		rulesChanged := changes.Rules != nil || (actual.Rules != nil && expected.Rules == nil)
		var rules linodego.FirewallRulesUpdateOptions
		if changes.InboundPolicy != nil || changes.OutboundPolicy != nil || rulesChanged {
			var err error
			rules, err = firewallRulesToUpdateOptions(expected.InboundPolicy, expected.OutboundPolicy, expected.Rules)
			if err != nil {
				return fmt.Errorf("building rules for Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
			}
		}

		if changes.Tags != nil {
			if _, err := client.UpdateFirewall(context.Background(), fi.ValueOf(actual.ID), linodego.FirewallUpdateOptions{Tags: expected.Tags}); err != nil {
				return fmt.Errorf("error updating Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
			}
		}
		if changes.InboundPolicy != nil || changes.OutboundPolicy != nil || rulesChanged {
			if _, err := client.UpdateFirewallRules(context.Background(), fi.ValueOf(actual.ID), rules); err != nil {
				return fmt.Errorf("error updating rules for Akamai (Linode) firewall %q: %w", fi.ValueOf(expected.Name), err)
			}
		}
	}

	if expected.ApplyTo != nil && (actual == nil || changes.ApplyTo != nil) {
		if err := applyFirewallToTargets(context.Background(), client, firewallID, expected.ApplyTo); err != nil {
			return err
		}
	}

	return nil
}

// FirewallRule represents a Firewall's rules.
type FirewallRule struct {
	Direction string
	Label     string
	SourceIPs []net.IPNet
	Protocol  string
	Action    *string
	Port      *string
}

var _ fi.CloudupHasDependencies = (*FirewallRule)(nil)

func (e *FirewallRule) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

func validateFirewallTargets(targets []*FirewallTarget) error {
	for _, target := range targets {
		if target == nil {
			return fmt.Errorf("Firewall.ApplyTo cannot contain nil targets")
		}
		if len(target.Tags) == 0 {
			return fmt.Errorf("Firewall.ApplyTo target tags are required")
		}
	}
	return nil
}

func firewallTargetsAreApplied(ctx context.Context, client linode.LinodeClient, firewallID int, targets []*FirewallTarget) (bool, error) {
	applied := true
	err := forEachFirewallTargetInterface(ctx, client, targets, func(instance linodego.Instance, iface linodego.LinodeInterface) error {
		interfaceApplied, err := firewallIsAppliedToInterface(ctx, client, firewallID, instance, iface)
		if err != nil {
			return err
		}
		if !interfaceApplied {
			applied = false
		}
		return nil
	})
	return applied, err
}

func applyFirewallToTargets(ctx context.Context, client linode.LinodeClient, firewallID int, targets []*FirewallTarget) error {
	return forEachFirewallTargetInterface(ctx, client, targets, func(instance linodego.Instance, iface linodego.LinodeInterface) error {
		interfaceApplied, err := firewallIsAppliedToInterface(ctx, client, firewallID, instance, iface)
		if err != nil {
			return err
		}
		if interfaceApplied {
			return nil
		}
		if _, err := client.CreateFirewallDevice(ctx, firewallID, linodego.FirewallDeviceCreateOptions{ID: iface.ID, Type: linodego.FirewallDeviceLinodeInterface}); err != nil {
			return fmt.Errorf("error attaching Akamai (Linode) firewall %d to interface %d on instance %q: %w", firewallID, iface.ID, instance.Label, err)
		}
		return nil
	})
}

func firewallIsAppliedToInterface(ctx context.Context, client linode.LinodeClient, firewallID int, instance linodego.Instance, iface linodego.LinodeInterface) (bool, error) {
	firewalls, err := client.ListInterfaceFirewalls(ctx, instance.ID, iface.ID, nil)
	if err != nil {
		return false, fmt.Errorf("error listing Akamai (Linode) firewalls for interface %d on instance %q: %w", iface.ID, instance.Label, err)
	}
	if len(firewalls) == 0 {
		return false, nil
	}
	if len(firewalls) != 1 || firewalls[0].ID != firewallID {
		return false, fmt.Errorf("interface %d on instance %q has firewall %q, expected firewall ID %d", iface.ID, instance.Label, firewalls[0].Label, firewallID)
	}
	return true, nil
}

func forEachFirewallTargetInterface(ctx context.Context, client linode.LinodeClient, targets []*FirewallTarget, visit func(linodego.Instance, linodego.LinodeInterface) error) error {
	if err := validateFirewallTargets(targets); err != nil {
		return err
	}
	seenInterfaceIDs := make(map[int]struct{})
	for _, target := range targets {
		listOptions, err := linode.ListOptionsForTags(target.Tags...)
		if err != nil {
			return err
		}
		instances, err := client.ListInstances(ctx, listOptions)
		if err != nil {
			return fmt.Errorf("error listing Akamai (Linode) instances for firewall target: %w", err)
		}
		for _, instance := range instances {
			interfaces, err := client.ListInterfaces(ctx, instance.ID, nil)
			if err != nil {
				return fmt.Errorf("error listing Akamai (Linode) interfaces for instance %q: %w", instance.Label, err)
			}
			for _, iface := range interfaces {
				if iface.Public == nil {
					continue
				}
				if _, found := seenInterfaceIDs[iface.ID]; found {
					continue
				}
				seenInterfaceIDs[iface.ID] = struct{}{}
				if err := visit(instance, iface); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// firewallRulesToLinode converts a slice of FirewallRule objects to Linode's FirewallRulesCreateOptions format.
func firewallRulesToLinode(inboundPolicy, outboundPolicy *string, rules []*FirewallRule) (linodego.FirewallRulesCreateOptions, error) {
	options := linodego.FirewallRulesCreateOptions{
		InboundPolicy:  fi.ValueOf(inboundPolicy),
		OutboundPolicy: fi.ValueOf(outboundPolicy),
	}
	for index, rule := range rules {
		if rule == nil {
			return linodego.FirewallRulesCreateOptions{}, fmt.Errorf("rule %d is nil", index+1)
		}
		addresses := firewallRuleAddresses(rule.SourceIPs)
		switch strings.ToLower(rule.Direction) {
		case "in":
			options.Inbound = append(options.Inbound, linodego.FirewallRuleInbound{
				Action:    fi.ValueOf(rule.Action),
				Label:     rule.Label,
				Ports:     fi.ValueOf(rule.Port),
				Protocol:  linodego.NetworkProtocol(strings.ToUpper(rule.Protocol)),
				Addresses: addresses,
			})
		case "out":
			options.Outbound = append(options.Outbound, linodego.FirewallRuleOutbound{
				Action:    fi.ValueOf(rule.Action),
				Label:     rule.Label,
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
func firewallRulesToUpdateOptions(inboundPolicy, outboundPolicy *string, rules []*FirewallRule) (linodego.FirewallRulesUpdateOptions, error) {
	createOptions, err := firewallRulesToLinode(inboundPolicy, outboundPolicy, rules)
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
		result = append(result, firewallRuleFromLinode("in", rule.Label, string(rule.Protocol), rule.Action, rule.Ports, rule.Addresses))
	}
	for _, rule := range rules.Outbound {
		result = append(result, firewallRuleFromLinode("out", rule.Label, string(rule.Protocol), rule.Action, rule.Ports, rule.Addresses))
	}
	return result
}

// firewallRuleFromLinode converts a single Linode firewall rule to a FirewallRule object.
func firewallRuleFromLinode(direction, label, protocol, action, ports string, addresses linodego.NetworkAddresses) *FirewallRule {
	rule := &FirewallRule{
		Direction: direction,
		Label:     label,
		Protocol:  strings.ToLower(protocol),
		Action:    new(action),
	}
	if ports != "" {
		rule.Port = new(ports)
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
