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

package mockelbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) DescribeTargetGroups(request *elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTargetGroups %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var tgs []*elbv2.TargetGroup
	for _, tg := range m.TargetGroups {
		match := false

		if len(request.TargetGroupArns) > 0 {
			for _, name := range request.TargetGroupArns {
				if aws.StringValue(tg.description.TargetGroupArn) == aws.StringValue(name) {
					match = true
				}
			}
		} else if request.LoadBalancerArn != nil {
			if len(tg.description.LoadBalancerArns) > 0 && aws.StringValue(tg.description.LoadBalancerArns[0]) == aws.StringValue(request.LoadBalancerArn) {
				match = true
			}
		} else if len(request.Names) > 0 {
			for _, name := range request.Names {
				if aws.StringValue(tg.description.TargetGroupName) == aws.StringValue(name) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			tgs = append(tgs, &tg.description)
		}
	}

	if len(tgs) == 0 && len(request.TargetGroupArns) > 0 || request.LoadBalancerArn != nil {
		return nil, awserr.New(elbv2.ErrCodeTargetGroupNotFoundException, "target group not found", nil)
	}

	return &elbv2.DescribeTargetGroupsOutput{
		TargetGroups: tgs,
	}, nil
}

func (m *MockELBV2) DescribeTargetGroupsPages(request *elbv2.DescribeTargetGroupsInput, callback func(p *elbv2.DescribeTargetGroupsOutput, lastPage bool) (shouldContinue bool)) error {
	page, err := m.DescribeTargetGroups(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockELBV2) CreateTargetGroup(request *elbv2.CreateTargetGroupInput) (*elbv2.CreateTargetGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateTargetGroup %v", request)

	tg := elbv2.TargetGroup{
		TargetGroupName:         request.Name,
		Port:                    request.Port,
		Protocol:                request.Protocol,
		VpcId:                   request.VpcId,
		HealthyThresholdCount:   request.HealthyThresholdCount,
		UnhealthyThresholdCount: request.UnhealthyThresholdCount,
	}

	m.tgCount++
	arn := fmt.Sprintf("arn:aws-test:elasticloadbalancing:us-test-1:000000000000:targetgroup/%v/%v", aws.StringValue(request.Name), m.tgCount)
	tg.TargetGroupArn = aws.String(arn)

	if m.TargetGroups == nil {
		m.TargetGroups = make(map[string]*targetGroup)
	}
	if m.Tags == nil {
		m.Tags = make(map[string]*elbv2.TagDescription)
	}

	m.TargetGroups[arn] = &targetGroup{description: tg}
	m.Tags[arn] = &elbv2.TagDescription{
		ResourceArn: aws.String(arn),
		Tags:        request.Tags,
	}
	return &elbv2.CreateTargetGroupOutput{TargetGroups: []*elbv2.TargetGroup{&tg}}, nil
}

func (m *MockELBV2) DeleteTargetGroup(request *elbv2.DeleteTargetGroupInput) (*elbv2.DeleteTargetGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteTargetGroup %v", request)

	arn := aws.StringValue(request.TargetGroupArn)
	delete(m.TargetGroups, arn)
	return &elbv2.DeleteTargetGroupOutput{}, nil
}

func (m *MockELBV2) DescribeTargetGroupAttributes(request *elbv2.DescribeTargetGroupAttributesInput) (*elbv2.DescribeTargetGroupAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTargetGroupAttributes %v", request)

	arn := aws.StringValue(request.TargetGroupArn)
	return &elbv2.DescribeTargetGroupAttributesOutput{Attributes: m.TargetGroups[arn].attributes}, nil
}

func (m *MockELBV2) ModifyTargetGroupAttributes(request *elbv2.ModifyTargetGroupAttributesInput) (*elbv2.ModifyTargetGroupAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ModifyTargetGroupAttributes %v", request)

	arn := aws.StringValue(request.TargetGroupArn)
	m.TargetGroups[arn].attributes = request.Attributes
	return &elbv2.ModifyTargetGroupAttributesOutput{Attributes: request.Attributes}, nil
}
