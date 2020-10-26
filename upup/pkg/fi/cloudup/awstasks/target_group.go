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
	"strings"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// +kops:fitask
type TargetGroup struct {
	Name           *string
	Lifecycle      *fi.Lifecycle
	VPC            *VPC
	Tags           map[string]string
	Port           *int64
	Protocol       *string
	TargetGroupArn *string
}

func (e *TargetGroup) Find(c *fi.Context) (*TargetGroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &elbv2.DescribeTargetGroupsInput{
		Names: []*string{e.Name},
	}

	response, err := cloud.ELBV2().DescribeTargetGroups(request)

	if err != nil {
		if !strings.Contains(err.Error(), "TargetGroupNotFound:") {
			return nil, fmt.Errorf("Error retrieving target group with name %s with err : %v", *e.Name, err)
		}
		return nil, nil
	}

	if len(response.TargetGroups) != 1 {
		return nil, fmt.Errorf("expected api describe target groups response to have 1 tg; instead recieved %v for name = %v:", len(response.TargetGroups), e.Name)
	}

	tg := response.TargetGroups[0]

	actual := &TargetGroup{}
	actual.Port = tg.Port
	actual.Protocol = tg.Protocol

	// Prevent spurious changes
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	if e.TargetGroupArn == nil {
		e.TargetGroupArn = actual.TargetGroupArn
	}

	return actual, nil
}

func (e *TargetGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *TargetGroup) CheckChanges(a, e, changes *TargetGroup) error {
	return nil
}

func (_ *TargetGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *TargetGroup) error {
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

// func (_ *TargetGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *TargetGroup) error {
// 	return nil
// }

// func (_ *TargetGroup) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *TargetGroup) error {
// 	return nil
// }
