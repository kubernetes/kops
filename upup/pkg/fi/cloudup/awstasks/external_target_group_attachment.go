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
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=ExternalTargetGroupAttachment
type ExternalTargetGroupAttachment struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	TargetGroupARN string

	AutoscalingGroup *AutoscalingGroup
}

func (e *ExternalTargetGroupAttachment) name() string {
	re := regexp.MustCompile("arn:.*:targetgroup/")
	return re.ReplaceAllString(*e.Name, "")
}

func (e *ExternalTargetGroupAttachment) Find(c *fi.Context) (*ExternalTargetGroupAttachment, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.TargetGroupARN == "" {
		return nil, fmt.Errorf("InstanceGroup did not have TargetGroupARNs set")
	}

	g, err := findAutoscalingGroup(cloud, *e.AutoscalingGroup.Name)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}

	for _, name := range g.TargetGroupARNs {
		if aws.StringValue(name) != e.TargetGroupARN {
			continue
		}

		actual := &ExternalTargetGroupAttachment{}
		actual.TargetGroupARN = e.TargetGroupARN
		actual.AutoscalingGroup = e.AutoscalingGroup

		// Prevent spurious changes
		actual.Name = e.Name // ELB attachments don't have tags
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *ExternalTargetGroupAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *ExternalTargetGroupAttachment) CheckChanges(a, e, changes *ExternalTargetGroupAttachment) error {
	if a == nil {
		if e.TargetGroupARN == "" {
			return fi.RequiredField("TargetGroupARN")
		}
		if e.AutoscalingGroup == nil {
			return fi.RequiredField("AutoscalingGroup")
		}
	}
	return nil
}

func (_ *ExternalTargetGroupAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *ExternalTargetGroupAttachment) error {
	if e.TargetGroupARN == "" {
		return fi.RequiredField("TargetGroupARN")
	}

	request := &autoscaling.AttachLoadBalancerTargetGroupsInput{}
	request.AutoScalingGroupName = e.AutoscalingGroup.Name
	request.TargetGroupARNs = aws.StringSlice([]string{e.TargetGroupARN})

	klog.V(2).Infof("Attaching autoscaling group %q to Target Group %q", fi.StringValue(e.AutoscalingGroup.Name), e.TargetGroupARN)
	_, err := t.Cloud.Autoscaling().AttachLoadBalancerTargetGroups(request)
	if err != nil {
		return fmt.Errorf("error attaching autoscaling group to ELB: %v", err)
	}

	return nil
}

type terraformExternalTargetGroupAttachment struct {
	TargetGroupARN   *terraform.Literal `json:"alb_target_group_arn,omitempty"`
	AutoscalingGroup *terraform.Literal `json:"autoscaling_group_name,omitempty"`
}

func (_ *ExternalTargetGroupAttachment) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ExternalTargetGroupAttachment) error {
	tf := &terraformExternalTargetGroupAttachment{
		TargetGroupARN:   terraform.LiteralFromStringValue(e.TargetGroupARN),
		AutoscalingGroup: e.AutoscalingGroup.TerraformLink(),
	}

	return t.RenderResource("aws_autoscaling_attachment", e.name(), tf)
}

func (e *ExternalTargetGroupAttachment) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_autoscaling_attachment", e.name(), "id")
}

func (_ *ExternalTargetGroupAttachment) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *ExternalTargetGroupAttachment) error {
	cfObj, ok := t.Find(e.AutoscalingGroup.CloudformationLink())
	if !ok {
		// topo-sort fail?
		return fmt.Errorf("AutoScalingGroup not yet rendered")
	}
	cf, ok := cfObj.(*cloudformationAutoscalingGroup)
	if !ok {
		return fmt.Errorf("unexpected type for CF record: %T", cfObj)
	}

	cf.TargetGroupARNs = append(cf.TargetGroupARNs, cloudformation.LiteralString(e.TargetGroupARN))
	return nil
}
