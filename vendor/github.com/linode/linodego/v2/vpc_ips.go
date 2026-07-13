package linodego

import (
	"context"
	"fmt"
)

// ListAllVPCIPAddresses gets the list of all IP addresses of all VPCs in the Linode account.
func (c *Client) ListAllVPCIPAddresses(
	ctx context.Context, opts *ListOptions,
) ([]VPCIP, error) {
	return getPaginatedResults[VPCIP](ctx, c, "vpcs/ips", opts)
}

// ListVPCIPAddresses gets the list of all IP addresses of a specific VPC.
func (c *Client) ListVPCIPAddresses(
	ctx context.Context, vpcID int, opts *ListOptions,
) ([]VPCIP, error) {
	return getPaginatedResults[VPCIP](ctx, c, fmt.Sprintf("vpcs/%d/ips", vpcID), opts)
}

// ListAllVPCIPv6Addresses gets a list of all IPv6 addresses related to all VPCs
// accessible by the current Linode account.
// NOTE: IPv6 VPCs may not currently be available to all users.
func (c *Client) ListAllVPCIPv6Addresses(
	ctx context.Context, opts *ListOptions,
) ([]VPCIP, error) {
	return getPaginatedResults[VPCIP](ctx, c, "vpcs/ipv6s", opts)
}

// ListVPCIPv6Addresses gets the list of all IPv6 addresses of a specific VPC.
// NOTE: IPv6 VPCs may not currently be available to all users.
func (c *Client) ListVPCIPv6Addresses(
	ctx context.Context, vpcID int, opts *ListOptions,
) ([]VPCIP, error) {
	return getPaginatedResults[VPCIP](ctx, c, fmt.Sprintf("vpcs/%d/ipv6s", vpcID), opts)
}
