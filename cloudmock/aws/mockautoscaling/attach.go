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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog"
)

func (m *MockAutoscaling) AttachLoadBalancers(request *autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error) {
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

func (m *MockAutoscaling) AttachLoadBalancersWithContext(aws.Context, *autoscaling.AttachLoadBalancersInput, ...request.Option) (*autoscaling.AttachLoadBalancersOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) AttachLoadBalancersRequest(*autoscaling.AttachLoadBalancersInput) (*request.Request, *autoscaling.AttachLoadBalancersOutput) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
