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
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

type launchTemplateInfo struct {
	data    *ec2types.ResponseLaunchTemplateData
	name    *string
	version int
}

// DescribeLaunchTemplates mocks the describing the launch templates
func (m *MockEC2) DescribeLaunchTemplates(ctx context.Context, request *ec2.DescribeLaunchTemplatesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplatesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DescribeLaunchTemplates: %v", request)

	o := &ec2.DescribeLaunchTemplatesOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}

	for id, ltInfo := range m.LaunchTemplates {
		launchTemplatetName := aws.ToString(ltInfo.name)

		allFiltersMatch := true
		for _, filter := range request.Filters {
			filterName := aws.ToString(filter.Name)
			filterValue := filter.Values[0]

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
			o.LaunchTemplates = append(o.LaunchTemplates, ec2types.LaunchTemplate{
				LaunchTemplateName: aws.String(launchTemplatetName),
				LaunchTemplateId:   aws.String(id),
			})
		}
	}

	return o, nil
}

// DescribeLaunchTemplateVersions mocks the retrieval of launch template versions - we don't use this at the moment so we can just return the template
func (m *MockEC2) DescribeLaunchTemplateVersions(ctx context.Context, request *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DescribeLaunchTemplateVersions: %v", request)

	o := &ec2.DescribeLaunchTemplateVersionsOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}

	for id, ltInfo := range m.LaunchTemplates {
		if aws.ToString(ltInfo.name) != aws.ToString(request.LaunchTemplateName) {
			continue
		}
		o.LaunchTemplateVersions = append(o.LaunchTemplateVersions, ec2types.LaunchTemplateVersion{
			DefaultVersion:     aws.Bool(true),
			LaunchTemplateId:   aws.String(id),
			LaunchTemplateData: ltInfo.data,
			LaunchTemplateName: request.LaunchTemplateName,
		})
	}
	return o, nil
}

// CreateLaunchTemplate mocks the ec2 create launch template
func (m *MockEC2) CreateLaunchTemplate(ctx context.Context, request *ec2.CreateLaunchTemplateInput, optFns ...func(*ec2.Options)) (*ec2.CreateLaunchTemplateOutput, error) {
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
	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeLaunchTemplate)...)

	return &ec2.CreateLaunchTemplateOutput{
		LaunchTemplate: &ec2types.LaunchTemplate{
			LaunchTemplateId: aws.String(id),
		},
	}, nil
}

func (m *MockEC2) CreateLaunchTemplateVersion(ctx context.Context, request *ec2.CreateLaunchTemplateVersionInput, optFns ...func(*ec2.Options)) (*ec2.CreateLaunchTemplateVersionOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock CreateLaunchTemplateVersion: %v", request)

	name := request.LaunchTemplateName
	found := false
	var ltVersion int
	var ltID string
	for id, ltInfo := range m.LaunchTemplates {
		if aws.ToString(ltInfo.name) == aws.ToString(name) {
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
		LaunchTemplateVersion: &ec2types.LaunchTemplateVersion{
			VersionNumber:    aws.Int64(int64(ltVersion)),
			LaunchTemplateId: &ltID,
		},
	}, nil
}

// DeleteLaunchTemplate mocks the deletion of a launch template
func (m *MockEC2) DeleteLaunchTemplate(ctx context.Context, request *ec2.DeleteLaunchTemplateInput, optFns ...func(*ec2.Options)) (*ec2.DeleteLaunchTemplateOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("Mock DeleteLaunchTemplate: %v", request)

	o := &ec2.DeleteLaunchTemplateOutput{}

	if m.LaunchTemplates == nil {
		return o, nil
	}
	for id := range m.LaunchTemplates {
		if id == aws.ToString(request.LaunchTemplateId) {
			delete(m.LaunchTemplates, id)
		}
	}

	return o, nil
}

func (m *MockEC2) ModifyLaunchTemplate(ctx context.Context, request *ec2.ModifyLaunchTemplateInput, optFns ...func(*ec2.Options)) (*ec2.ModifyLaunchTemplateOutput, error) {
	return &ec2.ModifyLaunchTemplateOutput{}, nil
}

func responseLaunchTemplateData(req *ec2types.RequestLaunchTemplateData) *ec2types.ResponseLaunchTemplateData {
	resp := &ec2types.ResponseLaunchTemplateData{
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
		resp.MetadataOptions = &ec2types.LaunchTemplateInstanceMetadataOptions{
			HttpTokens:              req.MetadataOptions.HttpTokens,
			HttpPutResponseHopLimit: req.MetadataOptions.HttpPutResponseHopLimit,
			HttpProtocolIpv6:        req.MetadataOptions.HttpProtocolIpv6,
		}
	}
	if req.Monitoring != nil {
		resp.Monitoring = &ec2types.LaunchTemplatesMonitoring{Enabled: req.Monitoring.Enabled}
	}
	if req.CpuOptions != nil {
		resp.CpuOptions = &ec2types.LaunchTemplateCpuOptions{
			CoreCount:      req.CpuOptions.CoreCount,
			ThreadsPerCore: req.CpuOptions.ThreadsPerCore,
		}
	}
	if len(req.BlockDeviceMappings) > 0 {
		for _, x := range req.BlockDeviceMappings {
			var ebs *ec2types.LaunchTemplateEbsBlockDevice
			if x.Ebs != nil {
				ebs = &ec2types.LaunchTemplateEbsBlockDevice{
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
			resp.BlockDeviceMappings = append(resp.BlockDeviceMappings, ec2types.LaunchTemplateBlockDeviceMapping{
				DeviceName:  x.DeviceName,
				Ebs:         ebs,
				NoDevice:    x.NoDevice,
				VirtualName: x.VirtualName,
			})
		}
	}
	if req.CreditSpecification != nil {
		resp.CreditSpecification = &ec2types.CreditSpecification{CpuCredits: req.CreditSpecification.CpuCredits}
	}
	if req.IamInstanceProfile != nil {
		resp.IamInstanceProfile = &ec2types.LaunchTemplateIamInstanceProfileSpecification{
			Arn:  req.IamInstanceProfile.Arn,
			Name: req.IamInstanceProfile.Name,
		}
	}
	if req.InstanceMarketOptions != nil {
		resp.InstanceMarketOptions = &ec2types.LaunchTemplateInstanceMarketOptions{
			MarketType: req.InstanceMarketOptions.MarketType,
			SpotOptions: &ec2types.LaunchTemplateSpotMarketOptions{
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
			resp.NetworkInterfaces = append(resp.NetworkInterfaces, ec2types.LaunchTemplateInstanceNetworkInterfaceSpecification{
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
			resp.TagSpecifications = append(resp.TagSpecifications, ec2types.LaunchTemplateTagSpecification{
				ResourceType: x.ResourceType,
				Tags:         x.Tags,
			})
		}
	}
	return resp
}
