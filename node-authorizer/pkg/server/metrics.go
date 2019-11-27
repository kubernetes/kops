/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	authorizerErrorMetric = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "node_authorizer_error_counter",
			Help: "The number of errors encountered by the authorizer",
		},
	)
	authorizeRequestLatencyMetric = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "node_request_latency_seconds",
			Help: "A summary of the latency for incoming node authorization requests",
		},
	)
	authorizerLatencyMetric = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "node_authorizer_latency_seconds",
			Help: "A summary of the latency for the authorizer in seconds",
		},
	)
	nodeAuthorizationMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "node_authorizer_counter",
			Help: "A counter of number node authorizations broken down by denied and allowed",
		},
		[]string{"action"},
	)
	tokenLatencyMetric = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "token_latency_seconds",
			Help: "A summary of the latency experienced when creating bootstrap tokens in seconds",
		},
	)
)

func init() {
	prometheus.MustRegister(authorizeRequestLatencyMetric)
	prometheus.MustRegister(authorizerErrorMetric)
	prometheus.MustRegister(authorizerLatencyMetric)
	prometheus.MustRegister(nodeAuthorizationMetric)
	prometheus.MustRegister(tokenLatencyMetric)
}
