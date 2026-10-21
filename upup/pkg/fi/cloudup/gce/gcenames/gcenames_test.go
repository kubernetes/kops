/*
Copyright The Kubernetes Authors.

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

package gcenames

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestSafeClusterName(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		want        string
	}{
		{
			name:        "dots replaced with dashes",
			clusterName: "my.cluster.k8s.local",
			want:        "my-cluster-k8s-local",
		},
		{
			name:        "already safe name",
			clusterName: "my-cluster-k8s-local",
			want:        "my-cluster-k8s-local",
		},
		{
			name:        "empty string",
			clusterName: "",
			want:        "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SafeClusterName(tc.clusterName); got != tc.want {
				t.Errorf("SafeClusterName(%q) = %q, want %q", tc.clusterName, got, tc.want)
			}
		})
	}
}

func TestClusterPrefixedName(t *testing.T) {
	tests := []struct {
		name        string
		objectName  string
		clusterName string
		maxLength   int
		want        string
	}{
		{
			name:        "short cluster name without truncation",
			objectName:  "node",
			clusterName: "test.k8s.local",
			maxLength:   63,
			want:        "test-k8s-local-node",
		},
		{
			name:        "long cluster name truncated with hash",
			objectName:  "k8s-io-role-node",
			clusterName: "very-long-cluster-name-that-exceeds-the-maximum-allowed-length.k8s.local",
			maxLength:   63,
			want:        "very-long-cluster-name-that-exceeds-the-3b7s2q-k8s-io-role-node",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClusterPrefixedName(tc.objectName, tc.clusterName, tc.maxLength); got != tc.want {
				t.Errorf("ClusterPrefixedName(%q, %q, %d) = %q, want %q", tc.objectName, tc.clusterName, tc.maxLength, got, tc.want)
			}
		})
	}
}

func TestTagForRole(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		role        kops.InstanceGroupRole
		want        string
	}{
		{
			name:        "node role",
			clusterName: "test.k8s.local",
			role:        kops.InstanceGroupRoleNode,
			want:        "test-k8s-local-k8s-io-role-node",
		},
		{
			name:        "control-plane role",
			clusterName: "test.k8s.local",
			role:        kops.InstanceGroupRoleControlPlane,
			want:        "test-k8s-local-k8s-io-role-control-plane",
		},
		{
			name:        "bastion role",
			clusterName: "test.k8s.local",
			role:        kops.InstanceGroupRoleBastion,
			want:        "test-k8s-local-k8s-io-role-bastion",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := TagForRole(tc.clusterName, tc.role); got != tc.want {
				t.Errorf("TagForRole(%q, %q) = %q, want %q", tc.clusterName, tc.role, got, tc.want)
			}
		})
	}
}

func TestUsesIPAliases(t *testing.T) {
	tests := []struct {
		name    string
		cluster *kops.Cluster
		want    bool
	}{
		{
			name: "GCP networking enabled",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						GCP: &kops.GCPNetworkingSpec{},
					},
				},
			},
			want: true,
		},
		{
			name: "Kindnet networking enabled",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						Kindnet: &kops.KindnetNetworkingSpec{},
					},
				},
			},
			want: true,
		},
		{
			name: "Kubenet networking does not use IP aliases",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						Kubenet: &kops.KubenetNetworkingSpec{},
					},
				},
			},
			want: false,
		},
		{
			name: "Cilium networking does not use IP aliases",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := UsesIPAliases(tc.cluster); got != tc.want {
				t.Errorf("UsesIPAliases() = %v, want %v", got, tc.want)
			}
		})
	}
}
