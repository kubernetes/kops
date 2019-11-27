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
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func TestVPCCreate(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockec2.MockEC2{}
	cloud.MockEC2 = c

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		vpc1 := &VPC{
			Name: s("vpc1"),
			CIDR: s("172.21.0.0/16"),
			Tags: map[string]string{"Name": "vpc1"},
		}
		return map[string]fi.Task{
			"vpc1": vpc1,
		}
	}

	{
		allTasks := buildTasks()
		vpc1 := allTasks["vpc1"].(*VPC)

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

		if fi.StringValue(vpc1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		if len(c.Vpcs) != 1 {
			t.Fatalf("Expected exactly one Vpc; found %v", c.Vpcs)
		}

		expected := &ec2.Vpc{
			CidrBlock: s("172.21.0.0/16"),
			IsDefault: fi.Bool(false),
			VpcId:     vpc1.ID,
			Tags: buildTags(map[string]string{
				"Name": "vpc1",
			}),
		}
		actual := c.FindVpc(*vpc1.ID)
		if actual == nil {
			t.Fatalf("VPC created but then not found")
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Unexpected VPC: expected=%v actual=%v", expected, actual)
		}
	}

	{
		allTasks := buildTasks()

		checkNoChanges(t, cloud, allTasks)
	}
}

func buildTags(tags map[string]string) []*ec2.Tag {
	var t []*ec2.Tag
	for k, v := range tags {
		t = append(t, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return t
}

// Test4758 is a sanity check for https://github.com/kubernetes/kops/issues/4758
func Test4758(t *testing.T) {
	a := &VPC{
		Name: s("cluster2.example.com"),
		Tags: map[string]string{},
	}

	e := &VPC{
		Name: s("cluster2.example.com"),
		Tags: map[string]string{},
	}

	changes := &VPC{}
	changed := fi.BuildChanges(a, e, changes)

	if changed {
		t.Errorf("expected changed=false")
	}

	expectedChanges := &VPC{}
	if !reflect.DeepEqual(changes, expectedChanges) {
		t.Errorf("unexpected changes: +%v", changes)
	}
}
