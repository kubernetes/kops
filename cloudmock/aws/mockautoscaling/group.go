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

package mockautoscaling

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"k8s.io/klog/v2"
)

func (m *MockAutoscaling) AttachInstances(ctx context.Context, input *autoscaling.AttachInstancesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.AttachInstancesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock AttachInstances %v", input)

	g := m.Groups[aws.ToString(input.AutoScalingGroupName)]
	if g == nil {
		return nil, fmt.Errorf("AutoScaling Group not found")
	}

	for _, instanceID := range input.InstanceIds {
		g.Instances = append(g.Instances, autoscalingtypes.Instance{InstanceId: aws.String(instanceID)})
	}

	return &autoscaling.AttachInstancesOutput{}, nil
}

func (m *MockAutoscaling) DetachInstances(ctx context.Context, input *autoscaling.DetachInstancesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DetachInstancesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock AttachInstances %v", input)

	g := m.Groups[aws.ToString(input.AutoScalingGroupName)]
	if g == nil {
		return nil, fmt.Errorf("AutoScaling Group not found")
	}

	instances := make([]autoscalingtypes.Instance, 0)
	for _, instance := range g.Instances {
		if !slices.Contains(input.InstanceIds, aws.ToString(instance.InstanceId)) {
			instances = append(instances, instance)
		}
	}
	g.Instances = instances

	return &autoscaling.DetachInstancesOutput{}, nil
}

func (m *MockAutoscaling) CreateAutoScalingGroup(ctx context.Context, input *autoscaling.CreateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock CreateAutoScalingGroup %v", input)

	createdTime := time.Now().UTC()

	g := &autoscalingtypes.AutoScalingGroup{
		AutoScalingGroupName: input.AutoScalingGroupName,
		AvailabilityZones:    input.AvailabilityZones,
		CreatedTime:          &createdTime,
		DefaultCooldown:      input.DefaultCooldown,
		DesiredCapacity:      input.DesiredCapacity,
		// EnabledMetrics:          input.EnabledMetrics,
		HealthCheckGracePeriod:           input.HealthCheckGracePeriod,
		HealthCheckType:                  input.HealthCheckType,
		Instances:                        []autoscalingtypes.Instance{},
		LaunchConfigurationName:          input.LaunchConfigurationName,
		LaunchTemplate:                   input.LaunchTemplate,
		LoadBalancerNames:                input.LoadBalancerNames,
		MaxSize:                          input.MaxSize,
		MinSize:                          input.MinSize,
		NewInstancesProtectedFromScaleIn: input.NewInstancesProtectedFromScaleIn,
		PlacementGroup:                   input.PlacementGroup,
		// Status:                           input.Status,
		SuspendedProcesses:  make([]autoscalingtypes.SuspendedProcess, 0),
		TargetGroupARNs:     input.TargetGroupARNs,
		TerminationPolicies: input.TerminationPolicies,
		VPCZoneIdentifier:   input.VPCZoneIdentifier,
		MaxInstanceLifetime: input.MaxInstanceLifetime,
	}

	if input.LaunchTemplate != nil {
		g.LaunchTemplate.LaunchTemplateName = input.AutoScalingGroupName
		if g.LaunchTemplate.LaunchTemplateId == nil {
			return nil, fmt.Errorf("AutoScalingGroup has LaunchTemplate without ID")
		}
	}

	for _, tag := range input.Tags {
		g.Tags = append(g.Tags, autoscalingtypes.TagDescription{
			Key:               tag.Key,
			PropagateAtLaunch: tag.PropagateAtLaunch,
			ResourceId:        tag.ResourceId,
			ResourceType:      tag.ResourceType,
			Value:             tag.Value,
		})
	}

	if m.Groups == nil {
		m.Groups = make(map[string]*autoscalingtypes.AutoScalingGroup)
	}
	m.Groups[*g.AutoScalingGroupName] = g

	return &autoscaling.CreateAutoScalingGroupOutput{}, nil
}

