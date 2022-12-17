/*
Copyright 2020 The Kubernetes Authors.

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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

const (
	// TargetGroupAttributeDeregistrationDelayConnectionTerminationEnabled indicates whether
	//the load balancer terminates connections at the end of the deregistration timeout.
	// https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#deregistration-delay
	TargetGroupAttributeDeregistrationDelayConnectionTerminationEnabled = "deregistration_delay.connection_termination.enabled"
	// TargetGroupAttributeDeregistrationDelayTimeoutSeconds is the amount of time for Elastic Load Balancing
	// to wait before changing the state of a deregistering target from draining to unused.
	// https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#deregistration-delay
	TargetGroupAttributeDeregistrationDelayTimeoutSeconds = "deregistration_delay.timeout_seconds"
)

// +kops:fitask
type TargetGroup struct {
	Name      *string
	Lifecycle fi.Lifecycle
	VPC       *VPC
	Tags      map[string]string
	Port      *int64
	Protocol  *string

	// ARN is the Amazon Resource Name for the Target Group
	ARN *string

	// Shared is set if this is an external LB (one we don't create or own)
	Shared *bool

	Attributes map[string]string

	Interval           *int64
	HealthyThreshold   *int64
	UnhealthyThreshold *int64
}

var _ fi.CompareWithID = &TargetGroup{}

func (e *TargetGroup) CompareWithID() *string {
	return e.ARN
}

func (e *TargetGroup) Find(c *fi.CloudupContext) (*TargetGroup, error) {
	cloud := c.T.Cloud.(awsup.AWSCloud)

	request := &elbv2.DescribeTargetGroupsInput{}
	if e.ARN != nil {
		request.TargetGroupArns = []*string{e.ARN}
	} else if e.Name != nil {
		request.Names = []*string{e.Name}
	}

	response, err := cloud.ELBV2().DescribeTargetGroups(request)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == elbv2.ErrCodeTargetGroupNotFoundException {
			if !fi.ValueOf(e.Shared) {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("error describing targetgroup %s: %v", *e.Name, err)
	}

	if len(response.TargetGroups) > 1 {
		return nil, fmt.Errorf("found %d TargetGroups with ID %q, expected 1", len(response.TargetGroups), fi.ValueOf(e.Name))
	} else if len(response.TargetGroups) == 0 {
		return nil, nil
	}

	tg := response.TargetGroups[0]

	actual := &TargetGroup{
		Name:               tg.TargetGroupName,
		Port:               tg.Port,
		Protocol:           tg.Protocol,
		ARN:                tg.TargetGroupArn,
		Interval:           tg.HealthCheckIntervalSeconds,
		HealthyThreshold:   tg.HealthyThresholdCount,
		UnhealthyThreshold: tg.UnhealthyThresholdCount,
		VPC:                &VPC{ID: tg.VpcId},
	}
	// Interval cannot be changed after TargetGroup creation
	e.Interval = actual.Interval

	e.ARN = tg.TargetGroupArn

	tagsResp, err := cloud.ELBV2().DescribeTags(&elbv2.DescribeTagsInput{
		ResourceArns: []*string{tg.TargetGroupArn},
	})
	if err != nil {
		return nil, err
	}
	tags := make(map[string]string)
	for _, tagDesc := range tagsResp.TagDescriptions {
		for _, tag := range tagDesc.Tags {
			tags[fi.ValueOf(tag.Key)] = fi.ValueOf(tag.Value)
		}
	}
	actual.Tags = tags

	attrResp, err := cloud.ELBV2().DescribeTargetGroupAttributes(&elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: tg.TargetGroupArn,
	})
	if err != nil {
		return nil, err
	}
	attributes := make(map[string]string)
	for _, attr := range attrResp.Attributes {
		if _, ok := e.Attributes[fi.ValueOf(attr.Key)]; ok {
			attributes[fi.ValueOf(attr.Key)] = fi.ValueOf(attr.Value)
		}
	}
	if len(attributes) > 0 {
		actual.Attributes = attributes
	}

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	return actual, nil
}

func (e *TargetGroup) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *TargetGroup) ShouldCreate(a, e, changes *TargetGroup) (bool, error) {
	if fi.ValueOf(e.Shared) {
		return false, nil
	}
	return true, nil
}

func (s *TargetGroup) CheckChanges(a, e, changes *TargetGroup) error {
	return nil
}

func (_ *TargetGroup) RenderAWS(ctx *fi.CloudupContext, t *awsup.AWSAPITarget, a, e, changes *TargetGroup) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		return nil
	}

	// You register targets for your Network Load Balancer with a target group. By default, the load balancer sends requests
	// to registered targets using the port and protocol that you specified for the target group. You can override this port
	// when you register each target with the target group.

	if a == nil {
		request := &elbv2.CreateTargetGroupInput{
			Name:                       e.Name,
			Port:                       e.Port,
			Protocol:                   e.Protocol,
			VpcId:                      e.VPC.ID,
			HealthCheckIntervalSeconds: e.Interval,
			HealthyThresholdCount:      e.HealthyThreshold,
			UnhealthyThresholdCount:    e.UnhealthyThreshold,
			Tags:                       awsup.ELBv2Tags(e.Tags),
		}

		klog.V(2).Infof("Creating Target Group for NLB")
		response, err := t.Cloud.ELBV2().CreateTargetGroup(request)
		if err != nil {
			return fmt.Errorf("Error creating target group for NLB : %v", err)
		}

		if err := ModifyTargetGroupAttributes(t.Cloud, response.TargetGroups[0].TargetGroupArn, e.Attributes); err != nil {
			return err
		}

		// Avoid spurious changes
		e.ARN = response.TargetGroups[0].TargetGroupArn

	} else {
		if a.ARN != nil {
			if err := t.AddELBV2Tags(fi.ValueOf(a.ARN), e.Tags); err != nil {
				return err
			}
			if err := ModifyTargetGroupAttributes(t.Cloud, a.ARN, e.Attributes); err != nil {
				return err
			}
		}
	}
	return nil
}

func ModifyTargetGroupAttributes(cloud awsup.AWSCloud, arn *string, attributes map[string]string) error {
	klog.V(2).Infof("Modifying Target Group attributes for NLB")
	attrReq := &elbv2.ModifyTargetGroupAttributesInput{
		Attributes:     []*elbv2.TargetGroupAttribute{},
		TargetGroupArn: arn,
	}
	for k, v := range attributes {
		attrReq.Attributes = append(attrReq.Attributes, &elbv2.TargetGroupAttribute{
			Key:   fi.PtrTo(k),
			Value: fi.PtrTo(v),
		})
	}
	if _, err := cloud.ELBV2().ModifyTargetGroupAttributes(attrReq); err != nil {
		return fmt.Errorf("error modifying target group attributes for NLB : %v", err)
	}
	return nil
}

// OrderTargetGroupsByName implements sort.Interface for []OrderTargetGroupsByName, based on port number
type OrderTargetGroupsByName []*TargetGroup

func (a OrderTargetGroupsByName) Len() int      { return len(a) }
func (a OrderTargetGroupsByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderTargetGroupsByName) Less(i, j int) bool {
	return fi.ValueOf(a[i].Name) < fi.ValueOf(a[j].Name)
}

type terraformTargetGroup struct {
	Name                  string                          `cty:"name"`
	Port                  int64                           `cty:"port"`
	Protocol              string                          `cty:"protocol"`
	VPCID                 *terraformWriter.Literal        `cty:"vpc_id"`
	ConnectionTermination string                          `cty:"connection_termination"`
	DeregistrationDelay   string                          `cty:"deregistration_delay"`
	Tags                  map[string]string               `cty:"tags"`
	HealthCheck           terraformTargetGroupHealthCheck `cty:"health_check"`
}

type terraformTargetGroupHealthCheck struct {
	Interval           int64  `cty:"interval"`
	HealthyThreshold   int64  `cty:"healthy_threshold"`
	UnhealthyThreshold int64  `cty:"unhealthy_threshold"`
	Protocol           string `cty:"protocol"`
}

func (_ *TargetGroup) RenderTerraform(ctx *fi.CloudupContext, t *terraform.TerraformTarget, a, e, changes *TargetGroup) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		return nil
	}

	if e.VPC == nil {
		return fmt.Errorf("Missing VPC task from target group:\n%v\n%v", e, e.VPC)
	}

	tf := &terraformTargetGroup{
		Name:     *e.Name,
		Port:     *e.Port,
		Protocol: *e.Protocol,
		VPCID:    e.VPC.TerraformLink(),
		Tags:     e.Tags,
		HealthCheck: terraformTargetGroupHealthCheck{
			Interval:           *e.Interval,
			HealthyThreshold:   *e.HealthyThreshold,
			UnhealthyThreshold: *e.UnhealthyThreshold,
			Protocol:           elbv2.ProtocolEnumTcp,
		},
	}

	for attr, val := range e.Attributes {
		if attr == TargetGroupAttributeDeregistrationDelayConnectionTerminationEnabled {
			tf.ConnectionTermination = val
		}
		if attr == TargetGroupAttributeDeregistrationDelayTimeoutSeconds {
			tf.DeregistrationDelay = val
		}
	}

	return t.RenderResource("aws_lb_target_group", *e.Name, tf)
}

func (e *TargetGroup) TerraformLink() *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		if e.ARN != nil {
			return terraformWriter.LiteralFromStringValue(*e.ARN)
		} else {
			klog.Warningf("ID not set on shared Target Group %v", e)
		}
	}
	return terraformWriter.LiteralProperty("aws_lb_target_group", *e.Name, "id")
}
