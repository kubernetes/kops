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
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func (m *MockEC2) CreateVolume(request *ec2.CreateVolumeInput) (*ec2.Volume, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateVolume: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	n := len(m.Volumes) + 1

	volume := &ec2.Volume{
		VolumeId:         s(fmt.Sprintf("vol-%d", n)),
		AvailabilityZone: request.AvailabilityZone,
		Encrypted:        request.Encrypted,
		Iops:             request.Iops,
		KmsKeyId:         request.KmsKeyId,
		Size:             request.Size,
		SnapshotId:       request.SnapshotId,
		VolumeType:       request.VolumeType,
	}

	for _, tags := range request.TagSpecifications {
		for _, tag := range tags.Tags {
			m.addTag(*volume.VolumeId, tag)
		}
	}
	if m.Volumes == nil {
		m.Volumes = make(map[string]*ec2.Volume)
	}
	m.Volumes[*volume.VolumeId] = volume

	copy := *volume
	copy.Tags = m.getTags(ec2.ResourceTypeVolume, *volume.VolumeId)

	// TODO: a few fields
	// // Information about the volume attachments.
	// Attachments []*VolumeAttachment `locationName:"attachmentSet" locationNameList:"item" type:"list"`

	// // The time stamp when volume creation was initiated.
	// CreateTime *time.Time `locationName:"createTime" type:"timestamp" timestampFormat:"iso8601"`

	// // The volume state.
	// State *string `locationName:"status" type:"string" enum:"VolumeState"`

	return &copy, nil
}

func (m *MockEC2) CreateVolumeWithContext(aws.Context, *ec2.CreateVolumeInput, ...request.Option) (*ec2.Volume, error) {
	panic("Not implemented")
}
func (m *MockEC2) CreateVolumeRequest(*ec2.CreateVolumeInput) (*request.Request, *ec2.Volume) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeVolumeAttributeRequest(*ec2.DescribeVolumeAttributeInput) (*request.Request, *ec2.DescribeVolumeAttributeOutput) {
	panic("MockEC2 DescribeVolumeAttributeRequest not implemented")
}
func (m *MockEC2) DescribeVolumeAttributeWithContext(aws.Context, *ec2.DescribeVolumeAttributeInput, ...request.Option) (*ec2.DescribeVolumeAttributeOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeVolumeAttribute(*ec2.DescribeVolumeAttributeInput) (*ec2.DescribeVolumeAttributeOutput, error) {
	panic("MockEC2 DescribeVolumeAttribute not implemented")
}
func (m *MockEC2) DescribeVolumeStatusRequest(*ec2.DescribeVolumeStatusInput) (*request.Request, *ec2.DescribeVolumeStatusOutput) {
	panic("MockEC2 DescribeVolumeStatusRequest not implemented")
}
func (m *MockEC2) DescribeVolumeStatusWithContext(aws.Context, *ec2.DescribeVolumeStatusInput, ...request.Option) (*ec2.DescribeVolumeStatusOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeVolumeStatus(*ec2.DescribeVolumeStatusInput) (*ec2.DescribeVolumeStatusOutput, error) {
	panic("MockEC2 DescribeVolumeStatus not implemented")
}
func (m *MockEC2) DescribeVolumeStatusPages(*ec2.DescribeVolumeStatusInput, func(*ec2.DescribeVolumeStatusOutput, bool) bool) error {
	panic("MockEC2 DescribeVolumeStatusPages not implemented")
}
func (m *MockEC2) DescribeVolumeStatusPagesWithContext(aws.Context, *ec2.DescribeVolumeStatusInput, func(*ec2.DescribeVolumeStatusOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}
func (m *MockEC2) DescribeVolumesRequest(*ec2.DescribeVolumesInput) (*request.Request, *ec2.DescribeVolumesOutput) {
	panic("MockEC2 DescribeVolumesRequest not implemented")
}
func (m *MockEC2) DescribeVolumesWithContext(aws.Context, *ec2.DescribeVolumesInput, ...request.Option) (*ec2.DescribeVolumesOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeVolumes(request *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeVolumes: %v", request)

	if request.VolumeIds != nil {
		klog.Fatalf("VolumeIds")
	}

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

func (m *MockEC2) DescribeVolumesPages(request *ec2.DescribeVolumesInput, callback func(*ec2.DescribeVolumesOutput, bool) bool) error {
	// For the mock, we just send everything in one page
	page, err := m.DescribeVolumes(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockEC2) DescribeVolumesPagesWithContext(aws.Context, *ec2.DescribeVolumesInput, func(*ec2.DescribeVolumesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockEC2) DescribeVolumesModifications(*ec2.DescribeVolumesModificationsInput) (*ec2.DescribeVolumesModificationsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeVolumesModificationsWithContext(aws.Context, *ec2.DescribeVolumesModificationsInput, ...request.Option) (*ec2.DescribeVolumesModificationsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeVolumesModificationsRequest(*ec2.DescribeVolumesModificationsInput) (*request.Request, *ec2.DescribeVolumesModificationsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteVolume(request *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteVolume: %v", request)

	id := aws.StringValue(request.VolumeId)
	o := m.Volumes[id]
	if o == nil {
		return nil, fmt.Errorf("Volume %q not found", id)
	}
	delete(m.Volumes, id)

	return &ec2.DeleteVolumeOutput{}, nil
}

func (m *MockEC2) DeleteVolumeWithContext(aws.Context, *ec2.DeleteVolumeInput, ...request.Option) (*ec2.DeleteVolumeOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DeleteVolumeRequest(*ec2.DeleteVolumeInput) (*request.Request, *ec2.DeleteVolumeOutput) {
	panic("Not implemented")
}
