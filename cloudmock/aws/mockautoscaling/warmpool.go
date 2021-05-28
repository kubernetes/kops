/*
Copyright 2021 The Kubernetes Authors.

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

import "github.com/aws/aws-sdk-go/service/autoscaling"

func (m *MockAutoscaling) DescribeWarmPool(input *autoscaling.DescribeWarmPoolInput) (*autoscaling.DescribeWarmPoolOutput, error) {
	instances, found := m.WarmPoolInstances[*input.AutoScalingGroupName]
	if !found {
		return &autoscaling.DescribeWarmPoolOutput{}, nil
	}
	ret := &autoscaling.DescribeWarmPoolOutput{
		Instances: instances,
	}
	return ret, nil
}

func (m *MockAutoscaling) DeleteWarmPool(*autoscaling.DeleteWarmPoolInput) (*autoscaling.DeleteWarmPoolOutput, error) {
	return &autoscaling.DeleteWarmPoolOutput{}, nil
}
