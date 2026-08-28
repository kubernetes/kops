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

import (
	"context"
	"reflect"
	"sort"
	"testing"

	compute "google.golang.org/api/compute/v1"
	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/pkg/resources"
)

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

func TestContainsOnlyListedIGMs(t *testing.T) {
	igms := []*resources.Resource{
		{Name: "nodes-igm"},
		{Name: "master-igm"},
	}

	tests := []struct {
		name    string
		service *compute.BackendService
		want    bool
	}{
		{
			name: "empty backends",
			service: &compute.BackendService{
				Backends: nil,
			},
			want: false,
		},
		{
			name: "all backends match listed igms",
			service: &compute.BackendService{
				Backends: []*compute.Backend{
					{Group: "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-a/instanceGroups/nodes-igm"},
					{Group: "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-b/instanceGroups/master-igm"},
				},
			},
			want: true,
		},
		{
			name: "contains backend from non listed igm",
			service: &compute.BackendService{
				Backends: []*compute.Backend{
					{Group: "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-a/instanceGroups/nodes-igm"},
					{Group: "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-b/instanceGroups/gke-default-igm"},
				},
			},
			want: false,
		},
		{
			name: "NEG backends from GKE Gateway do not match any IGM",
			service: &compute.BackendService{
				Backends: []*compute.Backend{
					{Group: "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-a/networkEndpointGroups/gkegw1-serve404-neg"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsOnlyListedIGMs(tt.service, igms)
			if got != tt.want {
				t.Fatalf("containsOnlyListedIGMs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// routeInserter is implemented by the mock compute client; the production RouteClient
// does not create routes, so tests assert to this interface to seed them.
type routeInserter interface {
	Insert(project string, route *compute.Route) (*compute.Operation, error)
}

func trackerKeys(trackers []*resources.Resource) []string {
	var keys []string
	for _, t := range trackers {
		keys = append(keys, t.Type+":"+t.ID)
	}
	sort.Strings(keys)
	return keys
}

// TestListFirewallRulesLoadBalancer verifies that we find all the resources of a Service
// LoadBalancer with externalTrafficPolicy: Local, whose health check is named for the load
// balancer rather than the cluster.
func TestListFirewallRulesLoadBalancer(t *testing.T) {
	ctx := context.TODO()

	const (
		project     = "testproject"
		region      = "us-test1"
		clusterName = "cluster.example.com"
		lbName      = "a1b2c3d4e5f67890a1b2c3d4e5f67890"
	)

	c := gcemock.InstallMockGCECloud(region, project)

	network := "https://www.googleapis.com/compute/v1/projects/" + project + "/global/networks/cluster-example-com"
	nodeTags := []string{"cluster-example-com-k8s-io-role-node"}

	for _, firewall := range []*compute.Firewall{
		{Name: "k8s-fw-" + lbName, Network: network, TargetTags: nodeTags},
		{Name: "k8s-" + lbName + "-http-hc", Network: network, TargetTags: nodeTags},
		{Name: "k8s-fw-ffffffffffffffffffffffffffffffff", Network: network, TargetTags: []string{"other-example-com-k8s-io-role-node"}},
	} {
		if _, err := c.Compute().Firewalls().Insert(project, firewall); err != nil {
			t.Fatalf("error creating firewall %q: %v", firewall.Name, err)
		}
	}

	if _, err := c.Compute().ForwardingRules().Insert(ctx, project, region, &compute.ForwardingRule{
		Name:   lbName,
		Target: "https://www.googleapis.com/compute/v1/projects/" + project + "/regions/" + region + "/targetPools/" + lbName,
	}); err != nil {
		t.Fatalf("error creating forwarding rule: %v", err)
	}

	if _, err := c.Compute().TargetPools().Insert(project, region, &compute.TargetPool{
		Name:         lbName,
		HealthChecks: []string{"https://www.googleapis.com/compute/v1/projects/" + project + "/global/httpHealthChecks/" + lbName},
	}); err != nil {
		t.Fatalf("error creating target pool: %v", err)
	}

	if _, err := c.Compute().HTTPHealthChecks().Insert(project, &compute.HttpHealthCheck{Name: lbName}); err != nil {
		t.Fatalf("error creating http health check: %v", err)
	}

	d := &clusterDiscoveryGCE{
		cloud:       c,
		gceCloud:    c,
		clusterName: clusterName,
	}

	trackers, err := d.listFirewallRules()
	if err != nil {
		t.Fatalf("error listing firewall rules: %v", err)
	}

	want := []string{
		"FirewallRule:k8s-" + lbName + "-http-hc",
		"FirewallRule:k8s-fw-" + lbName,
		"ForwardingRule:" + lbName,
		"HTTP HealthCheck:" + lbName,
		"TargetPool:" + lbName,
	}
	sort.Strings(want)
	if got := trackerKeys(trackers); !reflect.DeepEqual(got, want) {
		t.Errorf("listFirewallRules() got %v, want %v", got, want)
	}
}

// TestListRoutesBlockedOnInstanceGroupManagers verifies that we delete routes only after the
// managed instance groups, which would otherwise recreate a control-plane instance that
// recreates the routes.
func TestListRoutesBlockedOnInstanceGroupManagers(t *testing.T) {
	ctx := context.TODO()

	const (
		project     = "testproject"
		region      = "us-test1"
		zone        = "us-test1-a"
		clusterName = "cluster.example.com"
	)

	c := gcemock.InstallMockGCECloud(region, project)

	routes, ok := c.Compute().Routes().(routeInserter)
	if !ok {
		t.Fatalf("mock route client does not implement Insert")
	}
	for _, route := range []*compute.Route{
		{
			Name:            "cluster-example-com-51a343e2-c285-4e73-b933-18a6ea44c3e4",
			NextHopInstance: "https://www.googleapis.com/compute/v1/projects/" + project + "/zones/" + zone + "/instances/nodes-abcd",
		},
		{
			Name:            "other-example-com-51a343e2-c285-4e73-b933-18a6ea44c3e4",
			NextHopInstance: "https://www.googleapis.com/compute/v1/projects/" + project + "/zones/" + zone + "/instances/nodes-abcd",
		},
	} {
		if _, err := routes.Insert(project, route); err != nil {
			t.Fatalf("error creating route %q: %v", route.Name, err)
		}
	}

	resourceMap := map[string]*resources.Resource{
		"Instance:" + zone + "/nodes-abcd": {
			Name: "nodes-abcd",
			ID:   zone + "/nodes-abcd",
			Type: typeInstance,
		},
		"InstanceGroupManager:" + zone + "/a-nodes-" + zone: {
			Name: "a-nodes-" + zone,
			ID:   zone + "/a-nodes-" + zone,
			Type: typeInstanceGroupManager,
		},
	}

	d := &clusterDiscoveryGCE{
		cloud:       c,
		gceCloud:    c,
		clusterName: clusterName,
	}

	trackers, err := d.listRoutes(ctx, resourceMap)
	if err != nil {
		t.Fatalf("error listing routes: %v", err)
	}

	want := []string{"Route:cluster-example-com-51a343e2-c285-4e73-b933-18a6ea44c3e4"}
	if got := trackerKeys(trackers); !reflect.DeepEqual(got, want) {
		t.Fatalf("listRoutes() got %v, want %v", got, want)
	}

	wantBlocked := []string{
		"Instance:" + zone + "/nodes-abcd",
		"InstanceGroupManager:" + zone + "/a-nodes-" + zone,
	}
	got := append([]string{}, trackers[0].Blocked...)
	sort.Strings(got)
	if !reflect.DeepEqual(got, wantBlocked) {
		t.Errorf("route Blocked got %v, want %v", got, wantBlocked)
	}
}
