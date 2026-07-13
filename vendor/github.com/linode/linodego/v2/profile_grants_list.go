package linodego

import (
	"context"
)

type GrantsListResponse = UserGrants

func (c *Client) GrantsList(ctx context.Context) (*GrantsListResponse, error) {
	return doGETRequest[GrantsListResponse](ctx, c, "profile/grants")
}
