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

package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

//go:generate fitask -type=LoadBalancerAttachment
type LoadBalancerHealthChecks struct {
	Name         *string
	LoadBalancer *LoadBalancer

	Target *string

	HealthyThreshold   *int64
	UnhealthyThreshold *int64

	Interval *int64
	Timeout  *int64
}

func (e *LoadBalancerHealthChecks) Find(c *fi.Context) (*LoadBalancerHealthChecks, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	elbName := fi.StringValue(e.LoadBalancer.ID)

	lb, err := findLoadBalancer(cloud, elbName)
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	actual := &LoadBalancerHealthChecks{}
	actual.LoadBalancer = e.LoadBalancer

	if lb.HealthCheck != nil {
		actual.Target = lb.HealthCheck.Target
		actual.HealthyThreshold = lb.HealthCheck.HealthyThreshold
		actual.UnhealthyThreshold = lb.HealthCheck.UnhealthyThreshold
		actual.Interval = lb.HealthCheck.Interval
		actual.Timeout = lb.HealthCheck.Timeout
	}

	// Prevent spurious changes
	actual.Name = e.Name

	return actual, nil

}

func (e *LoadBalancerHealthChecks) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *LoadBalancerHealthChecks) CheckChanges(a, e, changes *LoadBalancerHealthChecks) error {
	if a == nil {
		if e.LoadBalancer == nil {
			return fi.RequiredField("LoadBalancer")
		}
		if e.Target == nil {
			return fi.RequiredField("Target")
		}
	}
	return nil
}

func (_ *LoadBalancerHealthChecks) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LoadBalancerHealthChecks) error {
	request := &elb.ConfigureHealthCheckInput{}
	request.LoadBalancerName = e.LoadBalancer.ID
	request.HealthCheck = &elb.HealthCheck{
		Target:             e.Target,
		HealthyThreshold:   e.HealthyThreshold,
		UnhealthyThreshold: e.UnhealthyThreshold,
		Interval:           e.Interval,
		Timeout:            e.Timeout,
	}

	glog.V(2).Infof("Configuring health checks on ELB %q", *e.LoadBalancer.ID)

	_, err := t.Cloud.ELB().ConfigureHealthCheck(request)
	if err != nil {
		return fmt.Errorf("error attaching autoscaling group to ELB: %v", err)
	}

	return nil
}
