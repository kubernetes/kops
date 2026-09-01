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

package gce

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestUsesIPAliases(t *testing.T) {
	tests := []struct {
		name              string
		networking        kops.NetworkingSpec
		nonMasqueradeCIDR string
		want              bool
	}{
		{
			name:       "GCP networking uses IP aliases",
			networking: kops.NetworkingSpec{GCP: &kops.GCPNetworkingSpec{}},
			want:       true,
		},
		{
			name:       "kindnet uses IP aliases",
			networking: kops.NetworkingSpec{Kindnet: &kops.KindnetNetworkingSpec{}},
			want:       true,
		},
		{
			name: "cni does not use IP aliases",
			networking: kops.NetworkingSpec{
				CNI: &kops.CNINetworkingSpec{},
			},
			want: false,
		},
		{
			// IP alias ranges are IPv4 secondary ranges; an IPv6-only cluster must not
			// use them, otherwise the IP-alias assignment path overwrites the IPv6
			// NonMasqueradeCIDR with IPv4 CIDRs.
			name:              "kindnet on an IPv6-only cluster does not use IP aliases",
			networking:        kops.NetworkingSpec{Kindnet: &kops.KindnetNetworkingSpec{}},
			nonMasqueradeCIDR: "::/0",
			want:              false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: tc.networking,
				},
			}
			cluster.Spec.Networking.NonMasqueradeCIDR = tc.nonMasqueradeCIDR

			if got := UsesIPAliases(cluster); got != tc.want {
				t.Errorf("UsesIPAliases() = %v, want %v", got, tc.want)
			}
		})
	}
}
