package linodego

import (
	"context"
)

// MonitorServiceToken represents a MonitorServiceToken object
type MonitorServiceToken struct {
	Token string `json:"token"`
}

// MonitorTokenCreateOptions contains create token options.
type MonitorTokenCreateOptions struct {
	// EntityIDs are expected to be type "any" as different service_types have different variable type for their entity_ids. For example, Linode has "int" entity_ids whereas object storage has "string" as entity_ids.
	EntityIDs []any `json:"entity_ids"`
}

// CreateMonitorServiceTokenForServiceType to create token for a given serviceType
func (c *Client) CreateMonitorServiceTokenForServiceType(
	ctx context.Context,
	serviceType string,
	opts MonitorTokenCreateOptions,
) (*MonitorServiceToken, error) {
	e := formatAPIPath("monitor/services/%s/token", serviceType)
	return doPOSTRequest[MonitorServiceToken](ctx, c, e, opts)
}
