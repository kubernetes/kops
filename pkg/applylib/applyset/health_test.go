/*
Copyright 2026 The Kubernetes Authors.

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

package applyset

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestIsHealthy(t *testing.T) {
	grid := []struct {
		name   string
		object map[string]interface{}
		want   bool
	}{
		{
			name: "no ready signal for kind",
			object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]interface{}{"name": "cm"},
			},
			want: true,
		},
		{
			name: "scheduled for deletion",
			object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":              "deploy",
					"deletionTimestamp": "2026-01-01T00:00:00Z",
				},
			},
			want: false,
		},
		{
			name: "no status conditions",
			object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]interface{}{"name": "deploy"},
			},
			want: true,
		},
		{
			name: "null status conditions",
			object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]interface{}{"name": "deploy"},
				"status":     map[string]interface{}{"conditions": nil},
			},
			want: true,
		},
		{
			name: "available deployment that is still progressing",
			object: deploymentWithConditions(
				condition("Available", "True"),
				condition("Progressing", "True"),
			),
			want: true,
		},
		{
			name: "unavailable deployment",
			object: deploymentWithConditions(
				condition("Available", "False"),
				condition("Progressing", "True"),
			),
			want: false,
		},
		{
			name: "deployment reporting a replica failure",
			object: deploymentWithConditions(
				condition("Available", "True"),
				// ReplicaFailure is abnormal-true, so it is not a readiness signal.
				condition("ReplicaFailure", "True"),
			),
			want: true,
		},
		{
			name: "available deployment that gave up on its rollout",
			object: deploymentWithConditions(
				condition("Available", "True"),
				condition("Progressing", "False"),
			),
			want: false,
		},
		{
			name: "progressing is only a readiness signal for Deployments",
			object: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "Widget",
				"metadata":   map[string]interface{}{"name": "widget"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("Ready", "True"),
						condition("Progressing", "False"),
					},
				},
			},
			want: true,
		},
		{
			name: "programmed gateway",
			object: gatewayWithConditions("Gateway",
				condition("Accepted", "True"),
				condition("Programmed", "True"),
			),
			want: true,
		},
		{
			name: "gateway that is accepted but not programmed",
			object: gatewayWithConditions("Gateway",
				condition("Accepted", "True"),
				condition("Programmed", "False"),
			),
			want: false,
		},
		{
			name: "accepted gateway class",
			object: gatewayWithConditions("GatewayClass",
				condition("Accepted", "True"),
			),
			want: true,
		},
		{
			name: "rejected gateway class",
			object: gatewayWithConditions("GatewayClass",
				condition("Accepted", "False"),
			),
			want: false,
		},
		{
			name: "healthy node with negative pressure conditions",
			object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Node",
				"metadata":   map[string]interface{}{"name": "node"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("MemoryPressure", "False"),
						condition("DiskPressure", "False"),
						condition("PIDPressure", "False"),
						condition("Ready", "True"),
					},
				},
			},
			want: true,
		},
		{
			name: "idle KEDA scaled object",
			object: map[string]interface{}{
				"apiVersion": "keda.sh/v1alpha1",
				"kind":       "ScaledObject",
				"metadata":   map[string]interface{}{"name": "scaler"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("Ready", "True"),
						// KEDA reports these on a perfectly healthy object.
						condition("Active", "False"),
						condition("Fallback", "False"),
						condition("Paused", "False"),
					},
				},
			},
			want: true,
		},
		{
			name: "KEDA scaled object that is not ready",
			object: map[string]interface{}{
				"apiVersion": "keda.sh/v1alpha1",
				"kind":       "ScaledObject",
				"metadata":   map[string]interface{}{"name": "scaler"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("Ready", "False"),
						condition("Paused", "False"),
					},
				},
			},
			want: false,
		},
		{
			name: "object with only unrecognized conditions",
			object: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "Widget",
				"metadata":   map[string]interface{}{"name": "widget"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("Paused", "False"),
					},
				},
			},
			want: true,
		},
		{
			name: "established CRD",
			object: map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinition",
				"metadata":   map[string]interface{}{"name": "widgets.example.com"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("Established", "True"),
						condition("NamesAccepted", "True"),
					},
				},
			},
			want: true,
		},
		{
			name: "CRD whose names were rejected",
			object: map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinition",
				"metadata":   map[string]interface{}{"name": "widgets.example.com"},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						condition("Established", "True"),
						condition("NamesAccepted", "False"),
					},
				},
			},
			want: false,
		},
		{
			name:   "unknown ready status is not treated as a failure",
			object: deploymentWithConditions(condition("Available", "Unknown")),
			want:   true,
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			u := &unstructured.Unstructured{Object: g.object}
			if got := isHealthy(u); got != g.want {
				t.Errorf("isHealthy() = %v, want %v", got, g.want)
			}
		})
	}
}

func condition(conditionType, status string) map[string]interface{} {
	return map[string]interface{}{
		"type":   conditionType,
		"status": status,
	}
}

func gatewayWithConditions(kind string, conditions ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       kind,
		"metadata":   map[string]interface{}{"name": "gw"},
		"status":     map[string]interface{}{"conditions": toList(conditions)},
	}
}

func deploymentWithConditions(conditions ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   map[string]interface{}{"name": "deploy"},
		"status":     map[string]interface{}{"conditions": toList(conditions)},
	}
}

func toList(conditions []map[string]interface{}) []interface{} {
	list := make([]interface{}, 0, len(conditions))
	for _, c := range conditions {
		list = append(list, c)
	}
	return list
}
