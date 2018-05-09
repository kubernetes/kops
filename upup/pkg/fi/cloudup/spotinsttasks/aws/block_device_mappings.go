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

package aws

import (
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"k8s.io/kops/upup/pkg/fi"
)

type BlockDeviceMapping struct {
	VirtualName *string

	EbsDeleteOnTermination *bool
	EbsVolumeSize          *int64
	EbsVolumeType          *string
	EbsVolumeIOPS          *int64
}

func BlockDeviceMappingFromGroup(i *aws.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{}
	o.VirtualName = i.VirtualName
	if i.EBS != nil {
		o.EbsDeleteOnTermination = i.EBS.DeleteOnTermination
		o.EbsVolumeSize = spotinst.Int64(int64(spotinst.IntValue(i.EBS.VolumeSize)))
		o.EbsVolumeType = i.EBS.VolumeType
		o.EbsVolumeIOPS = spotinst.Int64(int64(spotinst.IntValue(i.EBS.IOPS)))
	}
	return spotinst.StringValue(i.DeviceName), o
}

func (i *BlockDeviceMapping) ToGroup(deviceName string) *aws.BlockDeviceMapping {
	o := &aws.BlockDeviceMapping{}
	o.DeviceName = spotinst.String(deviceName)
	o.VirtualName = i.VirtualName

	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil {
		o.EBS = &aws.EBS{}
		o.EBS.DeleteOnTermination = i.EbsDeleteOnTermination
		o.EBS.VolumeSize = spotinst.Int(int(spotinst.Int64Value(i.EbsVolumeSize)))
		o.EBS.VolumeType = i.EbsVolumeType

		// The parameter IOPS is not supported for gp2 volumes.
		if spotinst.StringValue(i.EbsVolumeType) != "gp2" {
			o.EBS.IOPS = spotinst.Int(int(spotinst.Int64Value(i.EbsVolumeIOPS)))
		}
	}

	return o
}

var _ fi.HasDependencies = &BlockDeviceMapping{}

func (f *BlockDeviceMapping) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}
