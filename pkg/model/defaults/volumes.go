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

package defaults

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
)

const (
	DefaultVolumeSizeNode    = 128
	DefaultVolumeSizeMaster  = 64
	DefaultVolumeSizeBastion = 32
)

// FindDefaultVolumeSize returns the default volume size based on role.
func FindDefaultVolumeSize(volumeSize int32, role kops.InstanceGroupRole) (int32, error) {
	if volumeSize == 0 {
		switch role {
		case kops.InstanceGroupRoleMaster:
			volumeSize = DefaultVolumeSizeMaster
		case kops.InstanceGroupRoleNode:
			volumeSize = DefaultVolumeSizeNode
		case kops.InstanceGroupRoleBastion:
			volumeSize = DefaultVolumeSizeBastion
		default:
			return -1, fmt.Errorf("this case should not get hit, kops.Role not found %s", role)
		}
	}

	return volumeSize, nil
}
