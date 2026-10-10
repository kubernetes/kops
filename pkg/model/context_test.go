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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
)

func TestCloudTagsForInstanceGroup_Taints(t *testing.T) {
	grid := []struct {
		name     string
		taints   []string
		wantTags map[string]string
	}{
		{
			name:   "taint with value and effect",
			taints: []string{"foo=bar:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": "bar:NoSchedule",
			},
		},
		{
			name:   "taint without value (key:effect)",
			taints: []string{"foo:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": ":NoSchedule",
			},
		},
		{
			name: "mix of taint formats",
			taints: []string{
				"foo:NoSchedule",
				"bar=baz:PreferNoSchedule",
			},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": ":NoSchedule",
				"k8s.io/cluster-autoscaler/node-template/taint/bar": "baz:PreferNoSchedule",
			},
		},
		{
			name:     "taint without effect is skipped",
			taints:   []string{"foo"},
			wantTags: map[string]string{},
		},
	}

	for _, tc := range grid {
		t.Run(tc.name, func(t *testing.T) {
			cluster := testutils.BuildMinimalClusterAWS("testcluster.test.com")
			ig := &kops.InstanceGroup{}
			ig.ObjectMeta.Name = "nodes"
			ig.Spec.Role = kops.InstanceGroupRoleNode
			ig.Spec.Taints = tc.taints

			b := &KopsModelContext{
				IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
				AllInstanceGroups: []*kops.InstanceGroup{ig},
				InstanceGroups:    []*kops.InstanceGroup{ig},
			}

			tags, err := b.CloudTagsForInstanceGroup(ig)
			if err != nil {
				t.Fatalf("CloudTagsForInstanceGroup() error = %v", err)
			}

			for k, want := range tc.wantTags {
				if got := tags[k]; got != want {
					t.Errorf("tag %q = %q, want %q", k, got, want)
				}
			}
		})
	}
}

// TestServesWellKnownService covers the cluster-aware resolution of endpoint membership, in
// particular the historic behaviour that applies while no instance group configures the field.
func TestServesWellKnownService(t *testing.T) {
	ig := func(name string, role kops.InstanceGroupRole, services []kops.WellKnownService) *kops.InstanceGroup {
		g := &kops.InstanceGroup{}
		g.ObjectMeta.Name = name
		g.Spec.Role = role
		g.Spec.ServesWellKnownServices = services
		return g
	}

	const external = kops.WellKnownServiceKubeAPIServerExternal
	const internal = kops.WellKnownServiceKubeAPIServerInternal
	const kopsController = kops.WellKnownServiceKopsController
	const etcdMain = kops.WellKnownServiceEtcdMain

	grid := []struct {
		name   string
		groups []*kops.InstanceGroup
		// want maps instance group name to the endpoints it should serve.
		want map[string][]kops.WellKnownService
	}{
		{
			name: "plain cluster: the control plane serves everything",
			groups: []*kops.InstanceGroup{
				ig("master", kops.InstanceGroupRoleControlPlane, nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]kops.WellKnownService{
				"master": {external, internal, kopsController, etcdMain},
				"nodes":  {},
			},
		},
		{
			// The historic rule from #18496: a dedicated API server group displaces the
			// control plane on the externally reachable endpoint, which the role-derived
			// default cannot express on its own.
			name: "dedicated API server displaces the control plane externally",
			groups: []*kops.InstanceGroup{
				ig("master", kops.InstanceGroupRoleControlPlane, nil),
				ig("apiserver", kops.InstanceGroupRoleAPIServer, nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]kops.WellKnownService{
				"master":    {internal, kopsController, etcdMain},
				"apiserver": {external, internal, kopsController},
				"nodes":     {},
			},
		},
		{
			name: "configuring any group makes the field authoritative",
			groups: []*kops.InstanceGroup{
				ig("master", kops.InstanceGroupRoleControlPlane, nil),
				ig("apiserver", kops.InstanceGroupRoleAPIServer, []kops.WellKnownService{internal}),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]kops.WellKnownService{
				// No longer displaced, because the historic rule no longer applies.
				"master":    {external, internal, kopsController, etcdMain},
				"apiserver": {internal},
				"nodes":     {},
			},
		},
		{
			name: "split control plane with an external and an internal API server",
			groups: []*kops.InstanceGroup{
				ig("external", kops.InstanceGroupRoleAPIServer, []kops.WellKnownService{external}),
				ig("internal", kops.InstanceGroupRoleAPIServer, []kops.WellKnownService{internal, kopsController}),
				ig("etcd", kops.InstanceGroupRoleEtcd, nil),
				ig("scheduler", "APIServer,Scheduler", nil),
				ig("kcm", "APIServer,KubeControllerManager", nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]kops.WellKnownService{
				"external": {external},
				"internal": {internal, kopsController},
				"etcd":     {etcdMain},
				// The co-located API servers are reached on localhost only.
				"scheduler": {},
				"kcm":       {},
				"nodes":     {},
			},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			b := &KopsModelContext{
				AllInstanceGroups: g.groups,
				InstanceGroups:    g.groups,
			}

			for _, instanceGroup := range g.groups {
				want, ok := g.want[instanceGroup.Name]
				if !ok {
					t.Fatalf("test case does not cover instance group %q", instanceGroup.Name)
				}

				for _, service := range kops.AllWellKnownServices {
					expected := false
					for _, w := range want {
						if w == service {
							expected = true
						}
					}
					got := b.ServesWellKnownService(instanceGroup, service)
					if got != expected {
						t.Errorf("%s: ServesWellKnownService(%q) = %v, want %v",
							instanceGroup.Name, service, got, expected)
					}
				}
			}
		})
	}
}
