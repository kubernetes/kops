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

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

type AutoScalingAPI interface {
	AttachInstances(ctx context.Context, params *autoscaling.AttachInstancesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.AttachInstancesOutput, error)
	AttachLoadBalancers(ctx context.Context, params *autoscaling.AttachLoadBalancersInput, optFns ...func(*autoscaling.Options)) (*autoscaling.AttachLoadBalancersOutput, error)
	AttachLoadBalancerTargetGroups(ctx context.Context, params *autoscaling.AttachLoadBalancerTargetGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error)
	CompleteLifecycleAction(ctx context.Context, params *autoscaling.CompleteLifecycleActionInput, optFns ...func(*autoscaling.Options)) (*autoscaling.CompleteLifecycleActionOutput, error)
	CreateAutoScalingGroup(ctx context.Context, params *autoscaling.CreateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.CreateAutoScalingGroupOutput, error)
	CreateOrUpdateTags(ctx context.Context, params *autoscaling.CreateOrUpdateTagsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.CreateOrUpdateTagsOutput, error)
	DeleteAutoScalingGroup(ctx context.Context, params *autoscaling.DeleteAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteAutoScalingGroupOutput, error)
	DeleteLaunchConfiguration(ctx context.Context, params *autoscaling.DeleteLaunchConfigurationInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteLaunchConfigurationOutput, error)
	DeleteLifecycleHook(ctx context.Context, params *autoscaling.DeleteLifecycleHookInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteLifecycleHookOutput, error)
	DeleteTags(ctx context.Context, params *autoscaling.DeleteTagsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteTagsOutput, error)
	DeleteWarmPool(ctx context.Context, params *autoscaling.DeleteWarmPoolInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteWarmPoolOutput, error)
	DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	DescribeLifecycleHooks(ctx context.Context, params *autoscaling.DescribeLifecycleHooksInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeLifecycleHooksOutput, error)
	DescribeTags(ctx context.Context, params *autoscaling.DescribeTagsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeTagsOutput, error)
	DescribeWarmPool(ctx context.Context, params *autoscaling.DescribeWarmPoolInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeWarmPoolOutput, error)
	DetachInstances(ctx context.Context, params *autoscaling.DetachInstancesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DetachInstancesOutput, error)
	DetachLoadBalancers(ctx context.Context, params *autoscaling.DetachLoadBalancersInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DetachLoadBalancersOutput, error)
	DetachLoadBalancerTargetGroups(ctx context.Context, params *autoscaling.DetachLoadBalancerTargetGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error)
	EnableMetricsCollection(ctx context.Context, params *autoscaling.EnableMetricsCollectionInput, optFns ...func(*autoscaling.Options)) (*autoscaling.EnableMetricsCollectionOutput, error)
	PutLifecycleHook(ctx context.Context, params *autoscaling.PutLifecycleHookInput, optFns ...func(*autoscaling.Options)) (*autoscaling.PutLifecycleHookOutput, error)
	PutWarmPool(ctx context.Context, params *autoscaling.PutWarmPoolInput, optFns ...func(*autoscaling.Options)) (*autoscaling.PutWarmPoolOutput, error)
	ResumeProcesses(ctx context.Context, params *autoscaling.ResumeProcessesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.ResumeProcessesOutput, error)
	SuspendProcesses(ctx context.Context, params *autoscaling.SuspendProcessesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.SuspendProcessesOutput, error)
	TerminateInstanceInAutoScalingGroup(ctx context.Context, params *autoscaling.TerminateInstanceInAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
	UpdateAutoScalingGroup(ctx context.Context, params *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error)
}
