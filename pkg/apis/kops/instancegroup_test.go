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

package kops

import (
	"reflect"
	"slices"
	"testing"
)

func TestInstanceGroupRoleRoles(t *testing.T) {
	grid := []struct {
		role InstanceGroupRole
		want []InstanceGroupRole
	}{
		{role: "", want: nil},
		{role: InstanceGroupRoleNode, want: []InstanceGroupRole{InstanceGroupRoleNode}},
		{
			role: "APIServer,Scheduler",
			want: []InstanceGroupRole{InstanceGroupRoleAPIServer, InstanceGroupRoleScheduler},
		},
		{
			// Canonical ordering: input order must not matter.
			role: "Scheduler,APIServer",
			want: []InstanceGroupRole{InstanceGroupRoleAPIServer, InstanceGroupRoleScheduler},
		},
		{
			// Whitespace around separators is tolerated.
			role: " APIServer , KubeControllerManager ",
			want: []InstanceGroupRole{InstanceGroupRoleAPIServer, InstanceGroupRoleKubeControllerManager},
		},
		{
			// Duplicates collapse.
			role: "Etcd,Etcd",
			want: []InstanceGroupRole{InstanceGroupRoleEtcd},
		},
		{
			// Empty fields are dropped.
			role: "Etcd,,Scheduler,",
			want: []InstanceGroupRole{InstanceGroupRoleEtcd, InstanceGroupRoleScheduler},
		},
		{
			role: "ControlPlane,APIServer,Etcd,Scheduler,KubeControllerManager",
			want: []InstanceGroupRole{
				InstanceGroupRoleControlPlane,
				InstanceGroupRoleAPIServer,
				InstanceGroupRoleEtcd,
				InstanceGroupRoleScheduler,
				InstanceGroupRoleKubeControllerManager,
			},
		},
		{
			// Unrecognised roles are preserved, after the recognised ones, so that validation
			// can report them rather than silently dropping them.
			role: "Nonsense,APIServer",
			want: []InstanceGroupRole{InstanceGroupRoleAPIServer, "Nonsense"},
		},
		{
			role: "Nonsense",
			want: []InstanceGroupRole{"Nonsense"},
		},
	}

	for _, g := range grid {
		got := g.role.Roles()
		if !reflect.DeepEqual(got, g.want) {
			t.Errorf("InstanceGroupRole(%q).Roles() = %v, want %v", g.role, got, g.want)
		}
	}
}

func TestInstanceGroupRoleHasRole(t *testing.T) {
	grid := []struct {
		role InstanceGroupRole
		want map[InstanceGroupRole]bool
	}{
		{
			role: InstanceGroupRoleControlPlane,
			want: map[InstanceGroupRole]bool{InstanceGroupRoleControlPlane: true},
		},
		{
			role: "APIServer,Scheduler",
			want: map[InstanceGroupRole]bool{
				InstanceGroupRoleAPIServer: true,
				InstanceGroupRoleScheduler: true,
			},
		},
		{
			role: "APIServer,KubeControllerManager",
			want: map[InstanceGroupRole]bool{
				InstanceGroupRoleAPIServer:             true,
				InstanceGroupRoleKubeControllerManager: true,
			},
		},
		{
			role: "",
			want: map[InstanceGroupRole]bool{},
		},
	}

	for _, g := range grid {
		for _, role := range AllInstanceGroupRoles {
			if got := g.role.HasRole(role); got != g.want[role] {
				t.Errorf("InstanceGroupRole(%q).HasRole(%q) = %v, want %v", g.role, role, got, g.want[role])
			}
		}
	}
}

