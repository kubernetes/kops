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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type ec2Shim struct {
	ec2iface.EC2API
	mockAutoscaling *MockAutoscaling
}

func (m *MockAutoscaling) GetEC2Shim(e ec2iface.EC2API) ec2iface.EC2API {
	return &ec2Shim{
		EC2API:          e,
		mockAutoscaling: m,
	}
}

func (e *ec2Shim) TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	if input.DryRun != nil && *input.DryRun {
		return &ec2.TerminateInstancesOutput{}, nil
	}
	for _, id := range input.InstanceIds {
		request := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     id,
			ShouldDecrementDesiredCapacity: aws.Bool(false),
		}
		if _, err := e.mockAutoscaling.TerminateInstanceInAutoScalingGroup(request); err != nil {
			return nil, err
		}
	}
	return &ec2.TerminateInstancesOutput{}, nil
}
