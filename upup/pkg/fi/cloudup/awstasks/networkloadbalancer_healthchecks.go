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

package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type NetworkLoadBalancerHealthCheck struct {
	Target *string

	HealthyThreshold   *int64
	UnhealthyThreshold *int64

	Interval *int64
	Timeout  *int64

	Port     *string
	Protocol *string
}

var _ fi.HasDependencies = &LoadBalancerListener{}

func (e *NetworkLoadBalancerHealthCheck) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

func findNLBHealthCheck(cloud awsup.AWSCloud, lb *elbv2.LoadBalancer) (*NetworkLoadBalancerHealthCheck, error) {

	klog.V(2).Infof("Requesting Target Group for NLB with Name:%v", lb.LoadBalancerName)
	request := &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}
	response, err := cloud.ELBV2().DescribeTargetGroups(request)
	if err != nil {
		return nil, fmt.Errorf("error querying for target groups associated with LoadBalancerArn:%+v", lb.LoadBalancerArn)
	}

	if len(response.TargetGroups) != 1 {
		return nil, fmt.Errorf("error wrong # of target groups returned while querying for target groups associated with LoadBalancerArn:%+v", lb.LoadBalancerArn)
	}

	tg := response.TargetGroups[0]

	if lb == nil || tg == nil {
		return nil, nil
	}

	actual := &NetworkLoadBalancerHealthCheck{}
	if tg != nil {
		actual.UnhealthyThreshold = tg.UnhealthyThresholdCount
		actual.HealthyThreshold = tg.HealthyThresholdCount
		actual.Port = tg.HealthCheckPort
	}

	return actual, nil
}
