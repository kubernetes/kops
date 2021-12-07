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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
)

type launchTemplateInfo struct {
	data    *ec2.ResponseLaunchTemplateData
	name    *string
	version int
}

// DescribeLaunchTemplatesPages mocks the describing the launch templates
func (m *MockEC2) DescribeLaunchTemplatesPages(request *ec2.DescribeLaunchTemplatesInput, callback func(*ec2.DescribeLaunchTemplatesOutput, bool) bool) error {
	page, err := m.DescribeLaunchTemplates(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

// DescribeLaunchTemplates mocks the describing the launch templates
func (m *MockEC2) DescribeLaunchTemplates(request *ec2.DescribeLaunchTemplatesInput) (*ec2.DescribeLaunchTemplatesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DescribeLaunchTemplates: %v", request)

	o := &ec2.DescribeLaunchTemplatesOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}

	for id, ltInfo := range m.LaunchTemplates {
		launchTemplatetName := aws.StringValue(ltInfo.name)

		allFiltersMatch := true
		for _, filter := range request.Filters {
			filterName := aws.StringValue(filter.Name)
			filterValue := aws.StringValue(filter.Values[0])

			filterMatches := false
			if filterName == "tag:Name" && filterValue == launchTemplatetName {
				filterMatches = true
			}
			if strings.HasPrefix(filterName, "tag:kubernetes.io/cluster/") {
				filterMatches = true
			}

			if !filterMatches {
				allFiltersMatch = false
				break
			}
		}

		if allFiltersMatch {
			o.LaunchTemplates = append(o.LaunchTemplates, &ec2.LaunchTemplate{
				LaunchTemplateName: aws.String(launchTemplatetName),
				LaunchTemplateId:   aws.String(id),
			})
		}
	}

	return o, nil
}

// DescribeLaunchTemplateVersions mocks the retrieval of launch template versions - we don't use this at the moment so we can just return the template
func (m *MockEC2) DescribeLaunchTemplateVersions(request *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DescribeLaunchTemplateVersions: %v", request)

	o := &ec2.DescribeLaunchTemplateVersionsOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}

	for id, ltInfo := range m.LaunchTemplates {
		if aws.StringValue(ltInfo.name) != aws.StringValue(request.LaunchTemplateName) {
			continue
		}
		o.LaunchTemplateVersions = append(o.LaunchTemplateVersions, &ec2.LaunchTemplateVersion{
			DefaultVersion:     aws.Bool(true),
			LaunchTemplateId:   aws.String(id),
			LaunchTemplateData: ltInfo.data,
			LaunchTemplateName: request.LaunchTemplateName,
		})
	}
	return o, nil
}

// CreateLaunchTemplate mocks the ec2 create launch template
func (m *MockEC2) CreateLaunchTemplate(request *ec2.CreateLaunchTemplateInput) (*ec2.CreateLaunchTemplateOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock CreateLaunchTemplate: %v", request)

	m.launchTemplateNumber++
	n := m.launchTemplateNumber
	id := fmt.Sprintf("lt-%d", n)

	if m.LaunchTemplates == nil {
		m.LaunchTemplates = make(map[string]*launchTemplateInfo)
	}
	if m.LaunchTemplates[id] != nil {
		return nil, fmt.Errorf("duplicate LaunchTemplateId %s", id)
	}
	m.LaunchTemplates[id] = &launchTemplateInfo{
		data:    responseLaunchTemplateData(request.LaunchTemplateData),
		name:    request.LaunchTemplateName,
		version: 1,
	}
	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2.ResourceTypeLaunchTemplate)...)

	return &ec2.CreateLaunchTemplateOutput{
		LaunchTemplate: &ec2.LaunchTemplate{
			LaunchTemplateId: aws.String(id),
		},
	}, nil
}

func (m *MockEC2) CreateLaunchTemplateVersion(request *ec2.CreateLaunchTemplateVersionInput) (*ec2.CreateLaunchTemplateVersionOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock CreateLaunchTemplateVersion: %v", request)

	name := request.LaunchTemplateName
	found := false
	var ltVersion int
	var ltID string
	for id, ltInfo := range m.LaunchTemplates {
		if aws.StringValue(ltInfo.name) == aws.StringValue(name) {
			found = true
			ltInfo.data = responseLaunchTemplateData(request.LaunchTemplateData)
			ltInfo.version++
			ltVersion = ltInfo.version
			ltID = id
		}
	}
	if !found {
		return nil, nil // TODO: error
	}
	return &ec2.CreateLaunchTemplateVersionOutput{
		LaunchTemplateVersion: &ec2.LaunchTemplateVersion{
			VersionNumber:    aws.Int64(int64(ltVersion)),
			LaunchTemplateId: &ltID,
		},
	}, nil
}

