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
	"fmt"
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

func Test_GetResourceName32(t *testing.T) {
	grid := []struct {
		ClusterName string
		Prefix      string
		Expected    string
	}{
		{
			"mycluster",
			"bastion",
			"bastion-mycluster-vnrjie",
		},
		{
			"mycluster.example.com",
			"bastion",
			"bastion-mycluster-example-o8elkm",
		},
		{
			"this.is.a.very.long.cluster.example.com",
			"api",
			"api-this-is-a-very-long-c-q4ukp4",
		},
		{
			"this.is.a.very.long.cluster.example.com",
			"bastion",
			"bastion-this-is-a-very-lo-4ggpa2",
		},
	}
	for _, g := range grid {
		actual := GetResourceName32(g.ClusterName, g.Prefix)
		if actual != g.Expected {
			t.Errorf("unexpected result from %q+%q.  expected %q, got %q", g.Prefix, g.ClusterName, g.Expected, actual)
		}
	}
}

func TestTruncateString(t *testing.T) {
	grid := []struct {
		Input         string
		Expected      string
		MaxLength     int
		AlwaysAddHash bool
	}{
		{
			Input:     "foo",
			Expected:  "foo",
			MaxLength: 64,
		},
		{
			Input:     "this_string_is_33_characters_long",
			Expected:  "this_string_is_33_characters_long",
			MaxLength: 64,
		},
		{
			Input:         "this_string_is_33_characters_long",
			Expected:      "this_string_is_33_characters_long-t4mk8d",
			MaxLength:     64,
			AlwaysAddHash: true,
		},
		{
			Input:     "this_string_is_longer_it_is_46_characters_long",
			Expected:  "this_string_is_longer_it_-ha2gug",
			MaxLength: 32,
		},
		{
			Input:         "this_string_is_longer_it_is_46_characters_long",
			Expected:      "this_string_is_longer_it_-ha2gug",
			MaxLength:     32,
			AlwaysAddHash: true,
		},
		{
			Input:     "this_string_is_even_longer_due_to_extreme_verbosity_it_is_in_fact_84_characters_long",
			Expected:  "this_string_is_even_longer_due_to_extreme_verbosity_it_is-7mc0g6",
			MaxLength: 64,
		},
	}

	for _, g := range grid {
		t.Run(fmt.Sprintf("input:%s/maxLength:%d/alwaysAddHash:%v", g.Input, g.MaxLength, g.AlwaysAddHash), func(t *testing.T) {
			opt := TruncateStringOptions{MaxLength: g.MaxLength, AlwaysAddHash: g.AlwaysAddHash}
			actual := TruncateString(g.Input, opt)
			if actual != g.Expected {
				t.Errorf("TruncateString(%q, %+v) => %q, expected %q", g.Input, opt, actual, g.Expected)
			}
		})
	}
}
