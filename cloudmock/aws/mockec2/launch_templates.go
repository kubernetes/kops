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

package mockec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// DescribeLaunchTemplates mocks the describing the launch templates
func (m *MockEC2) DescribeLaunchTemplates(request *ec2.DescribeLaunchTemplatesInput) (*ec2.DescribeLaunchTemplatesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	o := &ec2.DescribeLaunchTemplatesOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}

	for name := range m.LaunchTemplates {
		o.LaunchTemplates = append(o.LaunchTemplates, &ec2.LaunchTemplate{
			LaunchTemplateName: aws.String(name),
		})
	}

	return o, nil
}

// DescribeLaunchTemplateVersions mocks the retrieval of launch template versions - we don't use this at the moment so we can just return the template
func (m *MockEC2) DescribeLaunchTemplateVersions(request *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	o := &ec2.DescribeLaunchTemplateVersionsOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}

	lt, found := m.LaunchTemplates[aws.StringValue(request.LaunchTemplateName)]
	if !found {
		return o, nil
	}
	o.LaunchTemplateVersions = append(o.LaunchTemplateVersions, &ec2.LaunchTemplateVersion{
		DefaultVersion:     aws.Bool(true),
		LaunchTemplateData: lt,
		LaunchTemplateName: request.LaunchTemplateName,
	})

	return o, nil
}

// CreateLaunchTemplate mocks the ec2 create launch template
func (m *MockEC2) CreateLaunchTemplate(request *ec2.CreateLaunchTemplateInput) (*ec2.CreateLaunchTemplateOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.LaunchTemplates == nil {
		m.LaunchTemplates = make(map[string]*ec2.ResponseLaunchTemplateData)
	}
	if m.LaunchTemplates[aws.StringValue(request.LaunchTemplateName)] != nil {
		return nil, fmt.Errorf("duplicate LaunchTemplateName %s", aws.StringValue(request.LaunchTemplateName))
	}
	resp := &ec2.ResponseLaunchTemplateData{
		DisableApiTermination: request.LaunchTemplateData.DisableApiTermination,
		EbsOptimized:          request.LaunchTemplateData.EbsOptimized,
		ImageId:               request.LaunchTemplateData.ImageId,
		InstanceType:          request.LaunchTemplateData.InstanceType,
		KeyName:               request.LaunchTemplateData.KeyName,
		SecurityGroupIds:      request.LaunchTemplateData.SecurityGroupIds,
		SecurityGroups:        request.LaunchTemplateData.SecurityGroups,
		UserData:              request.LaunchTemplateData.UserData,
	}
	m.LaunchTemplates[aws.StringValue(request.LaunchTemplateName)] = resp

	// @GOD DAMN AWS request vs response .. fu@@@@@ .. so much typing!!#@#@#

	if request.LaunchTemplateData.Monitoring != nil {
		resp.Monitoring = &ec2.LaunchTemplatesMonitoring{Enabled: request.LaunchTemplateData.Monitoring.Enabled}
	}
	if request.LaunchTemplateData.CpuOptions != nil {
		resp.CpuOptions = &ec2.LaunchTemplateCpuOptions{
			CoreCount:      request.LaunchTemplateData.CpuOptions.CoreCount,
			ThreadsPerCore: request.LaunchTemplateData.CpuOptions.ThreadsPerCore,
		}
	}
	if len(request.LaunchTemplateData.BlockDeviceMappings) > 0 {
		for _, x := range request.LaunchTemplateData.BlockDeviceMappings {
			resp.BlockDeviceMappings = append(resp.BlockDeviceMappings, &ec2.LaunchTemplateBlockDeviceMapping{
				DeviceName: x.DeviceName,
				Ebs: &ec2.LaunchTemplateEbsBlockDevice{
					DeleteOnTermination: x.Ebs.DeleteOnTermination,
					Encrypted:           x.Ebs.Encrypted,
					Iops:                x.Ebs.Iops,
					KmsKeyId:            x.Ebs.KmsKeyId,
					SnapshotId:          x.Ebs.SnapshotId,
					VolumeSize:          x.Ebs.VolumeSize,
					VolumeType:          x.Ebs.VolumeType,
				},
				NoDevice:    x.NoDevice,
				VirtualName: x.VirtualName,
			})
		}
	}
	if request.LaunchTemplateData.CreditSpecification != nil {
		resp.CreditSpecification = &ec2.CreditSpecification{CpuCredits: request.LaunchTemplateData.CreditSpecification.CpuCredits}
	}
	if request.LaunchTemplateData.IamInstanceProfile != nil {
		resp.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileSpecification{
			Arn:  request.LaunchTemplateData.IamInstanceProfile.Arn,
			Name: request.LaunchTemplateData.IamInstanceProfile.Name,
		}
	}
	if request.LaunchTemplateData.InstanceMarketOptions != nil {
		resp.InstanceMarketOptions = &ec2.LaunchTemplateInstanceMarketOptions{
			MarketType: request.LaunchTemplateData.InstanceMarketOptions.MarketType,
			SpotOptions: &ec2.LaunchTemplateSpotMarketOptions{
				BlockDurationMinutes:         request.LaunchTemplateData.InstanceMarketOptions.SpotOptions.BlockDurationMinutes,
				InstanceInterruptionBehavior: request.LaunchTemplateData.InstanceMarketOptions.SpotOptions.InstanceInterruptionBehavior,
				MaxPrice:                     request.LaunchTemplateData.InstanceMarketOptions.SpotOptions.MaxPrice,
				SpotInstanceType:             request.LaunchTemplateData.InstanceMarketOptions.SpotOptions.SpotInstanceType,
				ValidUntil:                   request.LaunchTemplateData.InstanceMarketOptions.SpotOptions.ValidUntil,
			},
		}
	}
	if len(request.LaunchTemplateData.NetworkInterfaces) > 0 {
		for _, x := range request.LaunchTemplateData.NetworkInterfaces {
			resp.NetworkInterfaces = append(resp.NetworkInterfaces, &ec2.LaunchTemplateInstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress:       x.AssociatePublicIpAddress,
				DeleteOnTermination:            x.DeleteOnTermination,
				Description:                    x.Description,
				DeviceIndex:                    x.DeviceIndex,
				Groups:                         x.Groups,
				Ipv6AddressCount:               x.Ipv6AddressCount,
				NetworkInterfaceId:             x.NetworkInterfaceId,
				PrivateIpAddress:               x.PrivateIpAddress,
				PrivateIpAddresses:             x.PrivateIpAddresses,
				SecondaryPrivateIpAddressCount: x.SecondaryPrivateIpAddressCount,
				SubnetId:                       x.SubnetId,
			})
		}
	}

	return &ec2.CreateLaunchTemplateOutput{}, nil
}

// DeleteLaunchTemplate mocks the deletion of a launch template
func (m *MockEC2) DeleteLaunchTemplate(request *ec2.DeleteLaunchTemplateInput) (*ec2.DeleteLaunchTemplateOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	o := &ec2.DeleteLaunchTemplateOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}
	delete(m.LaunchTemplates, aws.StringValue(request.LaunchTemplateName))

	return o, nil
}
