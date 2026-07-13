package linodego

import (
	"context"
)

// IPv6RangeCreateOptions fields are those accepted by CreateIPv6Range
type IPv6RangeCreateOptions struct {
	LinodeID     int    `json:"linode_id,omitzero"`
	PrefixLength int    `json:"prefix_length"`
	RouteTarget  string `json:"route_target,omitzero"`
}

// ListIPv6Ranges lists IPv6Ranges
func (c *Client) ListIPv6Ranges(ctx context.Context, opts *ListOptions) ([]IPv6Range, error) {
	return getPaginatedResults[IPv6Range](ctx, c, "networking/ipv6/ranges", opts)
}

// GetIPv6Range gets details about an IPv6 range
func (c *Client) GetIPv6Range(ctx context.Context, ipRange string) (*IPv6Range, error) {
	e := formatAPIPath("networking/ipv6/ranges/%s", ipRange)
	return doGETRequest[IPv6Range](ctx, c, e)
}

// CreateIPv6Range creates an IPv6 Range and assigns it based on the provided Linode or route target IPv6 SLAAC address.
func (c *Client) CreateIPv6Range(ctx context.Context, opts IPv6RangeCreateOptions) (*IPv6Range, error) {
	return doPOSTRequest[IPv6Range](ctx, c, "networking/ipv6/ranges", opts)
}

// DeleteIPv6Range deletes an IPv6 Range.
func (c *Client) DeleteIPv6Range(ctx context.Context, ipRange string) error {
	e := formatAPIPath("networking/ipv6/ranges/%s", ipRange)
	return doDELETERequest(ctx, c, e)
}
