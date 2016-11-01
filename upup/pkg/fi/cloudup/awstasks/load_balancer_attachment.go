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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"github.com/aws/aws-sdk-go/service/elb"
)

type LoadBalancerAttachment struct {
	LoadBalancer     *LoadBalancer

	// LoadBalancerAttachments now support ASGs or direct instances
	AutoscalingGroup *AutoscalingGroup
	Instance *Instance
}

func (e *LoadBalancerAttachment) String() string {
	return fi.TaskAsString(e)
}

func (e *LoadBalancerAttachment) Find(c *fi.Context) (*LoadBalancerAttachment, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	// Instance only
	if e.Instance != nil && e.AutoscalingGroup == nil {
		i, err := e.Instance.Find(c)
		if err != nil {
			return nil, fmt.Errorf("unable to find instance: %v", err)
		}
		actual := &LoadBalancerAttachment{}
		actual.LoadBalancer = e.LoadBalancer
		actual.Instance = i
		return actual, nil
	// ASG only
	}else if e.AutoscalingGroup != nil && e.Instance == nil {
		g, err := findAutoscalingGroup(cloud, *e.AutoscalingGroup.Name)
		if err != nil {
			return nil, err
		}
		if g == nil {
			return nil, nil
		}

		for _, name := range g.LoadBalancerNames {
			if aws.StringValue(name) != *e.LoadBalancer.ID {
				continue
			}

			actual := &LoadBalancerAttachment{}
			actual.LoadBalancer = e.LoadBalancer
			actual.AutoscalingGroup = e.AutoscalingGroup
			return actual, nil
		}
	}else{
	// Invalid request
		return nil, fmt.Errorf("Must specify either an instance or an ASG")
	}

	return nil, nil
}

func (e *LoadBalancerAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *LoadBalancerAttachment) CheckChanges(a, e, changes *LoadBalancerAttachment) error {
	if a == nil {
		if e.LoadBalancer == nil {
			return fi.RequiredField("LoadBalancer")
		}
		if e.AutoscalingGroup == nil {
			return fi.RequiredField("AutoscalingGroup")
		}
	}
	return nil
}

func (_ *LoadBalancerAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LoadBalancerAttachment) error {

	if e.AutoscalingGroup != nil && e.Instance == nil {
		request := &autoscaling.AttachLoadBalancersInput{}
		request.AutoScalingGroupName = e.AutoscalingGroup.Name
		request.LoadBalancerNames = []*string{e.LoadBalancer.ID}
		glog.V(2).Infof("Attaching autoscaling group %q to ELB %q", *e.AutoscalingGroup.Name, *e.LoadBalancer.Name)
		_, err := t.Cloud.Autoscaling().AttachLoadBalancers(request)
		if err != nil {
			return fmt.Errorf("error attaching autoscaling group to ELB: %v", err)
		}
	}else if e.AutoscalingGroup == nil && e.Instance != nil {
		request := &elb.RegisterInstancesWithLoadBalancerInput{}
		var instances []*elb.Instance
		i := &elb.Instance{
			InstanceId: e.Instance.ID,
		}
		instances = append(instances, i)
		request.Instances = instances
		_, err := t.Cloud.ELB().RegisterInstancesWithLoadBalancer(request)
		glog.V(2).Infof("Attaching instance %q to ELB %q", *e.AutoscalingGroup.Name, *e.LoadBalancer.Name)
		if err != nil {
			return fmt.Errorf("error attaching instance to ELB: %v", err)
		}
	}
	return nil
}
