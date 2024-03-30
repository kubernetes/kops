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

package mockautoscaling

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"k8s.io/klog/v2"
)

func (m *MockAutoscaling) AttachLoadBalancers(ctx context.Context, request *autoscaling.AttachLoadBalancersInput, optFns ...func(*autoscaling.Options)) (*autoscaling.AttachLoadBalancersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AttachLoadBalancers: %v", request)

	name := *request.AutoScalingGroupName

	asg := m.Groups[name]
	if asg == nil {
		return nil, fmt.Errorf("Group %q not found", name)
	}

	asg.LoadBalancerNames = request.LoadBalancerNames
	return &autoscaling.AttachLoadBalancersOutput{}, nil
}

func (m *MockAutoscaling) AttachLoadBalancerTargetGroups(ctx context.Context, request *autoscaling.AttachLoadBalancerTargetGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AttachLoadBalancers: %v", request)

	name := *request.AutoScalingGroupName

	asg := m.Groups[name]
	if asg == nil {
		return nil, fmt.Errorf("group %q not found", name)
	}

	asg.TargetGroupARNs = append(asg.TargetGroupARNs, request.TargetGroupARNs...)
	return &autoscaling.AttachLoadBalancerTargetGroupsOutput{}, nil
}
