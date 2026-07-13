package linodego

import (
	"context"
)

// InstanceIPAddressResponse contains the IPv4 and IPv6 details for an Instance
type InstanceIPAddressResponse struct {
	IPv4 *InstanceIPv4Response `json:"ipv4"`
	IPv6 *InstanceIPv6Response `json:"ipv6"`
}

// InstanceIPv4Response contains the details of all IPv4 addresses associated with an Instance
type InstanceIPv4Response struct {
	Public   []InstanceIP `json:"public"`
	Private  []InstanceIP `json:"private"`
	Shared   []InstanceIP `json:"shared"`
	Reserved []InstanceIP `json:"reserved"`
	VPC      []VPCIP      `json:"vpc"`
}

// InstanceIP represents an Instance IP with additional DNS and networking details
type InstanceIP struct {
	Address        string                    `json:"address"`
	Gateway        string                    `json:"gateway"`
	SubnetMask     string                    `json:"subnet_mask"`
	Prefix         int                       `json:"prefix"`
	Type           InstanceIPType            `json:"type"`
	Public         bool                      `json:"public"`
	RDNS           string                    `json:"rdns"`
	LinodeID       int                       `json:"linode_id"`
	InterfaceID    *int                      `json:"interface_id"`
	Region         string                    `json:"region"`
	VPCNAT1To1     *InstanceIPNAT1To1        `json:"vpc_nat_1_1"`
	Reserved       bool                      `json:"reserved"`
	Tags           []string                  `json:"tags"`
	AssignedEntity *ReservedIPAssignedEntity `json:"assigned_entity"`
}

type InstanceIPAddOptions struct {
	Type   string `json:"type"`
	Public bool   `json:"public"`
}

type InstanceIPAddressUpdateOptions struct {
	RDNS **string `json:"rdns,omitzero"`
}

// VPCIP represents a private IP address in a VPC subnet with additional networking details
type VPCIP struct {
	Address      *string `json:"address"`
	AddressRange *string `json:"address_range"`
	Gateway      string  `json:"gateway"`
	SubnetMask   string  `json:"subnet_mask"`
	Prefix       int     `json:"prefix"`
	LinodeID     int     `json:"linode_id"`
	Region       string  `json:"region"`
	Active       bool    `json:"active"`
	NAT1To1      *string `json:"nat_1_1"`
	VPCID        int     `json:"vpc_id"`
	SubnetID     int     `json:"subnet_id"`
	InterfaceID  int     `json:"interface_id"`
	// NOTE: NodeBalancerID and DatabaseID may not currently be available to all users.
	NodeBalancerID *int `json:"nodebalancer_id"`
	DatabaseID     *int `json:"database_id"`
	// NOTE: IPv6 VPCs may not currently be available to all users.
	IPv6Range     *string            `json:"ipv6_range"`
	IPv6IsPublic  *bool              `json:"ipv6_is_public"`
	IPv6Addresses []VPCIPIPv6Address `json:"ipv6_addresses"`

	// The type of this field will be made a pointer in the next major release of linodego.
	ConfigID int `json:"config_id"`
}

// VPCIPIPv6Address represents a single IPv6 address under a VPCIP.
// NOTE: IPv6 VPCs may not currently be available to all users.
type VPCIPIPv6Address struct {
	SLAACAddress string `json:"slaac_address"`
}

// InstanceIPv6Response contains the IPv6 addresses and ranges for an Instance
type InstanceIPv6Response struct {
	LinkLocal *InstanceIP `json:"link_local"`
	SLAAC     *InstanceIP `json:"slaac"`
	Global    []IPv6Range `json:"global"`
	// NOTE: IPv6 VPCs may not currently be available to all users.
	VPC []VPCIP `json:"vpc"`
}

// InstanceIPNAT1To1 contains information about the NAT 1:1 mapping
// of a public IP address to a VPC subnet.
type InstanceIPNAT1To1 struct {
	Address  string `json:"address"`
	SubnetID int    `json:"subnet_id"`
	VPCID    int    `json:"vpc_id"`
}

// IPv6Range represents a range of IPv6 addresses routed to a single Linode in a given Region
type IPv6Range struct {
	Range  string `json:"range"`
	Region string `json:"region"`
	Prefix int    `json:"prefix"`

	RouteTarget string `json:"route_target"`

	// These fields are only returned by GetIPv6Range(...)
	IsBGP   bool  `json:"is_bgp"`
	Linodes []int `json:"linodes"`
}

type InstanceReserveIPOptions struct {
	Type    string `json:"type"`
	Public  bool   `json:"public"`
	Address string `json:"address"`
}

// InstanceIPType constants start with IPType and include Linode Instance IP Types
type InstanceIPType string

// InstanceIPType constants represent the IP types an Instance IP may be
const (
	IPTypeIPv4      InstanceIPType = "ipv4"
	IPTypeIPv6      InstanceIPType = "ipv6"
	IPTypeIPv6Pool  InstanceIPType = "ipv6/pool"
	IPTypeIPv6Range InstanceIPType = "ipv6/range"
)

// GetInstanceIPAddresses gets the IPAddresses for a Linode instance
func (c *Client) GetInstanceIPAddresses(ctx context.Context, linodeID int) (*InstanceIPAddressResponse, error) {
	e := formatAPIPath("linode/instances/%d/ips", linodeID)
	return doGETRequest[InstanceIPAddressResponse](ctx, c, e)
}

// GetInstanceIPAddress gets the IPAddress for a Linode instance matching a supplied IP address
func (c *Client) GetInstanceIPAddress(ctx context.Context, linodeID int, ipaddress string) (*InstanceIP, error) {
	e := formatAPIPath("linode/instances/%d/ips/%s", linodeID, ipaddress)
	return doGETRequest[InstanceIP](ctx, c, e)
}

// AddInstanceIPAddress adds a public or private IP to a Linode instance
func (c *Client) AddInstanceIPAddress(ctx context.Context, linodeID int, opts InstanceIPAddOptions) (*InstanceIP, error) {
	opts.Type = "ipv4"
	e := formatAPIPath("linode/instances/%d/ips", linodeID)

	return doPOSTRequest[InstanceIP](ctx, c, e, opts)
}

// UpdateInstanceIPAddress updates the IPAddress with the specified instance id and IP address
func (c *Client) UpdateInstanceIPAddress(ctx context.Context, linodeID int, ipAddress string, opts InstanceIPAddressUpdateOptions) (*InstanceIP, error) {
	e := formatAPIPath("linode/instances/%d/ips/%s", linodeID, ipAddress)
	return doPUTRequest[InstanceIP](ctx, c, e, opts)
}

func (c *Client) DeleteInstanceIPAddress(ctx context.Context, linodeID int, ipAddress string) error {
	e := formatAPIPath("linode/instances/%d/ips/%s", linodeID, ipAddress)
	return doDELETERequest(ctx, c, e)
}

// AssignInstanceReservedIP adds additional reserved IPV4 addresses to an existing linode
func (c *Client) AssignInstanceReservedIP(ctx context.Context, linodeID int, opts InstanceReserveIPOptions) (*InstanceIP, error) {
	endpoint := formatAPIPath("linode/instances/%d/ips", linodeID)
	return doPOSTRequest[InstanceIP](ctx, c, endpoint, opts)
}
