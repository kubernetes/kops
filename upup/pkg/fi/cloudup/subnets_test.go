/*
Copyright 2019 The Kubernetes Authors.

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

package cloudup

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func Test_AssignSubnets(t *testing.T) {
	tests := []struct {
		subnets  []kops.ClusterSubnetSpec
		expected []string
	}{
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "10.1.0.0/16", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.1.0.0/16"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.32.0.0/11"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.32.0.0/11", "10.64.0.0/11"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.64.0.0/11", "10.32.0.0/11"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "10.64.0.0/11", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.64.0.0/11", "10.32.0.0/11"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "10.0.0.0/9", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.0.0.0/9", "10.160.0.0/11"},
		},

		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypeUtility},
			},
			expected: []string{"10.32.0.0/11", "10.0.0.0/14"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypeUtility},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePrivate},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypeUtility},
				{Name: "c", Zone: "c", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "c", Zone: "c", CIDR: "", Type: kops.SubnetTypePrivate},
				{Name: "c", Zone: "c", CIDR: "", Type: kops.SubnetTypeUtility},
			},
			expected: []string{
				"10.32.0.0/11", "10.0.0.0/14",
				"10.64.0.0/11", "10.96.0.0/11", "10.4.0.0/14",
				"10.128.0.0/11", "10.160.0.0/11", "10.8.0.0/14",
			},
		},
	}
	for i, test := range tests {
		c := &kops.Cluster{}
		c.Spec.NetworkCIDR = "10.0.0.0/8"
		c.Spec.Subnets = test.subnets

		err := assignCIDRsToSubnets(c)
		if err != nil {
			t.Fatalf("unexpected error on test %d: %v", i+1, err)
		}

		var actual []string
		for _, subnet := range c.Spec.Subnets {
			actual = append(actual, subnet.CIDR)
		}
		if !reflect.DeepEqual(actual, test.expected) {
			t.Fatalf("unexpected result of network allocation (#%d): actual=%v, expected=%v", i+1, actual, test.expected)
		}
	}
}