// DeleteLaunchTemplate mocks the deletion of a launch template
func (m *MockEC2) DeleteLaunchTemplate(request *ec2.DeleteLaunchTemplateInput) (*ec2.DeleteLaunchTemplateOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DeleteLaunchTemplate: %v", request)

	o := &ec2.DeleteLaunchTemplateOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}
	for id := range m.LaunchTemplates {
		if id == aws.StringValue(request.LaunchTemplateId) {
			delete(m.LaunchTemplates, id)
		}
	}

	return o, nil
}

func (m *MockEC2) ModifyLaunchTemplate(*ec2.ModifyLaunchTemplateInput) (*ec2.ModifyLaunchTemplateOutput, error) {
	return &ec2.ModifyLaunchTemplateOutput{}, nil
}

func responseLaunchTemplateData(req *ec2.RequestLaunchTemplateData) *ec2.ResponseLaunchTemplateData {
	resp := &ec2.ResponseLaunchTemplateData{
		DisableApiTermination: req.DisableApiTermination,
		EbsOptimized:          req.EbsOptimized,
		ImageId:               req.ImageId,
		InstanceType:          req.InstanceType,
		KeyName:               req.KeyName,
		SecurityGroupIds:      req.SecurityGroupIds,
		SecurityGroups:        req.SecurityGroups,
		UserData:              req.UserData,
	}

	if req.MetadataOptions != nil {
		resp.MetadataOptions = &ec2.LaunchTemplateInstanceMetadataOptions{
			HttpTokens:              req.MetadataOptions.HttpTokens,
			HttpPutResponseHopLimit: req.MetadataOptions.HttpPutResponseHopLimit,
			HttpProtocolIpv6:        req.MetadataOptions.HttpProtocolIpv6,
		}
	}
	if req.Monitoring != nil {
		resp.Monitoring = &ec2.LaunchTemplatesMonitoring{Enabled: req.Monitoring.Enabled}
	}
	if req.CpuOptions != nil {
		resp.CpuOptions = &ec2.LaunchTemplateCpuOptions{
			CoreCount:      req.CpuOptions.CoreCount,
			ThreadsPerCore: req.CpuOptions.ThreadsPerCore,
		}
	}
	if len(req.BlockDeviceMappings) > 0 {
		for _, x := range req.BlockDeviceMappings {
			var ebs *ec2.LaunchTemplateEbsBlockDevice
			if x.Ebs != nil {
				ebs = &ec2.LaunchTemplateEbsBlockDevice{
					DeleteOnTermination: x.Ebs.DeleteOnTermination,
					Encrypted:           x.Ebs.Encrypted,
					Iops:                x.Ebs.Iops,
					KmsKeyId:            x.Ebs.KmsKeyId,
					SnapshotId:          x.Ebs.SnapshotId,
					Throughput:          x.Ebs.Throughput,
					VolumeSize:          x.Ebs.VolumeSize,
					VolumeType:          x.Ebs.VolumeType,
				}
			}
			resp.BlockDeviceMappings = append(resp.BlockDeviceMappings, &ec2.LaunchTemplateBlockDeviceMapping{
				DeviceName:  x.DeviceName,
				Ebs:         ebs,
				NoDevice:    x.NoDevice,
				VirtualName: x.VirtualName,
			})
		}
	}
	if req.CreditSpecification != nil {
		resp.CreditSpecification = &ec2.CreditSpecification{CpuCredits: req.CreditSpecification.CpuCredits}
	}
	if req.IamInstanceProfile != nil {
		resp.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileSpecification{
			Arn:  req.IamInstanceProfile.Arn,
			Name: req.IamInstanceProfile.Name,
		}
	}
	if req.InstanceMarketOptions != nil {
		resp.InstanceMarketOptions = &ec2.LaunchTemplateInstanceMarketOptions{
			MarketType: req.InstanceMarketOptions.MarketType,
			SpotOptions: &ec2.LaunchTemplateSpotMarketOptions{
				BlockDurationMinutes:         req.InstanceMarketOptions.SpotOptions.BlockDurationMinutes,
				InstanceInterruptionBehavior: req.InstanceMarketOptions.SpotOptions.InstanceInterruptionBehavior,
				MaxPrice:                     req.InstanceMarketOptions.SpotOptions.MaxPrice,
				SpotInstanceType:             req.InstanceMarketOptions.SpotOptions.SpotInstanceType,
				ValidUntil:                   req.InstanceMarketOptions.SpotOptions.ValidUntil,
			},
		}
	}
	if len(req.NetworkInterfaces) > 0 {
		for _, x := range req.NetworkInterfaces {
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
	if len(req.TagSpecifications) > 0 {
		for _, x := range req.TagSpecifications {
			resp.TagSpecifications = append(resp.TagSpecifications, &ec2.LaunchTemplateTagSpecification{
				ResourceType: x.ResourceType,
				Tags:         x.Tags,
			})
		}
	}
	return resp
}
