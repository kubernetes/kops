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
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"

	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

var testRunTasksOptions = fi.RunTasksOptions{
	MaxTaskDuration:         2 * time.Second,
	WaitAfterAllTasksFailed: 500 * time.Millisecond,
}

func TestElasticIPCreate(t *testing.T) {
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
		eip1 := &ElasticIP{
			Name:        s("eip1"),
			TagOnSubnet: subnet1,
		}

		return map[string]fi.Task{
			"eip1":    eip1,
			"subnet1": subnet1,
			"vpc1":    vpc1,
		}
	}

	{
		allTasks := buildTasks()
		eip1 := allTasks["eip1"].(*ElasticIP)

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

		if fi.StringValue(eip1.ID) == "" {
			t.Fatalf("ID not set after create")
		}

		if len(c.Addresses) != 1 {
			t.Fatalf("Expected exactly one ElasticIP; found %v", c.Addresses)
		}

		expected := &ec2.Address{
			AllocationId: eip1.ID,
			Domain:       s("vpc"),
			PublicIp:     s("192.0.2.1"),
		}
		actual := c.Addresses[*eip1.ID]
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Unexpected ElasticIP: expected=%v actual=%v", expected, actual)
		}
	}

	{
		allTasks := buildTasks()
		checkNoChanges(t, cloud, allTasks)
	}
}

func checkNoChanges(t *testing.T, cloud fi.Cloud, allTasks map[string]fi.Task) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			KubernetesVersion: "v1.9.0",
		},
	}
	assetBuilder := assets.NewAssetBuilder(cluster, "")
	target := fi.NewDryRunTarget(assetBuilder, os.Stderr)
	context, err := fi.NewContext(target, nil, cloud, nil, nil, nil, true, allTasks)
	if err != nil {
		t.Fatalf("error building context: %v", err)
	}

	if err := context.RunTasks(testRunTasksOptions); err != nil {
		t.Fatalf("unexpected error during Run: %v", err)
	}

	if target.HasChanges() {
		var b bytes.Buffer
		if err := target.PrintReport(allTasks, &b); err != nil {
			t.Fatalf("error building report: %v", err)
		}
		t.Fatalf("Target had changes after executing: %v", b.String())
	}

}
