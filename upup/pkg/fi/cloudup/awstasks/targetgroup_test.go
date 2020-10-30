/*
Copyright 2020 The Kubernetes Authors.

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
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/kops/cloudmock/aws/mockelbv2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func TestReconcileTargetGroups(t *testing.T) {
	cases := []struct {
		name          string
		actualTGs     []*TargetGroup
		expectedTGs   []*TargetGroup
		reconciledTGs []*TargetGroup
	}{
		{
			name: "with external TGs",
			actualTGs: []*TargetGroup{
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg2/2")},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg3/3")},
			},
			expectedTGs: []*TargetGroup{
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg2/2"), Shared: fi.Bool(true)},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg3/3"), Shared: fi.Bool(true)},
			},
			reconciledTGs: []*TargetGroup{
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg2/2")},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg3/3")},
			},
		},
		{
			name: "with API TGs",
			actualTGs: []*TargetGroup{
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/api-foo1/1")},
			},
			expectedTGs: []*TargetGroup{
				{Name: aws.String("api-foo1"), Shared: fi.Bool(false)},
			},
			reconciledTGs: []*TargetGroup{
				{Name: aws.String("api-foo1"), Shared: fi.Bool(false)},
			},
		},
		{
			name: "with API and external TGs",
			actualTGs: []*TargetGroup{
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/api-foo1/1")},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg2/2")},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg3/3")},
			},
			expectedTGs: []*TargetGroup{
				{Name: aws.String("api-foo1"), Shared: fi.Bool(false)},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg2/2"), Shared: fi.Bool(true)},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg3/3"), Shared: fi.Bool(true)},
			},
			reconciledTGs: []*TargetGroup{
				{Name: aws.String("api-foo1"), Shared: fi.Bool(false)},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg2/2")},
				{ARN: aws.String("arn:aws:elasticloadbalancing:us-test-1:000000000000:targetgroup/tg3/3")},
			},
		},
	}
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockelbv2.MockELBV2{}
	cloud.MockELBV2 = c
	c.CreateTargetGroup(&elbv2.CreateTargetGroupInput{Name: aws.String("api-foo1")})
	c.CreateTargetGroup(&elbv2.CreateTargetGroupInput{Name: aws.String("tg2")})
	c.CreateTargetGroup(&elbv2.CreateTargetGroupInput{Name: aws.String("tg3")})

	resp, _ := c.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{})
	fmt.Printf("%+v\n", resp)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualReconciled, err := ReconcileTargetGroups(cloud, tc.actualTGs, tc.expectedTGs)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			if !reflect.DeepEqual(actualReconciled, tc.reconciledTGs) {
				t.Errorf("Reconciled TGs didn't match: %+v vs %+v\n", actualReconciled, tc.reconciledTGs)
				t.Fail()
			}
		})
	}
}
