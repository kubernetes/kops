package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

type VPC struct {
	ID          int    `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Region      string `json:"region"`

	// NOTE: IPv4 VPCs may not currently be available to all users.
	IPv4 []VPCIPv4Range `json:"ipv4"`
	// NOTE: IPv6 VPCs may not currently be available to all users.
	IPv6 []VPCIPv6Range `json:"ipv6"`

	Subnets []VPCSubnet `json:"subnets"`
	Created *time.Time  `json:"-"`
	Updated *time.Time  `json:"-"`
}

// VPCIPv4Range represents a single IPv4 range assigned to a VPC.
// NOTE: IPv4 VPCs may not currently be available to all users.
type VPCIPv4Range struct {
	Range string `json:"range"`
}

// VPCIPv6Range represents a single IPv6 range assigned to a VPC.
// NOTE: IPv6 VPCs may not currently be available to all users.
type VPCIPv6Range struct {
	Range string `json:"range"`
}

// VPCDefaultRanges represents the default settings for the internal and forbidden IPv4 address ranges in VPCs
// NOTE: VPCDefaultRanges may not currently be available to all users.
type VPCDefaultRanges struct {
	DefaultIPV4Ranges   []string `json:"default_ipv4_ranges"`
	ForbiddenIPV4Ranges []string `json:"forbidden_ipv4_ranges"`
}

type VPCCreateOptions struct {
	Label       string `json:"label"`
	Description string `json:"description,omitzero"`
	Region      string `json:"region"`

	// NOTE: IPv4 VPCs may not currently be available to all users.
	IPv4 []VPCCreateOptionsIPv4 `json:"ipv4,omitzero"`
	// NOTE: IPv6 VPCs may not currently be available to all users.
	IPv6 []VPCCreateOptionsIPv6 `json:"ipv6,omitzero"`

	Subnets []VPCSubnetCreateOptions `json:"subnets,omitzero"`
}

// VPCCreateOptionsIPv4 represents a single IPv4 range assigned to a VPC
// which is specified during a VPC's creation.
// NOTE: IPv4 VPCs may not currently be available to all users.
type VPCCreateOptionsIPv4 struct {
	Range *string `json:"range,omitzero"`
}

// VPCCreateOptionsIPv6 represents a single IPv6 range assigned to a VPC
// which is specified during a VPC's creation.
// NOTE: IPv6 VPCs may not currently be available to all users.
type VPCCreateOptionsIPv6 struct {
	Range           *string `json:"range,omitzero"`
	AllocationClass *string `json:"allocation_class,omitzero"`
}

type VPCUpdateOptions struct {
	Label       string                 `json:"label,omitzero"`
	Description string                 `json:"description,omitzero"`
	IPv4        []VPCUpdateOptionsIPv4 `json:"ipv4,omitzero"`
}

// VPCUpdateOptionsIPv4 represents a single IPv4 range assigned to a VPC
// which is specified during a VPC's update.
// NOTE: IPv4 VPCs may not currently be available to all users.
type VPCUpdateOptionsIPv4 struct {
	Range *string `json:"range,omitzero"`
}

func (v VPC) GetCreateOptions() VPCCreateOptions {
	subnetCreations := make([]VPCSubnetCreateOptions, len(v.Subnets))
	for i, s := range v.Subnets {
		subnetCreations[i] = s.GetCreateOptions()
	}

	return VPCCreateOptions{
		Label:       v.Label,
		Description: v.Description,
		Region:      v.Region,
		Subnets:     subnetCreations,
		IPv4: mapSlice(v.IPv4, func(i VPCIPv4Range) VPCCreateOptionsIPv4 {
			return VPCCreateOptionsIPv4{
				Range: copyValue(&i.Range),
			}
		}),
		IPv6: mapSlice(v.IPv6, func(i VPCIPv6Range) VPCCreateOptionsIPv6 {
			return VPCCreateOptionsIPv6{
				Range: copyValue(&i.Range),
			}
		}),
	}
}

func (v VPC) GetUpdateOptions() VPCUpdateOptions {
	return VPCUpdateOptions{
		Label:       v.Label,
		Description: v.Description,
		IPv4: mapSlice(v.IPv4, func(i VPCIPv4Range) VPCUpdateOptionsIPv4 {
			return VPCUpdateOptionsIPv4{
				Range: copyValue(&i.Range),
			}
		}),
	}
}

func (v *VPC) UnmarshalJSON(b []byte) error {
	type Mask VPC

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

func (c *Client) CreateVPC(
	ctx context.Context,
	opts VPCCreateOptions,
) (*VPC, error) {
	return doPOSTRequest[VPC](ctx, c, "vpcs", opts)
}

func (c *Client) GetVPC(ctx context.Context, vpcID int) (*VPC, error) {
	e := formatAPIPath("/vpcs/%d", vpcID)
	return doGETRequest[VPC](ctx, c, e)
}

func (c *Client) ListVPCs(ctx context.Context, opts *ListOptions) ([]VPC, error) {
	return getPaginatedResults[VPC](ctx, c, "vpcs", opts)
}

func (c *Client) UpdateVPC(
	ctx context.Context,
	vpcID int,
	opts VPCUpdateOptions,
) (*VPC, error) {
	e := formatAPIPath("vpcs/%d", vpcID)
	return doPUTRequest[VPC](ctx, c, e, opts)
}

func (c *Client) DeleteVPC(ctx context.Context, vpcID int) error {
	e := formatAPIPath("vpcs/%d", vpcID)
	return doDELETERequest(ctx, c, e)
}

// GetVPCDefaultRanges may not currently be available to all users.
func (c *Client) GetVPCDefaultRanges(ctx context.Context) (*VPCDefaultRanges, error) {
	return doGETRequest[VPCDefaultRanges](ctx, c, "/vpcs/default-ranges")
}
