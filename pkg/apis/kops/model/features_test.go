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

package model

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestUseCiliumEtcd(t *testing.T) {
	for _, tc := range []struct {
		cluster  *kops.Cluster
		expected bool
	}{
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Name: "cilium",
						},
					},
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							Version: "v1.8",
						},
					},
				},
			},
			expected: true,
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Name: "cilium",
						},
					},
					Networking: &kops.NetworkingSpec{},
				},
			},
			expected: false,
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Name: "calico",
						},
					},
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							Version: "v1.8",
						},
					},
				},
			},
			expected: false,
		},
	} {
		if !reflect.DeepEqual(tc.expected, UseCiliumEtcd(tc.cluster)) {
			t.Errorf("expected %v, but got %v", tc.expected, UseCiliumEtcd(tc.cluster))
		}
	}
}
