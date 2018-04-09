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
)

func (m *MockAutoscaling) AttachInstancesWithContext(aws.Context, *autoscaling.AttachInstancesInput, ...request.Option) (*autoscaling.AttachInstancesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) AttachInstancesRequest(*autoscaling.AttachInstancesInput) (*request.Request, *autoscaling.AttachInstancesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) AttachLoadBalancerTargetGroups(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) AttachLoadBalancerTargetGroupsWithContext(aws.Context, *autoscaling.AttachLoadBalancerTargetGroupsInput, ...request.Option) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) AttachLoadBalancerTargetGroupsRequest(*autoscaling.AttachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.AttachLoadBalancerTargetGroupsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) CompleteLifecycleAction(*autoscaling.CompleteLifecycleActionInput) (*autoscaling.CompleteLifecycleActionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) CompleteLifecycleActionWithContext(aws.Context, *autoscaling.CompleteLifecycleActionInput, ...request.Option) (*autoscaling.CompleteLifecycleActionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) CompleteLifecycleActionRequest(*autoscaling.CompleteLifecycleActionInput) (*request.Request, *autoscaling.CompleteLifecycleActionOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) CreateAutoScalingGroupWithContext(aws.Context, *autoscaling.CreateAutoScalingGroupInput, ...request.Option) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) CreateAutoScalingGroupRequest(*autoscaling.CreateAutoScalingGroupInput) (*request.Request, *autoscaling.CreateAutoScalingGroupOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) CreateOrUpdateTags(*autoscaling.CreateOrUpdateTagsInput) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) CreateOrUpdateTagsWithContext(aws.Context, *autoscaling.CreateOrUpdateTagsInput, ...request.Option) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) CreateOrUpdateTagsRequest(*autoscaling.CreateOrUpdateTagsInput) (*request.Request, *autoscaling.CreateOrUpdateTagsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DeleteLifecycleHook(*autoscaling.DeleteLifecycleHookInput) (*autoscaling.DeleteLifecycleHookOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteLifecycleHookWithContext(aws.Context, *autoscaling.DeleteLifecycleHookInput, ...request.Option) (*autoscaling.DeleteLifecycleHookOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteLifecycleHookRequest(*autoscaling.DeleteLifecycleHookInput) (*request.Request, *autoscaling.DeleteLifecycleHookOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DeleteNotificationConfiguration(*autoscaling.DeleteNotificationConfigurationInput) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteNotificationConfigurationWithContext(aws.Context, *autoscaling.DeleteNotificationConfigurationInput, ...request.Option) (*autoscaling.DeleteNotificationConfigurationOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteNotificationConfigurationRequest(*autoscaling.DeleteNotificationConfigurationInput) (*request.Request, *autoscaling.DeleteNotificationConfigurationOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DeletePolicy(*autoscaling.DeletePolicyInput) (*autoscaling.DeletePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeletePolicyWithContext(aws.Context, *autoscaling.DeletePolicyInput, ...request.Option) (*autoscaling.DeletePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeletePolicyRequest(*autoscaling.DeletePolicyInput) (*request.Request, *autoscaling.DeletePolicyOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DeleteScheduledAction(*autoscaling.DeleteScheduledActionInput) (*autoscaling.DeleteScheduledActionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteScheduledActionWithContext(aws.Context, *autoscaling.DeleteScheduledActionInput, ...request.Option) (*autoscaling.DeleteScheduledActionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteScheduledActionRequest(*autoscaling.DeleteScheduledActionInput) (*request.Request, *autoscaling.DeleteScheduledActionOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DeleteTags(*autoscaling.DeleteTagsInput) (*autoscaling.DeleteTagsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteTagsWithContext(aws.Context, *autoscaling.DeleteTagsInput, ...request.Option) (*autoscaling.DeleteTagsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DeleteTagsRequest(*autoscaling.DeleteTagsInput) (*request.Request, *autoscaling.DeleteTagsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeAccountLimits(*autoscaling.DescribeAccountLimitsInput) (*autoscaling.DescribeAccountLimitsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAccountLimitsWithContext(aws.Context, *autoscaling.DescribeAccountLimitsInput, ...request.Option) (*autoscaling.DescribeAccountLimitsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAccountLimitsRequest(*autoscaling.DescribeAccountLimitsInput) (*request.Request, *autoscaling.DescribeAccountLimitsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeAdjustmentTypes(*autoscaling.DescribeAdjustmentTypesInput) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAdjustmentTypesWithContext(aws.Context, *autoscaling.DescribeAdjustmentTypesInput, ...request.Option) (*autoscaling.DescribeAdjustmentTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAdjustmentTypesRequest(*autoscaling.DescribeAdjustmentTypesInput) (*request.Request, *autoscaling.DescribeAdjustmentTypesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeAutoScalingInstances(*autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAutoScalingInstancesWithContext(aws.Context, *autoscaling.DescribeAutoScalingInstancesInput, ...request.Option) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAutoScalingInstancesRequest(*autoscaling.DescribeAutoScalingInstancesInput) (*request.Request, *autoscaling.DescribeAutoScalingInstancesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeAutoScalingInstancesPages(*autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool) error {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAutoScalingInstancesPagesWithContext(aws.Context, *autoscaling.DescribeAutoScalingInstancesInput, func(*autoscaling.DescribeAutoScalingInstancesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeAutoScalingNotificationTypes(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAutoScalingNotificationTypesWithContext(aws.Context, *autoscaling.DescribeAutoScalingNotificationTypesInput, ...request.Option) (*autoscaling.DescribeAutoScalingNotificationTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeAutoScalingNotificationTypesRequest(*autoscaling.DescribeAutoScalingNotificationTypesInput) (*request.Request, *autoscaling.DescribeAutoScalingNotificationTypesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeLifecycleHookTypes(*autoscaling.DescribeLifecycleHookTypesInput) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeLifecycleHookTypesWithContext(aws.Context, *autoscaling.DescribeLifecycleHookTypesInput, ...request.Option) (*autoscaling.DescribeLifecycleHookTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLifecycleHookTypesRequest(*autoscaling.DescribeLifecycleHookTypesInput) (*request.Request, *autoscaling.DescribeLifecycleHookTypesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeLifecycleHooks(*autoscaling.DescribeLifecycleHooksInput) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLifecycleHooksWithContext(aws.Context, *autoscaling.DescribeLifecycleHooksInput, ...request.Option) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLifecycleHooksRequest(*autoscaling.DescribeLifecycleHooksInput) (*request.Request, *autoscaling.DescribeLifecycleHooksOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeLoadBalancerTargetGroups(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*autoscaling.DescribeLoadBalancerTargetGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLoadBalancerTargetGroupsWithContext(aws.Context, *autoscaling.DescribeLoadBalancerTargetGroupsInput, ...request.Option) (*autoscaling.DescribeLoadBalancerTargetGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLoadBalancerTargetGroupsRequest(*autoscaling.DescribeLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DescribeLoadBalancerTargetGroupsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeLoadBalancers(*autoscaling.DescribeLoadBalancersInput) (*autoscaling.DescribeLoadBalancersOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLoadBalancersWithContext(aws.Context, *autoscaling.DescribeLoadBalancersInput, ...request.Option) (*autoscaling.DescribeLoadBalancersOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeLoadBalancersRequest(*autoscaling.DescribeLoadBalancersInput) (*request.Request, *autoscaling.DescribeLoadBalancersOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeMetricCollectionTypes(*autoscaling.DescribeMetricCollectionTypesInput) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeMetricCollectionTypesWithContext(aws.Context, *autoscaling.DescribeMetricCollectionTypesInput, ...request.Option) (*autoscaling.DescribeMetricCollectionTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeMetricCollectionTypesRequest(*autoscaling.DescribeMetricCollectionTypesInput) (*request.Request, *autoscaling.DescribeMetricCollectionTypesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeNotificationConfigurations(*autoscaling.DescribeNotificationConfigurationsInput) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeNotificationConfigurationsWithContext(aws.Context, *autoscaling.DescribeNotificationConfigurationsInput, ...request.Option) (*autoscaling.DescribeNotificationConfigurationsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeNotificationConfigurationsRequest(*autoscaling.DescribeNotificationConfigurationsInput) (*request.Request, *autoscaling.DescribeNotificationConfigurationsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeNotificationConfigurationsPages(*autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool) error {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeNotificationConfigurationsPagesWithContext(aws.Context, *autoscaling.DescribeNotificationConfigurationsInput, func(*autoscaling.DescribeNotificationConfigurationsOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribePolicies(*autoscaling.DescribePoliciesInput) (*autoscaling.DescribePoliciesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribePoliciesWithContext(aws.Context, *autoscaling.DescribePoliciesInput, ...request.Option) (*autoscaling.DescribePoliciesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribePoliciesRequest(*autoscaling.DescribePoliciesInput) (*request.Request, *autoscaling.DescribePoliciesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribePoliciesPages(*autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool) error {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribePoliciesPagesWithContext(aws.Context, *autoscaling.DescribePoliciesInput, func(*autoscaling.DescribePoliciesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScalingActivitiesWithContext(aws.Context, *autoscaling.DescribeScalingActivitiesInput, ...request.Option) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScalingActivitiesRequest(*autoscaling.DescribeScalingActivitiesInput) (*request.Request, *autoscaling.DescribeScalingActivitiesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeScalingActivitiesPages(*autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool) error {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScalingActivitiesPagesWithContext(aws.Context, *autoscaling.DescribeScalingActivitiesInput, func(*autoscaling.DescribeScalingActivitiesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeScalingProcessTypes(*autoscaling.DescribeScalingProcessTypesInput) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScalingProcessTypesWithContext(aws.Context, *autoscaling.DescribeScalingProcessTypesInput, ...request.Option) (*autoscaling.DescribeScalingProcessTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScalingProcessTypesRequest(*autoscaling.DescribeScalingProcessTypesInput) (*request.Request, *autoscaling.DescribeScalingProcessTypesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeScheduledActions(*autoscaling.DescribeScheduledActionsInput) (*autoscaling.DescribeScheduledActionsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScheduledActionsWithContext(aws.Context, *autoscaling.DescribeScheduledActionsInput, ...request.Option) (*autoscaling.DescribeScheduledActionsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScheduledActionsRequest(*autoscaling.DescribeScheduledActionsInput) (*request.Request, *autoscaling.DescribeScheduledActionsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeScheduledActionsPages(*autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool) error {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeScheduledActionsPagesWithContext(aws.Context, *autoscaling.DescribeScheduledActionsInput, func(*autoscaling.DescribeScheduledActionsOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) DescribeTerminationPolicyTypes(*autoscaling.DescribeTerminationPolicyTypesInput) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeTerminationPolicyTypesWithContext(aws.Context, *autoscaling.DescribeTerminationPolicyTypesInput, ...request.Option) (*autoscaling.DescribeTerminationPolicyTypesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DescribeTerminationPolicyTypesRequest(*autoscaling.DescribeTerminationPolicyTypesInput) (*request.Request, *autoscaling.DescribeTerminationPolicyTypesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DetachInstances(*autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DetachInstancesWithContext(aws.Context, *autoscaling.DetachInstancesInput, ...request.Option) (*autoscaling.DetachInstancesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DetachInstancesRequest(*autoscaling.DetachInstancesInput) (*request.Request, *autoscaling.DetachInstancesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DetachLoadBalancerTargetGroups(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DetachLoadBalancerTargetGroupsWithContext(aws.Context, *autoscaling.DetachLoadBalancerTargetGroupsInput, ...request.Option) (*autoscaling.DetachLoadBalancerTargetGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DetachLoadBalancerTargetGroupsRequest(*autoscaling.DetachLoadBalancerTargetGroupsInput) (*request.Request, *autoscaling.DetachLoadBalancerTargetGroupsOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DetachLoadBalancers(*autoscaling.DetachLoadBalancersInput) (*autoscaling.DetachLoadBalancersOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DetachLoadBalancersWithContext(aws.Context, *autoscaling.DetachLoadBalancersInput, ...request.Option) (*autoscaling.DetachLoadBalancersOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DetachLoadBalancersRequest(*autoscaling.DetachLoadBalancersInput) (*request.Request, *autoscaling.DetachLoadBalancersOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) DisableMetricsCollection(*autoscaling.DisableMetricsCollectionInput) (*autoscaling.DisableMetricsCollectionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DisableMetricsCollectionWithContext(aws.Context, *autoscaling.DisableMetricsCollectionInput, ...request.Option) (*autoscaling.DisableMetricsCollectionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) DisableMetricsCollectionRequest(*autoscaling.DisableMetricsCollectionInput) (*request.Request, *autoscaling.DisableMetricsCollectionOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) EnableMetricsCollectionWithContext(aws.Context, *autoscaling.EnableMetricsCollectionInput, ...request.Option) (*autoscaling.EnableMetricsCollectionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) EnableMetricsCollectionRequest(*autoscaling.EnableMetricsCollectionInput) (*request.Request, *autoscaling.EnableMetricsCollectionOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) EnterStandby(*autoscaling.EnterStandbyInput) (*autoscaling.EnterStandbyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) EnterStandbyWithContext(aws.Context, *autoscaling.EnterStandbyInput, ...request.Option) (*autoscaling.EnterStandbyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) EnterStandbyRequest(*autoscaling.EnterStandbyInput) (*request.Request, *autoscaling.EnterStandbyOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) ExecutePolicy(*autoscaling.ExecutePolicyInput) (*autoscaling.ExecutePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) ExecutePolicyWithContext(aws.Context, *autoscaling.ExecutePolicyInput, ...request.Option) (*autoscaling.ExecutePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) ExecutePolicyRequest(*autoscaling.ExecutePolicyInput) (*request.Request, *autoscaling.ExecutePolicyOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) ExitStandby(*autoscaling.ExitStandbyInput) (*autoscaling.ExitStandbyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) ExitStandbyWithContext(aws.Context, *autoscaling.ExitStandbyInput, ...request.Option) (*autoscaling.ExitStandbyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) ExitStandbyRequest(*autoscaling.ExitStandbyInput) (*request.Request, *autoscaling.ExitStandbyOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) PutLifecycleHook(*autoscaling.PutLifecycleHookInput) (*autoscaling.PutLifecycleHookOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutLifecycleHookWithContext(aws.Context, *autoscaling.PutLifecycleHookInput, ...request.Option) (*autoscaling.PutLifecycleHookOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutLifecycleHookRequest(*autoscaling.PutLifecycleHookInput) (*request.Request, *autoscaling.PutLifecycleHookOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) PutNotificationConfiguration(*autoscaling.PutNotificationConfigurationInput) (*autoscaling.PutNotificationConfigurationOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutNotificationConfigurationWithContext(aws.Context, *autoscaling.PutNotificationConfigurationInput, ...request.Option) (*autoscaling.PutNotificationConfigurationOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutNotificationConfigurationRequest(*autoscaling.PutNotificationConfigurationInput) (*request.Request, *autoscaling.PutNotificationConfigurationOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) PutScalingPolicy(*autoscaling.PutScalingPolicyInput) (*autoscaling.PutScalingPolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutScalingPolicyWithContext(aws.Context, *autoscaling.PutScalingPolicyInput, ...request.Option) (*autoscaling.PutScalingPolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutScalingPolicyRequest(*autoscaling.PutScalingPolicyInput) (*request.Request, *autoscaling.PutScalingPolicyOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) PutScheduledUpdateGroupAction(*autoscaling.PutScheduledUpdateGroupActionInput) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutScheduledUpdateGroupActionWithContext(aws.Context, *autoscaling.PutScheduledUpdateGroupActionInput, ...request.Option) (*autoscaling.PutScheduledUpdateGroupActionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) PutScheduledUpdateGroupActionRequest(*autoscaling.PutScheduledUpdateGroupActionInput) (*request.Request, *autoscaling.PutScheduledUpdateGroupActionOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) RecordLifecycleActionHeartbeat(*autoscaling.RecordLifecycleActionHeartbeatInput) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) RecordLifecycleActionHeartbeatWithContext(aws.Context, *autoscaling.RecordLifecycleActionHeartbeatInput, ...request.Option) (*autoscaling.RecordLifecycleActionHeartbeatOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) RecordLifecycleActionHeartbeatRequest(*autoscaling.RecordLifecycleActionHeartbeatInput) (*request.Request, *autoscaling.RecordLifecycleActionHeartbeatOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) ResumeProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.ResumeProcessesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) ResumeProcessesWithContext(aws.Context, *autoscaling.ScalingProcessQuery, ...request.Option) (*autoscaling.ResumeProcessesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) ResumeProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.ResumeProcessesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) SetDesiredCapacity(*autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SetDesiredCapacityWithContext(aws.Context, *autoscaling.SetDesiredCapacityInput, ...request.Option) (*autoscaling.SetDesiredCapacityOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SetDesiredCapacityRequest(*autoscaling.SetDesiredCapacityInput) (*request.Request, *autoscaling.SetDesiredCapacityOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) SetInstanceHealth(*autoscaling.SetInstanceHealthInput) (*autoscaling.SetInstanceHealthOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SetInstanceHealthWithContext(aws.Context, *autoscaling.SetInstanceHealthInput, ...request.Option) (*autoscaling.SetInstanceHealthOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SetInstanceHealthRequest(*autoscaling.SetInstanceHealthInput) (*request.Request, *autoscaling.SetInstanceHealthOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) SetInstanceProtection(*autoscaling.SetInstanceProtectionInput) (*autoscaling.SetInstanceProtectionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SetInstanceProtectionWithContext(aws.Context, *autoscaling.SetInstanceProtectionInput, ...request.Option) (*autoscaling.SetInstanceProtectionOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SetInstanceProtectionRequest(*autoscaling.SetInstanceProtectionInput) (*request.Request, *autoscaling.SetInstanceProtectionOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) SuspendProcesses(*autoscaling.ScalingProcessQuery) (*autoscaling.SuspendProcessesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SuspendProcessesWithContext(aws.Context, *autoscaling.ScalingProcessQuery, ...request.Option) (*autoscaling.SuspendProcessesOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) SuspendProcessesRequest(*autoscaling.ScalingProcessQuery) (*request.Request, *autoscaling.SuspendProcessesOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroupWithContext(aws.Context, *autoscaling.TerminateInstanceInAutoScalingGroupInput, ...request.Option) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroupRequest(*autoscaling.TerminateInstanceInAutoScalingGroupInput) (*request.Request, *autoscaling.TerminateInstanceInAutoScalingGroupOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) UpdateAutoScalingGroup(*autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) UpdateAutoScalingGroupWithContext(aws.Context, *autoscaling.UpdateAutoScalingGroupInput, ...request.Option) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	panic("Not implemented")
}
func (m *MockAutoscaling) UpdateAutoScalingGroupRequest(*autoscaling.UpdateAutoScalingGroupInput) (*request.Request, *autoscaling.UpdateAutoScalingGroupOutput) {
	panic("Not implemented")
}

func (m *MockAutoscaling) WaitUntilGroupExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) WaitUntilGroupExistsWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.WaiterOption) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) WaitUntilGroupInService(*autoscaling.DescribeAutoScalingGroupsInput) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) WaitUntilGroupInServiceWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.WaiterOption) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) WaitUntilGroupNotExists(*autoscaling.DescribeAutoScalingGroupsInput) error {
	panic("Not implemented")
}

func (m *MockAutoscaling) WaitUntilGroupNotExistsWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.WaiterOption) error {
	panic("Not implemented")
}
