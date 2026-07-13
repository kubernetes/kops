package linodego

import (
	"context"
)

// NodeBalancerVPCConfig objects represent a VPC config for a NodeBalancer
// s
// NOTE: NodeBalancer VPC support may not currently be available to all users.
type NodeBalancerVPCConfig struct {
	ID             int    `json:"id"`
	IPv4Range      string `json:"ipv4_range"`
	IPv6Range      string `json:"ipv6_range,omitzero"`
	NodeBalancerID int    `json:"nodebalancer_id"`
	SubnetID       int    `json:"subnet_id"`
	VPCID          int    `json:"vpc_id"`
}

// ListNodeBalancerVPCConfigs lists NodeBalancer VPC configs
func (c *Client) ListNodeBalancerVPCConfigs(ctx context.Context, nodebalancerID int, opts *ListOptions) ([]NodeBalancerVPCConfig, error) {
	return getPaginatedResults[NodeBalancerVPCConfig](ctx, c, formatAPIPath("nodebalancers/%d/vpcs", nodebalancerID), opts)
}

// GetNodeBalancerVPCConfig gets the NodeBalancer VPC config with the specified id
func (c *Client) GetNodeBalancerVPCConfig(ctx context.Context, nodebalancerID int, vpcID int) (*NodeBalancerVPCConfig, error) {
	e := formatAPIPath("nodebalancers/%d/vpcs/%d", nodebalancerID, vpcID)
	return doGETRequest[NodeBalancerVPCConfig](ctx, c, e)
}
