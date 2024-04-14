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

func (m *MockEC2) CreateVolume(ctx context.Context, request *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateVolume: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	n := len(m.Volumes) + 1
	id := fmt.Sprintf("vol-%d", n)

	volume := &ec2types.Volume{
		VolumeId:         s(id),
		AvailabilityZone: request.AvailabilityZone,
		Encrypted:        request.Encrypted,
		Iops:             request.Iops,
		KmsKeyId:         request.KmsKeyId,
		Size:             request.Size,
		SnapshotId:       request.SnapshotId,
		Throughput:       request.Throughput,
		VolumeType:       request.VolumeType,
	}

	if m.Volumes == nil {
		m.Volumes = make(map[string]*ec2types.Volume)
	}
	m.Volumes[*volume.VolumeId] = volume

	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeVolume)...)

	copy := *volume
	copy.Tags = m.getTags(ec2types.ResourceTypeVolume, *volume.VolumeId)
	// TODO: a few fields
	// // Information about the volume attachments.
	// Attachments []*VolumeAttachment `locationName:"attachmentSet" locationNameList:"item" type:"list"`

	// // The time stamp when volume creation was initiated.
	// CreateTime *time.Time `locationName:"createTime" type:"timestamp" timestampFormat:"iso8601"`

	// // The volume state.
	// State *string `locationName:"status" type:"string" enum:"VolumeState"`

	return &ec2.CreateVolumeOutput{
		VolumeId: copy.VolumeId,
	}, nil
}

func (m *MockEC2) DescribeVolumes(ctx context.Context, request *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeVolumes: %v", request)

	if request.VolumeIds != nil {
		klog.Fatalf("VolumeIds")
	}

	var volumes []ec2types.Volume

	for _, volume := range m.Volumes {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2types.ResourceTypeVolume, *volume.VolumeId, filter)
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
		copy.Tags = m.getTags(ec2types.ResourceTypeVolume, *volume.VolumeId)
		volumes = append(volumes, copy)
	}

	response := &ec2.DescribeVolumesOutput{
		Volumes: volumes,
	}

	return response, nil
}

func (m *MockEC2) DeleteVolume(ctx context.Context, request *ec2.DeleteVolumeInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVolumeOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteVolume: %v", request)

	id := aws.ToString(request.VolumeId)
	o := m.Volumes[id]
	if o == nil {
		return nil, fmt.Errorf("Volume %q not found", id)
	}
	delete(m.Volumes, id)

	return &ec2.DeleteVolumeOutput{}, nil
}
