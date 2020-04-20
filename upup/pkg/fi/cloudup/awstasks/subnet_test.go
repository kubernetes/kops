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

package awstasks

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func Test_Subnet_ValidateRequired(t *testing.T) {
	var a *Subnet
	e := &Subnet{}

	changes := &Subnet{}
	fi.BuildChanges(a, e, changes)

	err := e.CheckChanges(a, e, changes)
	if err == nil {
		t.Errorf("validation error was expected")
	}
	if fmt.Sprintf("%v", err) != "Subnet.VPC: Required value: must specify a VPC" {
		t.Errorf("unexpected error: %v", err)
	}
}

func Test_Subnet_CannotChangeSubnet(t *testing.T) {
	a := &Subnet{VPC: &VPC{Name: s("defaultvpc")}, CIDR: s("192.168.0.0/16")}
	e := &Subnet{}
	*e = *a

	e.CIDR = s("192.168.0.1/16")

	changes := &Subnet{}
	fi.BuildChanges(a, e, changes)

	err := e.CheckChanges(a, e, changes)
	if err == nil {
		t.Errorf("validation error was expected")
	}
	if fmt.Sprintf("%v", err) != "Subnet.CIDR: Forbidden: field is immutable: old=\"192.168.0.1/16\" new=\"192.168.0.0/16\"" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubnetCreate(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockec2.MockEC2{}
	cloud.MockEC2 = c

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		vpc1 := &VPC{
			Name: s("vpc1"),
			CIDR: s("172.20.0.0/16"),
			Tags: map[string]string{"Name": "vpc1"},
		}
		subnet1 := &Subnet{
			Name: s("subnet1"),
			VPC:  vpc1,
			CIDR: s("172.20.1.0/24"),
			Tags: map[string]string{"Name": "subnet1"},
		}

		return map[string]fi.Task{
			"subnet1": subnet1,
			"vpc1":    vpc1,
		}
	}

	{
		allTasks := buildTasks()
		subnet1 := allTasks["subnet1"].(*Subnet)

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

		if fi.StringValue(subnet1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		if len(c.SubnetIds()) != 1 {
			t.Fatalf("Expected exactly one Subnet; found %v", c.SubnetIds())
		}

		expected := &ec2.Subnet{
			CidrBlock: aws.String("172.20.1.0/24"),
			SubnetId:  aws.String("subnet-1"),
			VpcId:     aws.String("vpc-1"),
			Tags: buildTags(map[string]string{
				"Name": "subnet1",
			}),
		}
		actual := c.FindSubnet(*subnet1.ID)
		if actual == nil {
			t.Fatalf("Subnet created but then not found")
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Unexpected Subnet: expected=%v actual=%v", expected, actual)
		}
	}

	{
		allTasks := buildTasks()
		checkNoChanges(t, cloud, allTasks)
	}
}

func TestSharedSubnetCreateDoesNotCreateNew(t *testing.T) {
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

	subnet, err := c.CreateSubnet(&ec2.CreateSubnetInput{
		VpcId:     vpc.Vpc.VpcId,
		CidrBlock: aws.String("172.20.1.0/24"),
	})
	if err != nil {
		t.Fatalf("error creating test subnet: %v", err)
	}

	_, err = c.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{subnet.Subnet.SubnetId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("ExistingSubnet"),
			},
		},
	})
	if err != nil {
		t.Fatalf("error tagging test subnet: %v", err)
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
		subnet1 := &Subnet{
			Name:   s("subnet1"),
			VPC:    vpc1,
			CIDR:   s("172.20.1.0/24"),
			Tags:   map[string]string{"kubernetes.io/cluster/cluster.example.com": "shared"},
			Shared: fi.Bool(true),
			ID:     subnet.Subnet.SubnetId,
		}

		return map[string]fi.Task{
			"subnet1": subnet1,
			"vpc1":    vpc1,
		}
	}

	{
		allTasks := buildTasks()
		subnet1 := allTasks["subnet1"].(*Subnet)

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

		if fi.StringValue(subnet1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		if len(c.SubnetIds()) != 1 {
			t.Fatalf("Expected exactly one Subnet; found %v", c.SubnetIds())
		}

		actual := c.FindSubnet(*subnet.Subnet.SubnetId)
		if actual == nil {
			t.Fatalf("Subnet created but then not found")
		}
		expected := &ec2.Subnet{
			CidrBlock: aws.String("172.20.1.0/24"),
			SubnetId:  aws.String("subnet-1"),
			VpcId:     aws.String("vpc-1"),
			Tags: buildTags(map[string]string{
				"Name": "ExistingSubnet",
				"kubernetes.io/cluster/cluster.example.com": "shared",
			}),
		}

		mockec2.SortTags(expected.Tags)
		mockec2.SortTags(actual.Tags)

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Unexpected Subnet: expected=%v actual=%v", expected, actual)
		}
	}

	{
		allTasks := buildTasks()
		checkNoChanges(t, cloud, allTasks)
	}
}
