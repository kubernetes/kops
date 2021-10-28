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

func TestMatchesClusterNameWithUUID(t *testing.T) {
	grid := []struct {
		Name        string
		ClusterName string
		Want        bool
	}{
		{
			Name:        "e2e-5e08b256bc-d3d02-k8s-l-51a343e2-c285-4e73-b933-0123456789ab",
			ClusterName: "e2e-5e08b256bc-d3d02.k8s.local",
			Want:        true,
		},
		{
			// UUID is one character too short
			Name:        "e2e-5e08b256bc-d3d02-k8s-l-51a343e2-c285-4e73-b933-0123456789a",
			ClusterName: "e2e-5e08b256bc-d3d02.k8s.local",
			Want:        false,
		},
		{
			// UUID is one character too short and prefix fills the gap
			Name:        "e2e-5e08b256bc-d3d02-k8s-lo-51a343e2-c285-4e73-b933-0123456789a",
			ClusterName: "e2e-5e08b256bc-d3d02.k8s.local",
			Want:        false,
		},
		{
			Name:        "example-k8s-local-51a343e2-c285-4e73-b933-0123456789ab",
			ClusterName: "example.k8s.local",
			Want:        true,
		},
		{
			Name:        "",
			ClusterName: "example.k8s.local",
			Want:        false,
		},
		{
			Name:        "51a343e2-c285-4e73-b933-0123456789ab",
			ClusterName: "example.k8s.local",
			Want:        false,
		},
	}
	for _, g := range grid {
		d := &clusterDiscoveryGCE{
			clusterName: g.ClusterName,
		}
		got := d.matchesClusterNameWithUUID(g.Name, maxGCERouteNameLength)
		if got != g.Want {
			t.Errorf("{clusterName=%q}.matchesClusterNameWithUUID(%q) got %v, want %v", g.ClusterName, g.Name, got, g.Want)
		}
	}
}
