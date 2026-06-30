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

package gcemodel

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// makeNetworkBuilder constructs a minimal NetworkModelBuilder for the given cluster.
// The cluster must have at least one private subnet to trigger router creation.
func makeNetworkBuilder(cluster *kops.Cluster) *NetworkModelBuilder {
	return &NetworkModelBuilder{
		GCEModelContext: &GCEModelContext{
			ProjectID: "test-project",
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext: iam.IAMModelContext{Cluster: cluster},
				Region:          "us-central1",
			},
		},
		Lifecycle: fi.LifecycleSync,
	}
}

// makeCluster builds a minimal cluster spec with one unshared private subnet.
// nonMasqueradeCIDR controls whether IsIPv6Only() returns true:
//   - IPv6 CIDR (e.g. "fd00::/56") → IPv6-only cluster
//   - IPv4 CIDR (e.g. "100.64.0.0/10") → normal cluster
func makeCluster(nonMasqueradeCIDR string) *kops.Cluster {
	return &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "test.k8s.local"},
		Spec: kops.ClusterSpec{
			Networking: kops.NetworkingSpec{
				NonMasqueradeCIDR: nonMasqueradeCIDR,
				// One unshared private subnet (no ID, no external egress) so the
				// CloudNAT router task is always created.
				Subnets: []kops.ClusterSubnetSpec{
					{
						Name: "private-us-central1-a",
						Type: kops.SubnetTypePrivate,
					},
				},
			},
		},
	}
}

// findRouterTask returns the single gcetasks.Router in the task map, or nil.
func findRouterTask(tasks map[string]fi.CloudupTask) *gcetasks.Router {
	for _, task := range tasks {
		if r, ok := task.(*gcetasks.Router); ok {
			return r
		}
	}
	return nil
}

func TestNetworkModelBuilder_RouterNAT64(t *testing.T) {
	tests := []struct {
		name              string
		nonMasqueradeCIDR string
		wantNAT64         *string // nil means the field must be nil
	}{
		{
			name:              "IPv6-only cluster sets NAT64",
			nonMasqueradeCIDR: "fd00::/56",
			wantNAT64:         fi.PtrTo(gcetasks.SourceSubnetworkIPRangesAllIPv6),
		},
		{
			name:              "non-IPv6 cluster leaves NAT64 nil",
			nonMasqueradeCIDR: "100.64.0.0/10",
			wantNAT64:         nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cluster := makeCluster(tc.nonMasqueradeCIDR)
			b := makeNetworkBuilder(cluster)
			ctx := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

			if err := b.Build(ctx); err != nil {
				t.Fatalf("Build() returned error: %v", err)
			}

			router := findRouterTask(ctx.Tasks)
			if router == nil {
				t.Fatalf("no Router task found in %d tasks; private subnet should have triggered CloudNAT router creation", len(ctx.Tasks))
			}

			got := router.SourceSubnetworkIPRangesToNAT64
			switch {
			case tc.wantNAT64 == nil && got != nil:
				t.Errorf("SourceSubnetworkIPRangesToNAT64: got %q, want nil", fi.ValueOf(got))
			case tc.wantNAT64 != nil && got == nil:
				t.Errorf("SourceSubnetworkIPRangesToNAT64: got nil, want %q", fi.ValueOf(tc.wantNAT64))
			case tc.wantNAT64 != nil && got != nil && *got != *tc.wantNAT64:
				t.Errorf("SourceSubnetworkIPRangesToNAT64: got %q, want %q", *got, *tc.wantNAT64)
			}
		})
	}
}
