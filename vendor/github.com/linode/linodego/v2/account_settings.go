package linodego

import (
	"context"
)

type InterfacesForNewLinodes string

const (
	LegacyConfigOnly                    InterfacesForNewLinodes = "legacy_config_only"
	LegacyConfigDefaultButLinodeAllowed InterfacesForNewLinodes = "legacy_config_default_but_linode_allowed"
	LinodeDefaultButLegacyConfigAllowed InterfacesForNewLinodes = "linode_default_but_legacy_config_allowed"
	LinodeOnly                          InterfacesForNewLinodes = "linode_only"
)

// AccountSettings are the account wide flags or plans that effect new resources
type AccountSettings struct {
	// The default backups enrollment status for all new Linodes for all users on the account.  When enabled, backups are mandatory per instance.
	BackupsEnabled bool `json:"backups_enabled"`

	// Whether or not Linode Managed service is enabled for the account.
	Managed bool `json:"managed"`

	// Whether or not the Network Helper is enabled for all new Linode Instance Configs on the account.
	NetworkHelper bool `json:"network_helper"`

	// A plan name like "longview-3"..."longview-100", or a nil value for to cancel any existing subscription plan.
	LongviewSubscription *string `json:"longview_subscription"`

	// A string like "disabled", "suspended", or "active" describing the status of this account’s Object Storage service enrollment.
	ObjectStorage *string `json:"object_storage"`

	// A new configuration flag defines whether new Linodes can use Linode and/or legacy config interfaces.
	InterfacesForNewLinodes InterfacesForNewLinodes `json:"interfaces_for_new_linodes"`

	// The slug of the maintenance policy associated with the account.
	MaintenancePolicy string `json:"maintenance_policy"`
}

// AccountSettingsUpdateOptions are the updateable account wide flags or plans that effect new resources.
type AccountSettingsUpdateOptions struct {
	// The default backups enrollment status for all new Linodes for all users on the account.  When enabled, backups are mandatory per instance.
	BackupsEnabled *bool `json:"backups_enabled,omitzero"`

	// The default network helper setting for all new Linodes and Linode Configs for all users on the account.
	NetworkHelper *bool `json:"network_helper,omitzero"`

	// NOTE: Interfaces for new linode setting may not currently be available to all users.
	// A new configuration flag defines whether new Linodes can use Linode and/or legacy config interfaces.
	InterfacesForNewLinodes *InterfacesForNewLinodes `json:"interfaces_for_new_linodes"`

	// The slug of the maintenance policy to set the account to.
	MaintenancePolicy *string `json:"maintenance_policy,omitzero"`
}

// GetAccountSettings gets the account wide flags or plans that effect new resources
func (c *Client) GetAccountSettings(ctx context.Context) (*AccountSettings, error) {
	return doGETRequest[AccountSettings](ctx, c, "account/settings")
}

// UpdateAccountSettings updates the settings associated with the account
func (c *Client) UpdateAccountSettings(ctx context.Context, opts AccountSettingsUpdateOptions) (*AccountSettings, error) {
	return doPUTRequest[AccountSettings](ctx, c, "account/settings", opts)
}
