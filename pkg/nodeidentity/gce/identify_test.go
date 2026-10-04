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

package gce

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// TestNodeLabelsForTags covers the mapping kops-controller uses to label nodes on GCE. The role
// labels BuildNodeLabels computes are applied here, not by the kubelet, so a role missing from
// this mapping is a role whose label never reaches the node.
func TestNodeLabelsForTags(t *testing.T) {
	const clusterName = "testcluster.example.com"

	tag := func(role kops.InstanceGroupRole) string {
		return gce.TagForRole(clusterName, role)
	}

	grid := []struct {
		name string
		tags []string
		want map[string]string
	}{
		{
			name: "control plane",
			tags: []string{tag(kops.InstanceGroupRoleControlPlane)},
			want: map[string]string{nodelabels.RoleLabelControlPlane20: ""},
		},
		{
			name: "node",
			tags: []string{tag(kops.InstanceGroupRoleNode)},
			want: map[string]string{nodelabels.RoleLabelNode16: ""},
		},
		{
			name: "dedicated API server",
			tags: []string{tag(kops.InstanceGroupRoleAPIServer)},
			want: map[string]string{nodelabels.RoleLabelAPIServer16: ""},
		},
		{
			name: "dedicated etcd",
			tags: []string{tag(kops.InstanceGroupRoleEtcd)},
			want: map[string]string{nodelabels.RoleLabelEtcd: ""},
		},
		{
			name: "dedicated scheduler",
			tags: []string{tag(kops.InstanceGroupRoleScheduler)},
			want: map[string]string{nodelabels.RoleLabelScheduler: ""},
		},
		{
			name: "dedicated kube-controller-manager",
			tags: []string{tag(kops.InstanceGroupRoleKubeControllerManager)},
			want: map[string]string{nodelabels.RoleLabelKubeControllerManager: ""},
		},
		{
			// An instance group carrying two roles has a tag for each, and should end up
			// labelled for both.
			name: "scheduler alongside an API server",
			tags: []string{tag(kops.InstanceGroupRoleScheduler), tag(kops.InstanceGroupRoleAPIServer)},
			want: map[string]string{
				nodelabels.RoleLabelScheduler:   "",
				nodelabels.RoleLabelAPIServer16: "",
			},
		},
		{
			name: "kube-controller-manager alongside an API server",
			tags: []string{tag(kops.InstanceGroupRoleKubeControllerManager), tag(kops.InstanceGroupRoleAPIServer)},
			want: map[string]string{
				nodelabels.RoleLabelKubeControllerManager: "",
				nodelabels.RoleLabelAPIServer16:           "",
			},
		},
		{
			name: "bastions are not cluster members",
			tags: []string{tag(kops.InstanceGroupRoleBastion)},
			want: map[string]string{},
		},
		{
			name: "tags from another cluster are ignored",
			tags: []string{gce.TagForRole("other.example.com", kops.InstanceGroupRoleControlPlane)},
			want: map[string]string{},
		},
		{
			name: "unrelated tags are ignored",
			tags: []string{"some-other-tag"},
			want: map[string]string{},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			got := nodeLabelsForTags(clusterName, g.tags, "selfLink")
			if !reflect.DeepEqual(got, g.want) {
				t.Errorf("nodeLabelsForTags(%v) = %v, want %v", g.tags, got, g.want)
			}
		})
	}
}

// TestNodeLabelsForTagsCoversEveryRole makes sure no role is silently unlabelled. Adding a role
// without extending the mapping would leave its nodes without a role label, which is how the
// split control plane roles were missed.
func TestNodeLabelsForTagsCoversEveryRole(t *testing.T) {
	const clusterName = "testcluster.example.com"

	for _, role := range kops.AllInstanceGroupRoles {
		if role.HasBastion() {
			// Bastions are deliberately unlabelled; they are not cluster members.
			continue
		}
		labels := nodeLabelsForTags(clusterName, []string{gce.TagForRole(clusterName, role)}, "selfLink")
		if len(labels) == 0 {
			t.Errorf("role %q produces no node label", role)
		}
	}
}
