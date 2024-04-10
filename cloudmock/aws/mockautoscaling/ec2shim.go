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

package mockautoscaling

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type ec2Shim struct {
	awsinterfaces.EC2API
	mockAutoscaling *MockAutoscaling
}

func (m *MockAutoscaling) GetEC2Shim(e awsinterfaces.EC2API) awsinterfaces.EC2API {
	return &ec2Shim{
		EC2API:          e,
		mockAutoscaling: m,
	}
}

func (e *ec2Shim) TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if input.DryRun != nil && *input.DryRun {
		return &ec2.TerminateInstancesOutput{}, nil
	}
	for _, id := range input.InstanceIds {
		request := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     aws.String(id),
			ShouldDecrementDesiredCapacity: aws.Bool(false),
		}
		if _, err := e.mockAutoscaling.TerminateInstanceInAutoScalingGroup(ctx, request); err != nil {
			return nil, err
		}
	}
	return &ec2.TerminateInstancesOutput{}, nil
}
