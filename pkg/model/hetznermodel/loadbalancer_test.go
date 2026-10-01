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

package hetznermodel

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

func TestLoadBalancerLocationWhenUpdatingSomeInstanceGroups(t *testing.T) {
	controlPlane := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "control-plane-fsn1"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane, Subnets: []string{"fsn1"}},
	}
	nodes := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-nbg1"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode, Subnets: []string{"nbg1"}},
	}
	bastions := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "bastions"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleBastion, Subnets: []string{"nbg1"}},
	}

	grid := []struct {
		desc           string
		instanceGroups []*kops.InstanceGroup
	}{
		{
			desc:           "all instance groups",
			instanceGroups: []*kops.InstanceGroup{controlPlane, nodes, bastions},
		},
		{
			desc:           "nodes in another location",
			instanceGroups: []*kops.InstanceGroup{nodes},
		},
		{
			desc: "no instance groups",
		},
	}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			b := &LoadBalancerModelBuilder{
				HetznerModelContext: &HetznerModelContext{
					KopsModelContext: &model.KopsModelContext{
						IAMModelContext: iam.IAMModelContext{
							Cluster: &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "minimal.example.com"}},
						},
						AllInstanceGroups: []*kops.InstanceGroup{controlPlane, nodes, bastions},
						InstanceGroups:    g.instanceGroups,
					},
				},
			}
			c := &fi.CloudupModelBuilderContext{
				Tasks: make(map[string]fi.CloudupTask),
			}
			if err := b.Build(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			lb, ok := c.Tasks["LoadBalancer/api.minimal.example.com"].(*hetznertasks.LoadBalancer)
			if !ok {
				t.Fatalf("load balancer task not found")
			}
			if lb.Location != "fsn1" {
				t.Errorf("expected load balancer location %q, but got %q", "fsn1", lb.Location)
			}
		})
	}
}
