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

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var (
	APIRequestMetrics = &OpenstackMetrics{
		Duration: metrics.NewHistogramVec(
			&metrics.HistogramOpts{
				Name: "openstack_api_request_duration_seconds",
				Help: "Latency of an OpenStack API call",
			}, []string{"request"}),
		Total: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "openstack_api_requests_total",
				Help: "Total number of OpenStack API calls",
			}, []string{"request"}),
		Errors: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "openstack_api_request_errors_total",
				Help: "Total number of errors for an OpenStack API call",
			}, []string{"request"}),
	}
)

// ObserveRequest records the request latency and counts the errors.
func (mc *MetricContext) ObserveRequest(err error) error {
	return mc.Observe(APIRequestMetrics, err)
}

var registerAPIMetrics sync.Once

// RegisterMetrics registers OpenStack metrics.
func doRegisterAPIMetrics() {
	registerAPIMetrics.Do(func() {
		legacyregistry.MustRegister(
			APIRequestMetrics.Duration,
			APIRequestMetrics.Total,
			APIRequestMetrics.Errors,
		)
	})
}
