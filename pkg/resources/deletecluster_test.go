/*
Copyright 2016 The Kubernetes Authors.

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

package resources

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func TestAddUntaggedRouteTables(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	resources := make(map[string]*ResourceTracker)

	clusterName := "me.example.com"

	c := &mockec2.MockEC2{}
	cloud.MockEC2 = c

	// Matches by vpc id
	c.RouteTables = append(c.RouteTables, &ec2.RouteTable{
		VpcId:        aws.String("vpc-1234"),
		RouteTableId: aws.String("rt-1234"),
	})

	// Skips main route tables
	c.RouteTables = append(c.RouteTables, &ec2.RouteTable{
		VpcId:        aws.String("vpc-1234"),
		RouteTableId: aws.String("rt-1234main"),
		Associations: []*ec2.RouteTableAssociation{
			{
				Main: aws.Bool(true),
			},
		},
	})

	// Skips route table tagged with other cluster
	c.RouteTables = append(c.RouteTables, &ec2.RouteTable{
		VpcId:        aws.String("vpc-1234"),
		RouteTableId: aws.String("rt-1234main"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(awsup.TagClusterName),
				Value: aws.String("other.example.com"),
			},
		},
	})

	// Ignores non-matching vpcs
	c.RouteTables = append(c.RouteTables, &ec2.RouteTable{
		VpcId:        aws.String("vpc-5555"),
		RouteTableId: aws.String("rt-5555"),
	})

	resources["vpc:vpc-1234"] = &ResourceTracker{}

	err := addUntaggedRouteTables(cloud, clusterName, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var keys []string
	for k := range resources {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	expected := []string{"route-table:rt-1234", "vpc:vpc-1234"}
	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("expected=%q, actual=%q", expected, keys)
	}
}
