package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// FirewallStatus enum type
type FirewallStatus string

// FirewallStatus enums start with Firewall
const (
	FirewallEnabled  FirewallStatus = "enabled"
	FirewallDisabled FirewallStatus = "disabled"
	FirewallDeleted  FirewallStatus = "deleted"
)

// A Firewall is a set of networking rules (iptables) applied to Devices with which it is associated
type Firewall struct {
	ID       int                    `json:"id"`
	Label    string                 `json:"label"`
	Status   FirewallStatus         `json:"status"`
	Tags     []string               `json:"tags"`
	Rules    FirewallRules          `json:"rules"`
	Entities []FirewallDeviceEntity `json:"entities"`
	Created  *time.Time             `json:"-"`
	Updated  *time.Time             `json:"-"`
}

// DevicesCreationOptions fields are used when adding devices during the Firewall creation process.
type DevicesCreationOptions struct {
	Linodes          []int `json:"linodes,omitzero"`
	NodeBalancers    []int `json:"nodebalancers,omitzero"`
	LinodeInterfaces []int `json:"linode_interfaces,omitzero"`
}

// FirewallCreateOptions fields are those accepted by CreateFirewall
type FirewallCreateOptions struct {
	Label   string                     `json:"label,omitzero"`
	Rules   FirewallRulesCreateOptions `json:"rules"`
	Tags    []string                   `json:"tags,omitzero"`
	Devices DevicesCreationOptions     `json:"devices,omitzero"`
}

type FirewallRulesCreateOptions struct {
	Inbound        []FirewallRuleInbound  `json:"inbound"`
	InboundPolicy  string                 `json:"inbound_policy"`
	Outbound       []FirewallRuleOutbound `json:"outbound"`
	OutboundPolicy string                 `json:"outbound_policy"`
}

// FirewallUpdateOptions is an options struct used when Updating a Firewall
type FirewallUpdateOptions struct {
	Label  string         `json:"label,omitzero"`
	Status FirewallStatus `json:"status,omitzero"`
	Tags   []string       `json:"tags,omitzero"`
}

// FirewallSettings represents the default firewalls for Linodes,
// Linode VPC and public interfaces, and NodeBalancers.
type FirewallSettings struct {
	DefaultFirewallIDs DefaultFirewallIDs `json:"default_firewall_ids"`
}

type DefaultFirewallIDs struct {
	Linode          *int `json:"linode"`
	NodeBalancer    *int `json:"nodebalancer"`
	PublicInterface *int `json:"public_interface"`
	VPCInterface    *int `json:"vpc_interface"`
}

// FirewallSettingsUpdateOptions is an options struct used when Updating FirewallSettings
type FirewallSettingsUpdateOptions struct {
	DefaultFirewallIDs *DefaultFirewallIDsOptions `json:"default_firewall_ids,omitzero"`
}

type DefaultFirewallIDsOptions struct {
	Linode          **int `json:"linode,omitzero"`
	NodeBalancer    **int `json:"nodebalancer,omitzero"`
	PublicInterface **int `json:"public_interface,omitzero"`
	VPCInterface    **int `json:"vpc_interface,omitzero"`
}

// GetUpdateOptions converts a Firewall to FirewallUpdateOptions for use in Client.UpdateFirewall.
func (f *Firewall) GetUpdateOptions() FirewallUpdateOptions {
	return FirewallUpdateOptions{
		Label:  f.Label,
		Status: f.Status,
		Tags:   f.Tags,
	}
}

// UnmarshalJSON for Firewall responses
func (f *Firewall) UnmarshalJSON(b []byte) error {
	type Mask Firewall

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(f),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	f.Created = (*time.Time)(p.Created)
	f.Updated = (*time.Time)(p.Updated)

	return nil
}

// ListFirewalls returns a paginated list of Cloud Firewalls
func (c *Client) ListFirewalls(ctx context.Context, opts *ListOptions) ([]Firewall, error) {
	return getPaginatedResults[Firewall](ctx, c, "networking/firewalls", opts)
}

// CreateFirewall creates a single Firewall with at least one set of inbound or outbound rules
func (c *Client) CreateFirewall(ctx context.Context, opts FirewallCreateOptions) (*Firewall, error) {
	return doPOSTRequest[Firewall](ctx, c, "networking/firewalls", opts)
}

// GetFirewall gets a single Firewall with the provided ID
func (c *Client) GetFirewall(ctx context.Context, firewallID int) (*Firewall, error) {
	e := formatAPIPath("networking/firewalls/%d", firewallID)
	return doGETRequest[Firewall](ctx, c, e)
}

// UpdateFirewall updates a Firewall with the given ID
func (c *Client) UpdateFirewall(ctx context.Context, firewallID int, opts FirewallUpdateOptions) (*Firewall, error) {
	e := formatAPIPath("networking/firewalls/%d", firewallID)
	return doPUTRequest[Firewall](ctx, c, e, opts)
}

// DeleteFirewall deletes a single Firewall with the provided ID
func (c *Client) DeleteFirewall(ctx context.Context, firewallID int) error {
	e := formatAPIPath("networking/firewalls/%d", firewallID)
	return doDELETERequest(ctx, c, e)
}

// GetFirewallSettings returns default firewalls for Linodes, Linode VPC and public interfaces, and NodeBalancers.
func (c *Client) GetFirewallSettings(ctx context.Context) (*FirewallSettings, error) {
	return doGETRequest[FirewallSettings](ctx, c, "networking/firewalls/settings")
}

// UpdateFirewallSettings updates the default firewalls for Linodes, Linode VPC and public interfaces, and NodeBalancers.
func (c *Client) UpdateFirewallSettings(ctx context.Context, opts FirewallSettingsUpdateOptions) (*FirewallSettings, error) {
	return doPUTRequest[FirewallSettings](ctx, c, "networking/firewalls/settings", opts)
}
