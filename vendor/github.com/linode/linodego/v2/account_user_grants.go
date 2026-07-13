package linodego

import (
	"context"
)

type GrantPermissionLevel string

const (
	AccessLevelReadOnly  GrantPermissionLevel = "read_only"
	AccessLevelReadWrite GrantPermissionLevel = "read_write"
)

type GlobalUserGrants struct {
	AccountAccess        *GrantPermissionLevel `json:"account_access"`
	AddDatabases         bool                  `json:"add_databases"`
	AddDomains           bool                  `json:"add_domains"`
	AddFirewalls         bool                  `json:"add_firewalls"`
	AddImages            bool                  `json:"add_images"`
	AddLinodes           bool                  `json:"add_linodes"`
	AddLongview          bool                  `json:"add_longview"`
	AddNodeBalancers     bool                  `json:"add_nodebalancers"`
	AddStackScripts      bool                  `json:"add_stackscripts"`
	AddVolumes           bool                  `json:"add_volumes"`
	AddVPCs              bool                  `json:"add_vpcs"`
	CancelAccount        bool                  `json:"cancel_account"`
	ChildAccountAccess   bool                  `json:"child_account_access"`
	LongviewSubscription bool                  `json:"longview_subscription"`
}

type EntityUserGrant struct {
	ID          int                   `json:"id"`
	Permissions *GrantPermissionLevel `json:"permissions"`
}

type GrantedEntity struct {
	ID          int                  `json:"id"`
	Label       string               `json:"label"`
	Permissions GrantPermissionLevel `json:"permissions"`
}

type UserGrants struct {
	Database       []GrantedEntity `json:"database"`
	Domain         []GrantedEntity `json:"domain"`
	Firewall       []GrantedEntity `json:"firewall"`
	Image          []GrantedEntity `json:"image"`
	Linode         []GrantedEntity `json:"linode"`
	Longview       []GrantedEntity `json:"longview"`
	NodeBalancer   []GrantedEntity `json:"nodebalancer"`
	PlacementGroup []GrantedEntity `json:"placement_group"`
	StackScript    []GrantedEntity `json:"stackscript"`
	Volume         []GrantedEntity `json:"volume"`
	VPC            []GrantedEntity `json:"vpc"`

	Global GlobalUserGrants `json:"global"`
}

type UserGrantsUpdateOptions struct {
	Database       []GrantedEntity   `json:"database,omitzero"`
	Domain         []EntityUserGrant `json:"domain,omitzero"`
	Firewall       []EntityUserGrant `json:"firewall,omitzero"`
	Image          []EntityUserGrant `json:"image,omitzero"`
	Linode         []EntityUserGrant `json:"linode,omitzero"`
	Longview       []EntityUserGrant `json:"longview,omitzero"`
	NodeBalancer   []EntityUserGrant `json:"nodebalancer,omitzero"`
	PlacementGroup []EntityUserGrant `json:"placement_group,omitzero"`
	StackScript    []EntityUserGrant `json:"stackscript,omitzero"`
	Volume         []EntityUserGrant `json:"volume,omitzero"`
	VPC            []EntityUserGrant `json:"vpc,omitzero"`

	Global GlobalUserGrants `json:"global"`
}

func (c *Client) GetUserGrants(ctx context.Context, username string) (*UserGrants, error) {
	e := formatAPIPath("account/users/%s/grants", username)
	return doGETRequest[UserGrants](ctx, c, e)
}

func (c *Client) UpdateUserGrants(ctx context.Context, username string, opts UserGrantsUpdateOptions) (*UserGrants, error) {
	e := formatAPIPath("account/users/%s/grants", username)
	return doPUTRequest[UserGrants](ctx, c, e, opts)
}
