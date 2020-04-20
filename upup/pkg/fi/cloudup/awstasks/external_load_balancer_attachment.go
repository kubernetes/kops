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
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=ExternalLoadBalancerAttachment
type ExternalLoadBalancerAttachment struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	LoadBalancerName string

	AutoscalingGroup *AutoscalingGroup
}

func (e *ExternalLoadBalancerAttachment) Find(c *fi.Context) (*ExternalLoadBalancerAttachment, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.LoadBalancerName == "" {
		return nil, fmt.Errorf("InstanceGroup did not have LoadBalancerNames set")
	}

	g, err := findAutoscalingGroup(cloud, *e.AutoscalingGroup.Name)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}

	for _, name := range g.LoadBalancerNames {
		if aws.StringValue(name) != e.LoadBalancerName {
			continue
		}

		actual := &ExternalLoadBalancerAttachment{}
		actual.LoadBalancerName = e.LoadBalancerName
		actual.AutoscalingGroup = e.AutoscalingGroup

		// Prevent spurious changes
		actual.Name = e.Name // ELB attachments don't have tags
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *ExternalLoadBalancerAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *ExternalLoadBalancerAttachment) CheckChanges(a, e, changes *ExternalLoadBalancerAttachment) error {
	if a == nil {
		if e.LoadBalancerName == "" {
			return fi.RequiredField("LoadBalancerName")
		}
		if e.AutoscalingGroup == nil {
			return fi.RequiredField("AutoscalingGroup")
		}
	}
	return nil
}

func (_ *ExternalLoadBalancerAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *ExternalLoadBalancerAttachment) error {
	if e.LoadBalancerName == "" {
		return fi.RequiredField("LoadBalancerName")
	}

	request := &autoscaling.AttachLoadBalancersInput{}
	request.AutoScalingGroupName = e.AutoscalingGroup.Name
	request.LoadBalancerNames = aws.StringSlice([]string{e.LoadBalancerName})

	klog.V(2).Infof("Attaching autoscaling group %q to ELB %q", fi.StringValue(e.AutoscalingGroup.Name), e.LoadBalancerName)
	_, err := t.Cloud.Autoscaling().AttachLoadBalancers(request)
	if err != nil {
		return fmt.Errorf("error attaching autoscaling group to ELB: %v", err)
	}

	return nil
}

type terraformExternalLoadBalancerAttachment struct {
	ELB              *terraform.Literal `json:"elb" cty:"elb"`
	AutoscalingGroup *terraform.Literal `json:"autoscaling_group_name,omitempty" cty:"autoscaling_group_name"`
}

func (_ *ExternalLoadBalancerAttachment) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ExternalLoadBalancerAttachment) error {
	tf := &terraformExternalLoadBalancerAttachment{
		ELB:              terraform.LiteralFromStringValue(e.LoadBalancerName),
		AutoscalingGroup: e.AutoscalingGroup.TerraformLink(),
	}

	return t.RenderResource("aws_autoscaling_attachment", *e.Name, tf)
}

func (e *ExternalLoadBalancerAttachment) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_autoscaling_attachment", e.LoadBalancerName, "id")
}

func (_ *ExternalLoadBalancerAttachment) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *ExternalLoadBalancerAttachment) error {
	cfObj, ok := t.Find(e.AutoscalingGroup.CloudformationLink())
	if !ok {
		// topo-sort fail?
		return fmt.Errorf("AutoScalingGroup not yet rendered")
	}
	cf, ok := cfObj.(*cloudformationAutoscalingGroup)
	if !ok {
		return fmt.Errorf("unexpected type for CF record: %T", cfObj)
	}

	cf.LoadBalancerNames = append(cf.LoadBalancerNames, cloudformation.LiteralString(e.LoadBalancerName))
	return nil
}
