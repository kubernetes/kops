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

package karpenter

import (
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
)

func TestBuildAWSPolicyCustomerManagedKeys(t *testing.T) {
	const key = "arn:aws:kms:us-test-1:123456789012:key/test"

	grid := []struct {
		desc     string
		spec     kops.InstanceGroupSpec
		expected bool
	}{
		{
			desc: "no encryption keys",
			spec: kops.InstanceGroupSpec{
				Volumes: []kops.VolumeSpec{{Device: "/dev/sdf", Size: 10}},
			},
		},
		{
			desc: "volume key",
			spec: kops.InstanceGroupSpec{
				Volumes: []kops.VolumeSpec{{Device: "/dev/sdf", Size: 10, Key: new(key)}},
			},
			expected: true,
		},
		{
			desc: "volume key on an instance group not managed by Karpenter",
			spec: kops.InstanceGroupSpec{
				Manager: kops.InstanceManagerCloudGroup,
				Volumes: []kops.VolumeSpec{{Device: "/dev/sdf", Size: 10, Key: new(key)}},
			},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			cluster := &kops.Cluster{}
			cluster.Name = "minimal.example.com"
			ig := &kops.InstanceGroup{Spec: g.spec}
			ig.Name = "nodes"
			ig.Spec.Role = kops.InstanceGroupRoleNode
			if ig.Spec.Manager == "" {
				ig.Spec.Manager = kops.InstanceManagerKarpenter
			}
			b := &iam.PolicyBuilder{
				Cluster:           cluster,
				AllInstanceGroups: []*kops.InstanceGroup{ig},
				Partition:         "aws",
				Region:            "us-test-1",
			}

			policy, err := (&ServiceAccount{}).BuildAWSPolicy(b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			policyJSON, err := policy.AsJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := strings.Contains(policyJSON, "kms:"); actual != g.expected {
				t.Errorf("expected KMS permissions %v, got %v in %s", g.expected, actual, policyJSON)
			}
		})
	}
}
