// Functions in the file are making calls to the monitor-api instead of linode domain.
// Please initialize a MonitorClient for using the endpoints below.

package linodego

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type MetricFilterOperator string

const (
	MetricFilterOperatorEq         MetricFilterOperator = "eq"
	MetricFilterOperatorNeq        MetricFilterOperator = "neq"
	MetricFilterOperatorStartsWith MetricFilterOperator = "startswith"
	MetricFilterOperatorEndsWith   MetricFilterOperator = "endswith"
	MetricFilterOperatorIn         MetricFilterOperator = "in"
)

type MetricTimeUnit string

const (
	MetricTimeUnitSec  MetricTimeUnit = "sec"
	MetricTimeUnitMin  MetricTimeUnit = "min"
	MetricTimeUnitHr   MetricTimeUnit = "hr"
	MetricTimeUnitDays MetricTimeUnit = "days"
)

// EntityMetrics is the response body of the metrics for the entities requested
type EntityMetrics struct {
	Data      EntityMetricsData  `json:"data"`
	IsPartial bool               `json:"is_partial"`
	Stats     EntityMetricsStats `json:"stats"`
	Status    string             `json:"status"`
}

// EntityMetricsData describes the result and type for the entity metrics
type EntityMetricsData struct {
	Result     []EntityMetricsDataResult `json:"result"`
	ResultType string                    `json:"result_type"`
}

// EntityMetricsDataResult contains the information of a metric and values
type EntityMetricsDataResult struct {
	Metric map[string]any `json:"metric"`
	Values [][]any        `json:"values"`
}

// EntityMetricsStats shows statistics info of the metrics fetched
type EntityMetricsStats struct {
	ExecutionTimeMsec int    `json:"executionTimeMsec"`
	SeriesFetched     string `json:"seriesFetched"`
}

// EntityMetricsFetchOptions are the options used to fetch metrics with the entity ids provided
type EntityMetricsFetchOptions struct {
	// EntityIDs are expected to be type "any" as different service_types have different variable type for their entity_ids. For example, Linode has "int" entity_ids whereas object storage has "string" as entity_ids.
	EntityIDs []any `json:"entity_ids"`

	Filters              []MetricFilter              `json:"filters,omitzero"`
	Metrics              []EntityMetric              `json:"metrics"`
	TimeGranularity      []MetricTimeGranularity     `json:"time_granularity,omitzero"`
	RelativeTimeDuration *MetricRelativeTimeDuration `json:"relative_time_duration,omitzero"`
	AbsoluteTimeDuration *MetricAbsoluteTimeDuration `json:"absolute_time_duration,omitzero"`
}

// MetricFilter describes individual objects that define dimension filters for the query.
type MetricFilter struct {
	DimensionLabel string               `json:"dimension_label"`
	Operator       MetricFilterOperator `json:"operator"`
	Value          string               `json:"value"`
}

// EntityMetric specifies a metric name and its corresponding aggregation function.
type EntityMetric struct {
	Name              string            `json:"name"`
	AggregateFunction AggregateFunction `json:"aggregate_function"`
}

// MetricTimeGranularity allows for an optional time granularity setting for metric data.
type MetricTimeGranularity struct {
	Unit  MetricTimeUnit `json:"unit"`
	Value int            `json:"value"`
}

// MetricRelativeTimeDuration specifies a relative time duration for data queries
type MetricRelativeTimeDuration struct {
	Unit  MetricTimeUnit `json:"unit"`
	Value int            `json:"value"`
}

// MetricAbsoluteTimeDuration specifies an absolute time range for data queries
type MetricAbsoluteTimeDuration struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// FetchEntityMetrics returns metrics information for the individual entities within a specific service type
func (mc *MonitorClient) FetchEntityMetrics(ctx context.Context, serviceType string, opts *EntityMetricsFetchOptions) (*EntityMetrics, error) {
	endpoint := formatAPIPath("monitor/services/%s/metrics", serviceType)

	var result EntityMetrics

	params := requestParams{
		Response: &result,
	}

	if opts != nil {
		body, err := json.Marshal(opts)
		if err != nil {
			return nil, err
		}

		params.Body = bytes.NewReader(body)
	}

	err := mc.doRequest(ctx, http.MethodPost, endpoint, params)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
