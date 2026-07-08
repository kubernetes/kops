/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cloudmock/aws/mockelbv2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// TestTargetGroupHealthCheckChange verifies that a health check change on an existing target group is applied,
// instead of recurring as an unapplied change on every run.
func TestTargetGroupHealthCheckChange(t *testing.T) {
	ctx := context.TODO()

	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockec2.MockEC2{}
	cloud.MockEC2 = c
	cloud.MockELBV2 = &mockelbv2.MockELBV2{EC2: c}

	// Pre-create the VPC
	vpc, err := c.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("172.20.0.0/16"),
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
		},
	})
	if err != nil {
		t.Fatalf("error tagging test vpc: %v", err)
	}

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func(healthCheckProtocol elbv2types.ProtocolEnum, healthCheckPath *string) map[string]fi.CloudupTask {
		vpc1 := &VPC{
			Name:      s("vpc1"),
			Lifecycle: fi.LifecycleSync,
			CIDR:      s("172.20.0.0/16"),
			Tags:      map[string]string{"kubernetes.io/cluster/cluster.example.com": "shared"},
			Shared:    new(true),
			ID:        vpc.Vpc.VpcId,
		}
		tg1 := &TargetGroup{
			Name:                s("tg1"),
			Lifecycle:           fi.LifecycleSync,
			VPC:                 vpc1,
			Tags:                map[string]string{"Name": "tg1"},
			Protocol:            elbv2types.ProtocolEnumTcp,
			Port:                new(int32(3988)),
			Interval:            new(int32(10)),
			HealthyThreshold:    new(int32(2)),
			UnhealthyThreshold:  new(int32(2)),
			HealthCheckProtocol: healthCheckProtocol,
			HealthCheckPath:     healthCheckPath,
			Shared:              new(false),
		}

		return map[string]fi.CloudupTask{
			"vpc1": vpc1,
			"tg1":  tg1,
		}
	}

	// Create the target group with a TCP health check (as an older kOps version would).
	runTasks(t, cloud, buildTasks(elbv2types.ProtocolEnumTcp, nil))

	// Upgrade to an HTTPS health check with a path.
	runTasks(t, cloud, buildTasks(elbv2types.ProtocolEnumHttps, s("/healthz")))

	// The change must have been applied, so a subsequent run sees no changes.
	checkNoChanges(t, ctx, cloud, buildTasks(elbv2types.ProtocolEnumHttps, s("/healthz")))
}
