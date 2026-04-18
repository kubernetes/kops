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

package linodetasks

import (
	"context"
	"net"
	"slices"
	"testing"

	"github.com/linode/linodego"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func newTestCloudupContext(t *testing.T, cloud linode.LinodeCloud) *fi.CloudupContext {
	t.Helper()

	ctx, err := fi.NewCloudupContext(
		context.Background(),
		fi.DeletionProcessingModeDeleteIncludingDeferred,
		linode.NewAPITarget(cloud),
		nil,
		cloud,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error creating context: %v", err)
	}

	return ctx
}

func TestNormalizedLoadBalancerLabel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "dots replaced",
			input: "api.kops-test.linode.k8s.local",
			want:  "api-kops-test-linode-k8s-local",
		},
		{
			name:  "invalid chars removed",
			input: "api@@@###",
			want:  "api",
		},
		{
			name:  "empty fallback",
			input: "...",
			want:  "kops-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := linode.NormalizedLoadBalancerLabel(tt.input); got != tt.want {
				t.Fatalf("unexpected normalized label: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadBalancerRenderLinodeCreateWithoutBackends(t *testing.T) {
	privateIP := net.ParseIP("192.168.210.226")
	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{
			{
				ID:   101,
				Tags: []string{kops.LabelClusterName + ":kops-test.linode.k8s.local", linode.TagKubernetesInstanceRole + ":ControlPlane"},
				IPv4: []*net.IP{&privateIP},
			},
		},
	}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	expected := &LoadBalancer{
		Name:   fi.PtrTo("api.kops-test.linode.k8s.local"),
		Region: fi.PtrTo("us-east"),
		Tags: []string{
			kops.LabelClusterName + ":kops-test.linode.k8s.local",
			linode.TagKubernetesInstanceRole + ":ControlPlane",
		},
	}

	if err := (&LoadBalancer{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}

	if got, want := client.CreateNodeBalancerCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}

	if got, want := fi.ValueOf(client.LastCreateNodeBalancerOpts.Label), "api-kops-test-linode-k8s-local"; got != want {
		t.Fatalf("unexpected nodebalancer label: got %q, want %q", got, want)
	}

	if got := len(client.LastCreateNodeBalancerOpts.Configs); got != 0 {
		t.Fatalf("expected no configs when no backends are discovered, got %d", got)
	}
	if got, want := client.CreateNodeBalancerConfigCalls, 2; got != want {
		t.Fatalf("expected configs to be reconciled after create: got %d, want %d", got, want)
	}
}

