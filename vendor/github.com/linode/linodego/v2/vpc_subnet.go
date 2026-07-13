package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// VPCSubnetLinodeInterface represents an interface on a Linode that is currently
// assigned to this VPC subnet.
type VPCSubnetLinodeInterface struct {
	ID       int  `json:"id"`
	Active   bool `json:"active"`
	ConfigID *int `json:"config_id"`
}

// VPCSubnetLinode represents a Linode currently assigned to a VPC subnet.
type VPCSubnetLinode struct {
	ID         int                        `json:"id"`
	Interfaces []VPCSubnetLinodeInterface `json:"interfaces"`
}

// VPCSubnetDatabase represents a Linode currently assigned to a VPC subnet.
type VPCSubnetDatabase struct {
	ID         int      `json:"id"`
	IPv4Range  *string  `json:"ipv4_range"`
	IPv6Ranges []string `json:"ipv6_ranges"`
}

// VPCSubnetNodebalancersRanges represents a single range assigned to a node balancer.
type VPCSubnetNodebalancersRanges struct {
	Range string `json:"range"`
}

// VPCSubnetNodebalancers represents a node balancer currently assigned to a VPC subnet.
type VPCSubnetNodebalancers struct {
	ID         int                            `json:"id"`
	Ipv4Range  string                         `json:"ipv4_range"`
	Ipv6Ranges []VPCSubnetNodebalancersRanges `json:"ipv6_ranges"`
}

type VPCSubnet struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	IPv4  string `json:"ipv4"`

	// NOTE: IPv6 VPCs may not currently be available to all users.
	IPv6 []VPCIPv6Range `json:"ipv6"`

	Linodes       []VPCSubnetLinode        `json:"linodes"`
	Databases     []VPCSubnetDatabase      `json:"databases"`
	Nodebalancers []VPCSubnetNodebalancers `json:"nodebalancers"`

	Created *time.Time `json:"-"`
	Updated *time.Time `json:"-"`
}

type VPCSubnetCreateOptions struct {
	Label string `json:"label"`
	IPv4  string `json:"ipv4"`

	// NOTE: IPv6 VPCs may not currently be available to all users.
	IPv6 []VPCSubnetCreateOptionsIPv6 `json:"ipv6,omitzero"`
}

// VPCSubnetCreateOptionsIPv6 represents a single IPv6 range assigned to a VPC
// which is specified during a VPC subnet's creation.
// NOTE: IPv6 VPCs may not currently be available to all users.
type VPCSubnetCreateOptionsIPv6 struct {
	Range *string `json:"range,omitzero"`
}

type VPCSubnetUpdateOptions struct {
	Label string `json:"label"`
}

func (v *VPCSubnet) UnmarshalJSON(b []byte) error {
	type Mask VPCSubnet

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(v),
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	v.Created = (*time.Time)(p.Created)
	v.Updated = (*time.Time)(p.Updated)

	return nil
}

func (v VPCSubnet) GetCreateOptions() VPCSubnetCreateOptions {
	return VPCSubnetCreateOptions{
		Label: v.Label,
		IPv4:  v.IPv4,
		IPv6: mapSlice(v.IPv6, func(i VPCIPv6Range) VPCSubnetCreateOptionsIPv6 {
			return VPCSubnetCreateOptionsIPv6{
				Range: copyValue(&i.Range),
			}
		}),
	}
}

func (v VPCSubnet) GetUpdateOptions() VPCSubnetUpdateOptions {
	return VPCSubnetUpdateOptions{Label: v.Label}
}

func (c *Client) CreateVPCSubnet(
	ctx context.Context,
	opts VPCSubnetCreateOptions,
	vpcID int,
) (*VPCSubnet, error) {
	e := formatAPIPath("vpcs/%d/subnets", vpcID)
	return doPOSTRequest[VPCSubnet](ctx, c, e, opts)
}

func (c *Client) GetVPCSubnet(
	ctx context.Context,
	vpcID int,
	subnetID int,
) (*VPCSubnet, error) {
	e := formatAPIPath("vpcs/%d/subnets/%d", vpcID, subnetID)
	return doGETRequest[VPCSubnet](ctx, c, e)
}

func (c *Client) ListVPCSubnets(
	ctx context.Context,
	vpcID int,
	opts *ListOptions,
) ([]VPCSubnet, error) {
	return getPaginatedResults[VPCSubnet](ctx, c, formatAPIPath("vpcs/%d/subnets", vpcID), opts)
}

func (c *Client) UpdateVPCSubnet(
	ctx context.Context,
	vpcID int,
	subnetID int,
	opts VPCSubnetUpdateOptions,
) (*VPCSubnet, error) {
	e := formatAPIPath("vpcs/%d/subnets/%d", vpcID, subnetID)
	return doPUTRequest[VPCSubnet](ctx, c, e, opts)
}

func (c *Client) DeleteVPCSubnet(ctx context.Context, vpcID int, subnetID int) error {
	e := formatAPIPath("vpcs/%d/subnets/%d", vpcID, subnetID)
	return doDELETERequest(ctx, c, e)
}