// TestInstanceGroupRoleHasHelpers checks the named helpers agree with HasRole, in particular
// that a composite role reports true for every role it carries.
func TestInstanceGroupRoleHasHelpers(t *testing.T) {
	grid := []struct {
		role        InstanceGroupRole
		controlPlan bool
		node        bool
		bastion     bool
		apiServer   bool
		etcd        bool
		scheduler   bool
		kcm         bool
	}{
		{role: InstanceGroupRoleControlPlane, controlPlan: true},
		{role: InstanceGroupRoleNode, node: true},
		{role: InstanceGroupRoleBastion, bastion: true},
		{role: InstanceGroupRoleAPIServer, apiServer: true},
		{role: InstanceGroupRoleEtcd, etcd: true},
		{role: InstanceGroupRoleScheduler, scheduler: true},
		{role: InstanceGroupRoleKubeControllerManager, kcm: true},
		{role: "APIServer,Scheduler", apiServer: true, scheduler: true},
		{role: "APIServer,KubeControllerManager", apiServer: true, kcm: true},
		{role: "APIServer,Etcd", apiServer: true, etcd: true},
	}

	for _, g := range grid {
		if got := g.role.HasControlPlane(); got != g.controlPlan {
			t.Errorf("%q.HasControlPlane() = %v, want %v", g.role, got, g.controlPlan)
		}
		if got := g.role.HasNode(); got != g.node {
			t.Errorf("%q.HasNode() = %v, want %v", g.role, got, g.node)
		}
		if got := g.role.HasBastion(); got != g.bastion {
			t.Errorf("%q.HasBastion() = %v, want %v", g.role, got, g.bastion)
		}
		if got := g.role.HasAPIServer(); got != g.apiServer {
			t.Errorf("%q.HasAPIServer() = %v, want %v", g.role, got, g.apiServer)
		}
		if got := g.role.HasEtcd(); got != g.etcd {
			t.Errorf("%q.HasEtcd() = %v, want %v", g.role, got, g.etcd)
		}
		if got := g.role.HasScheduler(); got != g.scheduler {
			t.Errorf("%q.HasScheduler() = %v, want %v", g.role, got, g.scheduler)
		}
		if got := g.role.HasKubeControllerManager(); got != g.kcm {
			t.Errorf("%q.HasKubeControllerManager() = %v, want %v", g.role, got, g.kcm)
		}
	}
}

func TestInstanceGroupRolePrimaryRole(t *testing.T) {
	grid := []struct {
		role InstanceGroupRole
		want InstanceGroupRole
	}{
		{role: "", want: ""},
		{role: InstanceGroupRoleNode, want: InstanceGroupRoleNode},
		{role: "APIServer,Scheduler", want: InstanceGroupRoleAPIServer},
		// Order of the input must not change the answer.
		{role: "Scheduler,APIServer", want: InstanceGroupRoleAPIServer},
		{role: "APIServer,KubeControllerManager", want: InstanceGroupRoleAPIServer},
		{role: "KubeControllerManager,APIServer", want: InstanceGroupRoleAPIServer},
		{role: "Scheduler,KubeControllerManager", want: InstanceGroupRoleScheduler},
		{role: "ControlPlane,Etcd", want: InstanceGroupRoleControlPlane},
		{role: "Etcd,ControlPlane", want: InstanceGroupRoleControlPlane},
	}

	for _, g := range grid {
		if got := g.role.PrimaryRole(); got != g.want {
			t.Errorf("InstanceGroupRole(%q).PrimaryRole() = %q, want %q", g.role, got, g.want)
		}
	}
}

// TestInstanceGroupRoleToLowerString pins the single-role spellings, which feed GCE label keys
// and values, cloud resource names and the igconfig state-store path. Changing any of them
// would silently move existing clusters' config to a different path.
func TestInstanceGroupRoleToLowerString(t *testing.T) {
	grid := []struct {
		role InstanceGroupRole
		want string
	}{
		{role: "", want: ""},
		{role: InstanceGroupRoleControlPlane, want: "control-plane"},
		{role: InstanceGroupRoleNode, want: "node"},
		{role: InstanceGroupRoleBastion, want: "bastion"},
		{role: InstanceGroupRoleAPIServer, want: "apiserver"},
		{role: InstanceGroupRoleEtcd, want: "etcd"},
		{role: InstanceGroupRoleScheduler, want: "scheduler"},
		{role: InstanceGroupRoleKubeControllerManager, want: "kubecontrollermanager"},
		// Composite values join with an underscore, in canonical order. A hyphen would be
		// ambiguous, because "control-plane" already contains one.
		{role: "APIServer,Scheduler", want: "apiserver_scheduler"},
		{role: "Scheduler,APIServer", want: "apiserver_scheduler"},
		{role: "APIServer,KubeControllerManager", want: "apiserver_kubecontrollermanager"},
		{role: "ControlPlane,Etcd", want: "control-plane_etcd"},
	}

	for _, g := range grid {
		if got := g.role.ToLowerString(); got != g.want {
			t.Errorf("InstanceGroupRole(%q).ToLowerString() = %q, want %q", g.role, got, g.want)
		}
	}
}

// TestInstanceGroupRoleToLowerStringRoundTrip checks every single role's lowercase form parses
// back to the same role, since ParseInstanceGroupRole backs the CLI's role filters.
func TestInstanceGroupRoleToLowerStringRoundTrip(t *testing.T) {
	for _, role := range AllInstanceGroupRoles {
		s := role.ToLowerString()
		got, ok := ParseInstanceGroupRole(s, false)
		if !ok {
			t.Errorf("ParseInstanceGroupRole(%q) failed for role %q", s, role)
			continue
		}
		if got != role {
			t.Errorf("ParseInstanceGroupRole(%q) = %q, want %q", s, got, role)
		}
	}
}

