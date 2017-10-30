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

package gce

import "testing"

func TestNameMatch(t *testing.T) {
	grid := []struct {
		Name  string
		Match bool
	}{
		{
			Name:  "nodeport-external-to-node-cluster-example-com",
			Match: true,
		},
		{
			Name:  "simple-cluster-example-com",
			Match: true,
		},
		{
			Name:  "-cluster-example-com",
			Match: false,
		},
		{
			Name:  "cluster-example-com",
			Match: false,
		},
		{
			Name:  "a-example-com",
			Match: false,
		},
		{
			Name:  "-example-com",
			Match: false,
		},
		{
			Name:  "",
			Match: false,
		},
	}
	for _, g := range grid {
		d := &clusterDiscoveryGCE{
			clusterName: "cluster.example.com",
		}
		match := d.matchesClusterNameMultipart(g.Name, maxPrefixTokens)
		if match != g.Match {
			t.Errorf("unexpected match value for %q, got %v, expected %v", g.Name, match, g.Match)
		}
	}
}
