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

package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func TestListVPCs(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	awsCloud := h.SetupMockAWS()

	mockEC2 := awsCloud.EC2().(*mockec2.MockEC2)

	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-legacy")
	mockEC2.CreateTags(&ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{"vpc-legacy"}),
		Tags: []*ec2.Tag{
			{Key: aws.String("KubernetesCluster"), Value: aws.String("legacy.example.com")},
		},
	})

	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-shared")
	mockEC2.CreateTags(&ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{"vpc-shared"}),
		Tags: []*ec2.Tag{
			{Key: aws.String("kubernetes.io/cluster/shared.example.com"), Value: aws.String("shared")},
		},
	})
	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-shared-with-legacy")
	mockEC2.CreateTags(&ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{"vpc-shared-with-legacy"}),
		Tags: []*ec2.Tag{
			{Key: aws.String("KubernetesCluster"), Value: aws.String("shared-with-legacy.example.com")},
			{Key: aws.String("kubernetes.io/cluster/shared-with-legacy.example.com"), Value: aws.String("shared")},
		},
	})

	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-owned")
	mockEC2.CreateTags(&ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{"vpc-owned"}),
		Tags: []*ec2.Tag{
			{Key: aws.String("kubernetes.io/cluster/owned.example.com"), Value: aws.String("owned")},
		},
	})

	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-owned-with-legacy")
	mockEC2.CreateTags(&ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{"vpc-owned-with-legacy"}),
		Tags: []*ec2.Tag{
			{Key: aws.String("KubernetesCluster"), Value: aws.String("owned-with-legacy.example.com")},
			{Key: aws.String("kubernetes.io/cluster/owned-with-legacy.example.com"), Value: aws.String("owned")},
		},
	})

	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-other")
	mockEC2.CreateTags(&ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{"vpc-other"}),
		Tags: []*ec2.Tag{
			{Key: aws.String("KubernetesCluster"), Value: aws.String("other.example.com")},
			{Key: aws.String("kubernetes.io/cluster/other.example.com"), Value: aws.String("shared")},
		},
	})

	grid := []struct {
		ClusterName string
		Expected    []string
		Shared      bool
	}{
		{ClusterName: "mismatch.example.com", Expected: []string{}},
		{ClusterName: "legacy.example.com", Expected: []string{"vpc-legacy"}},
		{ClusterName: "shared-with-legacy.example.com", Expected: []string{"vpc-shared-with-legacy"}, Shared: true},
		{ClusterName: "shared.example.com", Expected: []string{"vpc-shared"}, Shared: true},
		{ClusterName: "owned-with-legacy.example.com", Expected: []string{"vpc-owned-with-legacy"}},
		{ClusterName: "owned.example.com", Expected: []string{"vpc-owned"}},
	}

	for _, g := range grid {
		resources, err := ListVPCs(awsCloud, g.ClusterName)
		if err != nil {
			t.Errorf("unexpected error listing VPCs: %v", err)
			continue
		}

		var actual []string
		for _, r := range resources {
			actual = append(actual, r.ID)

			if r.Shared != g.Shared {
				t.Errorf("unexpected shared value for %s: actual=%v, expected=%v", g.ClusterName, r.Shared, g.Shared)
			}
		}

		if !utils.StringSlicesEqualIgnoreOrder(actual, g.Expected) {
			t.Errorf("unexpected vpcs for %s: actual=%v, expected=%v", g.ClusterName, actual, g.Expected)
		}
	}
}
