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

package mockelb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"k8s.io/klog/v2"
)

func (m *MockELB) ModifyLoadBalancerAttributes(ctx context.Context, request *elb.ModifyLoadBalancerAttributesInput, optFns ...func(*elb.Options)) (*elb.ModifyLoadBalancerAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ModifyLoadBalancerAttributes: %v", request)

	lb := m.LoadBalancers[aws.ToString(request.LoadBalancerName)]
	if lb == nil {
		return nil, fmt.Errorf("LoadBalancer not found")
	}

	lb.attributes = *request.LoadBalancerAttributes

	copy := lb.attributes

	return &elb.ModifyLoadBalancerAttributesOutput{
		LoadBalancerName:       request.LoadBalancerName,
		LoadBalancerAttributes: &copy,
	}, nil
}

func (m *MockELB) DescribeLoadBalancerAttributes(ctx context.Context, request *elb.DescribeLoadBalancerAttributesInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancerAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeLoadBalancerAttributes: %v", request)

	lb := m.LoadBalancers[aws.ToString(request.LoadBalancerName)]
	if lb == nil {
		return nil, fmt.Errorf("LoadBalancer not found")
	}

	copy := lb.attributes

	return &elb.DescribeLoadBalancerAttributesOutput{
		LoadBalancerAttributes: &copy,
	}, nil
}
