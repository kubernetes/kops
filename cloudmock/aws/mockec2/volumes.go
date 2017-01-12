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

package mockec2

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"strings"
)

func (m *MockEC2) DescribeVolumeAttributeRequest(*ec2.DescribeVolumeAttributeInput) (*request.Request, *ec2.DescribeVolumeAttributeOutput) {
	panic("MockEC2 DescribeVolumeAttributeRequest not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeVolumeAttribute(*ec2.DescribeVolumeAttributeInput) (*ec2.DescribeVolumeAttributeOutput, error) {
	panic("MockEC2 DescribeVolumeAttribute not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeVolumeStatusRequest(*ec2.DescribeVolumeStatusInput) (*request.Request, *ec2.DescribeVolumeStatusOutput) {
	panic("MockEC2 DescribeVolumeStatusRequest not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeVolumeStatus(*ec2.DescribeVolumeStatusInput) (*ec2.DescribeVolumeStatusOutput, error) {
	panic("MockEC2 DescribeVolumeStatus not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeVolumeStatusPages(*ec2.DescribeVolumeStatusInput, func(*ec2.DescribeVolumeStatusOutput, bool) bool) error {
	panic("MockEC2 DescribeVolumeStatusPages not implemented")
	return nil
}
func (m *MockEC2) DescribeVolumesRequest(*ec2.DescribeVolumesInput) (*request.Request, *ec2.DescribeVolumesOutput) {
	panic("MockEC2 DescribeVolumesRequest not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeVolumes(request *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	glog.Infof("DescribeVolumes: %v", request)

	var volumes []*ec2.Volume

	for _, volume := range m.Volumes {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeVolume, *volume.VolumeId, filter)
				} else {
					return nil, fmt.Errorf("unknown filter name: %q", *filter.Name)
				}
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := *volume
		copy.Tags = m.getTags(ec2.ResourceTypeVolume, *volume.VolumeId)
		volumes = append(volumes, &copy)
	}

	response := &ec2.DescribeVolumesOutput{
		Volumes: volumes,
	}

	return response, nil
}

func (m *MockEC2) DescribeVolumesPages(*ec2.DescribeVolumesInput, func(*ec2.DescribeVolumesOutput, bool) bool) error {
	panic("MockEC2 DescribeVolumesPages not implemented")
	return nil
}
