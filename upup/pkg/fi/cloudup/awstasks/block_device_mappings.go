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
	"k8s.io/kops/upup/pkg/fi"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// BlockDeviceMapping defines the specification for a device mapping
type BlockDeviceMapping struct {
	// DeviceName is the device name of the EBS
	DeviceName *string
	// EbsDeleteOnTermination indicates the volume should be delete on instance termination
	EbsDeleteOnTermination *bool
	// EbsEncrypted indicates the volume should be encrypted
	EbsEncrypted *bool
	// EbsVolumeIops is provisioned iops
	EbsVolumeIops *int64
	// EbsVolumeSize is the size of the volume
	EbsVolumeSize *int64
	// EbsVolumeType is the aws volume type
	EbsVolumeType *string
	// VirtualName is the device name
	VirtualName *string
}

// BlockDeviceMappingFromEC2 converts a e2c block mapping to internal block device mapping
func BlockDeviceMappingFromEC2(i *ec2.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{
		DeviceName:  i.DeviceName,
		VirtualName: i.VirtualName,
	}
	if i.Ebs != nil {
		o.EbsDeleteOnTermination = i.Ebs.DeleteOnTermination
		o.EbsEncrypted = i.Ebs.Encrypted
		o.EbsVolumeIops = i.Ebs.Iops
		o.EbsVolumeSize = i.Ebs.VolumeSize
		o.EbsVolumeType = i.Ebs.VolumeType
	}

	return aws.StringValue(i.DeviceName), o
}

// ToEC2 creates and returns an ec2 block mapping
func (i *BlockDeviceMapping) ToEC2(deviceName string) *ec2.BlockDeviceMapping {
	o := &ec2.BlockDeviceMapping{
		DeviceName:  aws.String(deviceName),
		VirtualName: i.VirtualName,
	}
	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil || i.EbsVolumeIops != nil {
		o.Ebs = &ec2.EbsBlockDevice{
			DeleteOnTermination: i.EbsDeleteOnTermination,
			Encrypted:           i.EbsEncrypted,
			VolumeSize:          i.EbsVolumeSize,
			VolumeType:          i.EbsVolumeType,
		}
		if fi.StringValue(o.Ebs.VolumeType) == ec2.VolumeTypeIo1 {
			o.Ebs.Iops = i.EbsVolumeIops
		}
	}

	return o
}

// BlockDeviceMappingFromAutoscaling converts an autoscaling block mapping to internal spec
func BlockDeviceMappingFromAutoscaling(i *autoscaling.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{
		DeviceName:  i.DeviceName,
		VirtualName: i.VirtualName,
	}
	if i.Ebs != nil {
		o.EbsDeleteOnTermination = i.Ebs.DeleteOnTermination
		o.EbsEncrypted = i.Ebs.Encrypted
		o.EbsVolumeSize = i.Ebs.VolumeSize
		o.EbsVolumeType = i.Ebs.VolumeType

		if fi.StringValue(o.EbsVolumeType) == ec2.VolumeTypeIo1 {
			o.EbsVolumeIops = i.Ebs.Iops
		}
	}

	return aws.StringValue(i.DeviceName), o
}

// ToAutoscaling converts the internal block mapping to autoscaling
func (i *BlockDeviceMapping) ToAutoscaling(deviceName string) *autoscaling.BlockDeviceMapping {
	o := &autoscaling.BlockDeviceMapping{
		DeviceName:  aws.String(deviceName),
		VirtualName: i.VirtualName,
	}
	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil {
		o.Ebs = &autoscaling.Ebs{
			DeleteOnTermination: i.EbsDeleteOnTermination,
			Encrypted:           i.EbsEncrypted,
			VolumeSize:          i.EbsVolumeSize,
			VolumeType:          i.EbsVolumeType,
		}
		if fi.StringValue(o.Ebs.VolumeType) == ec2.VolumeTypeIo1 {
			o.Ebs.Iops = i.EbsVolumeIops
		}
	}

	return o
}

// BlockDeviceMappingFromLaunchTemplateBootDeviceRequest coverts the launch template device mappings to an interval block device mapping
func BlockDeviceMappingFromLaunchTemplateBootDeviceRequest(i *ec2.LaunchTemplateBlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{
		DeviceName:  i.DeviceName,
		VirtualName: i.VirtualName,
	}
	if i.Ebs != nil {
		o.EbsDeleteOnTermination = i.Ebs.DeleteOnTermination
		o.EbsVolumeSize = i.Ebs.VolumeSize
		o.EbsVolumeType = i.Ebs.VolumeType
		o.EbsEncrypted = i.Ebs.Encrypted
	}

	return aws.StringValue(i.DeviceName), o
}

// ToLaunchTemplateBootDeviceRequest coverts in the internal block device mapping to a launcg template request
func (i *BlockDeviceMapping) ToLaunchTemplateBootDeviceRequest(deviceName string) *ec2.LaunchTemplateBlockDeviceMappingRequest {
	o := &ec2.LaunchTemplateBlockDeviceMappingRequest{
		DeviceName:  aws.String(deviceName),
		VirtualName: i.VirtualName,
	}
	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil || i.EbsVolumeIops != nil || i.EbsEncrypted != nil {
		o.Ebs = &ec2.LaunchTemplateEbsBlockDeviceRequest{
			DeleteOnTermination: i.EbsDeleteOnTermination,
			Encrypted:           i.EbsEncrypted,
			VolumeSize:          i.EbsVolumeSize,
			VolumeType:          i.EbsVolumeType,
			Iops:                i.EbsVolumeIops,
		}
	}

	return o
}

var _ fi.HasDependencies = &BlockDeviceMapping{}

// GetDependencies is for future use
func (i *BlockDeviceMapping) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}