func TestInstanceGroupRoleIsControlPlaneType(t *testing.T) {
	grid := []struct {
		role InstanceGroupRole
		want bool
	}{
		{role: "", want: false},
		{role: InstanceGroupRoleNode, want: false},
		{role: InstanceGroupRoleBastion, want: false},
		{role: InstanceGroupRoleControlPlane, want: true},
		{role: InstanceGroupRoleAPIServer, want: true},
		{role: InstanceGroupRoleEtcd, want: true},
		{role: InstanceGroupRoleScheduler, want: true},
		{role: InstanceGroupRoleKubeControllerManager, want: true},
		{role: "APIServer,Scheduler", want: true},
		{role: "APIServer,KubeControllerManager", want: true},
	}

	for _, g := range grid {
		if got := g.role.IsControlPlaneType(); got != g.want {
			t.Errorf("InstanceGroupRole(%q).IsControlPlaneType() = %v, want %v", g.role, got, g.want)
		}
	}
}

// TestInstanceGroupIsRoleOnlyVsRuns is the contract that makes composite roles safe: Is*Only
// asks whether a group is dedicated to one role, Runs* asks whether it runs that component at
// all. Conflating the two is what made the pre-composite helpers misleading, since IsAPIServerOnly
// was implemented as a plain "has the APIServer role" check.
func TestInstanceGroupIsRoleOnlyVsRuns(t *testing.T) {
	grid := []struct {
		role InstanceGroupRole

		apiServerOnly bool
		etcdOnly      bool
		schedulerOnly bool
		kcmOnly       bool

		runsAPIServer bool
		runsEtcd      bool
		runsScheduler bool
		runsKCM       bool
	}{
		{
			role:          InstanceGroupRoleAPIServer,
			apiServerOnly: true,
			runsAPIServer: true,
		},
		{
			role:     InstanceGroupRoleEtcd,
			etcdOnly: true,
			runsEtcd: true,
		},
		{
			role:          InstanceGroupRoleScheduler,
			schedulerOnly: true,
			runsScheduler: true,
		},
		{
			role:    InstanceGroupRoleKubeControllerManager,
			kcmOnly: true,
			runsKCM: true,
		},
		{
			// A full control-plane group runs every component but is dedicated to none of them.
			role:          InstanceGroupRoleControlPlane,
			runsAPIServer: true,
			runsEtcd:      true,
			runsScheduler: true,
			runsKCM:       true,
		},
		{
			// The case composite roles exist for: a scheduler and the API server it talks to.
			// It runs both, and is dedicated to neither.
			role:          "APIServer,Scheduler",
			runsAPIServer: true,
			runsScheduler: true,
		},
		{
			role:          "APIServer,KubeControllerManager",
			runsAPIServer: true,
			runsKCM:       true,
		},
		{
			role:          "APIServer,Etcd",
			runsAPIServer: true,
			runsEtcd:      true,
		},
		{
			role: InstanceGroupRoleNode,
		},
	}

	for _, g := range grid {
		ig := &InstanceGroup{Spec: InstanceGroupSpec{Role: g.role}}

		for _, tc := range []struct {
			name string
			got  bool
			want bool
		}{
			{"IsAPIServerOnly", ig.IsAPIServerOnly(), g.apiServerOnly},
			{"IsEtcdOnly", ig.IsEtcdOnly(), g.etcdOnly},
			{"IsSchedulerOnly", ig.IsSchedulerOnly(), g.schedulerOnly},
			{"IsKubeControllerManagerOnly", ig.IsKubeControllerManagerOnly(), g.kcmOnly},
			{"RunsAPIServer", ig.RunsAPIServer(), g.runsAPIServer},
			{"RunsEtcd", ig.RunsEtcd(), g.runsEtcd},
			{"RunsScheduler", ig.RunsScheduler(), g.runsScheduler},
			{"RunsKubeControllerManager", ig.RunsKubeControllerManager(), g.runsKCM},
		} {
			if tc.got != tc.want {
				t.Errorf("role %q: %s() = %v, want %v", g.role, tc.name, tc.got, tc.want)
			}
		}
	}
}

