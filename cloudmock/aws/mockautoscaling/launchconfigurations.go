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

func (m *MockAutoscaling) DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLaunchConfigurationsWithContext(aws.Context, *autoscaling.DescribeLaunchConfigurationsInput, ...request.Option) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLaunchConfigurationsRequest(*autoscaling.DescribeLaunchConfigurationsInput) (*request.Request, *autoscaling.DescribeLaunchConfigurationsOutput) {
	klog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLaunchConfigurationsPages(request *autoscaling.DescribeLaunchConfigurationsInput, callback func(*autoscaling.DescribeLaunchConfigurationsOutput, bool) bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if request.LaunchConfigurationNames != nil {
		klog.Fatalf("LaunchConfigurationNames not implemented")
	}
	if request.MaxRecords != nil {
		klog.Fatalf("MaxRecords not implemented")
	}
	if request.NextToken != nil {
		klog.Fatalf("NextToken not implemented")
	}

	// For the mock, we just send everything in one page
	page := &autoscaling.DescribeLaunchConfigurationsOutput{}

	for _, lc := range m.LaunchConfigurations {
		page.LaunchConfigurations = append(page.LaunchConfigurations, lc)
	}

	callback(page, false)

	return nil
}
func (m *MockAutoscaling) DescribeLaunchConfigurationsPagesWithContext(aws.Context, *autoscaling.DescribeLaunchConfigurationsInput, func(*autoscaling.DescribeLaunchConfigurationsOutput, bool) bool, ...request.Option) error {
	klog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) CreateLaunchConfiguration(request *autoscaling.CreateLaunchConfigurationInput) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateLaunchConfiguration: %v", request)

	createdTime := time.Now().UTC()
	lc := &autoscaling.LaunchConfiguration{
		AssociatePublicIpAddress:     request.AssociatePublicIpAddress,
		BlockDeviceMappings:          request.BlockDeviceMappings,
		ClassicLinkVPCId:             request.ClassicLinkVPCId,
		ClassicLinkVPCSecurityGroups: request.ClassicLinkVPCSecurityGroups,
		CreatedTime:                  &createdTime,
		EbsOptimized:                 request.EbsOptimized,
		IamInstanceProfile:           request.IamInstanceProfile,
		ImageId:                      request.ImageId,
		InstanceMonitoring:           request.InstanceMonitoring,
		InstanceType:                 request.InstanceType,
		KernelId:                     request.KernelId,
		KeyName:                      request.KeyName,
		// LaunchConfigurationARN:       request.LaunchConfigurationARN,
		LaunchConfigurationName: request.LaunchConfigurationName,
		PlacementTenancy:        request.PlacementTenancy,
		RamdiskId:               request.RamdiskId,
		SecurityGroups:          request.SecurityGroups,
		SpotPrice:               request.SpotPrice,
		UserData:                request.UserData,
	}

	if m.LaunchConfigurations == nil {
		m.LaunchConfigurations = make(map[string]*autoscaling.LaunchConfiguration)
	}
	if m.LaunchConfigurations[*lc.LaunchConfigurationName] != nil {
		return nil, fmt.Errorf("duplicate LaunchConfigurationName %s", *lc.LaunchConfigurationName)
	}
	m.LaunchConfigurations[*lc.LaunchConfigurationName] = lc

	return &autoscaling.CreateLaunchConfigurationOutput{}, nil
}
func (m *MockAutoscaling) CreateLaunchConfigurationWithContext(aws.Context, *autoscaling.CreateLaunchConfigurationInput, ...request.Option) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CreateLaunchConfigurationRequest(*autoscaling.CreateLaunchConfigurationInput) (*request.Request, *autoscaling.CreateLaunchConfigurationOutput) {
	klog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DeleteLaunchConfiguration(request *autoscaling.DeleteLaunchConfigurationInput) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteLaunchConfiguration: %v", request)

	id := aws.StringValue(request.LaunchConfigurationName)
	o := m.LaunchConfigurations[id]
	if o == nil {
		return nil, fmt.Errorf("LaunchConfiguration %q not found", id)
	}
	delete(m.LaunchConfigurations, id)

	return &autoscaling.DeleteLaunchConfigurationOutput{}, nil
}

func (m *MockAutoscaling) DeleteLaunchConfigurationWithContext(aws.Context, *autoscaling.DeleteLaunchConfigurationInput, ...request.Option) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DeleteLaunchConfigurationRequest(*autoscaling.DeleteLaunchConfigurationInput) (*request.Request, *autoscaling.DeleteLaunchConfigurationOutput) {
	klog.Fatalf("Not implemented")
	return nil, nil
}
