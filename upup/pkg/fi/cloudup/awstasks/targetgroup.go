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

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type TargetGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle
	VPC       *VPC
	Tags      map[string]string
	Port      *int64
	Protocol  *string

	// ARN is the Amazon Resource Name for the Target Group
	ARN *string

	// Shared is set if this is an external LB (one we don't create or own)
	Shared *bool
}

var _ fi.CompareWithID = &TargetGroup{}

func (e *TargetGroup) CompareWithID() *string {
	return e.ARN
}

func (e *TargetGroup) Find(c *fi.Context) (*TargetGroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &elbv2.DescribeTargetGroupsInput{}
	if e.ARN != nil {
		request.TargetGroupArns = []*string{e.ARN}
	}
	if e.Name != nil {
		request.Names = []*string{e.Name}
	}

	response, err := cloud.ELBV2().DescribeTargetGroups(request)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == elbv2.ErrCodeTargetGroupNotFoundException {
			if !fi.BoolValue(e.Shared) {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("error describing targetgroup %s: %v", *e.Name, err)
	}

	if len(response.TargetGroups) != 1 {
		return nil, fmt.Errorf("found %d TargetGroups with ID %q, expected 1", len(response.TargetGroups), fi.StringValue(e.Name))
	}

	tg := response.TargetGroups[0]

	actual := &TargetGroup{}
	actual.Port = tg.Port
	actual.Protocol = tg.Protocol
	actual.ARN = tg.TargetGroupArn

	// Prevent spurious changes
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *TargetGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *TargetGroup) ShouldCreate(a, e, changes *TargetGroup) (bool, error) {
	if fi.BoolValue(e.Shared) {
		return false, nil
	}
	return true, nil
}

func (s *TargetGroup) CheckChanges(a, e, changes *TargetGroup) error {
	return nil
}

func (_ *TargetGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *TargetGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		return nil
	}

	//You register targets for your Network Load Balancer with a target group. By default, the load balancer sends requests
	//to registered targets using the port and protocol that you specified for the target group. You can override this port
	//when you register each target with the target group.

	if a == nil {
		request := &elbv2.CreateTargetGroupInput{
			Name:     e.Name,
			Port:     e.Port,
			Protocol: e.Protocol,
			VpcId:    e.VPC.ID,
		}

		klog.V(2).Infof("Creating Target Group for NLB")
		response, err := t.Cloud.ELBV2().CreateTargetGroup(request)
		if err != nil {
			return fmt.Errorf("Error creating target group for NLB : %v", err)
		}

		targetGroupArn := *response.TargetGroups[0].TargetGroupArn

		if err := t.AddELBV2Tags(targetGroupArn, e.Tags); err != nil {
			return err
		}
	}
	return nil
}

func (_ *TargetGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *TargetGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		return nil
	}

	return fmt.Errorf("non shared Target Groups is not yet supported")
}

func (e *TargetGroup) TerraformLink(params ...string) *terraform.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ARN != nil {
			return terraform.LiteralFromStringValue(*e.ARN)
		} else {
			klog.Warningf("ID not set on shared Target Group %v", e)
		}
	}
	return terraform.LiteralProperty("aws_target_group", *e.Name, "id")
}

func (_ *TargetGroup) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *TargetGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		return nil
	}
	return fmt.Errorf("non shared Target Groups is not yet supported")
}

func (e *TargetGroup) CloudformationLink() *cloudformation.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ARN != nil {
			return cloudformation.LiteralString(*e.ARN)
		} else {
			klog.Warningf("ID not set on shared Target Group: %v", e)
		}
	}

	return cloudformation.Ref("AWS::ElasticLoadBalancingV2::TargetGroup", *e.Name)
}
