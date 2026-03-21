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
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func TestSharedEgressOnlyInternetGateway(t *testing.T) {
	ctx := context.TODO()

	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockec2.MockEC2{}
	cloud.MockEC2 = c

	// Pre-create multiple VPC & Egress Only Internet Gateways.
	// This is a broader test scenario that will also cover filtering Egress Only
	// Internet Gateways so that only the appropiate one for the VPC is selected.
	var vpcs []*ec2types.Vpc
	var eigws []*ec2types.EgressOnlyInternetGateway

	for index, cidr := range []string{"172.20.0.0/24", "172.20.1.0/24"} {
		vpc, err := c.CreateVpc(ctx, &ec2.CreateVpcInput{
			CidrBlock: aws.String(cidr),
		})

		if err != nil {
			t.Fatalf("error creating test VPC: %v", err)
		}

		_, err = c.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{aws.ToString(vpc.Vpc.VpcId)},
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("ExistingVPC"),
				},
				{
					Key:   aws.String("Index"),
					Value: aws.String(strconv.Itoa(index)),
				},
			},
		})

		if err != nil {
			t.Fatalf("error tagging test vpc: %v", err)
		}

		eigw, err := c.CreateEgressOnlyInternetGateway(ctx, &ec2.CreateEgressOnlyInternetGatewayInput{
			VpcId: vpc.Vpc.VpcId,
			TagSpecifications: awsup.EC2TagSpecification(ec2types.ResourceTypeEgressOnlyInternetGateway, map[string]string{
				"Name":  "ExistingInternetGateway",
				"Index": strconv.Itoa(index),
			}),
		})

		if err != nil {
			t.Fatalf("error creating test eigw: %v", err)
		}

		vpcs = append(vpcs, vpc.Vpc)
		eigws = append(eigws, eigw.EgressOnlyInternetGateway)
	}

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	// Only use the first cloud VPC and Egress Only Internet Gateway for KOPS tasks.
	buildTasks := func() map[string]fi.CloudupTask {
		vpc1 := &VPC{
			Name:      s("vpc1"),
			Lifecycle: fi.LifecycleSync,
			CIDR:      vpcs[0].CidrBlock,
			Tags:      map[string]string{"kubernetes.io/cluster/cluster.example.com": "shared"},
			Shared:    fi.PtrTo(true),
			ID:        vpcs[0].VpcId,
		}
		eigw1 := &EgressOnlyInternetGateway{
			Name:      s("eigw1"),
			Lifecycle: fi.LifecycleSync,
			VPC:       vpc1,
			Shared:    fi.PtrTo(true),
			ID:        eigws[0].EgressOnlyInternetGatewayId,
			Tags:      make(map[string]string),
		}

		return map[string]fi.CloudupTask{
			"eigw1": eigw1,
			"vpc1":  vpc1,
		}
	}

	{
		allTasks := buildTasks()
		eigw1 := allTasks["eigw1"].(*EgressOnlyInternetGateway)

		runTasks(t, cloud, allTasks)

		// Check the created Egress Only Internet Gateway has a valid ID
		if fi.ValueOf(eigw1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		// Check that there are still 2 Egress Only Internet Gateways and that none
		// extra have been created (or destroyed).
		if len(c.EgressOnlyInternetGatewayIds()) != 2 {
			t.Fatalf("Expected exactly two EgressOnlyInternetGateway; found %v", c.EgressOnlyInternetGatewayIds())
		}

		// Check the Egress Only Internet Gateway in our build context is the one
		// that we expect to be there.
		actual := c.FindEgressOnlyInternetGateway(*eigws[0].EgressOnlyInternetGatewayId)
		if actual == nil {
			t.Fatalf("EgressOnlyInternetGateway created but then not found")
		}
		expected := &ec2types.EgressOnlyInternetGateway{
			EgressOnlyInternetGatewayId: aws.String("eigw-1"),
			Tags: buildTags(map[string]string{
				"Name":  "ExistingInternetGateway",
				"Index": "0",
			}),
			Attachments: []ec2types.InternetGatewayAttachment{
				{
					VpcId: vpcs[0].VpcId,
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
		// Check the Egress Only Internet Gateway does not change on further runs.
		allTasks := buildTasks()
		checkNoChanges(t, ctx, cloud, allTasks)
	}
}

func runTasks(t *testing.T, cloud awsup.AWSCloud, allTasks map[string]fi.CloudupTask) {
	t.Helper()
	ctx := context.TODO()

	target := &awsup.AWSAPITarget{
		Cloud: cloud,
	}

	context, err := fi.NewCloudupContext(ctx, fi.DeletionProcessingModeDeleteIncludingDeferred, target, nil, cloud, nil, nil, nil, allTasks)
	if err != nil {
		t.Fatalf("error building context: %v", err)
	}

	if err := context.RunTasks(testRunTasksOptions); err != nil {
		t.Fatalf("unexpected error during Run: %v", err)
	}
}
