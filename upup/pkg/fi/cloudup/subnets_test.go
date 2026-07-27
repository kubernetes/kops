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
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
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
			expected: []string{"10.0.0.0/8"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.0.0.0/9", "10.128.0.0/9"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.128.0.0/9", "10.0.0.0/9"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "10.64.0.0/11", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.64.0.0/11", "10.128.0.0/9"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "10.0.0.0/9", Type: kops.SubnetTypePublic},
				{Name: "b", Zone: "b", CIDR: "", Type: kops.SubnetTypePublic},
			},
			expected: []string{"10.0.0.0/9", "10.128.0.0/9"},
		},

		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePublic},
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypeUtility},
			},
			expected: []string{"10.128.0.0/9", "10.0.0.0/12"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", IPv6CIDR: "/64#0", Type: kops.SubnetTypePublic},
				{Name: "a", Zone: "a", CIDR: "", IPv6CIDR: "/64#1", Type: kops.SubnetTypeUtility},
			},
			expected: []string{"10.128.0.0/9", "10.0.0.0/12"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "10.1.0.0/16", Type: kops.SubnetTypePrivate},
			},
			expected: []string{"10.1.0.0/16"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", Type: kops.SubnetTypePrivate},
			},
			expected: []string{"10.0.0.0/8"},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", IPv6CIDR: "/64#0", Type: kops.SubnetTypePrivate},
			},
			expected: []string{""},
		},
		{
			subnets: []kops.ClusterSubnetSpec{
				{Name: "a", Zone: "a", CIDR: "", IPv6CIDR: "/64#0", Type: kops.SubnetTypeDualStack},
			},
			expected: []string{"10.0.0.0/8"},
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
		t.Run(fmt.Sprintf("test-%d", i+1), func(t *testing.T) {
			c := &kops.Cluster{}
			c.Spec.Networking.NetworkCIDR = "10.0.0.0/8"
			c.Spec.Networking.Subnets = test.subnets

			err := assignCIDRsToSubnets(c, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var actual []string
			for _, subnet := range c.Spec.Networking.Subnets {
				actual = append(actual, subnet.CIDR)
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Fatalf("unexpected result of network allocation: actual=%v, expected=%v", actual, test.expected)
			}
		})
	}
}

func Test_AssignSubnetZones(t *testing.T) {
	buildCloud := func() *awsup.MockAWSCloud {
		cloud := awsup.BuildMockAWSCloud("us-test-1", "abc")
		mockEC2 := &mockec2.MockEC2{}
		cloud.MockEC2 = mockEC2

		mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
			CidrBlock: aws.String("172.20.0.0/16"),
		}, "vpc-12345678")
		mockEC2.CreateSubnetWithId(&ec2.CreateSubnetInput{
			VpcId:            aws.String("vpc-12345678"),
			AvailabilityZone: aws.String("us-test-1a"),
			CidrBlock:        aws.String("172.20.32.0/19"),
		}, "subnet-a")
		mockEC2.CreateSubnetWithId(&ec2.CreateSubnetInput{
			VpcId:            aws.String("vpc-12345678"),
			AvailabilityZone: aws.String("us-test-1b"),
			Ipv6CidrBlock:    aws.String("2001:db8::/64"),
		}, "subnet-b")
		return cloud
	}

	buildCluster := func(subnets []kops.ClusterSubnetSpec) *kops.Cluster {
		c := &kops.Cluster{}
		c.Spec.CloudProvider.AWS = &kops.AWSSpec{}
		c.Spec.Networking.NetworkID = "vpc-12345678"
		c.Spec.Networking.NetworkCIDR = "172.20.0.0/16"
		c.Spec.Networking.Subnets = subnets
		return c
	}

	t.Run("zone is backfilled from the cloud", func(t *testing.T) {
		c := buildCluster([]kops.ClusterSubnetSpec{
			{Name: "a", ID: "subnet-a", CIDR: "172.20.32.0/19", Type: kops.SubnetTypePublic},
		})

		if err := assignCIDRsToSubnets(c, buildCloud()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if zone := c.Spec.Networking.Subnets[0].Zone; zone != "us-test-1a" {
			t.Fatalf("unexpected zone: %q", zone)
		}
	})

	t.Run("IPv6-only private subnets do not require an IPv4 CIDR", func(t *testing.T) {
		c := buildCluster([]kops.ClusterSubnetSpec{
			{Name: "a", ID: "subnet-a", CIDR: "172.20.32.0/19", Type: kops.SubnetTypePublic},
			{Name: "b", ID: "subnet-b", IPv6CIDR: "2001:db8::/64", Zone: "us-test-1b", Type: kops.SubnetTypePrivate},
		})

		if err := assignCIDRsToSubnets(c, buildCloud()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if zone := c.Spec.Networking.Subnets[0].Zone; zone != "us-test-1a" {
			t.Fatalf("unexpected zone: %q", zone)
		}
	})

	t.Run("errors when the zone cannot be determined", func(t *testing.T) {
		c := buildCluster([]kops.ClusterSubnetSpec{
			{Name: "a", ID: "subnet-a", CIDR: "172.20.32.0/19", Type: kops.SubnetTypePublic},
		})
		c.Spec.Networking.NetworkID = ""

		if err := assignCIDRsToSubnets(c, buildCloud()); err == nil {
			t.Fatalf("expected error assigning zones")
		}
	})
}
