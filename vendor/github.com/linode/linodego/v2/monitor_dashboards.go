package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// MonitorDashboard represents an ACLP Dashboard object
type MonitorDashboard struct {
	ID          int               `json:"id"`
	Type        DashboardType     `json:"type"`
	ServiceType ServiceType       `json:"service_type"`
	Label       string            `json:"label"`
	GroupBy     []string          `json:"group_by"`
	Created     *time.Time        `json:"-"`
	Updated     *time.Time        `json:"-"`
	Widgets     []DashboardWidget `json:"widgets"`
}

// ServiceType is an enum object for serviceType
type ServiceType string

const (
	ServiceTypeLinode          ServiceType = "linode"
	ServiceTypeLKE             ServiceType = "lke"
	ServiceTypeDBaaS           ServiceType = "dbaas"
	ServiceTypeACLB            ServiceType = "aclb"
	ServiceTypeNodeBalancer    ServiceType = "nodebalancer"
	ServiceTypeObjectStorage   ServiceType = "object_storage"
	ServiceTypeVPC             ServiceType = "vpc"
	ServiceTypeFirewallService ServiceType = "firewall"
	ServiceTypeNetLoadBalancer ServiceType = "netloadbalancer"
)

// DashboardType is an enum object for DashboardType
type DashboardType string

const (
	DashboardTypeStandard DashboardType = "standard"
	DashboardTypeCustom   DashboardType = "custom"
)

// DashboardWidget represents an ACLP DashboardWidget object
type DashboardWidget struct {
	Metric            string            `json:"metric"`
	Unit              string            `json:"unit"`
	Label             string            `json:"label"`
	Color             string            `json:"color"`
	Size              int               `json:"size"`
	ChartType         ChartType         `json:"chart_type"`
	YLabel            string            `json:"y_label"`
	AggregateFunction AggregateFunction `json:"aggregate_function"`
	GroupBy           []string          `json:"group_by"`
	Filters           []DashboardFilter `json:"filters"`
}

// DashboardFilter represents a filter for dashboard widgets
type DashboardFilter struct {
	DimensionLabel string `json:"dimension_label"`
	Operator       string `json:"operator"`
	Value          string `json:"value"`
}

// AggregateFunction is an enum object for AggregateFunction
type AggregateFunction string

const (
	AggregateFunctionMin      AggregateFunction = "min"
	AggregateFunctionMax      AggregateFunction = "max"
	AggregateFunctionAvg      AggregateFunction = "avg"
	AggregateFunctionSum      AggregateFunction = "sum"
	AggregateFunctionRate     AggregateFunction = "rate"
	AggregateFunctionIncrease AggregateFunction = "increase"
	AggregateFunctionCount    AggregateFunction = "count"
	AggregateFunctionLast     AggregateFunction = "last"
)

// ChartType is an enum object for Chart type
type ChartType string

const (
	ChartTypeLine ChartType = "line"
	ChartTypeArea ChartType = "area"
)

// ListMonitorDashboards lists all the ACLP Monitor Dashboards
func (c *Client) ListMonitorDashboards(ctx context.Context, opts *ListOptions) ([]MonitorDashboard, error) {
	return getPaginatedResults[MonitorDashboard](ctx, c, "monitor/dashboards", opts)
}

// GetMonitorDashboard gets an ACLP Monitor Dashboard for a given dashboardID
func (c *Client) GetMonitorDashboard(ctx context.Context, dashboardID int) (*MonitorDashboard, error) {
	e := formatAPIPath("monitor/dashboards/%d", dashboardID)
	return doGETRequest[MonitorDashboard](ctx, c, e)
}

// ListMonitorDashboardsByServiceType lists ACLP Monitor Dashboards for a given serviceType
func (c *Client) ListMonitorDashboardsByServiceType(ctx context.Context, serviceType string, opts *ListOptions) ([]MonitorDashboard, error) {
	e := formatAPIPath("monitor/services/%s/dashboards", serviceType)
	return getPaginatedResults[MonitorDashboard](ctx, c, e, opts)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *MonitorDashboard) UnmarshalJSON(b []byte) error {
	type Mask MonitorDashboard

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
