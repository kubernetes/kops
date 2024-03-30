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

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

type ELBAPI interface {
	AddTags(ctx context.Context, params *elb.AddTagsInput, optFns ...func(*elb.Options)) (*elb.AddTagsOutput, error)
	ApplySecurityGroupsToLoadBalancer(ctx context.Context, params *elb.ApplySecurityGroupsToLoadBalancerInput, optFns ...func(*elb.Options)) (*elb.ApplySecurityGroupsToLoadBalancerOutput, error)
	AttachLoadBalancerToSubnets(ctx context.Context, params *elb.AttachLoadBalancerToSubnetsInput, optFns ...func(*elb.Options)) (*elb.AttachLoadBalancerToSubnetsOutput, error)
	ConfigureHealthCheck(ctx context.Context, params *elb.ConfigureHealthCheckInput, optFns ...func(*elb.Options)) (*elb.ConfigureHealthCheckOutput, error)
	CreateLoadBalancer(ctx context.Context, params *elb.CreateLoadBalancerInput, optFns ...func(*elb.Options)) (*elb.CreateLoadBalancerOutput, error)
	CreateLoadBalancerListeners(ctx context.Context, params *elb.CreateLoadBalancerListenersInput, optFns ...func(*elb.Options)) (*elb.CreateLoadBalancerListenersOutput, error)
	DeleteLoadBalancer(ctx context.Context, params *elb.DeleteLoadBalancerInput, optFns ...func(*elb.Options)) (*elb.DeleteLoadBalancerOutput, error)
	DeleteLoadBalancerListeners(ctx context.Context, params *elb.DeleteLoadBalancerListenersInput, optFns ...func(*elb.Options)) (*elb.DeleteLoadBalancerListenersOutput, error)
	DeregisterInstancesFromLoadBalancer(ctx context.Context, params *elb.DeregisterInstancesFromLoadBalancerInput, optFns ...func(*elb.Options)) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	DescribeInstanceHealth(ctx context.Context, params *elb.DescribeInstanceHealthInput, optFns ...func(*elb.Options)) (*elb.DescribeInstanceHealthOutput, error)
	DescribeLoadBalancerAttributes(ctx context.Context, params *elb.DescribeLoadBalancerAttributesInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancerAttributesOutput, error)
	DescribeLoadBalancers(ctx context.Context, params *elb.DescribeLoadBalancersInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTags(ctx context.Context, params *elb.DescribeTagsInput, optFns ...func(*elb.Options)) (*elb.DescribeTagsOutput, error)
	DetachLoadBalancerFromSubnets(ctx context.Context, params *elb.DetachLoadBalancerFromSubnetsInput, optFns ...func(*elb.Options)) (*elb.DetachLoadBalancerFromSubnetsOutput, error)
	ModifyLoadBalancerAttributes(ctx context.Context, params *elb.ModifyLoadBalancerAttributesInput, optFns ...func(*elb.Options)) (*elb.ModifyLoadBalancerAttributesOutput, error)
	RemoveTags(ctx context.Context, params *elb.RemoveTagsInput, optFns ...func(*elb.Options)) (*elb.RemoveTagsOutput, error)
}
