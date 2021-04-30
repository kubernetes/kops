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
	occmReconcileMetrics = &OpenstackMetrics{
		Duration: metrics.NewHistogramVec(
			&metrics.HistogramOpts{
				Name:    "cloudprovider_openstack_reconcile_duration_seconds",
				Help:    "Time taken by various parts of OpenStack cloud controller manager reconciliation loops",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
			}, []string{"operation"}),
		Total: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "cloudprovider_openstack_reconcile_total",
				Help: "Total number of OpenStack cloud controller manager reconciliations",
			}, []string{"operation"}),
		Errors: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Name: "cloudprovider_openstack_reconcile_errors_total",
				Help: "Total number of OpenStack cloud controller manager reconciliation errors",
			}, []string{"operation"}),
	}
)

// ObserveReconcile records the request reconciliation duration
func (mc *MetricContext) ObserveReconcile(err error) error {
	return mc.Observe(occmReconcileMetrics, err)
}

var registerOccmMetrics sync.Once

// RegisterMetrics registers OpenStack metrics.
func doRegisterOccmMetrics() {
	registerOccmMetrics.Do(func() {
		legacyregistry.MustRegister(
			occmReconcileMetrics.Duration,
			occmReconcileMetrics.Total,
			occmReconcileMetrics.Errors,
		)
	})
}
