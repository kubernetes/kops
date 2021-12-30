/*
Copyright 2017 The Kubernetes Authors.

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

package cloudup

import (
	"reflect"
	"sort"
	"testing"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"k8s.io/kops/pkg/apis/kops"
)

func TestPrecreateDNSNames(t *testing.T) {
	grid := []struct {
		cluster  *kops.Cluster
		expected []recordKey
	}{
		{
			cluster: &kops.Cluster{},
			expected: []recordKey{
				{"api.cluster1.example.com", rrstype.A},
				{"api.internal.cluster1.example.com", rrstype.A},
			},
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					NonMasqueradeCIDR: "::/0",
				},
			},
			expected: []recordKey{
				{"api.cluster1.example.com", rrstype.A},
				{"api.cluster1.example.com", rrstype.AAAA},
				{"api.internal.cluster1.example.com", rrstype.AAAA},
			},
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					API: &kops.AccessSpec{
						LoadBalancer: &kops.LoadBalancerAccessSpec{},
					},
				},
			},
			expected: []recordKey{
				{"api.internal.cluster1.example.com", rrstype.A},
			},
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					API: &kops.AccessSpec{
						LoadBalancer: &kops.LoadBalancerAccessSpec{},
					},
					NonMasqueradeCIDR: "::/0",
				},
			},
			expected: []recordKey{
				{"api.internal.cluster1.example.com", rrstype.AAAA},
			},
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					API: &kops.AccessSpec{
						LoadBalancer: &kops.LoadBalancerAccessSpec{
							UseForInternalAPI: true,
						},
					},
				},
			},
			expected: nil,
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
					KubernetesVersion: "1.22.0",
				},
			},
			expected: []recordKey{
				{"api.cluster1.example.com", rrstype.A},
				{"api.internal.cluster1.example.com", rrstype.A},
				{"kops-controller.internal.cluster1.example.com", rrstype.A},
			},
		},
		{
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
					KubernetesVersion: "1.22.0",
					NonMasqueradeCIDR: "::/0",
				},
			},
			expected: []recordKey{
				{"api.cluster1.example.com", rrstype.A},
				{"api.cluster1.example.com", rrstype.AAAA},
				{"api.internal.cluster1.example.com", rrstype.AAAA},
				{"kops-controller.internal.cluster1.example.com", rrstype.AAAA},
			},
		},
	}

	for _, g := range grid {
		cluster := g.cluster

		cluster.ObjectMeta.Name = "cluster1.example.com"
		cluster.Spec.MasterPublicName = "api." + cluster.ObjectMeta.Name
		cluster.Spec.MasterInternalName = "api.internal." + cluster.ObjectMeta.Name
		cluster.Spec.EtcdClusters = []kops.EtcdClusterSpec{
			{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{Name: "zone1"},
					{Name: "zone2"},
					{Name: "zone3"},
				},
			},
			{
				Name: "events",
				Members: []kops.EtcdMemberSpec{
					{Name: "zonea"},
					{Name: "zoneb"},
					{Name: "zonec"},
				},
			},
		}

		actual := buildPrecreateDNSHostnames(cluster)

		expected := g.expected
		sort.Slice(actual, func(i, j int) bool {
			if actual[i].hostname < actual[j].hostname {
				return true
			}
			if actual[i].hostname == actual[j].hostname && actual[i].rrsType < actual[j].rrsType {
				return true
			}
			return false
		})

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("unexpected records.  expected=%v actual=%v", expected, actual)
		}
	}
}
