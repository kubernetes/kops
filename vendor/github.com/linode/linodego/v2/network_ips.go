package linodego

import (
	"context"
)

// IPAddressUpdateOptions fields are those accepted by UpdateIPAddress.
// NOTE: An IP's RDNS can be reset to default using the following pattern:
//
//	IPAddressUpdateOptions{
//		RDNS: linodego.Pointer[*string](nil),
//	}
type IPAddressUpdateOptions struct {
	// The reverse DNS assigned to this address. For public IPv4 addresses, this will be set to a default value provided by Linode if set to nil.
	Reserved *bool    `json:"reserved,omitzero"`
	RDNS     **string `json:"rdns,omitzero"`
}

// LinodeIPAssignment stores an assignment between an IP address and a Linode instance.
type LinodeIPAssignment struct {
	Address  string `json:"address"`
	LinodeID int    `json:"linode_id"`
}

type AllocateReserveIPOptions struct {
	Type     string `json:"type"`
	Public   bool   `json:"public"`
	Reserved bool   `json:"reserved,omitzero"`
	Region   string `json:"region,omitzero"`
	LinodeID int    `json:"linode_id,omitzero"`
}

// LinodesAssignIPsOptions fields are those accepted by InstancesAssignIPs.
type LinodesAssignIPsOptions struct {
	Region string `json:"region"`

	Assignments []LinodeIPAssignment `json:"assignments"`
}

// IPAddressesShareOptions fields are those accepted by ShareIPAddresses.
type IPAddressesShareOptions struct {
	IPs      []string `json:"ips"`
	LinodeID int      `json:"linode_id"`
}

// ListIPAddressesQuery fields are those accepted as query params for the
// ListIPAddresses function.
type ListIPAddressesQuery struct {
	SkipIPv6RDNS bool `query:"skip_ipv6_rdns"`
}

// GetUpdateOptions converts a IPAddress to IPAddressUpdateOptions for use in UpdateIPAddress.
func (i InstanceIP) GetUpdateOptions() IPAddressUpdateOptions {
	rdns := copyString(&i.RDNS)

	return IPAddressUpdateOptions{
		RDNS:     &rdns,
		Reserved: copyBool(&i.Reserved),
	}
}

// ListIPAddresses lists IPAddresses.
func (c *Client) ListIPAddresses(ctx context.Context, opts *ListOptions) ([]InstanceIP, error) {
	return getPaginatedResults[InstanceIP](ctx, c, "networking/ips", opts)
}

// GetIPAddress gets the IPAddress with the provided IP.
func (c *Client) GetIPAddress(ctx context.Context, id string) (*InstanceIP, error) {
	e := formatAPIPath("networking/ips/%s", id)
	return doGETRequest[InstanceIP](ctx, c, e)
}

// UpdateIPAddress updates the IP address with the specified address.
func (c *Client) UpdateIPAddress(ctx context.Context, address string, opts IPAddressUpdateOptions) (*InstanceIP, error) {
	e := formatAPIPath("networking/ips/%s", address)
	return doPUTRequest[InstanceIP](ctx, c, e, opts)
}

// InstancesAssignIPs assigns multiple IPv4 addresses and/or IPv6 ranges to multiple Linodes in one Region.
// This allows swapping, shuffling, or otherwise reorganizing IPs to your Linodes.
func (c *Client) InstancesAssignIPs(ctx context.Context, opts LinodesAssignIPsOptions) error {
	return doPOSTRequestNoResponseBody(ctx, c, "networking/ips/assign", opts)
}

// ShareIPAddresses allows IP address reassignment (also referred to as IP failover)
// from one Linode to another if the primary Linode becomes unresponsive.
func (c *Client) ShareIPAddresses(ctx context.Context, opts IPAddressesShareOptions) error {
	return doPOSTRequestNoResponseBody(ctx, c, "networking/ips/share", opts)
}

// AllocateReserveIP allocates a new IPv4 address to the Account, with the option to reserve it
// and optionally assign it to a Linode.
func (c *Client) AllocateReserveIP(ctx context.Context, opts AllocateReserveIPOptions) (*InstanceIP, error) {
	return doPOSTRequest[InstanceIP](ctx, c, "networking/ips", opts)
}
