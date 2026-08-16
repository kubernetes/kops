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

package digitalocean

import (
	"testing"

	"github.com/digitalocean/godo"

	"k8s.io/kops/pkg/resources"
)

// DigitalOcean SSH keys are account-scoped, so every cluster in an account sees
// every other cluster's keys. Matching too loosely here deletes keys belonging to
// live clusters.
func TestFilterClusterSSHKeys(t *testing.T) {
	cases := []struct {
		name        string
		clusterName string
		keys        []godo.Key
		expected    []string
	}{
		{
			name:        "matches the key kops created for the cluster",
			clusterName: "cluster.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "kubernetes.cluster.k8s.local-c4:a8:11:cc:c1:0d:cd:ba:0f:4f:0d:c1:d1:a2:2a:cd"},
			},
			expected: []string{"kubernetes.cluster.k8s.local-c4:a8:11:cc:c1:0d:cd:ba:0f:4f:0d:c1:d1:a2:2a:cd"},
		},
		{
			name:        "matches every key the cluster left behind across runs",
			clusterName: "cluster.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "kubernetes.cluster.k8s.local-aa:bb"},
				{ID: 2, Name: "kubernetes.cluster.k8s.local-cc:dd"},
			},
			expected: []string{"kubernetes.cluster.k8s.local-aa:bb", "kubernetes.cluster.k8s.local-cc:dd"},
		},
		{
			name:        "ignores another cluster whose name extends this one",
			clusterName: "cluster.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "kubernetes.cluster.k8s.local.example.com-aa:bb"},
			},
			expected: nil,
		},
		{
			// e2e-kops-do-dns-none vs e2e-kops-do-dns-none-ha, which run concurrently.
			name:        "ignores a sibling cluster with a suffixed name",
			clusterName: "e2e-kops-do-dns-none.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "kubernetes.e2e-kops-do-dns-none-ha.k8s.local-aa:bb"},
			},
			expected: nil,
		},
		{
			name:        "ignores an unrelated cluster",
			clusterName: "cluster.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "kubernetes.other.k8s.local-aa:bb"},
			},
			expected: nil,
		},
		{
			// A user-supplied key via spec.sshKeyName keeps its own name; kops did not
			// create it and must not delete it.
			name:        "ignores a key kops did not name",
			clusterName: "cluster.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "my-personal-key"},
				{ID: 2, Name: "cluster.k8s.local"},
			},
			expected: nil,
		},
		{
			name:        "ignores the bare cluster name with no fingerprint",
			clusterName: "cluster.k8s.local",
			keys: []godo.Key{
				{ID: 1, Name: "kubernetes.cluster.k8s.local"},
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := filterClusterSSHKeys(tc.keys, tc.clusterName)

			if len(actual) != len(tc.expected) {
				t.Fatalf("matched %d keys, expected %d: %v", len(actual), len(tc.expected), names(actual))
			}
			for i, want := range tc.expected {
				if actual[i].Name != want {
					t.Errorf("key[%d] = %q, expected %q", i, actual[i].Name, want)
				}
				if actual[i].Type != resourceTypeSSHKey {
					t.Errorf("key[%d] type = %q, expected %q", i, actual[i].Type, resourceTypeSSHKey)
				}
				if actual[i].Deleter == nil {
					t.Errorf("key[%d] has no deleter", i)
				}
			}
		})
	}
}

func names(rs []*resources.Resource) []string {
	var out []string
	for _, r := range rs {
		out = append(out, r.Name)
	}
	return out
}
