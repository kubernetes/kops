package linodego

import (
	"context"
	"encoding/json"
)

// NetworkProtocol is used for firewall rule protocol fields.
//
// This can also represent arbitrary protocols, for example an IP protocol number
// like `NetworkProtocol("50")`.
//
// NOTE: ALL and numeric protocols may not yet
// be available to all users.
type NetworkProtocol string

// NetworkProtocol enum values
const (
	TCP     NetworkProtocol = "TCP"
	UDP     NetworkProtocol = "UDP"
	ICMP    NetworkProtocol = "ICMP"
	IPENCAP NetworkProtocol = "IPENCAP"

	AllNetworkProtocols NetworkProtocol = "ALL"
)

// NetworkAddresses are arrays of ipv4 and v6 addresses
type NetworkAddresses struct {
	IPv4 []string `json:"ipv4,omitzero"`
	IPv6 []string `json:"ipv6,omitzero"`
}

type FirewallRuleInbound struct {
	Action      string           `json:"action"`
	Label       string           `json:"label"`
	Description string           `json:"description,omitzero"`
	Ports       string           `json:"ports,omitzero"`
	Protocol    NetworkProtocol  `json:"protocol"`
	Addresses   NetworkAddresses `json:"addresses"`

	// FirewallRuleInbound references one `Rule Set` by ID. When provided, this entry
	// represents a reference and should be mutually exclusive with ordinary
	// rule fields according to the API contract.
	RuleSet int `json:"ruleset,omitzero"`
}

type FirewallRuleOutbound struct {
	Action      string           `json:"action"`
	Label       string           `json:"label"`
	Description string           `json:"description,omitzero"`
	Ports       string           `json:"ports,omitzero"`
	Protocol    NetworkProtocol  `json:"protocol"`
	Addresses   NetworkAddresses `json:"addresses"`

	// FirewallRuleOutbound references one `Rule Set` by ID. When provided, this entry
	// represents a reference and should be mutually exclusive with ordinary
	// rule fields according to the API contract.
	RuleSet int `json:"ruleset,omitzero"`
}

// MarshalJSON ensures that when a rule references a Rule Set (FirewallRuleSet != 0),
// only the reference shape { "ruleset": <id> } is emitted. Otherwise, the
// ordinary rule fields are emitted without the ruleset key.
func (r FirewallRuleInbound) MarshalJSON() ([]byte, error) {
	if r.RuleSet != 0 {
		type rulesetOnly struct {
			RuleSet int `json:"ruleset"`
		}

		return json.Marshal(rulesetOnly{RuleSet: r.RuleSet})
	}

	type normal struct {
		Action      string           `json:"action"`
		Label       string           `json:"label"`
		Description string           `json:"description,omitzero"`
		Ports       string           `json:"ports,omitzero"`
		Protocol    NetworkProtocol  `json:"protocol"`
		Addresses   NetworkAddresses `json:"addresses"`
	}

	return json.Marshal(normal{
		Action:      r.Action,
		Label:       r.Label,
		Description: r.Description,
		Ports:       r.Ports,
		Protocol:    r.Protocol,
		Addresses:   r.Addresses,
	})
}

// MarshalJSON ensures that when a rule references a Rule Set (FirewallRuleSet != 0),
// only the reference shape { "ruleset": <id> } is emitted. Otherwise, the
// ordinary rule fields are emitted without the ruleset key.
func (r FirewallRuleOutbound) MarshalJSON() ([]byte, error) {
	if r.RuleSet != 0 {
		type rulesetOnly struct {
			RuleSet int `json:"ruleset"`
		}

		return json.Marshal(rulesetOnly{RuleSet: r.RuleSet})
	}

	type normal struct {
		Action      string           `json:"action"`
		Label       string           `json:"label"`
		Description string           `json:"description,omitzero"`
		Ports       string           `json:"ports,omitzero"`
		Protocol    NetworkProtocol  `json:"protocol"`
		Addresses   NetworkAddresses `json:"addresses"`
	}

	return json.Marshal(normal{
		Action:      r.Action,
		Label:       r.Label,
		Description: r.Description,
		Ports:       r.Ports,
		Protocol:    r.Protocol,
		Addresses:   r.Addresses,
	})
}

// FirewallRules is a pair of inbound and outbound rules that specify what network traffic should be allowed.
type FirewallRules struct {
	Inbound        []FirewallRuleInbound  `json:"inbound"`
	InboundPolicy  string                 `json:"inbound_policy"`
	Outbound       []FirewallRuleOutbound `json:"outbound"`
	OutboundPolicy string                 `json:"outbound_policy"`
	Version        int                    `json:"version,omitzero"`
	Fingerprint    string                 `json:"fingerprint,omitzero"`
}
type FirewallRulesUpdateOptions struct {
	Inbound        []FirewallRuleInbound  `json:"inbound"`
	InboundPolicy  string                 `json:"inbound_policy"`
	Outbound       []FirewallRuleOutbound `json:"outbound"`
	OutboundPolicy string                 `json:"outbound_policy"`
}

// GetFirewallRules gets the FirewallRules for the given Firewall.
func (c *Client) GetFirewallRules(ctx context.Context, firewallID int) (*FirewallRules, error) {
	e := formatAPIPath("networking/firewalls/%d/rules", firewallID)
	return doGETRequest[FirewallRules](ctx, c, e)
}

// GetFirewallRulesExpansion gets the expanded FirewallRules for the given Firewall.
func (c *Client) GetFirewallRulesExpansion(ctx context.Context, firewallID int) (*FirewallRules, error) {
	e := formatAPIPath("networking/firewalls/%d/rules/expansion", firewallID)
	return doGETRequest[FirewallRules](ctx, c, e)
}

// UpdateFirewallRules updates the FirewallRules for the given Firewall
func (c *Client) UpdateFirewallRules(ctx context.Context, firewallID int, rules FirewallRulesUpdateOptions) (*FirewallRules, error) {
	e := formatAPIPath("networking/firewalls/%d/rules", firewallID)
	return doPUTRequest[FirewallRules](ctx, c, e, rules)
}
