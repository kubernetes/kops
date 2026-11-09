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
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
)

func TestUseIPv6ForBastionWhenUpdatingOnlyBastions(t *testing.T) {
	cluster := testutils.BuildMinimalClusterAWS("minimal.example.com")
	cluster.Spec.Networking.Subnets = []kops.ClusterSubnetSpec{
		{Name: "utility-us-test-1a", Zone: "us-test-1a", CIDR: "172.20.4.0/22", Type: kops.SubnetTypeUtility},
		{Name: "us-test-1a", Zone: "us-test-1a", CIDR: "172.20.32.0/19", IPv6CIDR: "/64#1", Type: kops.SubnetTypePrivate},
	}
	bastions := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "bastions"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleBastion, Subnets: []string{"utility-us-test-1a"}},
	}
	nodes := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode, Subnets: []string{"us-test-1a"}},
	}

	// The bastion load balancer is dual-stack when any instance group has IPv6.
	b := &BastionModelBuilder{
		AWSModelContext: &AWSModelContext{
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
				AllInstanceGroups: []*kops.InstanceGroup{bastions, nodes},
				InstanceGroups:    []*kops.InstanceGroup{bastions},
			},
		},
	}
	if !useIPv6ForBastion(b) {
		t.Errorf("expected the bastion load balancer to use IPv6, because the nodes have IPv6")
	}
}
