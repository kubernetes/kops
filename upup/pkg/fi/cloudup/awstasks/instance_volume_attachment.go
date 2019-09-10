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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type InstanceVolumeAttachment struct {
	Instance *Instance
	Volume   *EBSVolume
	Device   *string
}

func (e *InstanceVolumeAttachment) String() string {
	return fi.TaskAsString(e)
}

func (e *InstanceVolumeAttachment) Find(c *fi.Context) (*InstanceVolumeAttachment, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	instanceID := e.Instance.ID
	volumeID := e.Volume.ID

	if instanceID == nil || volumeID == nil {
		return nil, nil
	}

	instance, err := cloud.DescribeInstance(*instanceID)
	if err != nil {
		return nil, err
	}

	for _, bdm := range instance.BlockDeviceMappings {
		if bdm.Ebs == nil {
			continue
		}
		if aws.StringValue(bdm.Ebs.VolumeId) != *volumeID {
			continue
		}

		actual := &InstanceVolumeAttachment{
			Device:   bdm.DeviceName,
			Instance: &Instance{ID: instance.InstanceId},
			Volume:   &EBSVolume{ID: bdm.Ebs.VolumeId},
		}

		klog.V(2).Infof("found matching InstanceVolumeAttachment")
		return actual, nil
	}

	return nil, nil
}

func (e *InstanceVolumeAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *InstanceVolumeAttachment) CheckChanges(a, e, changes *InstanceVolumeAttachment) error {
	if a != nil {
		if changes.Device != nil {
			// TODO: Support this?
			return fi.CannotChangeField("Device")
		}
	}

	if a == nil {
		if e.Device == nil {
			return fi.RequiredField("Device")
		}
	}
	return nil
}

func (_ *InstanceVolumeAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InstanceVolumeAttachment) error {
	if a == nil {
		err := t.WaitForInstanceRunning(*e.Instance.ID)
		if err != nil {
			return err
		}

		request := &ec2.AttachVolumeInput{
			InstanceId: e.Instance.ID,
			VolumeId:   e.Volume.ID,
			Device:     e.Device,
		}

		_, err = t.Cloud.EC2().AttachVolume(request)
		if err != nil {
			return fmt.Errorf("error creating InstanceVolumeAttachment: %v", err)
		}
	}

	return nil // no tags
}
