/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestBlockDeviceMappingVolumeInitializationRate(t *testing.T) {
	mapping := &BlockDeviceMapping{
		EbsDeleteOnTermination:      new(true),
		EbsVolumeSize:               new(int32(64)),
		EbsVolumeType:               ec2types.VolumeTypeGp3,
		EbsVolumeInitializationRate: new(int32(300)),
	}

	request := mapping.ToLaunchTemplateBootDeviceRequest("/dev/xvda")
	if got := request.Ebs.VolumeInitializationRate; got == nil || *got != 300 {
		t.Errorf("launch template request rate = %v, want 300", got)
	}

	_, actual := BlockDeviceMappingFromLaunchTemplateBootDeviceRequest(ec2types.LaunchTemplateBlockDeviceMapping{
		DeviceName: new("/dev/xvda"),
		Ebs:        &ec2types.LaunchTemplateEbsBlockDevice{VolumeInitializationRate: new(int32(300))},
	})
	if got := actual.EbsVolumeInitializationRate; got == nil || *got != 300 {
		t.Errorf("launch template mapping rate = %v, want 300", got)
	}

	terraform := createTerraformLaunchTemplateBlockDevice("/dev/xvda", mapping)
	if got := terraform.EBS[0].VolumeInitializationRate; got == nil || *got != 300 {
		t.Errorf("terraform rate = %v, want 300", got)
	}
}
