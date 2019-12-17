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

package defaults

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
)

const (
	// DefaultVolumeSizeBastion is the default root disk size of a bastion
	DefaultVolumeSizeBastion = 32
	// DefaultVolumeSizeMaster is the default root disk size of a master
	DefaultVolumeSizeMaster = 64
	// DefaultVolumeSizeNode is the default root disk size of a node
	DefaultVolumeSizeNode = 128
)

// DefaultInstanceGroupVolumeSize returns the default volume size for nodes in an InstanceGroup with the specified role
func DefaultInstanceGroupVolumeSize(role kops.InstanceGroupRole) (int32, error) {
	switch role {
	case kops.InstanceGroupRoleMaster:
		return DefaultVolumeSizeMaster, nil
	case kops.InstanceGroupRoleNode:
		return DefaultVolumeSizeNode, nil
	case kops.InstanceGroupRoleBastion:
		return DefaultVolumeSizeBastion, nil
	default:
		return -1, fmt.Errorf("unknown InstanceGroup Role %s", role)
	}
}
