/*
Copyright 2024 The Kubernetes Authors.

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

package awsinterfaces

import (
	"context"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type ELBV2API interface {
	AddTags(ctx context.Context, input *elbv2.AddTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.AddTagsOutput, error)
	CreateListener(ctx context.Context, input *elbv2.CreateListenerInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateListenerOutput, error)
	CreateLoadBalancer(ctx context.Context, input *elbv2.CreateLoadBalancerInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateLoadBalancerOutput, error)
	CreateTargetGroup(ctx context.Context, input *elbv2.CreateTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateTargetGroupOutput, error)
	DeleteListener(ctx context.Context, input *elbv2.DeleteListenerInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteListenerOutput, error)
	DeleteLoadBalancer(ctx context.Context, input *elbv2.DeleteLoadBalancerInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteLoadBalancerOutput, error)
	DeleteTargetGroup(ctx context.Context, input *elbv2.DeleteTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteTargetGroupOutput, error)
	DeregisterTargets(ctx context.Context, input *elbv2.DeregisterTargetsInput, optFns ...func(*elbv2.Options)) (*elbv2.DeregisterTargetsOutput, error)
	DescribeListeners(ctx context.Context, input *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error)
	DescribeLoadBalancerAttributes(ctx context.Context, input *elbv2.DescribeLoadBalancerAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancerAttributesOutput, error)
	DescribeLoadBalancers(ctx context.Context, input *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(ctx context.Context, input *elbv2.DescribeTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTagsOutput, error)
	DescribeTargetGroupAttributes(ctx context.Context, input *elbv2.DescribeTargetGroupAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupAttributesOutput, error)
	DescribeTargetGroups(ctx context.Context, input *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealth(ctx context.Context, input *elbv2.DescribeTargetHealthInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetHealthOutput, error)
	ModifyLoadBalancerAttributes(ctx context.Context, input *elbv2.ModifyLoadBalancerAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyLoadBalancerAttributesOutput, error)
	ModifyTargetGroupAttributes(ctx context.Context, input *elbv2.ModifyTargetGroupAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyTargetGroupAttributesOutput, error)
	RemoveTags(ctx context.Context, input *elbv2.RemoveTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.RemoveTagsOutput, error)
	SetIpAddressType(ctx context.Context, input *elbv2.SetIpAddressTypeInput, optFns ...func(*elbv2.Options)) (*elbv2.SetIpAddressTypeOutput, error)
	SetSecurityGroups(ctx context.Context, input *elbv2.SetSecurityGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.SetSecurityGroupsOutput, error)
	SetSubnets(ctx context.Context, input *elbv2.SetSubnetsInput, optFns ...func(*elbv2.Options)) (*elbv2.SetSubnetsOutput, error)
}
