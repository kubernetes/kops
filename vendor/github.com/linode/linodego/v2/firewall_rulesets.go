package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// FirewallRuleSetType represents the type of rules a Rule Set contains.
// Valid values are "inbound" and "outbound".
type FirewallRuleSetType string

const (
	FirewallRuleSetTypeInbound  FirewallRuleSetType = "inbound"
	FirewallRuleSetTypeOutbound FirewallRuleSetType = "outbound"
)

// FirewallRuleSet represents the Rule Set resource.
// Note: created/updated/deleted are parsed via UnmarshalJSON into time.Time pointers.
type FirewallRuleSet struct {
	ID               int                   `json:"id"`
	Label            string                `json:"label"`
	Description      string                `json:"description,omitzero"`
	Type             FirewallRuleSetType   `json:"type"`
	Rules            []FirewallRuleSetRule `json:"rules"`
	IsServiceDefined bool                  `json:"is_service_defined"`
	Version          int                   `json:"version"`

	Created *time.Time `json:"-"`
	Updated *time.Time `json:"-"`
	Deleted *time.Time `json:"-"`
}

// A FirewallRuleSetRule is a whitelist of ports, protocols, and addresses for which traffic should be allowed.
// The ipv4/ipv6 address lists may contain Prefix List tokens (for example, "pl::..." or "pl:system:...")
// in addition to literal IP addresses.
type FirewallRuleSetRule struct {
	Action    string           `json:"action"`
	Label     string           `json:"label"`
	Ports     string           `json:"ports,omitzero"`
	Protocol  NetworkProtocol  `json:"protocol"`
	Addresses NetworkAddresses `json:"addresses"`
}

// FirewallRuleSetRuleCreateOptions fields accepted in Firewall Rule Set create payloads.
type FirewallRuleSetRuleCreateOptions struct {
	Action    string           `json:"action"`
	Label     string           `json:"label"`
	Ports     string           `json:"ports,omitzero"`
	Protocol  NetworkProtocol  `json:"protocol"`
	Addresses NetworkAddresses `json:"addresses"`
}

// FirewallRuleSetRuleUpdateOptions fields accepted in Firewall Rule Set update payloads.
type FirewallRuleSetRuleUpdateOptions struct {
	Action    string           `json:"action"`
	Label     string           `json:"label"`
	Ports     string           `json:"ports,omitzero"`
	Protocol  NetworkProtocol  `json:"protocol"`
	Addresses NetworkAddresses `json:"addresses"`
}

// UnmarshalJSON implements custom timestamp parsing for FirewallRuleSet.
func (r *FirewallRuleSet) UnmarshalJSON(b []byte) error {
	type Mask FirewallRuleSet

	aux := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
		Deleted *parseabletime.ParseableTime `json:"deleted"`
	}{
		Mask: (*Mask)(r),
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	if aux.Created != nil {
		r.Created = (*time.Time)(aux.Created)
	}

	if aux.Updated != nil {
		r.Updated = (*time.Time)(aux.Updated)
	}

	if aux.Deleted != nil {
		r.Deleted = (*time.Time)(aux.Deleted)
	}

	return nil
}

// FirewallRuleSetCreateOptions fields accepted by CreateRuleSet.
type FirewallRuleSetCreateOptions struct {
	Label       string                             `json:"label"`
	Description string                             `json:"description,omitzero"`
	Type        FirewallRuleSetType                `json:"type"`
	Rules       []FirewallRuleSetRuleCreateOptions `json:"rules"`
}

// FirewallRuleSetUpdateOptions fields accepted by UpdateRuleSet.
// Omit a top-level field to leave it unchanged. If Rules is provided, it
// replaces the entire ordered rules array.
type FirewallRuleSetUpdateOptions struct {
	Label       *string                            `json:"label,omitzero"`
	Description *string                            `json:"description,omitzero"`
	Rules       []FirewallRuleSetRuleUpdateOptions `json:"rules,omitzero"`
}

// ListFirewallRuleSets returns a paginated list of Rule Sets.
// Supports filtering (e.g., by label) via ListOptions.Filter.
func (c *Client) ListFirewallRuleSets(ctx context.Context, opts *ListOptions) ([]FirewallRuleSet, error) {
	return getPaginatedResults[FirewallRuleSet](ctx, c, "networking/firewalls/rulesets", opts)
}

// CreateFirewallRuleSet creates a new Rule Set.
func (c *Client) CreateFirewallRuleSet(ctx context.Context, opts FirewallRuleSetCreateOptions) (*FirewallRuleSet, error) {
	return doPOSTRequest[FirewallRuleSet](ctx, c, "networking/firewalls/rulesets", opts)
}

// GetFirewallRuleSet fetches a Rule Set by ID.
func (c *Client) GetFirewallRuleSet(ctx context.Context, rulesetID int) (*FirewallRuleSet, error) {
	e := formatAPIPath("networking/firewalls/rulesets/%d", rulesetID)
	return doGETRequest[FirewallRuleSet](ctx, c, e)
}

// UpdateFirewallRuleSet updates a Rule Set by ID.
func (c *Client) UpdateFirewallRuleSet(ctx context.Context, rulesetID int, opts FirewallRuleSetUpdateOptions) (*FirewallRuleSet, error) {
	e := formatAPIPath("networking/firewalls/rulesets/%d", rulesetID)
	return doPUTRequest[FirewallRuleSet](ctx, c, e, opts)
}

// DeleteFirewallRuleSet deletes a Rule Set by ID.
func (c *Client) DeleteFirewallRuleSet(ctx context.Context, rulesetID int) error {
	e := formatAPIPath("networking/firewalls/rulesets/%d", rulesetID)
	return doDELETERequest(ctx, c, e)
}
