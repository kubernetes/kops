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
	"strconv"
	"testing"
	"time"

	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestLaunchConfigurationGarbageCollection(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockEC2 := &mockec2.MockEC2{}
	cloud.MockEC2 = mockEC2
	as := &mockautoscaling.MockAutoscaling{}
	cloud.MockAutoscaling = as

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2016-10-21T20:07:19.000Z"),
		ImageId:        aws.String("ami-12345678"),
		Name:           aws.String("k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-10-21"),
		OwnerId:        aws.String(awsup.WellKnownAccountKopeio),
		RootDeviceName: aws.String("/dev/xvda"),
	})

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func(spotPrice string) map[string]fi.Task {
		lc := &LaunchConfiguration{
			Name:           s("lc1"),
			SpotPrice:      spotPrice,
			ImageID:        s("ami-12345678"),
			InstanceType:   s("m3.medium"),
			SecurityGroups: []*SecurityGroup{},
		}

		return map[string]fi.Task{
			"lc1": lc,
		}
	}

	// We change the launch configuration 5 times, verifying that new launch configurations are created,
	// and that older ones are eventually GCed
	for i := 0; i < 5; i++ {
		spotPrice := strconv.Itoa(i + 1)
		{
			allTasks := buildTasks(spotPrice)
			lc1 := allTasks["lc1"].(*LaunchConfiguration)

			target := &awsup.AWSAPITarget{
				Cloud: cloud,
			}

			context, err := fi.NewContext(target, nil, cloud, nil, nil, nil, true, allTasks)
			if err != nil {
				t.Fatalf("error building context: %v", err)
			}

			// We use a longer deadline because we know we often need to
			// retry here, because we create different versions of
			// launchconfigurations using the timestamp, but only to
			// per-second granularity.  This normally works out because we
			// retry for O(minutes), so after a few retries the clock has
			// advanced.  But if we use too short a deadline in our tests we
			// don't get this behaviour.
			options := testRunTasksOptions
			options.MaxTaskDuration = 5 * time.Second
			if err := context.RunTasks(options); err != nil {
				t.Fatalf("unexpected error during Run: %v", err)
			}

			if fi.StringValue(lc1.ID) == "" {
				t.Fatalf("ID not set after create")
			}

			expectedCount := i + 1
			if expectedCount > RetainLaunchConfigurationCount() {
				expectedCount = RetainLaunchConfigurationCount()
			}
			if len(as.LaunchConfigurations) != expectedCount {
				t.Fatalf("Expected exactly %d LaunchConfigurations; found %v", expectedCount, as.LaunchConfigurations)
			}

			// TODO: verify that we retained the N latest

			actual := as.LaunchConfigurations[*lc1.ID]
			if aws.StringValue(actual.SpotPrice) != spotPrice {
				t.Fatalf("Unexpected spotPrice: expected=%v actual=%v", spotPrice, aws.StringValue(actual.SpotPrice))
			}
		}

		{
			allTasks := buildTasks(spotPrice)
			checkNoChanges(t, cloud, allTasks)
		}
	}
}
