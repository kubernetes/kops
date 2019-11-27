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

package awstasks

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
)

type cloudformationLaunchTemplateNetworkInterfaces struct {
	// AssociatePublicIPAddress associates a public ip address with the network interface. Boolean value.
	AssociatePublicIPAddress *bool `json:"AssociatePublicIpAddress,omitempty"`
	// DeleteOnTermination indicates whether the network interface should be destroyed on instance termination.
	DeleteOnTermination *bool `json:"DeleteOnTermination,omitempty"`
}

type cloudformationLaunchTemplateMonitoring struct {
	// Enabled indicates that monitoring is enabled
	Enabled *bool `json:"Enabled,omitempty"`
}

type cloudformationLaunchTemplatePlacement struct {
	// Affinity is he affinity setting for an instance on a Dedicated Host.
	Affinity *string `json:"Affinity,omitempty"`
	// AvailabilityZone is the Availability Zone for the instance.
	AvailabilityZone *string `json:"AvailabilityZone,omitempty"`
	// GroupName is the name of the placement group for the instance.
	GroupName *string `json:"GroupName,omitempty"`
	// HostID is the ID of the Dedicated Host for the instance.
	HostID *string `json:"HostId,omitempty"`
	// SpreadDomain are reserved for future use.
	SpreadDomain *string `json:"SpreadDomain,omitempty"`
	// Tenancy ist he tenancy of the instance. Can be default, dedicated, or host.
	Tenancy *string `json:"Tenancy,omitempty"`
}

type cloudformationLaunchTemplateIAMProfile struct {
	// Name is the name of the profile
	Name *cloudformation.Literal `json:"Name,omitempty"`
}

type cloudformationLaunchTemplateMarketOptionsSpotOptions struct {
	// InstancesInterruptionBehavior is the behavior when a Spot Instance is interrupted. Can be hibernate, stop, or terminate
	InstancesInterruptionBehavior *string `json:"InstancesInterruptionBehavior,omitempty"`
	// MaxPrice is the maximum hourly price you're willing to pay for the Spot Instances
	MaxPrice *string `json:"MaxPrice,omitempty"`
	// SpotInstanceType is the Spot Instance request type. Can be one-time, or persistent
	SpotInstanceType *string `json:"SpotInstanceType,omitempty"`
}

type cloudformationLaunchTemplateMarketOptions struct {
	// MarketType is the option type
	MarketType *string `json:"MarketType,omitempty"`
	// SpotOptions are the set of options
	SpotOptions []*cloudformationLaunchTemplateMarketOptionsSpotOptions `json:"Options,omitempty"`
}

type cloudformationLaunchTemplateBlockDeviceEBS struct {
	// VolumeType is the ebs type to use
	VolumeType *string `json:"VolumeType,omitempty"`
	// VolumeSize is the volume size
	VolumeSize *int64 `json:"VolumeSize,omitempty"`
	// IOPS is the provisioned iops
	IOPS *int64 `json:"Iops,omitempty"`
	// DeleteOnTermination indicates the volume should die with the instance
	DeleteOnTermination *bool `json:"DeleteOnTermination,omitempty"`
	// Encrypted indicates the device is encrypted
	Encrypted *bool `json:"Encrypted,omitempty"`
}

type cloudformationLaunchTemplateBlockDevice struct {
	// DeviceName is the name of the device
	DeviceName *string `json:"DeviceName,omitempty"`
	// VirtualName is used for the ephemeral devices
	VirtualName *string `json:"VirtualName,omitempty"`
	// EBS defines the ebs spec
	EBS *cloudformationLaunchTemplateBlockDeviceEBS `json:"EBS,omitempty"`
}

type cloudformationLaunchTemplateData struct {
	// BlockDeviceMappings is the device mappings
	BlockDeviceMappings []*cloudformationLaunchTemplateBlockDevice `json:"BlockDeviceMappings,omitempty"`
	// EBSOptimized indicates if the root device is ebs optimized
	EBSOptimized *bool `json:"EbsOptimized,omitempty"`
	// IAMInstanceProfile is the IAM profile to assign to the nodes
	IAMInstanceProfile *cloudformationLaunchTemplateIAMProfile `json:"IamInstanceProfile,omitempty"`
	// ImageID is the ami to use for the instances
	ImageID *string `json:"ImageId,omitempty"`
	// InstanceType is the type of instance
	InstanceType *string `json:"InstanceType,omitempty"`
	// KeyName is the ssh key to use
	KeyName *string `json:"KeyName,omitempty"`
	// MarketOptions are the spot pricing options
	MarketOptions *cloudformationLaunchTemplateMarketOptions `json:"InstanceMarketOptions,omitempty"`
	// Monitoring are the instance monitoring options
	Monitoring *cloudformationLaunchTemplateMonitoring `json:"Monitoring,omitempty"`
	// NetworkInterfaces are the networking options
	NetworkInterfaces []*cloudformationLaunchTemplateNetworkInterfaces `json:"NetworkInterfaces,omitempty"`
	// Placement are the tenancy options
	Placement []*cloudformationLaunchTemplatePlacement `json:"Placement,omitempty"`
	// UserData is the user data for the instances
	UserData *string `json:"UserData,omitempty"`
	// VpcSecurityGroupIDs is a list of security group ids
	VpcSecurityGroupIDs []*cloudformation.Literal `json:"SecurityGroup,omitempty"`
}

