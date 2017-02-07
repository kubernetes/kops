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

func AttachInstancesRequest(*autoscaling.AttachInstancesInput) (*request.Request, *autoscaling.AttachInstancesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func AttachLoadBalancerTargetGroupsRequest(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.AttachLoadBalancerTargetGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func AttachLoadBalancerTargetGroups(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func AttachLoadBalancersRequest(*autoscaling.AttachLoadBalancersInput) (*request.Request, *autoscaling.AttachLoadBalancersOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func AttachLoadBalancers(*autoscaling.AttachLoadBalancersInput) (*autoscaling.AttachLoadBalancersOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CompleteLifecycleActionRequest(*autoscaling.CompleteLifecycleActionInput) (*request.Request, *autoscaling.CompleteLifecycleActionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CompleteLifecycleAction(*autoscaling.CompleteLifecycleActionInput) (*autoscaling.CompleteLifecycleActionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CreateAutoScalingGroupRequest(*autoscaling.CreateAutoScalingGroupInput) (*request.Request, *autoscaling.CreateAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CreateLaunchConfigurationRequest(*autoscaling.CreateLaunchConfigurationInput) (*request.Request, *autoscaling.CreateLaunchConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CreateLaunchConfiguration(*autoscaling.CreateLaunchConfigurationInput) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CreateOrUpdateTagsRequest(*autoscaling.CreateOrUpdateTagsInput) (*request.Request, *autoscaling.CreateOrUpdateTagsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func CreateOrUpdateTags(*autoscaling.CreateOrUpdateTagsInput) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteAutoScalingGroupRequest(*autoscaling.DeleteAutoScalingGroupInput) (*request.Request, *autoscaling.DeleteAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteAutoScalingGroup(*autoscaling.DeleteAutoScalingGroupInput) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteLaunchConfigurationRequest(*autoscaling.DeleteLaunchConfigurationInput) (*request.Request, *autoscaling.DeleteLaunchConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteLaunchConfiguration(*autoscaling.DeleteLaunchConfigurationInput) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteLifecycleHookRequest(*autoscaling.DeleteLifecycleHookInput) (*request.Request, *autoscaling.DeleteLifecycleHookOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteLifecycleHook(*autoscaling.DeleteLifecycleHookInput) (*autoscaling.DeleteLifecycleHookOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteNotificationConfigurationRequest(*autoscaling.DeleteNotificationConfigurationInput) (*request.Request, *autoscaling.DeleteNotificationConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteNotificationConfiguration(*autoscaling.DeleteNotificationConfigurationInput) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeletePolicyRequest(*autoscaling.DeletePolicyInput) (*request.Request, *autoscaling.DeletePolicyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeletePolicy(*autoscaling.DeletePolicyInput) (*autoscaling.DeletePolicyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteScheduledActionRequest(*autoscaling.DeleteScheduledActionInput) (*request.Request, *autoscaling.DeleteScheduledActionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteScheduledAction(*autoscaling.DeleteScheduledActionInput) (*autoscaling.DeleteScheduledActionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteTagsRequest(*autoscaling.DeleteTagsInput) (*request.Request, *autoscaling.DeleteTagsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DeleteTags(*autoscaling.DeleteTagsInput) (*autoscaling.DeleteTagsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAccountLimitsRequest(*autoscaling.DescribeAccountLimitsInput) (*request.Request, *autoscaling.DescribeAccountLimitsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAccountLimits(*autoscaling.DescribeAccountLimitsInput) (*autoscaling.DescribeAccountLimitsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAdjustmentTypesRequest(*autoscaling.DescribeAdjustmentTypesInput) (*request.Request, *autoscaling.DescribeAdjustmentTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAdjustmentTypes(*autoscaling.DescribeAdjustmentTypesInput) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAutoScalingGroupsRequest(*autoscaling.DescribeAutoScalingGroupsInput) (*request.Request, *autoscaling.DescribeAutoScalingGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAutoScalingGroups(*autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAutoScalingGroupsPages(*autoscaling.DescribeAutoScalingGroupsInput, func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeAutoScalingInstancesRequest(*autoscaling.DescribeAutoScalingInstancesInput) (*request.Request, *autoscaling.DescribeAutoScalingInstancesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAutoScalingInstances(*autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAutoScalingInstancesPages(*autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeAutoScalingNotificationTypesRequest(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*request.Request, *autoscaling.DescribeAutoScalingNotificationTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeAutoScalingNotificationTypes(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLaunchConfigurationsRequest(*autoscaling.DescribeLaunchConfigurationsInput) (*request.Request, *autoscaling.DescribeLaunchConfigurationsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLaunchConfigurationsPages(*autoscaling.DescribeLaunchConfigurationsInput, func(*autoscaling.DescribeLaunchConfigurationsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeLifecycleHookTypesRequest(*autoscaling.DescribeLifecycleHookTypesInput) (*request.Request, *autoscaling.DescribeLifecycleHookTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLifecycleHookTypes(*autoscaling.DescribeLifecycleHookTypesInput) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLifecycleHooksRequest(*autoscaling.DescribeLifecycleHooksInput) (*request.Request, *autoscaling.DescribeLifecycleHooksOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLifecycleHooks(*autoscaling.DescribeLifecycleHooksInput) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLoadBalancerTargetGroupsRequest(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DescribeLoadBalancerTargetGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLoadBalancerTargetGroups(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*autoscaling.DescribeLoadBalancerTargetGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLoadBalancersRequest(*autoscaling.DescribeLoadBalancersInput) (*request.Request, *autoscaling.DescribeLoadBalancersOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeLoadBalancers(*autoscaling.DescribeLoadBalancersInput) (*autoscaling.DescribeLoadBalancersOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeMetricCollectionTypesRequest(*autoscaling.DescribeMetricCollectionTypesInput) (*request.Request, *autoscaling.DescribeMetricCollectionTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeMetricCollectionTypes(*autoscaling.DescribeMetricCollectionTypesInput) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeNotificationConfigurationsRequest(*autoscaling.DescribeNotificationConfigurationsInput) (*request.Request, *autoscaling.DescribeNotificationConfigurationsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeNotificationConfigurations(*autoscaling.DescribeNotificationConfigurationsInput) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeNotificationConfigurationsPages(*autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribePoliciesRequest(*autoscaling.DescribePoliciesInput) (*request.Request, *autoscaling.DescribePoliciesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribePolicies(*autoscaling.DescribePoliciesInput) (*autoscaling.DescribePoliciesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribePoliciesPages(*autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeScalingActivitiesRequest(*autoscaling.DescribeScalingActivitiesInput) (*request.Request, *autoscaling.DescribeScalingActivitiesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeScalingActivitiesPages(*autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeScalingProcessTypesRequest(*autoscaling.DescribeScalingProcessTypesInput) (*request.Request, *autoscaling.DescribeScalingProcessTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeScalingProcessTypes(*autoscaling.DescribeScalingProcessTypesInput) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeScheduledActionsRequest(*autoscaling.DescribeScheduledActionsInput) (*request.Request, *autoscaling.DescribeScheduledActionsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeScheduledActions(*autoscaling.DescribeScheduledActionsInput) (*autoscaling.DescribeScheduledActionsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeScheduledActionsPages(*autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeTagsRequest(*autoscaling.DescribeTagsInput) (*request.Request, *autoscaling.DescribeTagsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeTags(*autoscaling.DescribeTagsInput) (*autoscaling.DescribeTagsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeTagsPages(*autoscaling.DescribeTagsInput, func(*autoscaling.DescribeTagsOutput, bool) bool) error {
	log.Fatal("Not implemented")
	return nil
}

func DescribeTerminationPolicyTypesRequest(*autoscaling.DescribeTerminationPolicyTypesInput) (*request.Request, *autoscaling.DescribeTerminationPolicyTypesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DescribeTerminationPolicyTypes(*autoscaling.DescribeTerminationPolicyTypesInput) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DetachInstancesRequest(*autoscaling.DetachInstancesInput) (*request.Request, *autoscaling.DetachInstancesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DetachInstances(*autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DetachLoadBalancerTargetGroupsRequest(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DetachLoadBalancerTargetGroupsOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DetachLoadBalancerTargetGroups(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DetachLoadBalancersRequest(*autoscaling.DetachLoadBalancersInput) (*request.Request, *autoscaling.DetachLoadBalancersOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DetachLoadBalancers(*autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DisableMetricsCollectionRequest(*autoscaling.DisableMetricsCollectionInput) (*request.Request, *autoscaling.DisableMetricsCollectionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func DisableMetricsCollection(*autoscaling.DisableMetricsCollectionInput) (*autoscaling.DisableMetricsCollectionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func EnableMetricsCollectionRequest(*autoscaling.EnableMetricsCollectionInput) (*request.Request, *autoscaling.EnableMetricsCollectionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func EnableMetricsCollection(*autoscaling.EnableMetricsCollectionInput) (*autoscaling.EnableMetricsCollectionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func EnterStandbyRequest(*autoscaling.EnterStandbyInput) (*request.Request, *autoscaling.EnterStandbyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func EnterStandby(*autoscaling.EnterStandbyInput) (*autoscaling.EnterStandbyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func ExecutePolicyRequest(*autoscaling.ExecutePolicyInput) (*request.Request, *autoscaling.ExecutePolicyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func ExecutePolicy(*autoscaling.ExecutePolicyInput) (*autoscaling.ExecutePolicyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func ExitStandbyRequest(*autoscaling.ExitStandbyInput) (*request.Request, *autoscaling.ExitStandbyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func ExitStandby(*autoscaling.ExitStandbyInput) (*autoscaling.ExitStandbyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutLifecycleHookRequest(*autoscaling.PutLifecycleHookInput) (*request.Request, *autoscaling.PutLifecycleHookOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutLifecycleHook(*autoscaling.PutLifecycleHookInput) (*autoscaling.PutLifecycleHookOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutNotificationConfigurationRequest(*autoscaling.PutNotificationConfigurationInput) (*request.Request, *autoscaling.PutNotificationConfigurationOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutNotificationConfiguration(*autoscaling.PutNotificationConfigurationInput) (*autoscaling.PutNotificationConfigurationOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutScalingPolicyRequest(*autoscaling.PutScalingPolicyInput) (*request.Request, *autoscaling.PutScalingPolicyOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutScalingPolicy(*autoscaling.PutScalingPolicyInput) (*autoscaling.PutScalingPolicyOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutScheduledUpdateGroupActionRequest(*autoscaling.PutScheduledUpdateGroupActionInput) (*request.Request, *autoscaling.PutScheduledUpdateGroupActionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func PutScheduledUpdateGroupAction(*autoscaling.PutScheduledUpdateGroupActionInput) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func RecordLifecycleActionHeartbeatRequest(*autoscaling.RecordLifecycleActionHeartbeatInput) (*request.Request, *autoscaling.RecordLifecycleActionHeartbeatOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func RecordLifecycleActionHeartbeat(*autoscaling.RecordLifecycleActionHeartbeatInput) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func ResumeProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.ResumeProcessesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func ResumeProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.ResumeProcessesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SetDesiredCapacityRequest(*autoscaling.SetDesiredCapacityInput) (*request.Request, *autoscaling.SetDesiredCapacityOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SetDesiredCapacity(*autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SetInstanceHealthRequest(*autoscaling.SetInstanceHealthInput) (*request.Request, *autoscaling.SetInstanceHealthOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SetInstanceHealth(*autoscaling.SetInstanceHealthInput) (*autoscaling.SetInstanceHealthOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SetInstanceProtectionRequest(*autoscaling.SetInstanceProtectionInput) (*request.Request, *autoscaling.SetInstanceProtectionOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SetInstanceProtection(*autoscaling.SetInstanceProtectionInput) (*autoscaling.SetInstanceProtectionOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SuspendProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.SuspendProcessesOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func SuspendProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.SuspendProcessesOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func TerminateInstanceInAutoScalingGroupRequest(*autoscaling.TerminateInstanceInAutoScalingGroupInput) (*request.Request, *autoscaling.TerminateInstanceInAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func UpdateAutoScalingGroupRequest(*autoscaling.UpdateAutoScalingGroupInput) (*request.Request, *autoscaling.UpdateAutoScalingGroupOutput) {
	log.Fatal("Not implemented")
	return nil, nil
}

func UpdateAutoScalingGroup(*autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	log.Fatal("Not implemented")
	return nil, nil
}

func WaitUntilGroupExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	log.Fatal("Not implemented")
	return nil
}

func WaitUntilGroupInService(*autoscaling.DescribeAutoScalingGroupsInput) error {
	log.Fatal("Not implemented")
	return nil
}

func WaitUntilGroupNotExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	log.Fatal("Not implemented")
	return nil
}
