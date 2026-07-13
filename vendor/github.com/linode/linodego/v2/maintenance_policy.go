package linodego

import (
	"context"
)

type MaintenancePolicy struct {
	Slug                  string `json:"slug"`
	Label                 string `json:"label"`
	Description           string `json:"description"`
	Type                  string `json:"type"`
	NotificationPeriodSec int    `json:"notification_period_sec"`
	IsDefault             bool   `json:"is_default"`
}

// ListMaintenancePolicies lists all available maintenance policies that can be applied to Linodes.
func (c *Client) ListMaintenancePolicies(ctx context.Context, opts *ListOptions) ([]MaintenancePolicy, error) {
	return getPaginatedResults[MaintenancePolicy](ctx, c, "maintenance/policies", opts)
}
