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
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog"
)

func (m *MockAutoscaling) AttachInstances(input *autoscaling.AttachInstancesInput) (*autoscaling.AttachInstancesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("AttachInstances %v", input)

	g := m.Groups[aws.StringValue(input.AutoScalingGroupName)]
	if g == nil {
		return nil, fmt.Errorf("AutoScaling Group not found")
	}

	for _, instanceID := range input.InstanceIds {
		g.Instances = append(g.Instances, &autoscaling.Instance{InstanceId: instanceID})
	}

	return &autoscaling.AttachInstancesOutput{}, nil
}

func (m *MockAutoscaling) CreateAutoScalingGroup(input *autoscaling.CreateAutoScalingGroupInput) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("CreateAutoScalingGroup %v", input)
	createdTime := time.Now().UTC()

	g := &autoscaling.Group{
		AutoScalingGroupName: input.AutoScalingGroupName,
		AvailabilityZones:    input.AvailabilityZones,
		CreatedTime:          &createdTime,
		DefaultCooldown:      input.DefaultCooldown,
		DesiredCapacity:      input.DesiredCapacity,
		// EnabledMetrics:          input.EnabledMetrics,
		HealthCheckGracePeriod:           input.HealthCheckGracePeriod,
		HealthCheckType:                  input.HealthCheckType,
		Instances:                        []*autoscaling.Instance{},
		LaunchConfigurationName:          input.LaunchConfigurationName,
		LoadBalancerNames:                input.LoadBalancerNames,
		MaxSize:                          input.MaxSize,
		MinSize:                          input.MinSize,
		NewInstancesProtectedFromScaleIn: input.NewInstancesProtectedFromScaleIn,
		PlacementGroup:                   input.PlacementGroup,
		// Status:                           input.Status,
		// SuspendedProcesses:               input.SuspendedProcesses,
		// Tags:                input.Tags,
		TargetGroupARNs:     input.TargetGroupARNs,
		TerminationPolicies: input.TerminationPolicies,
		VPCZoneIdentifier:   input.VPCZoneIdentifier,
	}

	for _, tag := range input.Tags {
		g.Tags = append(g.Tags, &autoscaling.TagDescription{
			Key:               tag.Key,
			PropagateAtLaunch: tag.PropagateAtLaunch,
			ResourceId:        tag.ResourceId,
			ResourceType:      tag.ResourceType,
			Value:             tag.Value,
		})
	}

	if m.Groups == nil {
		m.Groups = make(map[string]*autoscaling.Group)
	}
	m.Groups[*g.AutoScalingGroupName] = g

	return &autoscaling.CreateAutoScalingGroupOutput{}, nil
}

func (m *MockAutoscaling) EnableMetricsCollection(request *autoscaling.EnableMetricsCollectionInput) (*autoscaling.EnableMetricsCollectionOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("EnableMetricsCollection: %v", request)

	g := m.Groups[*request.AutoScalingGroupName]
	if g == nil {
		return nil, fmt.Errorf("AutoScalingGroup not found")
	}

	metrics := make(map[string]*autoscaling.EnabledMetric)
	for _, m := range g.EnabledMetrics {
		metrics[*m.Metric] = m
	}
	for _, m := range request.Metrics {
		metrics[*m] = &autoscaling.EnabledMetric{
			Metric:      m,
			Granularity: request.Granularity,
		}
	}

	g.EnabledMetrics = nil
	for _, m := range metrics {
		g.EnabledMetrics = append(g.EnabledMetrics, m)
	}

	response := &autoscaling.EnableMetricsCollectionOutput{}

	return response, nil
}

func (m *MockAutoscaling) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	groups := []*autoscaling.Group{}
	for _, group := range m.Groups {
		match := false

		if len(input.AutoScalingGroupNames) > 0 {
			for _, inputGroupName := range input.AutoScalingGroupNames {
				if aws.StringValue(group.AutoScalingGroupName) == aws.StringValue(inputGroupName) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			groups = append(groups, group)
		}
	}

	return &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: groups,
	}, nil
}

func (m *MockAutoscaling) TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, group := range m.Groups {
		for i := range group.Instances {
			if aws.StringValue(group.Instances[i].InstanceId) == aws.StringValue(input.InstanceId) {
				group.Instances = append(group.Instances[:i], group.Instances[i+1:]...)
				return &autoscaling.TerminateInstanceInAutoScalingGroupOutput{
					Activity: nil, // TODO
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("Instance not found")
}

func (m *MockAutoscaling) DescribeAutoScalingGroupsWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, ...request.Option) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeAutoScalingGroupsRequest(*autoscaling.DescribeAutoScalingGroupsInput) (*request.Request, *autoscaling.DescribeAutoScalingGroupsOutput) {
	klog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeAutoScalingGroupsPages(request *autoscaling.DescribeAutoScalingGroupsInput, callback func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	if request.MaxRecords != nil {
		klog.Fatalf("MaxRecords not implemented")
	}
	if request.NextToken != nil {
		klog.Fatalf("NextToken not implemented")
	}

	// For the mock, we just send everything in one page
	page, err := m.DescribeAutoScalingGroups(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockAutoscaling) DescribeAutoScalingGroupsPagesWithContext(aws.Context, *autoscaling.DescribeAutoScalingGroupsInput, func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool, ...request.Option) error {
	klog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) DeleteAutoScalingGroup(request *autoscaling.DeleteAutoScalingGroupInput) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteAutoScalingGroup: %v", request)

	id := aws.StringValue(request.AutoScalingGroupName)
	o := m.Groups[id]
	if o == nil {
		return nil, fmt.Errorf("AutoScalingGroup %q not found", id)
	}
	delete(m.Groups, id)

	return &autoscaling.DeleteAutoScalingGroupOutput{}, nil
}

func (m *MockAutoscaling) DeleteAutoScalingGroupWithContext(aws.Context, *autoscaling.DeleteAutoScalingGroupInput, ...request.Option) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteAutoScalingGroupRequest(*autoscaling.DeleteAutoScalingGroupInput) (*request.Request, *autoscaling.DeleteAutoScalingGroupOutput) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
