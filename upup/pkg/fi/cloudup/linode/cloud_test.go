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

package linode

import (
	"errors"
	"net"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/linode/linodego"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

func ptrIP(s string) *net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil
	}
	return &ip
}

func TestNewCloud(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		region     string
		wantErrSub string
	}{
		{
			name:       "requires LINODE_TOKEN",
			region:     "us-east",
			wantErrSub: "LINODE_TOKEN is required",
		},
		{
			name:       "requires region",
			token:      "test-token",
			wantErrSub: "region is required",
		},
		{
			name:   "builds cloud",
			token:  "test-token",
			region: "us-east",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LINODE_TOKEN", tt.token)

			cloud, err := NewCloud(tt.region)
			if tt.wantErrSub != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErrSub)
				}
				if !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewCloud returned error: %v", err)
			}
			if got, want := cloud.ProviderID(), kops.CloudProviderLinode; got != want {
				t.Fatalf("provider mismatch: got %q, want %q", got, want)
			}
			if got, want := cloud.Region(), tt.region; got != want {
				t.Fatalf("region mismatch: got %q, want %q", got, want)
			}
			if got, want := cloud.AccessToken(), tt.token; got != want {
				t.Fatalf("access token mismatch: got %q, want %q", got, want)
			}
			if cloud.Client() == nil {
				t.Fatalf("expected Linode client to be initialized")
			}
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	t.Run("deletes instance", func(t *testing.T) {
		client := &MockLinodeClient{}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DeleteInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err != nil {
			t.Fatalf("DeleteInstance returned error: %v", err)
		}

		if got, want := client.DeletedInstanceIDs, []int{42}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected deleted IDs: got %v, want %v", got, want)
		}
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		client := &MockLinodeClient{}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DeleteInstance(&cloudinstances.CloudInstance{ID: "not-a-number"})
		if err == nil {
			t.Fatalf("expected invalid ID error")
		}
		if !strings.Contains(err.Error(), "invalid Linode (Akamai) instance ID") {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(client.DeletedInstanceIDs) != 0 {
			t.Fatalf("expected no delete calls, got %v", client.DeletedInstanceIDs)
		}
	})

	t.Run("ignores not found", func(t *testing.T) {
		client := &MockLinodeClient{DeleteInstanceErrByID: map[int]error{
			42: &linodego.Error{Code: 404, Message: "not found"},
		}}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DeleteInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err != nil {
			t.Fatalf("DeleteInstance returned error: %v", err)
		}

		if got, want := client.DeletedInstanceIDs, []int{42}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected deleted IDs: got %v, want %v", got, want)
		}
	})

	t.Run("returns API errors", func(t *testing.T) {
		client := &MockLinodeClient{DeleteInstanceErrByID: map[int]error{
			42: errors.New("api unavailable"),
		}}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DeleteInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err == nil {
			t.Fatalf("expected delete error")
		}
		if !strings.Contains(err.Error(), "error deleting Linode (Akamai) instance") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDeleteGroup(t *testing.T) {
	t.Run("deletes instances for group and cluster", func(t *testing.T) {
		client := &MockLinodeClient{
			ListInstancesResponse: []linodego.Instance{
				{ID: 101, Tags: []string{"kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/cluster:example.k8s.local"}},
				{ID: 102, Tags: []string{"kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/cluster:example.k8s.local"}},
				{ID: 103, Tags: []string{"kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/cluster:other.k8s.local"}},
				{ID: 104, Tags: []string{"kops.k8s.io/instance-group:control-plane-us-east", "kops.k8s.io/cluster:example.k8s.local"}},
			},
		}
		cloud := &Cloud{region: "us-east", client: client}

		group := &cloudinstances.CloudInstanceGroup{
			InstanceGroup: &kops.InstanceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodes-us-east",
					Labels: map[string]string{
						kops.LabelClusterName: "example.k8s.local",
					},
				},
			},
		}

		err := cloud.DeleteGroup(group)
		if err != nil {
			t.Fatalf("DeleteGroup returned error: %v", err)
		}

		sort.Ints(client.DeletedInstanceIDs)
		if got, want := client.DeletedInstanceIDs, []int{101, 102}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected deleted IDs: got %v, want %v", got, want)
		}
	})

	t.Run("deletes by instance-group tag when cluster label is missing", func(t *testing.T) {
		client := &MockLinodeClient{
			ListInstancesResponse: []linodego.Instance{
				{ID: 101, Tags: []string{"kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/cluster:example.k8s.local"}},
				{ID: 102, Tags: []string{"kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/cluster:other.k8s.local"}},
				{ID: 103, Tags: []string{"kops.k8s.io/instance-group:control-plane-us-east", "kops.k8s.io/cluster:example.k8s.local"}},
			},
		}
		cloud := &Cloud{region: "us-east", client: client}

		group := &cloudinstances.CloudInstanceGroup{
			InstanceGroup: &kops.InstanceGroup{ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"}},
		}

		err := cloud.DeleteGroup(group)
		if err != nil {
			t.Fatalf("DeleteGroup returned error: %v", err)
		}

		sort.Ints(client.DeletedInstanceIDs)
		if got, want := client.DeletedInstanceIDs, []int{101, 102}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected deleted IDs: got %v, want %v", got, want)
		}
	})

	t.Run("returns list errors", func(t *testing.T) {
		client := &MockLinodeClient{ListInstancesError: errors.New("api unavailable")}
		cloud := &Cloud{region: "us-east", client: client}

		group := &cloudinstances.CloudInstanceGroup{InstanceGroup: &kops.InstanceGroup{ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"}}}
		err := cloud.DeleteGroup(group)
		if err == nil {
			t.Fatalf("expected list error")
		}
		if !strings.Contains(err.Error(), "error listing Linode (Akamai) instances") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns delete errors", func(t *testing.T) {
		client := &MockLinodeClient{
			ListInstancesResponse: []linodego.Instance{{ID: 101, Tags: []string{"kops.k8s.io/instance-group:nodes-us-east"}}},
			DeleteInstanceErrByID: map[int]error{101: errors.New("api unavailable")},
		}
		cloud := &Cloud{region: "us-east", client: client}

		group := &cloudinstances.CloudInstanceGroup{InstanceGroup: &kops.InstanceGroup{ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"}}}
		err := cloud.DeleteGroup(group)
		if err == nil {
			t.Fatalf("expected delete error")
		}
		if !strings.Contains(err.Error(), "error deleting Linode (Akamai) instance") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDetachInstance(t *testing.T) {
	t.Run("removes instance-role tag", func(t *testing.T) {
		client := &MockLinodeClient{
			ListInstancesResponse: []linodego.Instance{{
				ID:   42,
				Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/instance-role:Node", "env:test"},
			}},
		}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err != nil {
			t.Fatalf("DetachInstance returned error: %v", err)
		}

		if got, want := client.UpdatedInstanceIDs, []int{42}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected updated IDs: got %v, want %v", got, want)
		}

		wantTags := []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east", "env:test"}
		if got := client.UpdatedTagsByID[42]; !reflect.DeepEqual(got, wantTags) {
			t.Fatalf("unexpected updated tags: got %v, want %v", got, wantTags)
		}
	})

	t.Run("returns nil when instance already missing", func(t *testing.T) {
		client := &MockLinodeClient{ListInstancesResponse: []linodego.Instance{{ID: 101, Tags: []string{"kops.k8s.io/instance-role:Node"}}}}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err != nil {
			t.Fatalf("DetachInstance returned error: %v", err)
		}
		if len(client.UpdatedInstanceIDs) != 0 {
			t.Fatalf("expected no update calls, got %v", client.UpdatedInstanceIDs)
		}
	})

	t.Run("returns nil when no instance-role tag", func(t *testing.T) {
		client := &MockLinodeClient{ListInstancesResponse: []linodego.Instance{{
			ID:   42,
			Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east"},
		}}}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err != nil {
			t.Fatalf("DetachInstance returned error: %v", err)
		}
		if len(client.UpdatedInstanceIDs) != 0 {
			t.Fatalf("expected no update calls, got %v", client.UpdatedInstanceIDs)
		}
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		client := &MockLinodeClient{}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "not-a-number"})
		if err == nil {
			t.Fatalf("expected invalid ID error")
		}
		if !strings.Contains(err.Error(), "invalid Linode (Akamai) instance ID") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns list errors", func(t *testing.T) {
		client := &MockLinodeClient{ListInstancesError: errors.New("api unavailable")}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err == nil {
			t.Fatalf("expected list error")
		}
		if !strings.Contains(err.Error(), "error listing Linode (Akamai) instances for detach") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ignores not found update errors", func(t *testing.T) {
		client := &MockLinodeClient{
			ListInstancesResponse: []linodego.Instance{{ID: 42, Tags: []string{"kops.k8s.io/instance-role:Node"}}},
			UpdateInstanceErrByID: map[int]error{42: &linodego.Error{Code: 404, Message: "not found"}},
		}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err != nil {
			t.Fatalf("DetachInstance returned error: %v", err)
		}
	})

	t.Run("returns update errors", func(t *testing.T) {
		client := &MockLinodeClient{
			ListInstancesResponse: []linodego.Instance{{ID: 42, Tags: []string{"kops.k8s.io/instance-role:Node"}}},
			UpdateInstanceErrByID: map[int]error{42: errors.New("api unavailable")},
		}
		cloud := &Cloud{region: "us-east", client: client}

		err := cloud.DetachInstance(&cloudinstances.CloudInstance{ID: "42"})
		if err == nil {
			t.Fatalf("expected update error")
		}
		if !strings.Contains(err.Error(), "error detaching Linode (Akamai) instance") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetApiIngressStatus(t *testing.T) {
	tests := []struct {
		name         string
		publicName   string
		loadBalancer bool
		client       LinodeClient
		wantHosts    []string
		wantIPs      []string
		wantErrSub   string
	}{
		{
			name:       "hostname ingress",
			publicName: "api.example.test",
			wantHosts:  []string{"api.example.test"},
		},
		{
			name:       "ip ingress",
			publicName: "203.0.113.15",
			wantIPs:    []string{"203.0.113.15"},
		},
		{
			name:         "load balancer ingress",
			loadBalancer: true,
			client: &MockLinodeClient{
				ListNodeBalancersResponse: []linodego.NodeBalancer{
					{
						Label: fi.PtrTo("api-example-k8s-local"),
						IPv4:  fi.PtrTo("203.0.113.20"),
					},
				},
			},
			wantIPs: []string{"203.0.113.20"},
		},
		{
			name:         "empty ingress when load balancer not created yet",
			loadBalancer: true,
			client: &MockLinodeClient{ListInstancesResponse: []linodego.Instance{
				{
					Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:ControlPlane"},
					IPv4: []*net.IP{ptrIP("10.0.0.10"), ptrIP("198.51.100.10")},
				},
				{
					Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:APIServer"},
					IPv4: []*net.IP{ptrIP("198.51.100.11")},
				},
			}},
			// When LoadBalancer is configured but doesn't exist yet, return empty (not fallback to instances)
		},
		{
			name: "control-plane ingress from instances when no load balancer",
			client: &MockLinodeClient{ListInstancesResponse: []linodego.Instance{
				{
					Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:ControlPlane"},
					IPv4: []*net.IP{ptrIP("10.0.0.10"), ptrIP("198.51.100.10")},
				},
				{
					Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:APIServer"},
					IPv4: []*net.IP{ptrIP("198.51.100.11")},
				},
				{
					Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-role:Node"},
					IPv4: []*net.IP{ptrIP("198.51.100.12")},
				},
			}},
			wantIPs: []string{"198.51.100.10", "198.51.100.11"},
		},
		{
			name:         "empty ingress when no api endpoint data",
			loadBalancer: true,
			client:       &MockLinodeClient{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud := &Cloud{region: "us-east", client: tt.client}
			cluster := &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
				Spec:       kops.ClusterSpec{API: kops.APISpec{PublicName: tt.publicName}},
			}
			if tt.loadBalancer {
				cluster.Spec.API.LoadBalancer = &kops.LoadBalancerAccessSpec{}
			}

			ingresses, err := cloud.GetApiIngressStatus(cluster)
			if tt.wantErrSub != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErrSub)
				}
				if !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetApiIngressStatus returned error: %v", err)
			}

			var gotHosts []string
			var gotIPs []string
			for _, ingress := range ingresses {
				if ingress.Hostname != "" {
					gotHosts = append(gotHosts, ingress.Hostname)
				}
				if ingress.IP != "" {
					gotIPs = append(gotIPs, ingress.IP)
				}
			}
			sort.Strings(gotHosts)
			sort.Strings(gotIPs)

			wantHosts := append([]string(nil), tt.wantHosts...)
			wantIPs := append([]string(nil), tt.wantIPs...)
			sort.Strings(wantHosts)
			sort.Strings(wantIPs)

			if !reflect.DeepEqual(gotHosts, wantHosts) {
				t.Fatalf("unexpected hostnames: got %v, want %v", gotHosts, wantHosts)
			}
			if !reflect.DeepEqual(gotIPs, wantIPs) {
				t.Fatalf("unexpected IPs: got %v, want %v", gotIPs, wantIPs)
			}
		})
	}
}

func TestGetCloudGroups(t *testing.T) {
	cloud := &Cloud{region: "us-east"}
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Linode: &kops.LinodeSpec{},
			},
		},
	}

	instanceGroups := []*kops.InstanceGroup{
		{ObjectMeta: metav1.ObjectMeta{Name: "control-plane-us-east"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"}},
	}

	nodes := []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cp-1",
				Labels: map[string]string{
					kops.NodeLabelInstanceGroup: "control-plane-us-east",
				},
			},
			Spec: v1.NodeSpec{ProviderID: "linode:///111"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					kops.NodeLabelInstanceGroup: "nodes-us-east",
				},
				Annotations: map[string]string{
					"kops.k8s.io/needs-update": "yes",
				},
			},
			Spec: v1.NodeSpec{ProviderID: "linode:///222"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "orphan-node",
				Labels: map[string]string{
					kops.NodeLabelInstanceGroup: "unknown-group",
				},
			},
			Spec: v1.NodeSpec{ProviderID: "linode:///333"},
		},
	}

	groups, err := cloud.GetCloudGroups(cluster, instanceGroups, true, nodes)
	if err != nil {
		t.Fatalf("GetCloudGroups returned error: %v", err)
	}

	cpGroup := groups["control-plane-us-east"]
	if cpGroup == nil {
		t.Fatalf("missing control-plane group")
	}
	if got, want := len(cpGroup.Ready), 1; got != want {
		t.Fatalf("unexpected control-plane ready count: got %d, want %d", got, want)
	}
	if got, want := len(cpGroup.NeedUpdate), 0; got != want {
		t.Fatalf("unexpected control-plane need-update count: got %d, want %d", got, want)
	}

	nodesGroup := groups["nodes-us-east"]
	if nodesGroup == nil {
		t.Fatalf("missing nodes group")
	}
	if got, want := len(nodesGroup.Ready), 0; got != want {
		t.Fatalf("unexpected nodes ready count: got %d, want %d", got, want)
	}
	if got, want := len(nodesGroup.NeedUpdate), 1; got != want {
		t.Fatalf("unexpected nodes need-update count: got %d, want %d", got, want)
	}
}
