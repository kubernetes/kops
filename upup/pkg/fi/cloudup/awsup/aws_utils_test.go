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

package awsup

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/pkg/apis/kops"
)

func TestValidateRegion(t *testing.T) {
	allRegions = []*ec2.Region{
		{
			RegionName: aws.String("us-mock-1"),
		},
		{
			RegionName: aws.String("us-mock-2"),
		},
	}
	for _, region := range []string{"us-mock-1", "us-mock-2"} {
		err := ValidateRegion(region)
		if err != nil {
			t.Fatalf("unexpected error validating region %q: %v", region, err)
		}
	}

	for _, region := range []string{"is-lost-1", "no-road-2", "no-real-3"} {
		err := ValidateRegion(region)
		if err == nil {
			t.Fatalf("expected error validating region %q", region)
		}
	}
}

func TestFindRegion(t *testing.T) {
	for _, zone := range []string{"us-east-1a", "us-east-1b", "us-east-1c", "us-east-2a", "us-east-2b", "us-east-2c"} {
		c := &kops.Cluster{}
		c.Spec.Subnets = append(c.Spec.Subnets, kops.ClusterSubnetSpec{Name: "subnet-" + zone, Zone: zone})

		region, err := FindRegion(c)
		if err != nil {
			t.Fatalf("unexpected error finding region for %q: %v", zone, err)
		}

		expected := zone[:len(zone)-1]
		if region != expected {
			t.Fatalf("unexpected region for zone: %q vs %q", expected, region)
		}
	}

}

func TestEC2TagSpecification(t *testing.T) {
	cases := []struct {
		Name          string
		ResourceType  string
		Tags          map[string]string
		Specification []*ec2.TagSpecification
	}{
		{
			Name: "No tags",
		},
		{
			Name:         "simple tag",
			ResourceType: "vpc",
			Tags: map[string]string{
				"foo": "bar",
			},
			Specification: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("vpc"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("foo"),
							Value: aws.String("bar"),
						},
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			s := EC2TagSpecification(tc.ResourceType, tc.Tags)
			if !reflect.DeepEqual(s, tc.Specification) {
				t.Fatalf("tag specifications did not match: %q vs %q", s, tc.Specification)
			}
		})
	}
}
