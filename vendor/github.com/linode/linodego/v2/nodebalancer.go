package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// NodeBalancer represents a NodeBalancer object
type NodeBalancer struct {
	// This NodeBalancer's unique ID.
	ID int `json:"id"`
	// This NodeBalancer's label. These must be unique on your Account.
	Label *string `json:"label"`
	// The Region where this NodeBalancer is located. NodeBalancers only support backends in the same Region.
	Region string `json:"region"`
	// This NodeBalancer's hostname, ending with .nodebalancer.linode.com
	Hostname *string `json:"hostname"`
	// This NodeBalancer's public IPv4 address.
	IPv4 *string `json:"ipv4"`
	// This NodeBalancer's public IPv6 address.
	IPv6 *string `json:"ipv6"`
	// Throttle connections per second (0-20). Set to 0 (zero) to disable throttling.
	ClientConnThrottle int `json:"client_conn_throttle"`

	// ClientUDPSessThrottle throttles UDP sessions per second. Set to 0 (zero) to disable throttling.
	// NOTE: ClientUDPSessThrottle may not currently be available to all users.
	ClientUDPSessThrottle int `json:"client_udp_sess_throttle"`

	// Information about the amount of transfer this NodeBalancer has had so far this month.
	Transfer NodeBalancerTransfer `json:"transfer"`
	// This NodeBalancer's plan Type
	Type NodeBalancerPlanType `json:"type"`

	// An array of tags applied to this object. Tags are for organizational purposes only.
	Tags []string `json:"tags"`

	// This NodeBalancer's related LKE cluster, if any. The value is null if this NodeBalancer is not related to an LKE cluster.
	LKECluster *NodeBalancerLKECluster `json:"lke_cluster"`

	// An array of locks applied to this NodeBalancer for deletion protection.
	// Locks prevent the NodeBalancer or its subresources from being deleted.
	// NOTE: Locks can only be used with v4beta.
	Locks []LockType `json:"locks"`

	Created *time.Time `json:"-"`
	Updated *time.Time `json:"-"`
}

// NodeBalancerTransfer contains information about the amount of transfer a NodeBalancer has had in the current month
type NodeBalancerTransfer struct {
	// The total transfer, in MB, used by this NodeBalancer this month.
	Total *float64 `json:"total"`
	// The total inbound transfer, in MB, used for this NodeBalancer this month.
	Out *float64 `json:"out"`
	// The total outbound transfer, in MB, used for this NodeBalancer this month.
	In *float64 `json:"in"`
}

type NodeBalancerVPCOptions struct {
	IPv4Range           string `json:"ipv4_range,omitzero"`
	IPv6Range           string `json:"ipv6_range,omitzero"`
	SubnetID            int    `json:"subnet_id"`
	IPv4RangeAutoAssign bool   `json:"ipv4_range_auto_assign,omitzero"`
}

// NodeBalancerCreateOptions are the options permitted for CreateNodeBalancer
type NodeBalancerCreateOptions struct {
	Label              *string `json:"label,omitzero"`
	Region             string  `json:"region,omitzero"`
	ClientConnThrottle *int    `json:"client_conn_throttle,omitzero"`

	// NOTE: ClientUDPSessThrottle may not currently be available to all users.
	ClientUDPSessThrottle *int `json:"client_udp_sess_throttle,omitzero"`

	Configs    []NodeBalancerConfigCreateOptions `json:"configs,omitzero"`
	Tags       []string                          `json:"tags"`
	FirewallID int                               `json:"firewall_id,omitzero"`
	Type       NodeBalancerPlanType              `json:"type,omitzero"`
	VPCs       []NodeBalancerVPCOptions          `json:"vpcs,omitzero"`
	IPv4       *string                           `json:"ipv4,omitzero"`
}

// NodeBalancerUpdateOptions are the options permitted for UpdateNodeBalancer
type NodeBalancerUpdateOptions struct {
	Label              *string `json:"label,omitzero"`
	ClientConnThrottle *int    `json:"client_conn_throttle,omitzero"`

	// NOTE: ClientUDPSessThrottle may not currently be available to all users.
	ClientUDPSessThrottle *int `json:"client_udp_sess_throttle,omitzero"`

	Tags []string `json:"tags,omitzero"`
}

type NodeBalancerLKECluster struct {
	// The ID of the related LKE cluster.
	ID int `json:"id"`
	// The label of the related LKE cluster.
	Label string `json:"label"`
	// The type for LKE clusters.
	Type string `json:"type"`
	// The URL where you can access the related LKE cluster.
	URL string `json:"url"`
}

// NodeBalancerPlanType constants start with NBType and include Linode API NodeBalancer's plan types
type NodeBalancerPlanType string

// NodeBalancerPlanType constants reflect the plan type used by a NodeBalancer Config
const (
	NBTypePremium     NodeBalancerPlanType = "premium"
	NBTypePremium40GB NodeBalancerPlanType = "premium_40gb"
	NBTypeCommon      NodeBalancerPlanType = "common"
)

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *NodeBalancer) UnmarshalJSON(b []byte) error {
	type Mask NodeBalancer

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// GetCreateOptions converts a NodeBalancer to NodeBalancerCreateOptions for use in CreateNodeBalancer
func (i NodeBalancer) GetCreateOptions() NodeBalancerCreateOptions {
	return NodeBalancerCreateOptions{
		Label:                 i.Label,
		Region:                i.Region,
		ClientConnThrottle:    &i.ClientConnThrottle,
		ClientUDPSessThrottle: &i.ClientUDPSessThrottle,
		Type:                  i.Type,
		Tags:                  i.Tags,
	}
}

// GetUpdateOptions converts a NodeBalancer to NodeBalancerUpdateOptions for use in UpdateNodeBalancer
func (i NodeBalancer) GetUpdateOptions() NodeBalancerUpdateOptions {
	return NodeBalancerUpdateOptions{
		Label:                 i.Label,
		ClientConnThrottle:    &i.ClientConnThrottle,
		ClientUDPSessThrottle: &i.ClientUDPSessThrottle,
		Tags:                  i.Tags,
	}
}

// ListNodeBalancers lists NodeBalancers
func (c *Client) ListNodeBalancers(ctx context.Context, opts *ListOptions) ([]NodeBalancer, error) {
	return getPaginatedResults[NodeBalancer](ctx, c, "nodebalancers", opts)
}

// GetNodeBalancer gets the NodeBalancer with the provided ID
func (c *Client) GetNodeBalancer(ctx context.Context, nodebalancerID int) (*NodeBalancer, error) {
	e := formatAPIPath("nodebalancers/%d", nodebalancerID)
	return doGETRequest[NodeBalancer](ctx, c, e)
}

// CreateNodeBalancer creates a NodeBalancer
func (c *Client) CreateNodeBalancer(ctx context.Context, opts NodeBalancerCreateOptions) (*NodeBalancer, error) {
	return doPOSTRequest[NodeBalancer](ctx, c, "nodebalancers", opts)
}

// UpdateNodeBalancer updates the NodeBalancer with the specified id
func (c *Client) UpdateNodeBalancer(ctx context.Context, nodebalancerID int, opts NodeBalancerUpdateOptions) (*NodeBalancer, error) {
	e := formatAPIPath("nodebalancers/%d", nodebalancerID)
	return doPUTRequest[NodeBalancer](ctx, c, e, opts)
}

// DeleteNodeBalancer deletes the NodeBalancer with the specified id
func (c *Client) DeleteNodeBalancer(ctx context.Context, nodebalancerID int) error {
	e := formatAPIPath("nodebalancers/%d", nodebalancerID)
	return doDELETERequest(ctx, c, e)
}
