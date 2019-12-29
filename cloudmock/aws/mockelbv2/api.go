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

package mockelbv2

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"k8s.io/klog"
)

type MockELBV2 struct {
	elbv2iface.ELBV2API

	mutex sync.Mutex

	LoadBalancers map[string]*loadBalancer
	TargetGroups  map[string]*targetGroup
}

type loadBalancer struct {
	description elbv2.LoadBalancer
}

type targetGroup struct {
	description elbv2.TargetGroup
}

func (m *MockELBV2) DescribeLoadBalancers(request *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("DescribeLoadBalancers v2 %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var elbs []*elbv2.LoadBalancer
	for _, elb := range m.LoadBalancers {
		match := false

		if len(request.LoadBalancerArns) > 0 {
			for _, name := range request.LoadBalancerArns {
				if aws.StringValue(elb.description.LoadBalancerArn) == aws.StringValue(name) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			elbs = append(elbs, &elb.description)
		}
	}

	return &elbv2.DescribeLoadBalancersOutput{
		LoadBalancers: elbs,
	}, nil
}

func (m *MockELBV2) DescribeLoadBalancersPages(request *elbv2.DescribeLoadBalancersInput, callback func(p *elbv2.DescribeLoadBalancersOutput, lastPage bool) (shouldContinue bool)) error {
	// For the mock, we just send everything in one page
	page, err := m.DescribeLoadBalancers(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockELBV2) DescribeTargetGroups(request *elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("DescribeTargetGroups %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var tgs []*elbv2.TargetGroup
	for _, tg := range m.TargetGroups {
		match := false

		if len(request.TargetGroupArns) > 0 {
			for _, name := range request.TargetGroupArns {
				if aws.StringValue(tg.description.TargetGroupArn) == aws.StringValue(name) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			tgs = append(tgs, &tg.description)
		}
	}

	return &elbv2.DescribeTargetGroupsOutput{
		TargetGroups: tgs,
	}, nil
}

func (m *MockELBV2) DescribeTargetGroupsPages(request *elbv2.DescribeTargetGroupsInput, callback func(p *elbv2.DescribeTargetGroupsOutput, lastPage bool) (shouldContinue bool)) error {
	page, err := m.DescribeTargetGroups(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}
