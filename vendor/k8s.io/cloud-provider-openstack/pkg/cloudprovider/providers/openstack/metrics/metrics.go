/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"sync"
	"time"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

type openstackMetrics struct {
	duration *metrics.HistogramVec
	total    *metrics.CounterVec
	errors   *metrics.CounterVec
}

var (
	reconcileMetrics = &openstackMetrics{
		duration: metrics.NewHistogramVec(
			&metrics.HistogramOpts{
				Name:    "cloudprovider_openstack_reconcile_duration_seconds",
				Help:    "Time taken by various parts of OpenStack cloud controller manager reconciliation loops",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
			}, []string{"operation"}),
		total: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "cloudprovider_openstack_reconcile_total",
				Help: "Total number of OpenStack cloud controller manager reconciliations",
			}, []string{"operation"}),
		errors: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "cloudprovider_openstack_reconcile_errors_total",
				Help: "Total number of OpenStack cloud controller manager reconciliation errors",
			}, []string{"operation"}),
	}
	requestMetrics = &openstackMetrics{
		duration: metrics.NewHistogramVec(
			&metrics.HistogramOpts{
				Name: "openstack_api_request_duration_seconds",
				Help: "Latency of an OpenStack API call",
			}, []string{"request"}),
		total: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "openstack_api_requests_total",
				Help: "Total number of OpenStack API calls",
			}, []string{"request"}),
		errors: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "openstack_api_request_errors_total",
				Help: "Total number of errors for an OpenStack API call",
			}, []string{"request"}),
	}
)

// MetricContext indicates the context for OpenStack metrics.
type MetricContext struct {
	start      time.Time
	attributes []string
}

// NewMetricContext creates a new MetricContext.
func NewMetricContext(resource string, request string) *MetricContext {
	return &MetricContext{
		start:      time.Now(),
		attributes: []string{resource + "_" + request},
	}
}

// ObserveReconcile records reconciliation duration,
// frequency and number of errors.
func (mc *MetricContext) ObserveReconcile(err error) error {
	reconcileMetrics.duration.WithLabelValues(mc.attributes...).Observe(
		time.Since(mc.start).Seconds())
	reconcileMetrics.total.WithLabelValues(mc.attributes...).Inc()
	if err != nil {
		reconcileMetrics.errors.WithLabelValues(mc.attributes...).Inc()
	}
	return err
}

// ObserveRequest records the request latency and counts the errors.
func (mc *MetricContext) ObserveRequest(err error) error {
	requestMetrics.duration.WithLabelValues(mc.attributes...).Observe(
		time.Since(mc.start).Seconds())
	requestMetrics.total.WithLabelValues(mc.attributes...).Inc()
	if err != nil {
		requestMetrics.errors.WithLabelValues(mc.attributes...).Inc()
	}
	return err
}

var registerMetrics sync.Once

// RegisterMetrics registers OpenStack metrics.
func RegisterMetrics() {
	registerMetrics.Do(func() {
		legacyregistry.MustRegister(
			reconcileMetrics.duration,
			reconcileMetrics.total,
			reconcileMetrics.errors,
			requestMetrics.duration,
			requestMetrics.total,
			requestMetrics.errors,
		)
	})
}
