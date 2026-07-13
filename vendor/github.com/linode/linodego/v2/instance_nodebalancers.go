package linodego

import (
	"context"
)

// ListInstanceNodeBalancers lists NodeBalancers that the provided instance is a node in
func (c *Client) ListInstanceNodeBalancers(ctx context.Context, linodeID int, opts *ListOptions) ([]NodeBalancer, error) {
	return getPaginatedResults[NodeBalancer](ctx, c, formatAPIPath("linode/instances/%d/nodebalancers", linodeID), opts)
}
