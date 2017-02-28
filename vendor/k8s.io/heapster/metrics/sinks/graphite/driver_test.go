// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graphite

import (
	"testing"

	"k8s.io/heapster/metrics/core"

	"github.com/stretchr/testify/assert"
)

var metricsTestCases = []struct {
	metric graphiteMetric
	path   string
	value  string
}{
	{
		graphiteMetric{},
		"",
		"0",
	},
	{
		graphiteMetric{
			name:  "metric/avg",
			value: core.MetricValue{IntValue: 100, ValueType: core.ValueInt64},
			labels: map[string]string{
				"hostname":       "example",
				"type":           "pod_container",
				"namespace_name": "namespace",
				"pod_name":       "pod-name-12345",
				"container_name": "container",
			},
		},
		"nodes.example.pods.namespace.pod-name-12345.containers.container.metric.avg",
		"100",
	},
	{
		graphiteMetric{
			name:  "metric/avg",
			value: core.MetricValue{IntValue: 100, ValueType: core.ValueInt64},
			labels: map[string]string{
				"hostname":       "example",
				"type":           "sys_container",
				"container_name": "container",
			},
		},
		"nodes.example.sys-containers.container.metric.avg",
		"100",
	},
	{
		graphiteMetric{
			name:  "metric/avg",
			value: core.MetricValue{IntValue: 100, ValueType: core.ValueInt64},
			labels: map[string]string{
				"hostname":       "example",
				"type":           "pod",
				"namespace_name": "namespace",
				"pod_name":       "pod-name-12345",
			},
		},
		"nodes.example.pods.namespace.pod-name-12345.metric.avg",
		"100",
	},
	{
		graphiteMetric{
			name:  "metric/avg",
			value: core.MetricValue{IntValue: 100, ValueType: core.ValueInt64},
			labels: map[string]string{
				"type":           "ns",
				"namespace_name": "namespace",
			},
		},
		"namespaces.namespace.metric.avg",
		"100",
	},
	{
		graphiteMetric{
			name:  "metric/avg",
			value: core.MetricValue{IntValue: 100, ValueType: core.ValueInt64},
			labels: map[string]string{
				"hostname": "example",
				"type":     "node",
			},
		},
		"nodes.example.metric.avg",
		"100",
	},
	{
		graphiteMetric{
			name:  "metric/avg",
			value: core.MetricValue{IntValue: 100, ValueType: core.ValueInt64},
			labels: map[string]string{
				"hostname": "example",
				"type":     "cluster",
			},
		},
		"cluster.metric.avg",
		"100",
	},
}

func TestGraphitePathMetrics(t *testing.T) {
	for _, c := range metricsTestCases {
		m := c.metric.Metric()
		assert.Equal(t, c.path, m.Name)
		assert.Equal(t, c.value, m.Value)
	}
}
