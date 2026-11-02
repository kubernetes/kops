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

package awsmodel

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
)

func TestChooseBestSubnetForELBWhenUpdatingOnlyNodes(t *testing.T) {
	subnets := []*kops.ClusterSubnetSpec{
		{Name: "a-nodes", Zone: "us-test-1a", Type: kops.SubnetTypePrivate},
		{Name: "b-control-plane", Zone: "us-test-1a", Type: kops.SubnetTypePrivate},
	}
	controlPlane := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "control-plane"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane, Subnets: []string{"b-control-plane"}},
	}
	nodes := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode, Subnets: []string{"a-nodes"}},
	}

	b := &APILoadBalancerBuilder{
		AWSModelContext: &AWSModelContext{
			KopsModelContext: &model.KopsModelContext{
				AllInstanceGroups: []*kops.InstanceGroup{controlPlane, nodes},
				InstanceGroups:    []*kops.InstanceGroup{nodes},
			},
		},
	}
	if subnet := b.chooseBestSubnetForELB("us-test-1a", subnets); subnet.Name != "b-control-plane" {
		t.Errorf("expected the subnet of the control plane %q, but got %q", "b-control-plane", subnet.Name)
	}
}
