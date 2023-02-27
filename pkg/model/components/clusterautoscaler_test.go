/*
Copyright 2021 The Kubernetes Authors.

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

package components

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func buildSpec(smc *api.ClusterAutoscalerServiceMonitorConfig) *api.ClusterSpec {
	return &api.ClusterSpec{
		ClusterAutoscaler: &api.ClusterAutoscalerConfig{
			Enabled:        fi.PtrTo(true),
			ServiceMonitor: smc,
		},
		KubernetesVersion: "v1.20.0",
	}
}

func Test_ClusterAutoscalerOptionsBuilder_ServiceMonitorConfig(t *testing.T) {
	testCases := []struct {
		description string
		inSM        *api.ClusterAutoscalerServiceMonitorConfig
		want        api.ClusterAutoscalerServiceMonitorConfig
	}{
		{
			description: "ServiceMonitor nil",
			inSM:        nil,
			want: api.ClusterAutoscalerServiceMonitorConfig{
				Enabled: fi.PtrTo(false),
			},
		},
		{
			description: "ServiceMonitor empty",
			inSM:        &api.ClusterAutoscalerServiceMonitorConfig{},
			want: api.ClusterAutoscalerServiceMonitorConfig{
				Enabled:        fi.PtrTo(false),
				Namespace:      fi.PtrTo("kube-system"),
				ScrapeInterval: fi.PtrTo("30s"),
			},
		},
		{
			description: "All possible values set",
			inSM: &api.ClusterAutoscalerServiceMonitorConfig{
				Enabled:        fi.PtrTo(true),
				Namespace:      fi.PtrTo("monitoring"),
				ScrapeInterval: fi.PtrTo("10s"),
			},
			want: api.ClusterAutoscalerServiceMonitorConfig{
				Enabled:        fi.PtrTo(true),
				Namespace:      fi.PtrTo("monitoring"),
				ScrapeInterval: fi.PtrTo("10s"),
			},
		},
		{
			description: "Fields partly set",
			inSM: &api.ClusterAutoscalerServiceMonitorConfig{
				Enabled: fi.PtrTo(true),
			},
			want: api.ClusterAutoscalerServiceMonitorConfig{
				Enabled:        fi.PtrTo(true),
				Namespace:      fi.PtrTo("kube-system"),
				ScrapeInterval: fi.PtrTo("30s"),
			},
		},
	}

	ob := ClusterAutoscalerOptionsBuilder{}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			clusterSpec := buildSpec(test.inSM)

			err := ob.BuildOptions(clusterSpec)
			if err != nil {
				t.Fatal(err)
			}

			got := *clusterSpec.ClusterAutoscaler.ServiceMonitor

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
