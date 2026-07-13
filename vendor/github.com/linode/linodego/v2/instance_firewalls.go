package linodego

import (
	"context"
)

// ListInstanceFirewalls returns a paginated list of Cloud Firewalls for linodeID
func (c *Client) ListInstanceFirewalls(ctx context.Context, linodeID int, opts *ListOptions) ([]Firewall, error) {
	return getPaginatedResults[Firewall](ctx, c, formatAPIPath("linode/instances/%d/firewalls", linodeID), opts)
}

type InstanceFirewallUpdateOptions struct {
	FirewallIDs []int `json:"firewall_ids"`
}

// UpdateInstanceFirewalls updates the Cloud Firewalls for a Linode instance
// Followup this call with `ListInstanceFirewalls` to verify the changes if necessary.
func (c *Client) UpdateInstanceFirewalls(ctx context.Context, linodeID int, opts InstanceFirewallUpdateOptions) ([]Firewall, error) {
	return putPaginatedResults[Firewall, InstanceFirewallUpdateOptions](ctx, c, formatAPIPath("linode/instances/%d/firewalls", linodeID), nil, opts)
}
