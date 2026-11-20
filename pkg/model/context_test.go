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
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func TestCloudTagsForInstanceGroup_Taints(t *testing.T) {
	const clusterName = "testcluster.example.com"

	// clusterFor builds a minimal cluster for providers without a testutils builder.
	clusterFor := func(cloudProvider kops.CloudProviderSpec) *kops.Cluster {
		c := testutils.BuildMinimalClusterAWS(clusterName)
		c.Spec.CloudProvider = cloudProvider
		return c
	}

	grid := []struct {
		name    string
		cluster *kops.Cluster
		taints  []string
		// wantTags must be set. No other cluster-autoscaler taint tag may be set.
		wantTags map[string]string
	}{
		{
			name:    "taint with value and effect",
			cluster: testutils.BuildMinimalClusterAWS(clusterName),
			taints:  []string{"foo=bar:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": "bar:NoSchedule",
			},
		},
		{
			name:    "taint without value (key:effect)",
			cluster: testutils.BuildMinimalClusterAWS(clusterName),
			taints:  []string{"foo:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": ":NoSchedule",
			},
		},
		{
			name:    "mix of taint formats",
			cluster: testutils.BuildMinimalClusterAWS(clusterName),
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
			cluster:  testutils.BuildMinimalClusterAWS(clusterName),
			taints:   []string{"foo"},
			wantTags: map[string]string{},
		},
		{
			name:    "openstack keeps the taint tags",
			cluster: clusterFor(kops.CloudProviderSpec{Openstack: &kops.OpenstackSpec{}}),
			taints:  []string{"nvidia.com/gpu=present:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/nvidia.com/gpu": "present:NoSchedule",
				nodeidentityaws.CloudTagInstanceGroupName:                      "nodes",
			},
		},
		{
			name:    "gce gets no taint tags",
			cluster: testutils.BuildMinimalClusterGCE(clusterName, "testproject"),
			taints:  []string{"nvidia.com/gpu=present:NoSchedule"},
			wantTags: map[string]string{
				gce.GceLabelNameInstanceGroup: "nodes",
			},
		},
		{
			name:    "hetzner gets no taint tags",
			cluster: clusterFor(kops.CloudProviderSpec{Hetzner: &kops.HetznerSpec{}}),
			taints:  []string{"nvidia.com/gpu=present:NoSchedule"},
			wantTags: map[string]string{
				hetzner.TagKubernetesInstanceGroup: "nodes",
			},
		},
		{
			name:    "linode gets no taint tags",
			cluster: clusterFor(kops.CloudProviderSpec{Linode: &kops.LinodeSpec{}}),
			taints:  []string{"nvidia.com/gpu=present:NoSchedule"},
			wantTags: map[string]string{
				linode.TagKubernetesInstanceGroup: "nodes",
			},
		},
		{
			// Providers have to be listed to get the taint tags.
			name:    "other providers get no taint tags",
			cluster: clusterFor(kops.CloudProviderSpec{DO: &kops.DOSpec{}}),
			taints:  []string{"nvidia.com/gpu=present:NoSchedule"},
			wantTags: map[string]string{
				nodeidentityaws.CloudTagInstanceGroupName: "nodes",
			},
		},
	}

	for _, tc := range grid {
		t.Run(tc.name, func(t *testing.T) {
			ig := &kops.InstanceGroup{}
			ig.ObjectMeta.Name = "nodes"
			ig.Spec.Role = kops.InstanceGroupRoleNode
			ig.Spec.Taints = tc.taints

			b := &KopsModelContext{
				IAMModelContext:   iam.IAMModelContext{Cluster: tc.cluster},
				AllInstanceGroups: []*kops.InstanceGroup{ig},
				InstanceGroups:    []*kops.InstanceGroup{ig},
			}

			tags, err := b.CloudTagsForInstanceGroup(ig)
			if err != nil {
				t.Fatalf("CloudTagsForInstanceGroup() error = %v", err)
			}

			for k, want := range tc.wantTags {
				if got, ok := tags[k]; !ok || got != want {
					t.Errorf("tag %q = %q, want %q", k, got, want)
				}
			}
			for k, got := range tags {
				if !strings.HasPrefix(k, clusterAutoscalerNodeTemplateTaint) {
					continue
				}
				if _, ok := tc.wantTags[k]; !ok {
					t.Errorf("unexpected cluster-autoscaler taint tag %q = %q", k, got)
				}
			}
		})
	}
}

func TestUsesSSHBastionWhenUpdatingOtherInstanceGroups(t *testing.T) {
	bastions := &kops.InstanceGroup{}
	bastions.ObjectMeta.Name = "bastions"
	bastions.Spec.Role = kops.InstanceGroupRoleBastion
	nodes := &kops.InstanceGroup{}
	nodes.ObjectMeta.Name = "nodes"
	nodes.Spec.Role = kops.InstanceGroupRoleNode

	// Without the bastions, the SSH access rules would allow SSH directly to the instances.
	b := &KopsModelContext{
		AllInstanceGroups: []*kops.InstanceGroup{bastions, nodes},
		InstanceGroups:    []*kops.InstanceGroup{nodes},
	}
	if !b.UsesSSHBastion() {
		t.Errorf("UsesSSHBastion() = false when updating only the nodes, want true")
	}
}
