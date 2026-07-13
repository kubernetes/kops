package linodego

import (
	"context"
)

// ReservedIPAssignedEntity represents the entity that a reserved IP is assigned to.
// NOTE: Reserved IP feature may not currently be available to all users.
type ReservedIPAssignedEntity struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// ReserveIPOptions represents the options for reserving an IP address
// NOTE: Reserved IP feature may not currently be available to all users.
type ReserveIPOptions struct {
	Region string   `json:"region"`
	Tags   []string `json:"tags,omitzero"`
}

// UpdateReservedIPOptions represents the options for updating a reserved IP address
// NOTE: Reserved IP feature may not currently be available to all users.
type UpdateReservedIPOptions struct {
	Tags []string `json:"tags,omitzero"`
}

// ReservedIPPrice represents the pricing information for a reserved IP type.
// It is an alias of the shared baseTypePrice to keep pricing consistent across resources.
type ReservedIPPrice = baseTypePrice

// ReservedIPRegionPrice represents region-specific pricing for a reserved IP type.
// It is an alias of the shared baseTypeRegionPrice to keep region pricing consistent across resources.
type ReservedIPRegionPrice = baseTypeRegionPrice

// ReservedIPType represents a reserved IP type with pricing information.
// It reuses the generic baseType to avoid duplicating type/pricing structures.
type ReservedIPType = baseType[ReservedIPPrice, ReservedIPRegionPrice]

// ListReservedIPAddresses retrieves a list of reserved IP addresses
// NOTE: Reserved IP feature may not currently be available to all users.
func (c *Client) ListReservedIPAddresses(ctx context.Context, opts *ListOptions) ([]InstanceIP, error) {
	e := formatAPIPath("networking/reserved/ips")
	return getPaginatedResults[InstanceIP](ctx, c, e, opts)
}

// GetReservedIPAddress retrieves details of a specific reserved IP address
// NOTE: Reserved IP feature may not currently be available to all users.
func (c *Client) GetReservedIPAddress(ctx context.Context, ipAddress string) (*InstanceIP, error) {
	e := formatAPIPath("networking/reserved/ips/%s", ipAddress)
	return doGETRequest[InstanceIP](ctx, c, e)
}

// ReserveIPAddress reserves a new IP address
// NOTE: Reserved IP feature may not currently be available to all users.
func (c *Client) ReserveIPAddress(ctx context.Context, opts ReserveIPOptions) (*InstanceIP, error) {
	return doPOSTRequest[InstanceIP](ctx, c, "networking/reserved/ips", opts)
}

// UpdateReservedIPAddress updates the tags of a reserved IP address
// NOTE: Reserved IP feature may not currently be available to all users.
func (c *Client) UpdateReservedIPAddress(ctx context.Context, address string, opts UpdateReservedIPOptions) (*InstanceIP, error) {
	e := formatAPIPath("networking/reserved/ips/%s", address)
	return doPUTRequest[InstanceIP](ctx, c, e, opts)
}

// DeleteReservedIPAddress deletes a reserved IP address
// NOTE: Reserved IP feature may not currently be available to all users.
func (c *Client) DeleteReservedIPAddress(ctx context.Context, ipAddress string) error {
	e := formatAPIPath("networking/reserved/ips/%s", ipAddress)
	return doDELETERequest(ctx, c, e)
}

// ListReservedIPTypes retrieves a list of reserved IP types with pricing information
// NOTE: Reserved IP feature may not currently be available to all users.
func (c *Client) ListReservedIPTypes(ctx context.Context, opts *ListOptions) ([]ReservedIPType, error) {
	return getPaginatedResults[ReservedIPType](ctx, c, "networking/reserved/ips/types", opts)
}