func TestInstanceGroupIsRoleOnly(t *testing.T) {
	grid := []struct {
		role  InstanceGroupRole
		query InstanceGroupRole
		want  bool
	}{
		{role: InstanceGroupRoleAPIServer, query: InstanceGroupRoleAPIServer, want: true},
		{role: InstanceGroupRoleAPIServer, query: InstanceGroupRoleScheduler, want: false},
		{role: "APIServer,Scheduler", query: InstanceGroupRoleAPIServer, want: false},
		{role: "APIServer,Scheduler", query: InstanceGroupRoleScheduler, want: false},
		// Duplicates still describe a group dedicated to one role.
		{role: "Etcd,Etcd", query: InstanceGroupRoleEtcd, want: true},
		{role: "", query: InstanceGroupRoleNode, want: false},
	}

	for _, g := range grid {
		ig := &InstanceGroup{Spec: InstanceGroupSpec{Role: g.role}}
		if got := ig.IsRoleOnly(g.query); got != g.want {
			t.Errorf("role %q: IsRoleOnly(%q) = %v, want %v", g.role, g.query, got, g.want)
		}
	}
}

// TestServedWellKnownServices pins the role-derived defaults. The first four cases are the
// topologies that existed before the field, and must keep behaving identically.
func TestServedWellKnownServices(t *testing.T) {
	grid := []struct {
		name string
		role InstanceGroupRole
		set  []WellKnownService
		want []WellKnownService
	}{
		{
			name: "full control plane serves everything",
			role: InstanceGroupRoleControlPlane,
			want: []WellKnownService{
				WellKnownServiceKubeAPIServerExternal,
				WellKnownServiceKubeAPIServerInternal,
				WellKnownServiceKopsController,
				WellKnownServiceEtcdMain,
			},
		},
		{
			name: "dedicated API server fronts both endpoints",
			role: InstanceGroupRoleAPIServer,
			want: []WellKnownService{
				WellKnownServiceKubeAPIServerExternal,
				WellKnownServiceKubeAPIServerInternal,
				WellKnownServiceKopsController,
			},
		},
		{
			name: "dedicated etcd serves only etcd",
			role: InstanceGroupRoleEtcd,
			want: []WellKnownService{WellKnownServiceEtcdMain},
		},
		{
			name: "node serves nothing",
			role: InstanceGroupRoleNode,
			want: nil,
		},
		{
			name: "bastion serves nothing",
			role: InstanceGroupRoleBastion,
			want: nil,
		},
		{
			name: "scheduler alone serves nothing",
			role: InstanceGroupRoleScheduler,
			want: nil,
		},
		{
			name: "kube-controller-manager alone serves nothing",
			role: InstanceGroupRoleKubeControllerManager,
			want: nil,
		},
		{
			// The reason the derived default matters: this API server exists for the scheduler
			// beside it, so nothing should be routed to it.
			name: "API server co-located with the scheduler is local only",
			role: "APIServer,Scheduler",
			want: nil,
		},
		{
			name: "API server co-located with kube-controller-manager is local only",
			role: "APIServer,KubeControllerManager",
			want: nil,
		},
		{
			// etcd beside an API server still has to be reachable by the other API servers.
			name: "API server co-located with etcd serves both",
			role: "APIServer,Etcd",
			want: []WellKnownService{
				WellKnownServiceKubeAPIServerExternal,
				WellKnownServiceKubeAPIServerInternal,
				WellKnownServiceKopsController,
				WellKnownServiceEtcdMain,
			},
		},
		{
			name: "explicit value wins over the derived default",
			role: InstanceGroupRoleAPIServer,
			set:  []WellKnownService{WellKnownServiceKubeAPIServerInternal},
			want: []WellKnownService{WellKnownServiceKubeAPIServerInternal},
		},
		{
			name: "explicitly empty means serves nothing",
			role: InstanceGroupRoleAPIServer,
			set:  []WellKnownService{},
			want: []WellKnownService{},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			ig := &InstanceGroup{Spec: InstanceGroupSpec{
				Role:                    g.role,
				ServesWellKnownServices: g.set,
			}}

			got := ig.ServedWellKnownServices()
			if !reflect.DeepEqual(got, g.want) {
				t.Errorf("role %q: ServedWellKnownServices() = %v, want %v", g.role, got, g.want)
			}

			// ServesWellKnownService must agree with the list.
			for _, service := range AllWellKnownServices {
				want := slices.Contains(g.want, service)
				if got := ig.ServesWellKnownService(service); got != want {
					t.Errorf("role %q: ServesWellKnownService(%q) = %v, want %v", g.role, service, got, want)
				}
			}
		})
	}
}
