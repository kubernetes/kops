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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) DescribeTargetGroups(ctx context.Context, request *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTargetGroups %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var tgs []elbv2types.TargetGroup
	for _, tg := range m.TargetGroups {
		match := false
		switch {
		case len(request.TargetGroupArns) > 0:
			for _, name := range request.TargetGroupArns {
				if aws.ToString(tg.description.TargetGroupArn) == name {
					match = true
				}
			}
		case request.LoadBalancerArn != nil:
			if len(tg.description.LoadBalancerArns) > 0 && tg.description.LoadBalancerArns[0] == aws.ToString(request.LoadBalancerArn) {
				match = true
			}
		case len(request.Names) > 0:
			for _, name := range request.Names {
				if aws.ToString(tg.description.TargetGroupName) == name {
					match = true
				}
			}
		default:
			match = true
		}

		if match {
			tgs = append(tgs, tg.description)
		}
	}

	if len(tgs) == 0 && len(request.TargetGroupArns) > 0 || request.LoadBalancerArn != nil {
		return nil, &elbv2types.TargetGroupNotFoundException{}
	}

	return &elbv2.DescribeTargetGroupsOutput{
		TargetGroups: tgs,
	}, nil
}

func (m *MockELBV2) CreateTargetGroup(ctx context.Context, request *elbv2.CreateTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateTargetGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateTargetGroup %v", request)

	tg := elbv2types.TargetGroup{
		TargetGroupName:         request.Name,
		Port:                    request.Port,
		Protocol:                request.Protocol,
		VpcId:                   request.VpcId,
		HealthyThresholdCount:   request.HealthyThresholdCount,
		UnhealthyThresholdCount: request.UnhealthyThresholdCount,
	}

	m.tgCount++
	arn := fmt.Sprintf("arn:aws-test:elasticloadbalancing:us-test-1:000000000000:targetgroup/%v/%v", aws.ToString(request.Name), m.tgCount)
	tg.TargetGroupArn = aws.String(arn)

	if m.TargetGroups == nil {
		m.TargetGroups = make(map[string]*targetGroup)
	}
	if m.Tags == nil {
		m.Tags = make(map[string]elbv2types.TagDescription)
	}

	m.TargetGroups[arn] = &targetGroup{description: tg}
	m.Tags[arn] = elbv2types.TagDescription{
		ResourceArn: aws.String(arn),
		Tags:        request.Tags,
	}
	return &elbv2.CreateTargetGroupOutput{TargetGroups: []elbv2types.TargetGroup{tg}}, nil
}

func (m *MockELBV2) DeleteTargetGroup(ctx context.Context, request *elbv2.DeleteTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteTargetGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteTargetGroup %v", request)

	arn := aws.ToString(request.TargetGroupArn)
	delete(m.TargetGroups, arn)
	return &elbv2.DeleteTargetGroupOutput{}, nil
}

func (m *MockELBV2) DescribeTargetGroupAttributes(ctx context.Context, request *elbv2.DescribeTargetGroupAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTargetGroupAttributes %v", request)

	arn := aws.ToString(request.TargetGroupArn)
	return &elbv2.DescribeTargetGroupAttributesOutput{Attributes: m.TargetGroups[arn].attributes}, nil
}

func (m *MockELBV2) ModifyTargetGroupAttributes(ctx context.Context, request *elbv2.ModifyTargetGroupAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyTargetGroupAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ModifyTargetGroupAttributes %v", request)

	arn := aws.ToString(request.TargetGroupArn)
	m.TargetGroups[arn].attributes = request.Attributes
	return &elbv2.ModifyTargetGroupAttributesOutput{Attributes: request.Attributes}, nil
}
