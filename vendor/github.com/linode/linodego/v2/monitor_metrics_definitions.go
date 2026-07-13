package linodego

import (
	"context"
)

// MonitorMetricsDefinition represents an ACLP MetricsDefinition object
type MonitorMetricsDefinition struct {
	AvailableAggregateFunctions []AggregateFunction `json:"available_aggregate_functions"`
	Dimensions                  []MonitorDimension  `json:"dimensions"`
	IsAlertable                 bool                `json:"is_alertable"`
	Label                       string              `json:"label"`
	Metric                      string              `json:"metric"`
	MetricType                  MetricType          `json:"metric_type"`
	ScrapeInterval              string              `json:"scrape_interval"`
	Unit                        MetricUnit          `json:"unit"`
}

// MetricType is an enum object for MetricType
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeSummary   MetricType = "summary"
)

// MetricUnit is an enum object for Unit
type MetricUnit string

const (
	MetricUnitCount              MetricUnit = "count"
	MetricUnitPercent            MetricUnit = "percent"
	MetricUnitByte               MetricUnit = "byte"
	MetricUnitSecond             MetricUnit = "second"
	MetricUnitBitsPerSecond      MetricUnit = "bits_per_second"
	MetricUnitMillisecond        MetricUnit = "millisecond"
	MetricUnitKB                 MetricUnit = "KB"
	MetricUnitMB                 MetricUnit = "MB"
	MetricUnitGB                 MetricUnit = "GB"
	MetricUnitRate               MetricUnit = "rate"
	MetricUnitBytesPerSecond     MetricUnit = "bytes_per_second"
	MetricUnitPercentile         MetricUnit = "percentile"
	MetricUnitRatio              MetricUnit = "ratio"
	MetricUnitOpsPerSecond       MetricUnit = "ops_per_second"
	MetricUnitIops               MetricUnit = "iops"
	MetricUnitKiloBytesPerSecond MetricUnit = "kilo_bytes_per_second"
	MetricUnitSessionsPerSecond  MetricUnit = "sessions_per_second"
	MetricUnitPacketsPerSecond   MetricUnit = "packets_per_second"
	MetricUnitKiloBitsPerSecond  MetricUnit = "kilo_bits_per_second"
)

// MonitorDimension represents an ACLP MonitorDimension object
type MonitorDimension struct {
	DimensionLabel string   `json:"dimension_label"`
	Label          string   `json:"label"`
	Values         []string `json:"values"`
}

// ListMonitorMetricsDefinitionByServiceType lists metric definitions
func (c *Client) ListMonitorMetricsDefinitionByServiceType(ctx context.Context, serviceType string, opts *ListOptions) ([]MonitorMetricsDefinition, error) {
	e := formatAPIPath("monitor/services/%s/metric-definitions", serviceType)
	return getPaginatedResults[MonitorMetricsDefinition](ctx, c, e, opts)
}
