/*
Copyright 2016 The Kubernetes Authors.

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

package mockautoscaling

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
)

func (m *MockAutoscaling) AttachInstancesWithContext(aws.Context, *autoscaling.AttachInstancesInput, ...request.Option) (*autoscaling.AttachInstancesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) AttachInstancesRequest(*autoscaling.AttachInstancesInput) (*request.Request, *autoscaling.AttachInstancesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) AttachLoadBalancerTargetGroups(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) AttachLoadBalancerTargetGroupsWithContext(aws.Context, *autoscaling.AttachLoadBalancerTargetGroupsInput, ...request.Option) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) AttachLoadBalancerTargetGroupsRequest(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.AttachLoadBalancerTargetGroupsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CompleteLifecycleAction(*autoscaling.CompleteLifecycleActionInput) (*autoscaling.CompleteLifecycleActionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CompleteLifecycleActionWithContext(aws.Context, *autoscaling.CompleteLifecycleActionInput, ...request.Option) (*autoscaling.CompleteLifecycleActionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CompleteLifecycleActionRequest(*autoscaling.CompleteLifecycleActionInput) (*request.Request, *autoscaling.CompleteLifecycleActionOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateAutoScalingGroupWithContext(aws.Context, *autoscaling.CreateAutoScalingGroupInput, ...request.Option) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CreateAutoScalingGroupRequest(*autoscaling.CreateAutoScalingGroupInput) (*request.Request, *autoscaling.CreateAutoScalingGroupOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateOrUpdateTags(*autoscaling.CreateOrUpdateTagsInput) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CreateOrUpdateTagsWithContext(aws.Context, *autoscaling.CreateOrUpdateTagsInput, ...request.Option) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CreateOrUpdateTagsRequest(*autoscaling.CreateOrUpdateTagsInput) (*request.Request, *autoscaling.CreateOrUpdateTagsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteAutoScalingGroup(*autoscaling.DeleteAutoScalingGroupInput) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteAutoScalingGroupWithContext(aws.Context, *autoscaling.DeleteAutoScalingGroupInput, ...request.Option) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteAutoScalingGroupRequest(*autoscaling.DeleteAutoScalingGroupInput) (*request.Request, *autoscaling.DeleteAutoScalingGroupOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLaunchConfiguration(*autoscaling.DeleteLaunchConfigurationInput) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteLaunchConfigurationWithContext(aws.Context, *autoscaling.DeleteLaunchConfigurationInput, ...request.Option) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteLaunchConfigurationRequest(*autoscaling.DeleteLaunchConfigurationInput) (*request.Request, *autoscaling.DeleteLaunchConfigurationOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLifecycleHook(*autoscaling.DeleteLifecycleHookInput) (*autoscaling.DeleteLifecycleHookOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteLifecycleHookWithContext(aws.Context, *autoscaling.DeleteLifecycleHookInput, ...request.Option) (*autoscaling.DeleteLifecycleHookOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteLifecycleHookRequest(*autoscaling.DeleteLifecycleHookInput) (*request.Request, *autoscaling.DeleteLifecycleHookOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteNotificationConfiguration(*autoscaling.DeleteNotificationConfigurationInput) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteNotificationConfigurationWithContext(aws.Context, *autoscaling.DeleteNotificationConfigurationInput, ...request.Option) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteNotificationConfigurationRequest(*autoscaling.DeleteNotificationConfigurationInput) (*request.Request, *autoscaling.DeleteNotificationConfigurationOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeletePolicy(*autoscaling.DeletePolicyInput) (*autoscaling.DeletePolicyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeletePolicyWithContext(aws.Context, *autoscaling.DeletePolicyInput, ...request.Option) (*autoscaling.DeletePolicyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeletePolicyRequest(*autoscaling.DeletePolicyInput) (*request.Request, *autoscaling.DeletePolicyOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteScheduledAction(*autoscaling.DeleteScheduledActionInput) (*autoscaling.DeleteScheduledActionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteScheduledActionWithContext(aws.Context, *autoscaling.DeleteScheduledActionInput, ...request.Option) (*autoscaling.DeleteScheduledActionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteScheduledActionRequest(*autoscaling.DeleteScheduledActionInput) (*request.Request, *autoscaling.DeleteScheduledActionOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteTags(*autoscaling.DeleteTagsInput) (*autoscaling.DeleteTagsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteTagsWithContext(aws.Context, *autoscaling.DeleteTagsInput, ...request.Option) (*autoscaling.DeleteTagsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteTagsRequest(*autoscaling.DeleteTagsInput) (*request.Request, *autoscaling.DeleteTagsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAccountLimits(*autoscaling.DescribeAccountLimitsInput) (*autoscaling.DescribeAccountLimitsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAccountLimitsWithContext(aws.Context, *autoscaling.DescribeAccountLimitsInput, ...request.Option) (*autoscaling.DescribeAccountLimitsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAccountLimitsRequest(*autoscaling.DescribeAccountLimitsInput) (*request.Request, *autoscaling.DescribeAccountLimitsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAdjustmentTypes(*autoscaling.DescribeAdjustmentTypesInput) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAdjustmentTypesWithContext(aws.Context, *autoscaling.DescribeAdjustmentTypesInput, ...request.Option) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAdjustmentTypesRequest(*autoscaling.DescribeAdjustmentTypesInput) (*request.Request, *autoscaling.DescribeAdjustmentTypesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingInstances(*autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAutoScalingInstancesWithContext(aws.Context, *autoscaling.DescribeAutoScalingInstancesInput, ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAutoScalingInstancesRequest(*autoscaling.DescribeAutoScalingInstancesInput) (*request.Request, *autoscaling.DescribeAutoScalingInstancesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingInstancesPages(*autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) DescribeAutoScalingInstancesPagesWithContext(aws.Context, *autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool, ...request.Option) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeAutoScalingNotificationTypes(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAutoScalingNotificationTypesWithContext(aws.Context, *autoscaling.DescribeAutoScalingNotificationTypesInput, ...request.Option) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAutoScalingNotificationTypesRequest(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*request.Request, *autoscaling.DescribeAutoScalingNotificationTypesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLifecycleHookTypes(*autoscaling.DescribeLifecycleHookTypesInput) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLifecycleHookTypesWithContext(aws.Context, *autoscaling.DescribeLifecycleHookTypesInput, ...request.Option) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLifecycleHookTypesRequest(*autoscaling.DescribeLifecycleHookTypesInput) (*request.Request, *autoscaling.DescribeLifecycleHookTypesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLifecycleHooks(*autoscaling.DescribeLifecycleHooksInput) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLifecycleHooksWithContext(aws.Context, *autoscaling.DescribeLifecycleHooksInput, ...request.Option) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLifecycleHooksRequest(*autoscaling.DescribeLifecycleHooksInput) (*request.Request, *autoscaling.DescribeLifecycleHooksOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLoadBalancerTargetGroups(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*autoscaling.DescribeLoadBalancerTargetGroupsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLoadBalancerTargetGroupsWithContext(aws.Context, *autoscaling.DescribeLoadBalancerTargetGroupsInput, ...request.Option) (*autoscaling.DescribeLoadBalancerTargetGroupsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLoadBalancerTargetGroupsRequest(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DescribeLoadBalancerTargetGroupsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLoadBalancers(*autoscaling.DescribeLoadBalancersInput) (*autoscaling.DescribeLoadBalancersOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLoadBalancersWithContext(aws.Context, *autoscaling.DescribeLoadBalancersInput, ...request.Option) (*autoscaling.DescribeLoadBalancersOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLoadBalancersRequest(*autoscaling.DescribeLoadBalancersInput) (*request.Request, *autoscaling.DescribeLoadBalancersOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeMetricCollectionTypes(*autoscaling.DescribeMetricCollectionTypesInput) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeMetricCollectionTypesWithContext(aws.Context, *autoscaling.DescribeMetricCollectionTypesInput, ...request.Option) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeMetricCollectionTypesRequest(*autoscaling.DescribeMetricCollectionTypesInput) (*request.Request, *autoscaling.DescribeMetricCollectionTypesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeNotificationConfigurations(*autoscaling.DescribeNotificationConfigurationsInput) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeNotificationConfigurationsWithContext(aws.Context, *autoscaling.DescribeNotificationConfigurationsInput, ...request.Option) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeNotificationConfigurationsRequest(*autoscaling.DescribeNotificationConfigurationsInput) (*request.Request, *autoscaling.DescribeNotificationConfigurationsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeNotificationConfigurationsPages(*autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) DescribeNotificationConfigurationsPagesWithContext(aws.Context, *autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool, ...request.Option) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribePolicies(*autoscaling.DescribePoliciesInput) (*autoscaling.DescribePoliciesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribePoliciesWithContext(aws.Context, *autoscaling.DescribePoliciesInput, ...request.Option) (*autoscaling.DescribePoliciesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribePoliciesRequest(*autoscaling.DescribePoliciesInput) (*request.Request, *autoscaling.DescribePoliciesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribePoliciesPages(*autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) DescribePoliciesPagesWithContext(aws.Context, *autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool, ...request.Option) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeScalingActivitiesWithContext(aws.Context, *autoscaling.DescribeScalingActivitiesInput, ...request.Option) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeScalingActivitiesRequest(*autoscaling.DescribeScalingActivitiesInput) (*request.Request, *autoscaling.DescribeScalingActivitiesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScalingActivitiesPages(*autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) DescribeScalingActivitiesPagesWithContext(aws.Context, *autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool, ...request.Option) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeScalingProcessTypes(*autoscaling.DescribeScalingProcessTypesInput) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeScalingProcessTypesWithContext(aws.Context, *autoscaling.DescribeScalingProcessTypesInput, ...request.Option) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeScalingProcessTypesRequest(*autoscaling.DescribeScalingProcessTypesInput) (*request.Request, *autoscaling.DescribeScalingProcessTypesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScheduledActions(*autoscaling.DescribeScheduledActionsInput) (*autoscaling.DescribeScheduledActionsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeScheduledActionsWithContext(aws.Context, *autoscaling.DescribeScheduledActionsInput, ...request.Option) (*autoscaling.DescribeScheduledActionsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeScheduledActionsRequest(*autoscaling.DescribeScheduledActionsInput) (*request.Request, *autoscaling.DescribeScheduledActionsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScheduledActionsPages(*autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) DescribeScheduledActionsPagesWithContext(aws.Context, *autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool, ...request.Option) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeTags(*autoscaling.DescribeTagsInput) (*autoscaling.DescribeTagsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeTagsWithContext(aws.Context, *autoscaling.DescribeTagsInput, ...request.Option) (*autoscaling.DescribeTagsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeTagsRequest(*autoscaling.DescribeTagsInput) (*request.Request, *autoscaling.DescribeTagsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeTagsPages(*autoscaling.DescribeTagsInput, func(*autoscaling.DescribeTagsOutput, bool) bool) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) DescribeTagsPagesWithContext(aws.Context, *autoscaling.DescribeTagsInput, func(*autoscaling.DescribeTagsOutput, bool) bool, ...request.Option) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeTerminationPolicyTypes(*autoscaling.DescribeTerminationPolicyTypesInput) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeTerminationPolicyTypesWithContext(aws.Context, *autoscaling.DescribeTerminationPolicyTypesInput, ...request.Option) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeTerminationPolicyTypesRequest(*autoscaling.DescribeTerminationPolicyTypesInput) (*request.Request, *autoscaling.DescribeTerminationPolicyTypesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachInstances(*autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DetachInstancesWithContext(aws.Context, *autoscaling.DetachInstancesInput, ...request.Option) (*autoscaling.DetachInstancesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DetachInstancesRequest(*autoscaling.DetachInstancesInput) (*request.Request, *autoscaling.DetachInstancesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachLoadBalancerTargetGroups(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DetachLoadBalancerTargetGroupsWithContext(aws.Context, *autoscaling.DetachLoadBalancerTargetGroupsInput, ...request.Option) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DetachLoadBalancerTargetGroupsRequest(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DetachLoadBalancerTargetGroupsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachLoadBalancers(*autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DetachLoadBalancersWithContext(aws.Context, *autoscaling.DetachLoadBalancersInput, ...request.Option) (*autoscaling.DetachLoadBalancersOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DetachLoadBalancersRequest(*autoscaling.DetachLoadBalancersInput) (*request.Request, *autoscaling.DetachLoadBalancersOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DisableMetricsCollection(*autoscaling.DisableMetricsCollectionInput) (*autoscaling.DisableMetricsCollectionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DisableMetricsCollectionWithContext(aws.Context, *autoscaling.DisableMetricsCollectionInput, ...request.Option) (*autoscaling.DisableMetricsCollectionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DisableMetricsCollectionRequest(*autoscaling.DisableMetricsCollectionInput) (*request.Request, *autoscaling.DisableMetricsCollectionOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) EnableMetricsCollectionWithContext(aws.Context, *autoscaling.EnableMetricsCollectionInput, ...request.Option) (*autoscaling.EnableMetricsCollectionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) EnableMetricsCollectionRequest(*autoscaling.EnableMetricsCollectionInput) (*request.Request, *autoscaling.EnableMetricsCollectionOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) EnterStandby(*autoscaling.EnterStandbyInput) (*autoscaling.EnterStandbyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) EnterStandbyWithContext(aws.Context, *autoscaling.EnterStandbyInput, ...request.Option) (*autoscaling.EnterStandbyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) EnterStandbyRequest(*autoscaling.EnterStandbyInput) (*request.Request, *autoscaling.EnterStandbyOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ExecutePolicy(*autoscaling.ExecutePolicyInput) (*autoscaling.ExecutePolicyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) ExecutePolicyWithContext(aws.Context, *autoscaling.ExecutePolicyInput, ...request.Option) (*autoscaling.ExecutePolicyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) ExecutePolicyRequest(*autoscaling.ExecutePolicyInput) (*request.Request, *autoscaling.ExecutePolicyOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ExitStandby(*autoscaling.ExitStandbyInput) (*autoscaling.ExitStandbyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) ExitStandbyWithContext(aws.Context, *autoscaling.ExitStandbyInput, ...request.Option) (*autoscaling.ExitStandbyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) ExitStandbyRequest(*autoscaling.ExitStandbyInput) (*request.Request, *autoscaling.ExitStandbyOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutLifecycleHook(*autoscaling.PutLifecycleHookInput) (*autoscaling.PutLifecycleHookOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutLifecycleHookWithContext(aws.Context, *autoscaling.PutLifecycleHookInput, ...request.Option) (*autoscaling.PutLifecycleHookOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutLifecycleHookRequest(*autoscaling.PutLifecycleHookInput) (*request.Request, *autoscaling.PutLifecycleHookOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutNotificationConfiguration(*autoscaling.PutNotificationConfigurationInput) (*autoscaling.PutNotificationConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutNotificationConfigurationWithContext(aws.Context, *autoscaling.PutNotificationConfigurationInput, ...request.Option) (*autoscaling.PutNotificationConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutNotificationConfigurationRequest(*autoscaling.PutNotificationConfigurationInput) (*request.Request, *autoscaling.PutNotificationConfigurationOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutScalingPolicy(*autoscaling.PutScalingPolicyInput) (*autoscaling.PutScalingPolicyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutScalingPolicyWithContext(aws.Context, *autoscaling.PutScalingPolicyInput, ...request.Option) (*autoscaling.PutScalingPolicyOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutScalingPolicyRequest(*autoscaling.PutScalingPolicyInput) (*request.Request, *autoscaling.PutScalingPolicyOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutScheduledUpdateGroupAction(*autoscaling.PutScheduledUpdateGroupActionInput) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutScheduledUpdateGroupActionWithContext(aws.Context, *autoscaling.PutScheduledUpdateGroupActionInput, ...request.Option) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) PutScheduledUpdateGroupActionRequest(*autoscaling.PutScheduledUpdateGroupActionInput) (*request.Request, *autoscaling.PutScheduledUpdateGroupActionOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) RecordLifecycleActionHeartbeat(*autoscaling.RecordLifecycleActionHeartbeatInput) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) RecordLifecycleActionHeartbeatWithContext(aws.Context, *autoscaling.RecordLifecycleActionHeartbeatInput, ...request.Option) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) RecordLifecycleActionHeartbeatRequest(*autoscaling.RecordLifecycleActionHeartbeatInput) (*request.Request, *autoscaling.RecordLifecycleActionHeartbeatOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ResumeProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.ResumeProcessesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) ResumeProcessesWithContext(aws.Context, *autoscaling.ScalingProcessQuery, ...request.Option) (*autoscaling.ResumeProcessesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) ResumeProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.ResumeProcessesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetDesiredCapacity(*autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SetDesiredCapacityWithContext(aws.Context, *autoscaling.SetDesiredCapacityInput, ...request.Option) (*autoscaling.SetDesiredCapacityOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SetDesiredCapacityRequest(*autoscaling.SetDesiredCapacityInput) (*request.Request, *autoscaling.SetDesiredCapacityOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetInstanceHealth(*autoscaling.SetInstanceHealthInput) (*autoscaling.SetInstanceHealthOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SetInstanceHealthWithContext(aws.Context, *autoscaling.SetInstanceHealthInput, ...request.Option) (*autoscaling.SetInstanceHealthOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SetInstanceHealthRequest(*autoscaling.SetInstanceHealthInput) (*request.Request, *autoscaling.SetInstanceHealthOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetInstanceProtection(*autoscaling.SetInstanceProtectionInput) (*autoscaling.SetInstanceProtectionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SetInstanceProtectionWithContext(aws.Context, *autoscaling.SetInstanceProtectionInput, ...request.Option) (*autoscaling.SetInstanceProtectionOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SetInstanceProtectionRequest(*autoscaling.SetInstanceProtectionInput) (*request.Request, *autoscaling.SetInstanceProtectionOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SuspendProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.SuspendProcessesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SuspendProcessesWithContext(aws.Context, *autoscaling.ScalingProcessQuery, ...request.Option) (*autoscaling.SuspendProcessesOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) SuspendProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.SuspendProcessesOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroupWithContext(aws.Context, *autoscaling.TerminateInstanceInAutoScalingGroupInput, ...request.Option) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroupRequest(*autoscaling.TerminateInstanceInAutoScalingGroupInput) (*request.Request, *autoscaling.TerminateInstanceInAutoScalingGroupOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) UpdateAutoScalingGroup(*autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) UpdateAutoScalingGroupWithContext(aws.Context, *autoscaling.UpdateAutoScalingGroupInput, ...request.Option) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) UpdateAutoScalingGroupRequest(*autoscaling.UpdateAutoScalingGroupInput) (*request.Request, *autoscaling.UpdateAutoScalingGroupOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) WaitUntilGroupExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) WaitUntilGroupExistsWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.WaiterOption) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) WaitUntilGroupInService(*autoscaling.DescribeAutoScalingGroupsInput) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) WaitUntilGroupInServiceWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.WaiterOption) error {
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) WaitUntilGroupNotExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	glog.Fatalf("Not implemented")
	return nil
}
func (m *MockAutoscaling) WaitUntilGroupNotExistsWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.WaiterOption) error {
	glog.Fatalf("Not implemented")
	return nil
}
