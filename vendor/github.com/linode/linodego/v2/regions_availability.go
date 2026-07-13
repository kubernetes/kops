package linodego

import (
	"context"
)

// RegionAvailability represents a linode region object.
type RegionAvailability struct {
	Region    string `json:"region"`
	Plan      string `json:"plan"`
	Available bool   `json:"available"`
}

// RegionVPCAvailability represents a linode region vpc availability object.
type RegionVPCAvailability struct {
	Region                     string `json:"region"`
	Available                  bool   `json:"available"`
	AvailableIPV6PrefixLengths []int  `json:"available_ipv6_prefix_lengths"`
}

// ListRegionsAvailability lists Regions. This endpoint is cached by default.
func (c *Client) ListRegionsAvailability(ctx context.Context, opts *ListOptions) ([]RegionAvailability, error) {
	e := "regions/availability"

	endpoint, err := generateListCacheURL(e, opts)
	if err != nil {
		return nil, err
	}

	if result := c.getCachedResponse(endpoint); result != nil {
		return result.([]RegionAvailability), nil
	}

	response, err := getPaginatedResults[RegionAvailability](ctx, c, e, opts)
	if err != nil {
		return nil, err
	}

	c.addCachedResponse(endpoint, response, &cacheExpiryTime)

	return response, nil
}

// GetRegionAvailability gets availability for all plans in the provided region. This endpoint is cached by default.
func (c *Client) GetRegionAvailability(ctx context.Context, regionID string) ([]RegionAvailability, error) {
	e := formatAPIPath("regions/%s/availability", regionID)

	if result := c.getCachedResponse(e); result != nil {
		return result.([]RegionAvailability), nil
	}

	response, err := doGETRequest[[]RegionAvailability](ctx, c, e)
	if err != nil {
		return nil, err
	}

	c.addCachedResponse(e, *response, &cacheExpiryTime)

	return *response, nil
}

// ListRegionsVPCAvailability lists VPC availability data for all regions.
// NOTE: IPv6 VPCs may not currently be available to all users.
func (c *Client) ListRegionsVPCAvailability(ctx context.Context, opts *ListOptions) ([]RegionVPCAvailability, error) {
	e := "regions/vpc-availability"
	return getPaginatedResults[RegionVPCAvailability](ctx, c, e, opts)
}

// GetRegionVPCAvailability gets VPC availability data for a single region.
// NOTE: IPv6 VPCs may not currently be available to all users.
func (c *Client) GetRegionVPCAvailability(ctx context.Context, regionID string) (*RegionVPCAvailability, error) {
	e := formatAPIPath("regions/%s/vpc-availability", regionID)
	return doGETRequest[RegionVPCAvailability](ctx, c, e)
}
