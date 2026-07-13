package linodego

import (
	"context"
)

// ListNodeBalancerFirewalls returns a paginated list of Cloud Firewalls for nodebalancerID
func (c *Client) ListNodeBalancerFirewalls(ctx context.Context, nodebalancerID int, opts *ListOptions) ([]Firewall, error) {
	return getPaginatedResults[Firewall](ctx, c, formatAPIPath("nodebalancers/%d/firewalls", nodebalancerID), opts)
}
