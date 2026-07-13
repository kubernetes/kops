package linodego

import (
	"context"
)

// MonitorService represents a MonitorService object
type MonitorService struct {
	Label       string               `json:"label"`
	ServiceType string               `json:"service_type"`
	Alert       *MonitorServiceAlert `json:"alert"`
}

// MonitorServiceAlert represents the alert configuration for a monitor service
type MonitorServiceAlert struct {
	PollingIntervalSeconds  []int    `json:"polling_interval_seconds"`
	EvaluationPeriodSeconds []int    `json:"evaluation_period_seconds"`
	Scope                   []string `json:"scope"`
}

// ListMonitorServices lists all the registered ACLP MonitorServices
func (c *Client) ListMonitorServices(ctx context.Context, opts *ListOptions) ([]MonitorService, error) {
	return getPaginatedResults[MonitorService](ctx, c, "monitor/services", opts)
}

// GetMonitorServiceByType gets a monitor service by a given service_type
func (c *Client) GetMonitorServiceByType(ctx context.Context, serviceType string) (*MonitorService, error) {
	e := formatAPIPath("monitor/services/%s", serviceType)

	return doGETRequest[MonitorService](ctx, c, e)
}
