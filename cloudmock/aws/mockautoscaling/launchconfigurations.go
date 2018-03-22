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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
)

func (m *MockAutoscaling) DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLaunchConfigurationsWithContext(aws.Context, *autoscaling.DescribeLaunchConfigurationsInput, ...request.Option) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) DescribeLaunchConfigurationsRequest(*autoscaling.DescribeLaunchConfigurationsInput) (*request.Request, *autoscaling.DescribeLaunchConfigurationsOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}

func (m *MockAutoscaling) DescribeLaunchConfigurationsPages(request *autoscaling.DescribeLaunchConfigurationsInput, callback func(*autoscaling.DescribeLaunchConfigurationsOutput, bool) bool) error {
	if request.LaunchConfigurationNames != nil {
		glog.Fatalf("LaunchConfigurationNames not implemented")
	}
	if request.MaxRecords != nil {
		glog.Fatalf("MaxRecords not implemented")
	}
	if request.NextToken != nil {
		glog.Fatalf("NextToken not implemented")
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
	glog.Fatalf("Not implemented")
	return nil
}

func (m *MockAutoscaling) CreateLaunchConfiguration(request *autoscaling.CreateLaunchConfigurationInput) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	glog.Infof("CreateLaunchConfiguration: %v", request)

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
	m.LaunchConfigurations[*lc.LaunchConfigurationName] = lc

	return &autoscaling.CreateLaunchConfigurationOutput{}, nil
}
func (m *MockAutoscaling) CreateLaunchConfigurationWithContext(aws.Context, *autoscaling.CreateLaunchConfigurationInput, ...request.Option) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
func (m *MockAutoscaling) CreateLaunchConfigurationRequest(*autoscaling.CreateLaunchConfigurationInput) (*request.Request, *autoscaling.CreateLaunchConfigurationOutput) {
	glog.Fatalf("Not implemented")
	return nil, nil
}
