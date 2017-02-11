/*
Copyright 2016 The Kubernetes Authors.

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
)

func (m *MockAutoscaling) AttachInstances(input *autoscaling.AttachInstancesInput) (*autoscaling.AttachInstancesOutput, error) {
	for _, group := range m.Groups {
		if group.AutoScalingGroupName == input.AutoScalingGroupName {
			for _, instanceID := range input.InstanceIds {
				group.Instances = append(group.Instances, &autoscaling.Instance{InstanceId: instanceID})
			}
		}
	}
	return nil, nil
}

func (m *MockAutoscaling) CreateAutoScalingGroup(input *autoscaling.CreateAutoScalingGroupInput) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	newGroup := &autoscaling.Group{
		AutoScalingGroupName: input.AutoScalingGroupName,
		MinSize:              input.MinSize,
		MaxSize:              input.MaxSize,
		Instances:            []*autoscaling.Instance{},
	}

	m.Groups = append(m.Groups, newGroup)
	return nil, nil
}

func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	for _, group := range m.Groups {
		for i := range group.Instances {
			if aws.StringValue(group.Instances[i].InstanceId) == aws.StringValue(input.InstanceId) {
				group.Instances = append(group.Instances[:i], group.Instances[i+1:]...)
			}
		}
	}
	return nil, nil
}
