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
	"github.com/aws/aws-sdk-go/service/elb"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=LoadBalancerAttachment
type LoadBalancerAttachment struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	LoadBalancer *LoadBalancer

	// LoadBalancerAttachments now support ASGs or direct instances
	AutoscalingGroup *AutoscalingGroup
	Subnet           *Subnet

	// Here be dragons..
	// This will *NOT* unmarshal.. for some reason this pointer is initiated as nil
	// instead of a pointer to Instance with nil members..
	Instance *Instance
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
	} else if e.AutoscalingGroup != nil && e.Instance == nil {
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

		for _, name := range g.LoadBalancerNames {
			if aws.StringValue(name) != *e.LoadBalancer.LoadBalancerName {
				continue
			}

			actual := &LoadBalancerAttachment{}
			actual.LoadBalancer = e.LoadBalancer
			actual.AutoscalingGroup = e.AutoscalingGroup

			// Prevent spurious changes
			actual.Name = e.Name // ELB attachments don't have tags
			actual.Lifecycle = e.Lifecycle

			return actual, nil
		}
	} else {
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
	if e.LoadBalancer == nil {
		return fi.RequiredField("LoadBalancer")
	}
	loadBalancerName := fi.StringValue(e.LoadBalancer.LoadBalancerName)
	if loadBalancerName == "" {
		return fi.RequiredField("LoadBalancer.LoadBalancerName")
	}

	if e.AutoscalingGroup != nil && e.Instance == nil {
		request := &autoscaling.AttachLoadBalancersInput{}
		request.AutoScalingGroupName = e.AutoscalingGroup.Name
		request.LoadBalancerNames = aws.StringSlice([]string{loadBalancerName})

		klog.V(2).Infof("Attaching autoscaling group %q to ELB %q", fi.StringValue(e.AutoscalingGroup.Name), loadBalancerName)
		_, err := t.Cloud.Autoscaling().AttachLoadBalancers(request)
		if err != nil {
			return fmt.Errorf("error attaching autoscaling group to ELB: %v", err)
		}
	} else if e.AutoscalingGroup == nil && e.Instance != nil {
		request := &elb.RegisterInstancesWithLoadBalancerInput{}
		request.Instances = append(request.Instances, &elb.Instance{InstanceId: e.Instance.ID})
		request.LoadBalancerName = aws.String(loadBalancerName)

		klog.V(2).Infof("Attaching instance %q to ELB %q", fi.StringValue(e.Instance.ID), loadBalancerName)
		_, err := t.Cloud.ELB().RegisterInstancesWithLoadBalancer(request)
		if err != nil {
			return fmt.Errorf("error attaching instance to ELB: %v", err)
		}
	}
	return nil
}

type terraformLoadBalancerAttachment struct {
	ELB              *terraform.Literal `json:"elb" cty:"elb"`
	Instance         *terraform.Literal `json:"instance,omitempty" cty:"instance"`
	AutoscalingGroup *terraform.Literal `json:"autoscaling_group_name,omitempty" cty:"autoscaling_group_name"`
}

func (_ *LoadBalancerAttachment) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancerAttachment) error {
	tf := &terraformLoadBalancerAttachment{
		ELB: e.LoadBalancer.TerraformLink(),
	}

	if e.AutoscalingGroup != nil && e.Instance == nil {
		tf.AutoscalingGroup = e.AutoscalingGroup.TerraformLink()
		return t.RenderResource("aws_autoscaling_attachment", *e.AutoscalingGroup.Name, tf)
	} else if e.AutoscalingGroup == nil && e.Instance != nil {
		tf.Instance = e.Instance.TerraformLink()
		return t.RenderResource("aws_elb_attachment", *e.LoadBalancer.Name, tf)
	}
	return nil
}

func (e *LoadBalancerAttachment) TerraformLink() *terraform.Literal {
	if e.AutoscalingGroup != nil && e.Instance == nil {
		return terraform.LiteralProperty("aws_autoscaling_attachment", *e.AutoscalingGroup.Name, "id")
	} else if e.AutoscalingGroup == nil && e.Instance != nil {
		return terraform.LiteralProperty("aws_elb_attachment", *e.LoadBalancer.Name, "id")
	}
	return nil
}

func (_ *LoadBalancerAttachment) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *LoadBalancerAttachment) error {
	if e.AutoscalingGroup != nil {
		cfObj, ok := t.Find(e.AutoscalingGroup.CloudformationLink())
		if !ok {
			// topo-sort fail?
			return fmt.Errorf("AutoScalingGroup not yet rendered")
		}
		cf, ok := cfObj.(*cloudformationAutoscalingGroup)
		if !ok {
			return fmt.Errorf("unexpected type for CF record: %T", cfObj)
		}

		cf.LoadBalancerNames = append(cf.LoadBalancerNames, e.LoadBalancer.CloudformationLink())
	}
	if e.Instance != nil {
		return fmt.Errorf("expected Instance to be nil")
	}
	return nil
}
