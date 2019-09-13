/*
Copyright 2017 The Kubernetes Authors.

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

package awstasks

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func TestSharedInternetGatewayDoesNotRename(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockec2.MockEC2{}
	cloud.MockEC2 = c

	// Pre-create the vpc / subnet
	vpc, err := c.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String("172.20.0.0/16"),
	})
	if err != nil {
		t.Fatalf("error creating test VPC: %v", err)
	}
	_, err = c.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{vpc.Vpc.VpcId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("ExistingVPC"),
			},
		},
	})
	if err != nil {
		t.Fatalf("error tagging test vpc: %v", err)
	}

	internetGateway, err := c.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		t.Fatalf("error creating test igw: %v", err)
	}

	_, err = c.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{internetGateway.InternetGateway.InternetGatewayId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("ExistingInternetGateway"),
			},
		},
	})
	if err != nil {
		t.Fatalf("error tagging test igw: %v", err)
	}

	_, err = c.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: internetGateway.InternetGateway.InternetGatewayId,
		VpcId:             vpc.Vpc.VpcId,
	})
	if err != nil {
		t.Fatalf("error attaching igw: %v", err)
	}

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		vpc1 := &VPC{
			Name:   s("vpc1"),
			CIDR:   s("172.20.0.0/16"),
			Tags:   map[string]string{"kubernetes.io/cluster/cluster.example.com": "shared"},
			Shared: fi.Bool(true),
			ID:     vpc.Vpc.VpcId,
		}
		igw1 := &InternetGateway{
			Name:   s("igw1"),
			VPC:    vpc1,
			Shared: fi.Bool(true),
			ID:     internetGateway.InternetGateway.InternetGatewayId,
			Tags:   make(map[string]string),
		}

		return map[string]fi.Task{
			"igw1": igw1,
			"vpc1": vpc1,
		}
	}

	{
		allTasks := buildTasks()
		igw1 := allTasks["igw1"].(*InternetGateway)

		target := &awsup.AWSAPITarget{
			Cloud: cloud,
		}

		context, err := fi.NewContext(target, nil, cloud, nil, nil, nil, true, allTasks)
		if err != nil {
			t.Fatalf("error building context: %v", err)
		}

		if err := context.RunTasks(testRunTasksOptions); err != nil {
			t.Fatalf("unexpected error during Run: %v", err)
		}

		if fi.StringValue(igw1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		if len(c.InternetGatewayIds()) != 1 {
			t.Fatalf("Expected exactly one InternetGateway; found %v", c.InternetGatewayIds())
		}

		actual := c.FindInternetGateway(*internetGateway.InternetGateway.InternetGatewayId)
		if actual == nil {
			t.Fatalf("InternetGateway created but then not found")
		}
		expected := &ec2.InternetGateway{
			InternetGatewayId: aws.String("igw-1"),
			Tags: buildTags(map[string]string{
				"Name": "ExistingInternetGateway",
			}),
			Attachments: []*ec2.InternetGatewayAttachment{
				{
					VpcId: vpc.Vpc.VpcId,
				},
			},
		}

		mockec2.SortTags(expected.Tags)
		mockec2.SortTags(actual.Tags)

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Unexpected InternetGateway: expected=%v actual=%v", expected, actual)
		}
	}

	{
		allTasks := buildTasks()
		checkNoChanges(t, cloud, allTasks)
	}
}
