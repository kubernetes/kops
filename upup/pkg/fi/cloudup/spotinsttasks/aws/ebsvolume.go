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
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=EBSVolume
type EBSVolume struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID               *string
	AvailabilityZone *string
	VolumeType       *string
	SizeGB           *int64
	KmsKeyId         *string
	Encrypted        *bool
	Tags             map[string]string
}

var _ fi.CompareWithID = &EBSVolume{}

func (e *EBSVolume) CompareWithID() *string {
	return e.ID
}

type TaggableResource interface {
	FindResourceID(c fi.Cloud) (*string, error)
}

var _ TaggableResource = &EBSVolume{}

func (e *EBSVolume) FindResourceID(c fi.Cloud) (*string, error) {
	actual, err := e.find(c.(spotinst.Cloud).Cloud().(awsup.AWSCloud))
	if err != nil {
		return nil, fmt.Errorf("error querying for EBSVolume: %v", err)
	}
	if actual == nil {
		return nil, nil
	}
	return actual.ID, nil
}

func (e *EBSVolume) Find(c *fi.Context) (*EBSVolume, error) {
	actual, err := e.find(c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud))
	if actual != nil && err == nil {
		e.ID = actual.ID
	}
	return actual, err
}

func (e *EBSVolume) find(cloud awsup.AWSCloud) (*EBSVolume, error) {
	filters := cloud.BuildFilters(e.Name)
	request := &ec2.DescribeVolumesInput{
		Filters: filters,
	}

	response, err := cloud.EC2().DescribeVolumes(request)
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %v", err)
	}

	if response == nil || len(response.Volumes) == 0 {
		return nil, nil
	}

	if len(response.Volumes) != 1 {
		return nil, fmt.Errorf("found multiple Volumes with name: %s", *e.Name)
	}
	glog.V(2).Info("found existing volume")
	v := response.Volumes[0]
	actual := &EBSVolume{
		ID:               v.VolumeId,
		AvailabilityZone: v.AvailabilityZone,
		VolumeType:       v.VolumeType,
		SizeGB:           v.Size,
		KmsKeyId:         v.KmsKeyId,
		Encrypted:        v.Encrypted,
		Name:             e.Name,
	}

	actual.Tags = mapEC2TagsToMap(v.Tags)

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *EBSVolume) Run(c *fi.Context) error {
	c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddTags(e.Name, e.Tags)
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *EBSVolume) CheckChanges(a, e, changes *EBSVolume) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	if a != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
	}
	return nil
}

func (_ *EBSVolume) Render(t *spotinst.Target, a, e, changes *EBSVolume) error {
	if a == nil {
		glog.V(2).Infof("Creating PersistentVolume with Name:%q", *e.Name)

		request := &ec2.CreateVolumeInput{
			Size:             e.SizeGB,
			AvailabilityZone: e.AvailabilityZone,
			VolumeType:       e.VolumeType,
			KmsKeyId:         e.KmsKeyId,
			Encrypted:        e.Encrypted,
		}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateVolume(request)
		if err != nil {
			return fmt.Errorf("error creating PersistentVolume: %v", err)
		}

		e.ID = response.VolumeId
	}

	return t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.ID, e.Tags)
}
