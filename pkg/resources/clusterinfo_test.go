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

package resources

import "testing"

func TestClusterInfoPublishesDNSRecords(t *testing.T) {
	tests := []struct {
		name        string
		clusterInfo ClusterInfo
		want        bool
	}{
		{
			name:        "hosted DNS",
			clusterInfo: ClusterInfo{Name: "cluster.example.com"},
			want:        true,
		},
		{
			name:        "hosted DNS with k8s.local name",
			clusterInfo: ClusterInfo{Name: "cluster.k8s.local"},
			want:        true,
		},
		{
			name:        "None DNS with hosted-style name",
			clusterInfo: ClusterInfo{Name: "cluster.example.com", UsesNoneDNS: true},
			want:        false,
		},
		{
			name:        "None DNS with k8s.local name",
			clusterInfo: ClusterInfo{Name: "cluster.k8s.local", UsesNoneDNS: true},
			want:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.clusterInfo.PublishesDNSRecords(); got != test.want {
				t.Errorf("PublishesDNSRecords() = %t, want %t", got, test.want)
			}
		})
	}
}