func (m *MockAutoscaling) UpdateAutoScalingGroup(ctx context.Context, request *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	klog.V(2).Infof("Mock UpdateAutoScalingGroup %v", request)

	if _, ok := m.Groups[*request.AutoScalingGroupName]; !ok {
		return nil, fmt.Errorf("Autoscaling group not found: %v", *request.AutoScalingGroupName)
	}
	group := m.Groups[*request.AutoScalingGroupName]

	if request.AvailabilityZones != nil {
		group.AvailabilityZones = request.AvailabilityZones
	}
	if request.CapacityRebalance != nil {
		group.CapacityRebalance = request.CapacityRebalance
	}
	if request.DesiredCapacity != nil {
		group.DesiredCapacity = request.DesiredCapacity
	}
	if request.HealthCheckGracePeriod != nil {
		group.HealthCheckGracePeriod = request.HealthCheckGracePeriod
	}
	if request.HealthCheckType != nil {
		group.HealthCheckType = request.HealthCheckType
	}
	if request.LaunchConfigurationName != nil {
		group.LaunchConfigurationName = request.LaunchConfigurationName
	}
	if request.LaunchTemplate != nil {
		group.LaunchTemplate = request.LaunchTemplate
	}
	if request.MaxInstanceLifetime != nil {
		group.MaxInstanceLifetime = request.MaxInstanceLifetime
	}
	if request.MaxSize != nil {
		group.MaxSize = request.MaxSize
	}
	if request.MinSize != nil {
		group.MinSize = request.MinSize
	}
	if request.MixedInstancesPolicy != nil {
		group.MixedInstancesPolicy = request.MixedInstancesPolicy
	}
	if request.NewInstancesProtectedFromScaleIn != nil {
		group.NewInstancesProtectedFromScaleIn = request.NewInstancesProtectedFromScaleIn
	}
	if request.PlacementGroup != nil {
		group.PlacementGroup = request.PlacementGroup
	}
	if request.ServiceLinkedRoleARN != nil {
		group.ServiceLinkedRoleARN = request.ServiceLinkedRoleARN
	}
	if request.TerminationPolicies != nil {
		group.TerminationPolicies = request.TerminationPolicies
	}
	if request.VPCZoneIdentifier != nil {
		group.VPCZoneIdentifier = request.VPCZoneIdentifier
	}
	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

func (m *MockAutoscaling) EnableMetricsCollection(ctx context.Context, request *autoscaling.EnableMetricsCollectionInput, optFns ...func(*autoscaling.Options)) (*autoscaling.EnableMetricsCollectionOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock EnableMetricsCollection: %v", request)

	g := m.Groups[*request.AutoScalingGroupName]
	if g == nil {
		return nil, fmt.Errorf("AutoScalingGroup not found")
	}

	metrics := make(map[string]*autoscalingtypes.EnabledMetric)
	for _, m := range g.EnabledMetrics {
		metrics[*m.Metric] = &m
	}
	for _, m := range request.Metrics {
		metrics[m] = &autoscalingtypes.EnabledMetric{
			Metric:      &m,
			Granularity: request.Granularity,
		}
	}

	g.EnabledMetrics = nil
	for _, m := range metrics {
		g.EnabledMetrics = append(g.EnabledMetrics, *m)
	}

	response := &autoscaling.EnableMetricsCollectionOutput{}

	return response, nil
}

func (m *MockAutoscaling) SuspendProcesses(ctx context.Context, input *autoscaling.SuspendProcessesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.SuspendProcessesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock SuspendProcesses: %v", input)

	g := m.Groups[*input.AutoScalingGroupName]
	if g == nil {
		return nil, fmt.Errorf("AutoScalingGroup not found")
	}

	for _, p := range input.ScalingProcesses {
		found := false
		for _, asgProc := range g.SuspendedProcesses {
			if aws.ToString(asgProc.ProcessName) == p {
				found = true
			}
		}
		if !found {
			g.SuspendedProcesses = append(g.SuspendedProcesses, autoscalingtypes.SuspendedProcess{
				ProcessName: aws.String(p),
			})
		}
	}

	return &autoscaling.SuspendProcessesOutput{}, nil
}

func (m *MockAutoscaling) DescribeAutoScalingGroups(ctx context.Context, input *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DescribeAutoScalingGroups: %v", input)

	groups := []autoscalingtypes.AutoScalingGroup{}
	for _, group := range m.Groups {
		match := false

		if len(input.AutoScalingGroupNames) > 0 {
			for _, inputGroupName := range input.AutoScalingGroupNames {
				if aws.ToString(group.AutoScalingGroupName) == inputGroupName {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			groups = append(groups, *group)
		}
	}

	return &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: groups,
	}, nil
}

func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroup(ctx context.Context, input *autoscaling.TerminateInstanceInAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, group := range m.Groups {
		for i := range group.Instances {
			if aws.ToString(group.Instances[i].InstanceId) == aws.ToString(input.InstanceId) {
				group.Instances = append(group.Instances[:i], group.Instances[i+1:]...)
				return &autoscaling.TerminateInstanceInAutoScalingGroupOutput{
					Activity: nil, // TODO
				}, nil
			}
		}
		wp := m.WarmPoolInstances[*group.AutoScalingGroupName]
		for i := range wp {
			if aws.ToString(wp[i].InstanceId) == aws.ToString(input.InstanceId) {
				m.WarmPoolInstances[*group.AutoScalingGroupName] = append(wp[:i], wp[i+1:]...)
				return &autoscaling.TerminateInstanceInAutoScalingGroupOutput{
					Activity: nil, // TODO
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("Instance not found")
}

func (m *MockAutoscaling) DeleteAutoScalingGroup(ctx context.Context, request *autoscaling.DeleteAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DeleteAutoScalingGroup: %v", request)

	id := aws.ToString(request.AutoScalingGroupName)
	o := m.Groups[id]
	if o == nil {
		return nil, fmt.Errorf("AutoScalingGroup %q not found", id)
	}
	delete(m.Groups, id)

	return &autoscaling.DeleteAutoScalingGroupOutput{}, nil
}

func (m *MockAutoscaling) PutLifecycleHook(ctx context.Context, input *autoscaling.PutLifecycleHookInput, optFns ...func(*autoscaling.Options)) (*autoscaling.PutLifecycleHookOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	hook := &autoscalingtypes.LifecycleHook{
		AutoScalingGroupName:  input.AutoScalingGroupName,
		DefaultResult:         input.DefaultResult,
		GlobalTimeout:         input.HeartbeatTimeout,
		HeartbeatTimeout:      input.HeartbeatTimeout,
		LifecycleHookName:     input.LifecycleHookName,
		LifecycleTransition:   input.LifecycleTransition,
		NotificationMetadata:  input.NotificationMetadata,
		NotificationTargetARN: input.NotificationTargetARN,
		RoleARN:               input.RoleARN,
	}

	if m.LifecycleHooks == nil {
		m.LifecycleHooks = make(map[string]*autoscalingtypes.LifecycleHook)
	}
	name := *input.AutoScalingGroupName + "::" + *input.LifecycleHookName
	m.LifecycleHooks[name] = hook

	return &autoscaling.PutLifecycleHookOutput{}, nil
}

func (m *MockAutoscaling) DescribeLifecycleHooks(ctx context.Context, input *autoscaling.DescribeLifecycleHooksInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &autoscaling.DescribeLifecycleHooksOutput{}
	for _, lifecycleHookName := range input.LifecycleHookNames {
		name := *input.AutoScalingGroupName + "::" + lifecycleHookName

		hook := m.LifecycleHooks[name]
		if hook != nil {
			response.LifecycleHooks = append(response.LifecycleHooks, *hook)
		}
	}
	return response, nil
}
