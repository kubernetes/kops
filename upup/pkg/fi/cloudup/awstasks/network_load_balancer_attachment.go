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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

//go:generate fitask -type=NetworkLoadBalancerAttachment
type NetworkLoadBalancerAttachment struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	LoadBalancer *NetworkLoadBalancer

	// LoadBalancerAttachments now support ASGs or direct instances
	AutoscalingGroup *AutoscalingGroup
}

func (e *NetworkLoadBalancerAttachment) Find(c *fi.Context) (*NetworkLoadBalancerAttachment, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.AutoscalingGroup != nil {
		if aws.StringValue(e.LoadBalancer.LoadBalancerName) == "" {
			return nil, fmt.Errorf("LoadBalancer did not have LoadBalancerName set")
		}

		g, err := findAutoscalingGroup(cloud, *e.AutoscalingGroup.Name)
		if err != nil {
			return nil, err
		}
		if g == nil {
			return nil, nil
		}

		tg, err := findTargetGroupByLoadBalancerName(cloud, *e.LoadBalancer.Name)
		if err != nil {
			return nil, err
		}
		if tg == nil { //should this return e.AutoscalingGroup w/ e.LoadBalancer?
			return nil, nil
		}

		for _, arn := range g.TargetGroupARNs {
			if aws.StringValue(arn) != *tg.TargetGroupArn {
				continue
			}

			actual := &NetworkLoadBalancerAttachment{}
			actual.LoadBalancer = e.LoadBalancer
			actual.AutoscalingGroup = e.AutoscalingGroup

			// Prevent spurious changes
			actual.Name = e.Name // ELB attachments don't have tags
			actual.Lifecycle = e.Lifecycle

			return actual, nil
		}
	} else {
		// Invalid request
		return nil, fmt.Errorf("Must specify an ASG")
	}

	return nil, nil
}

func (e *NetworkLoadBalancerAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *NetworkLoadBalancerAttachment) CheckChanges(a, e, changes *NetworkLoadBalancerAttachment) error {
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

func (_ *NetworkLoadBalancerAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NetworkLoadBalancerAttachment) error {
	if e.LoadBalancer == nil {
		return fi.RequiredField("LoadBalancer")
	}
	if e.AutoscalingGroup != nil {
		return nil
	}

	loadBalancerName := fi.StringValue(e.LoadBalancer.LoadBalancerName)
	if loadBalancerName == "" {
		return fi.RequiredField("LoadBalancer.LoadBalancerName")
	}

	tg, err := findTargetGroupByLoadBalancerName(t.Cloud, *e.LoadBalancer.Name)
	if err != nil {
		return err
	}
	if tg == nil {
		return nil
	}

	request := &autoscaling.AttachLoadBalancerTargetGroupsInput{}
	request.TargetGroupARNs = []*string{tg.TargetGroupArn}
	request.AutoScalingGroupName = e.AutoscalingGroup.Name

	klog.V(2).Infof("Attaching autoscaling group %q to NLB %q's target group", fi.StringValue(e.AutoscalingGroup.Name), loadBalancerName)
	if _, err = t.Cloud.Autoscaling().AttachLoadBalancerTargetGroups(request); err != nil {
		return fmt.Errorf("error attaching autoscaling group to NLB's target group: %v", err)
	}

	return nil
}
