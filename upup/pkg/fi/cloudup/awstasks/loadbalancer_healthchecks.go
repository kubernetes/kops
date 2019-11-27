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
	"github.com/aws/aws-sdk-go/service/elb"
	"k8s.io/kops/upup/pkg/fi"
)

type LoadBalancerHealthCheck struct {
	Target *string

	HealthyThreshold   *int64
	UnhealthyThreshold *int64

	Interval *int64
	Timeout  *int64
}

var _ fi.HasDependencies = &LoadBalancerListener{}

func (e *LoadBalancerHealthCheck) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

func findHealthCheck(lb *elb.LoadBalancerDescription) (*LoadBalancerHealthCheck, error) {
	if lb == nil || lb.HealthCheck == nil {
		return nil, nil
	}

	actual := &LoadBalancerHealthCheck{}
	if lb.HealthCheck != nil {
		actual.Target = lb.HealthCheck.Target
		actual.HealthyThreshold = lb.HealthCheck.HealthyThreshold
		actual.UnhealthyThreshold = lb.HealthCheck.UnhealthyThreshold
		actual.Interval = lb.HealthCheck.Interval
		actual.Timeout = lb.HealthCheck.Timeout
	}

	return actual, nil
}
