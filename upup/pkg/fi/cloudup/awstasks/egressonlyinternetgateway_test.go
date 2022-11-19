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

func TestSharedEgressOnlyInternetGatewayDoesNotRename(t *testing.T) {
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

	internetGateway, err := c.CreateEgressOnlyInternetGateway(&ec2.CreateEgressOnlyInternetGatewayInput{
		VpcId: vpc.Vpc.VpcId,
		TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeEgressOnlyInternetGateway, map[string]string{
			"Name": "ExistingInternetGateway",
		}),
	})
	if err != nil {
		t.Fatalf("error creating test eigw: %v", err)
	}

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		vpc1 := &VPC{
			Name:      s("vpc1"),
			Lifecycle: fi.LifecycleSync,
			CIDR:      s("172.20.0.0/16"),
			Tags:      map[string]string{"kubernetes.io/cluster/cluster.example.com": "shared"},
			Shared:    fi.PtrTo(true),
			ID:        vpc.Vpc.VpcId,
		}
		eigw1 := &EgressOnlyInternetGateway{
			Name:      s("eigw1"),
			Lifecycle: fi.LifecycleSync,
			VPC:       vpc1,
			Shared:    fi.PtrTo(true),
			ID:        internetGateway.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId,
			Tags:      make(map[string]string),
		}

		return map[string]fi.Task{
			"eigw1": eigw1,
			"vpc1":  vpc1,
		}
	}

	{
		allTasks := buildTasks()
		eigw1 := allTasks["eigw1"].(*EgressOnlyInternetGateway)

		target := &awsup.AWSAPITarget{
			Cloud: cloud,
		}

		context, err := fi.NewContext(target, nil, cloud, nil, nil, nil, true, allTasks)
		if err != nil {
			t.Fatalf("error building context: %v", err)
		}
		defer context.Close()

		if err := context.RunTasks(testRunTasksOptions); err != nil {
			t.Fatalf("unexpected error during Run: %v", err)
		}

		if fi.ValueOf(eigw1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		if len(c.EgressOnlyInternetGatewayIds()) != 1 {
			t.Fatalf("Expected exactly one EgressOnlyInternetGateway; found %v", c.EgressOnlyInternetGatewayIds())
		}

		actual := c.FindEgressOnlyInternetGateway(*internetGateway.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId)
		if actual == nil {
			t.Fatalf("EgressOnlyInternetGateway created but then not found")
		}
		expected := &ec2.EgressOnlyInternetGateway{
			EgressOnlyInternetGatewayId: aws.String("eigw-1"),
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
			t.Fatalf("Unexpected EgressOnlyInternetGateway: expected=%v actual=%v", expected, actual)
		}
	}

	{
		allTasks := buildTasks()
		checkNoChanges(t, cloud, allTasks)
	}
}