type cloudformationLaunchTemplate struct {
	// LaunchTemplateName is the name of the launch template
	LaunchTemplateName *string `json:"LaunchTemplateName,omitempty"`
	// LaunchTemplateData is the data request
	LaunchTemplateData *cloudformationLaunchTemplateData `json:"LaunchTemplateData,omitempty"`
}

// CloudformationLink returns the cloudformation link for us
func (t *LaunchTemplate) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::EC2::LaunchTemplate", fi.StringValue(t.Name))
}

// RenderCloudformation is responsible for rendering the cloudformation json
func (t *LaunchTemplate) RenderCloudformation(target *cloudformation.CloudformationTarget, a, e, changes *LaunchTemplate) error {
	var err error

	cloud := target.Cloud.(awsup.AWSCloud)

	var image *string
	if e.ImageID != nil {
		im, err := cloud.ResolveImage(fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}
		image = im.ImageId
	}

	cf := &cloudformationLaunchTemplate{
		LaunchTemplateName: fi.String(fi.StringValue(e.Name)),
		LaunchTemplateData: &cloudformationLaunchTemplateData{
			EBSOptimized: e.RootVolumeOptimization,
			ImageID:      image,
			InstanceType: e.InstanceType,
			NetworkInterfaces: []*cloudformationLaunchTemplateNetworkInterfaces{
				{AssociatePublicIPAddress: e.AssociatePublicIP,
					DeleteOnTermination: fi.Bool(true)},
			},
		},
	}
	data := cf.LaunchTemplateData

	if e.SpotPrice != "" {
		data.MarketOptions = &cloudformationLaunchTemplateMarketOptions{
			MarketType:  fi.String("spot"),
			SpotOptions: []*cloudformationLaunchTemplateMarketOptionsSpotOptions{{MaxPrice: fi.String(e.SpotPrice)}},
		}
	}
	for _, x := range e.SecurityGroups {
		data.VpcSecurityGroupIDs = append(data.VpcSecurityGroupIDs, x.CloudformationLink())
	}
	if e.SSHKey != nil {
		data.KeyName = e.SSHKey.Name
	}
	if e.Tenancy != nil {
		data.Placement = []*cloudformationLaunchTemplatePlacement{{Tenancy: e.Tenancy}}
	}
	if e.IAMInstanceProfile != nil {
		data.IAMInstanceProfile = &cloudformationLaunchTemplateIAMProfile{
			Name: e.IAMInstanceProfile.CloudformationLink(),
		}
	}
	if e.UserData != nil {
		d, err := e.UserData.AsBytes()
		if err != nil {
			return err
		}
		data.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
	}
	devices, err := e.buildRootDevice(cloud)
	if err != nil {
		return err
	}
	additionals, err := buildAdditionalDevices(e.BlockDeviceMappings)
	if err != nil {
		return err
	}
	for name, x := range devices {
		data.BlockDeviceMappings = append(data.BlockDeviceMappings, &cloudformationLaunchTemplateBlockDevice{
			DeviceName: fi.String(name),
			EBS: &cloudformationLaunchTemplateBlockDeviceEBS{
				DeleteOnTermination: fi.Bool(true),
				IOPS:                x.EbsVolumeIops,
				VolumeSize:          x.EbsVolumeSize,
				VolumeType:          x.EbsVolumeType,
			},
		})
	}
	for name, x := range additionals {
		data.BlockDeviceMappings = append(data.BlockDeviceMappings, &cloudformationLaunchTemplateBlockDevice{
			DeviceName: fi.String(name),
			EBS: &cloudformationLaunchTemplateBlockDeviceEBS{
				DeleteOnTermination: fi.Bool(true),
				IOPS:                x.EbsVolumeIops,
				VolumeSize:          x.EbsVolumeSize,
				VolumeType:          x.EbsVolumeType,
				Encrypted:           x.EbsEncrypted,
			},
		})
	}

	devices, err = buildEphemeralDevices(cloud, fi.StringValue(e.InstanceType))
	if err != nil {
		return err
	}
	for n, x := range devices {
		data.BlockDeviceMappings = append(data.BlockDeviceMappings, &cloudformationLaunchTemplateBlockDevice{
			VirtualName: x.VirtualName,
			DeviceName:  fi.String(n),
		})
	}

	return target.RenderResource("AWS::EC2::LaunchTemplate", fi.StringValue(e.Name), cf)
}