func TestLoadBalancerFindMatchesNormalizedLabel(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListNodeBalancersResponse: []linodego.NodeBalancer{
			{ID: 7, Label: fi.PtrTo("api-kops-test-linode-k8s-local"), Region: "us-east"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &LoadBalancer{Name: fi.PtrTo("api.kops-test.linode.k8s.local")}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find nodebalancer by normalized label")
	}
	if got, want := fi.ValueOf(actual.ID), 7; got != want {
		t.Fatalf("unexpected nodebalancer id: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Region), "us-east"; got != want {
		t.Fatalf("unexpected nodebalancer region: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(task.ID), 7; got != want {
		t.Fatalf("task ID should be propagated after Find: got %d, want %d", got, want)
	}
}

func TestLinodeDiscoverControlPlaneBackendsPublicFallback(t *testing.T) {
	publicIP := net.ParseIP("203.0.113.10")
	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{
			{
				ID:   101,
				Tags: []string{kops.LabelClusterName + ":kops-test.linode.k8s.local", linode.TagKubernetesInstanceRole + ":ControlPlane"},
				IPv4: []*net.IP{&publicIP},
			},
		},
	}

	backends, err := linodeDiscoverControlPlaneBackends(client, []string{kops.LabelClusterName + ":kops-test.linode.k8s.local"})
	if err != nil {
		t.Fatalf("linodeDiscoverControlPlaneBackends returned error: %v", err)
	}
	if got, want := len(backends), 1; got != want {
		t.Fatalf("unexpected backend count: got %d, want %d", got, want)
	}
	if got, want := backends[0], "203.0.113.10"; got != want {
		t.Fatalf("unexpected backend IP: got %q, want %q", got, want)
	}
}

func TestEnsureLoadBalancerConfigsCreatesMissingPorts(t *testing.T) {
	client := &linode.MockLinodeClient{}
	backends := []string{"192.168.210.226"}

	err := ensureLoadBalancerConfigs(client, 2085634, "api.kops-test.linode.k8s.local", backends)
	if err != nil {
		t.Fatalf("ensureLoadBalancerConfigs returned error: %v", err)
	}

	if got, want := client.CreateNodeBalancerConfigCalls, 2; got != want {
		t.Fatalf("unexpected config create calls: got %d, want %d", got, want)
	}
	if got, want := client.CreateNodeBalancerNodeCalls, 2; got != want {
		t.Fatalf("unexpected node create calls: got %d, want %d", got, want)
	}

	ports := []int{client.CreateNodeBalancerConfigOpts[0].Port, client.CreateNodeBalancerConfigOpts[1].Port}
	slices.Sort(ports)
	if got, want := ports, []int{wellknownports.KubeAPIServer, wellknownports.KopsControllerPort}; !slices.Equal(got, want) {
		t.Fatalf("unexpected created ports: got %v, want %v", got, want)
	}
}

func TestEnsureLoadBalancerConfigsRebuildsExistingPorts(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListNodeBalancerConfigsResponse: []linodego.NodeBalancerConfig{
			{ID: 11, Port: wellknownports.KubeAPIServer},
			{ID: 12, Port: wellknownports.KopsControllerPort},
		},
	}
	backends := []string{"192.168.210.226"}

	err := ensureLoadBalancerConfigs(client, 2085634, "api.kops-test.linode.k8s.local", backends)
	if err != nil {
		t.Fatalf("ensureLoadBalancerConfigs returned error: %v", err)
	}

	if got, want := client.RebuildNodeBalancerConfigCalls, 2; got != want {
		t.Fatalf("unexpected config rebuild calls: got %d, want %d", got, want)
	}
	if got, want := client.CreateNodeBalancerNodeCalls, 2; got != want {
		t.Fatalf("expected missing nodes to be created during rebuild path: got %d, want %d", got, want)
	}
	if got := client.CreateNodeBalancerConfigCalls; got != 0 {
		t.Fatalf("did not expect create calls when configs already exist, got %d", got)
	}
}

func TestExtractClusterTag(t *testing.T) {
	tags := []string{"foo:bar", kops.LabelClusterName + ":kops-test.linode.k8s.local"}
	if got, want := extractClusterTag(tags), kops.LabelClusterName+":kops-test.linode.k8s.local"; got != want {
		t.Fatalf("unexpected cluster tag: got %q, want %q", got, want)
	}
}

func TestLoadBalancerRunWithoutBackendsDoesNotBlock(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListNodeBalancersResponse: []linodego.NodeBalancer{
			{ID: 1, Label: fi.PtrTo("api-kops-test-linode-k8s-local"), Region: "us-east"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &LoadBalancer{
		Name:      fi.PtrTo("api.kops-test.linode.k8s.local"),
		ID:        fi.PtrTo(1),
		Lifecycle: fi.LifecycleSync,
		Tags:      []string{kops.LabelClusterName + ":kops-test.linode.k8s.local"},
	}

	err := task.Run(ctx)
	if err != nil {
		t.Fatalf("expected non-blocking run when no control-plane backends are discovered, got %v", err)
	}
}

func TestLoadBalancerBackendsRunWithoutLoadBalancerDoesNotBlock(t *testing.T) {
	cloud := &linode.MockLinodeCloud{Client_: &linode.MockLinodeClient{}}
	ctx := newTestCloudupContext(t, cloud)

	task := &LoadBalancerBackends{
		Name: fi.PtrTo("backends.api.kops-test.linode.k8s.local"),
		LoadBalancer: &LoadBalancer{
			Name: fi.PtrTo("api.kops-test.linode.k8s.local"),
			Tags: []string{kops.LabelClusterName + ":kops-test.linode.k8s.local"},
		},
	}

	err := task.Run(ctx)
	// With APITarget (actual execution), should retry when LB not ready
	if _, ok := err.(*fi.TryAgainLaterError); !ok {
		t.Fatalf("expected TryAgainLaterError when LB not ready, got %v", err)
	}
}

func TestLoadBalancerBackendsRunWithoutBackendsDoesNotBlock(t *testing.T) {
	client := &linode.MockLinodeClient{}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &LoadBalancerBackends{
		Name: fi.PtrTo("backends.api.kops-test.linode.k8s.local"),
		LoadBalancer: &LoadBalancer{
			Name: fi.PtrTo("api.kops-test.linode.k8s.local"),
			ID:   fi.PtrTo(1),
			Tags: []string{kops.LabelClusterName + ":kops-test.linode.k8s.local"},
		},
	}

	err := task.Run(ctx)
	// With APITarget (actual execution), should retry when backends not ready
	if _, ok := err.(*fi.TryAgainLaterError); !ok {
		t.Fatalf("expected TryAgainLaterError when backends not ready, got %v", err)
	}
}

func TestLoadBalancerBackendsRunReconcilesWhenBackendsReady(t *testing.T) {
	privateIP := net.ParseIP("192.168.210.226")
	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{
			{
				ID:   101,
				Tags: []string{kops.LabelClusterName + ":kops-test.linode.k8s.local", linode.TagKubernetesInstanceRole + ":ControlPlane"},
				IPv4: []*net.IP{&privateIP},
			},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &LoadBalancerBackends{
		Name: fi.PtrTo("backends.api.kops-test.linode.k8s.local"),
		LoadBalancer: &LoadBalancer{
			Name: fi.PtrTo("api.kops-test.linode.k8s.local"),
			ID:   fi.PtrTo(1),
			Tags: []string{kops.LabelClusterName + ":kops-test.linode.k8s.local"},
		},
	}

	if err := task.Run(ctx); err != nil {
		t.Fatalf("unexpected error reconciling backends: %v", err)
	}

	if got, want := client.CreateNodeBalancerConfigCalls, 2; got != want {
		t.Fatalf("expected backend reconcile to create both configs: got %d, want %d", got, want)
	}
}

func TestLoadBalancerBackendsGetDependenciesOnlyLoadBalancer(t *testing.T) {
	lb := &LoadBalancer{Name: fi.PtrTo("api.kops-test.linode.k8s.local")}

	tasks := map[string]fi.CloudupTask{}

	deps := (&LoadBalancerBackends{LoadBalancer: lb}).GetDependencies(tasks)
	if got, want := len(deps), 1; got != want {
		t.Fatalf("unexpected dependency count: got %d, want %d", got, want)
	}

	if !slices.Contains(deps, fi.CloudupTask(lb)) {
		t.Fatalf("expected load balancer dependency")
	}
}
