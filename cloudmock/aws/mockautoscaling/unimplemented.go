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
	"log"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func (m *MockAutoscaling) AttachInstancesRequest(*autoscaling.AttachInstancesInput) (*request.Request, *autoscaling.AttachInstancesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) AttachLoadBalancerTargetGroupsRequest(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.AttachLoadBalancerTargetGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) AttachLoadBalancerTargetGroups(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) AttachLoadBalancersRequest(*autoscaling.AttachLoadBalancersInput) (*request.Request, *autoscaling.AttachLoadBalancersOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) AttachLoadBalancers(*autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CompleteLifecycleActionRequest(*autoscaling.CompleteLifecycleActionInput) (*request.Request, *autoscaling.CompleteLifecycleActionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CompleteLifecycleAction(*autoscaling.CompleteLifecycleActionInput) (*autoscaling.CompleteLifecycleActionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateAutoScalingGroupRequest(*autoscaling.CreateAutoScalingGroupInput) (*request.Request, *autoscaling.CreateAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateLaunchConfigurationRequest(*autoscaling.CreateLaunchConfigurationInput) (*request.Request, *autoscaling.CreateLaunchConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateLaunchConfiguration(*autoscaling.CreateLaunchConfigurationInput) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateOrUpdateTagsRequest(*autoscaling.CreateOrUpdateTagsInput) (*request.Request, *autoscaling.CreateOrUpdateTagsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) CreateOrUpdateTags(*autoscaling.CreateOrUpdateTagsInput) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteAutoScalingGroupRequest(*autoscaling.DeleteAutoScalingGroupInput) (*request.Request, *autoscaling.DeleteAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteAutoScalingGroup(*autoscaling.DeleteAutoScalingGroupInput) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLaunchConfigurationRequest(*autoscaling.DeleteLaunchConfigurationInput) (*request.Request, *autoscaling.DeleteLaunchConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLaunchConfiguration(*autoscaling.DeleteLaunchConfigurationInput) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLifecycleHookRequest(*autoscaling.DeleteLifecycleHookInput) (*request.Request, *autoscaling.DeleteLifecycleHookOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLifecycleHook(*autoscaling.DeleteLifecycleHookInput) (*autoscaling.DeleteLifecycleHookOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteNotificationConfigurationRequest(*autoscaling.DeleteNotificationConfigurationInput) (*request.Request, *autoscaling.DeleteNotificationConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteNotificationConfiguration(*autoscaling.DeleteNotificationConfigurationInput) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeletePolicyRequest(*autoscaling.DeletePolicyInput) (*request.Request, *autoscaling.DeletePolicyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeletePolicy(*autoscaling.DeletePolicyInput) (*autoscaling.DeletePolicyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteScheduledActionRequest(*autoscaling.DeleteScheduledActionInput) (*request.Request, *autoscaling.DeleteScheduledActionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteScheduledAction(*autoscaling.DeleteScheduledActionInput) (*autoscaling.DeleteScheduledActionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteTagsRequest(*autoscaling.DeleteTagsInput) (*request.Request, *autoscaling.DeleteTagsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteTags(*autoscaling.DeleteTagsInput) (*autoscaling.DeleteTagsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAccountLimitsRequest(*autoscaling.DescribeAccountLimitsInput) (*request.Request, *autoscaling.DescribeAccountLimitsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAccountLimits(*autoscaling.DescribeAccountLimitsInput) (*autoscaling.DescribeAccountLimitsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAdjustmentTypesRequest(*autoscaling.DescribeAdjustmentTypesInput) (*request.Request, *autoscaling.DescribeAdjustmentTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAdjustmentTypes(*autoscaling.DescribeAdjustmentTypesInput) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingGroupsRequest(*autoscaling.DescribeAutoScalingGroupsInput) (*request.Request, *autoscaling.DescribeAutoScalingGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingGroupsPages(*autoscaling.DescribeAutoScalingGroupsInput, func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeAutoScalingInstancesRequest(*autoscaling.DescribeAutoScalingInstancesInput) (*request.Request, *autoscaling.DescribeAutoScalingInstancesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingInstances(*autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingInstancesPages(*autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeAutoScalingNotificationTypesRequest(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*request.Request, *autoscaling.DescribeAutoScalingNotificationTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingNotificationTypes(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLaunchConfigurationsRequest(*autoscaling.DescribeLaunchConfigurationsInput) (*request.Request, *autoscaling.DescribeLaunchConfigurationsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLaunchConfigurationsPages(*autoscaling.DescribeLaunchConfigurationsInput, func(*autoscaling.DescribeLaunchConfigurationsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeLifecycleHookTypesRequest(*autoscaling.DescribeLifecycleHookTypesInput) (*request.Request, *autoscaling.DescribeLifecycleHookTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLifecycleHookTypes(*autoscaling.DescribeLifecycleHookTypesInput) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLifecycleHooksRequest(*autoscaling.DescribeLifecycleHooksInput) (*request.Request, *autoscaling.DescribeLifecycleHooksOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLifecycleHooks(*autoscaling.DescribeLifecycleHooksInput) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLoadBalancerTargetGroupsRequest(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DescribeLoadBalancerTargetGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLoadBalancerTargetGroups(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*autoscaling.DescribeLoadBalancerTargetGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLoadBalancersRequest(*autoscaling.DescribeLoadBalancersInput) (*request.Request, *autoscaling.DescribeLoadBalancersOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLoadBalancers(*autoscaling.DescribeLoadBalancersInput) (*autoscaling.DescribeLoadBalancersOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeMetricCollectionTypesRequest(*autoscaling.DescribeMetricCollectionTypesInput) (*request.Request, *autoscaling.DescribeMetricCollectionTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeMetricCollectionTypes(*autoscaling.DescribeMetricCollectionTypesInput) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeNotificationConfigurationsRequest(*autoscaling.DescribeNotificationConfigurationsInput) (*request.Request, *autoscaling.DescribeNotificationConfigurationsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeNotificationConfigurations(*autoscaling.DescribeNotificationConfigurationsInput) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeNotificationConfigurationsPages(*autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribePoliciesRequest(*autoscaling.DescribePoliciesInput) (*request.Request, *autoscaling.DescribePoliciesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribePolicies(*autoscaling.DescribePoliciesInput) (*autoscaling.DescribePoliciesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribePoliciesPages(*autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeScalingActivitiesRequest(*autoscaling.DescribeScalingActivitiesInput) (*request.Request, *autoscaling.DescribeScalingActivitiesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScalingActivitiesPages(*autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeScalingProcessTypesRequest(*autoscaling.DescribeScalingProcessTypesInput) (*request.Request, *autoscaling.DescribeScalingProcessTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScalingProcessTypes(*autoscaling.DescribeScalingProcessTypesInput) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScheduledActionsRequest(*autoscaling.DescribeScheduledActionsInput) (*request.Request, *autoscaling.DescribeScheduledActionsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScheduledActions(*autoscaling.DescribeScheduledActionsInput) (*autoscaling.DescribeScheduledActionsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeScheduledActionsPages(*autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeTagsRequest(*autoscaling.DescribeTagsInput) (*request.Request, *autoscaling.DescribeTagsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeTags(*autoscaling.DescribeTagsInput) (*autoscaling.DescribeTagsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeTagsPages(*autoscaling.DescribeTagsInput, func(*autoscaling.DescribeTagsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) DescribeTerminationPolicyTypesRequest(*autoscaling.DescribeTerminationPolicyTypesInput) (*request.Request, *autoscaling.DescribeTerminationPolicyTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeTerminationPolicyTypes(*autoscaling.DescribeTerminationPolicyTypesInput) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachInstancesRequest(*autoscaling.DetachInstancesInput) (*request.Request, *autoscaling.DetachInstancesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachInstances(*autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachLoadBalancerTargetGroupsRequest(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DetachLoadBalancerTargetGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachLoadBalancerTargetGroups(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachLoadBalancersRequest(*autoscaling.DetachLoadBalancersInput) (*request.Request, *autoscaling.DetachLoadBalancersOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DetachLoadBalancers(*autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DisableMetricsCollectionRequest(*autoscaling.DisableMetricsCollectionInput) (*request.Request, *autoscaling.DisableMetricsCollectionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DisableMetricsCollection(*autoscaling.DisableMetricsCollectionInput) (*autoscaling.DisableMetricsCollectionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) EnableMetricsCollectionRequest(*autoscaling.EnableMetricsCollectionInput) (*request.Request, *autoscaling.EnableMetricsCollectionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) EnableMetricsCollection(*autoscaling.EnableMetricsCollectionInput) (*autoscaling.EnableMetricsCollectionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) EnterStandbyRequest(*autoscaling.EnterStandbyInput) (*request.Request, *autoscaling.EnterStandbyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) EnterStandby(*autoscaling.EnterStandbyInput) (*autoscaling.EnterStandbyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ExecutePolicyRequest(*autoscaling.ExecutePolicyInput) (*request.Request, *autoscaling.ExecutePolicyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ExecutePolicy(*autoscaling.ExecutePolicyInput) (*autoscaling.ExecutePolicyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ExitStandbyRequest(*autoscaling.ExitStandbyInput) (*request.Request, *autoscaling.ExitStandbyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ExitStandby(*autoscaling.ExitStandbyInput) (*autoscaling.ExitStandbyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutLifecycleHookRequest(*autoscaling.PutLifecycleHookInput) (*request.Request, *autoscaling.PutLifecycleHookOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutLifecycleHook(*autoscaling.PutLifecycleHookInput) (*autoscaling.PutLifecycleHookOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutNotificationConfigurationRequest(*autoscaling.PutNotificationConfigurationInput) (*request.Request, *autoscaling.PutNotificationConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutNotificationConfiguration(*autoscaling.PutNotificationConfigurationInput) (*autoscaling.PutNotificationConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutScalingPolicyRequest(*autoscaling.PutScalingPolicyInput) (*request.Request, *autoscaling.PutScalingPolicyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutScalingPolicy(*autoscaling.PutScalingPolicyInput) (*autoscaling.PutScalingPolicyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutScheduledUpdateGroupActionRequest(*autoscaling.PutScheduledUpdateGroupActionInput) (*request.Request, *autoscaling.PutScheduledUpdateGroupActionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) PutScheduledUpdateGroupAction(*autoscaling.PutScheduledUpdateGroupActionInput) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) RecordLifecycleActionHeartbeatRequest(*autoscaling.RecordLifecycleActionHeartbeatInput) (*request.Request, *autoscaling.RecordLifecycleActionHeartbeatOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) RecordLifecycleActionHeartbeat(*autoscaling.RecordLifecycleActionHeartbeatInput) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ResumeProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.ResumeProcessesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) ResumeProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.ResumeProcessesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetDesiredCapacityRequest(*autoscaling.SetDesiredCapacityInput) (*request.Request, *autoscaling.SetDesiredCapacityOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetDesiredCapacity(*autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetInstanceHealthRequest(*autoscaling.SetInstanceHealthInput) (*request.Request, *autoscaling.SetInstanceHealthOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetInstanceHealth(*autoscaling.SetInstanceHealthInput) (*autoscaling.SetInstanceHealthOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetInstanceProtectionRequest(*autoscaling.SetInstanceProtectionInput) (*request.Request, *autoscaling.SetInstanceProtectionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SetInstanceProtection(*autoscaling.SetInstanceProtectionInput) (*autoscaling.SetInstanceProtectionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SuspendProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.SuspendProcessesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) SuspendProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.SuspendProcessesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroupRequest(*autoscaling.TerminateInstanceInAutoScalingGroupInput) (*request.Request, *autoscaling.TerminateInstanceInAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) UpdateAutoScalingGroupRequest(*autoscaling.UpdateAutoScalingGroupInput) (*request.Request, *autoscaling.UpdateAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) UpdateAutoScalingGroup(*autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) WaitUntilGroupExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) WaitUntilGroupInService(*autoscaling.DescribeAutoScalingGroupsInput) error {
	log.Fatal("Not implemented")
	return nil
}

func (m *MockAutoscaling) WaitUntilGroupNotExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	log.Fatal("Not implemented")
	return nil
}
