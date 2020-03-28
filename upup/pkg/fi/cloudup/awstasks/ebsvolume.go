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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

//go:generate fitask -type=EBSVolume
type EBSVolume struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	AvailabilityZone *string
	Encrypted        *bool
	ID               *string
	KmsKeyId         *string
	SizeGB           *int64
	Tags             map[string]string
	VolumeIops       *int64
	VolumeType       *string
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
	actual, err := e.find(c.(awsup.AWSCloud))
	if err != nil {
		return nil, fmt.Errorf("error querying for EBSVolume: %v", err)
	}
	if actual == nil {
		return nil, nil
	}

	return actual.ID, nil
}

func (e *EBSVolume) Find(context *fi.Context) (*EBSVolume, error) {
	actual, err := e.find(context.Cloud.(awsup.AWSCloud))
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
	klog.V(2).Info("found existing volume")
	v := response.Volumes[0]
	actual := &EBSVolume{
		ID:               v.VolumeId,
		AvailabilityZone: v.AvailabilityZone,
		VolumeType:       v.VolumeType,
		SizeGB:           v.Size,
		KmsKeyId:         v.KmsKeyId,
		Encrypted:        v.Encrypted,
		Name:             e.Name,
		VolumeIops:       v.Iops,
	}

	actual.Tags = mapEC2TagsToMap(v.Tags)

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *EBSVolume) Run(c *fi.Context) error {
	c.Cloud.(awsup.AWSCloud).AddTags(e.Name, e.Tags)
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

func (_ *EBSVolume) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *EBSVolume) error {
	if a == nil {
		klog.V(2).Infof("Creating PersistentVolume with Name:%q", *e.Name)

		request := &ec2.CreateVolumeInput{
			Size:             e.SizeGB,
			AvailabilityZone: e.AvailabilityZone,
			VolumeType:       e.VolumeType,
			KmsKeyId:         e.KmsKeyId,
			Encrypted:        e.Encrypted,
			Iops:             e.VolumeIops,
		}

		if len(e.Tags) != 0 {
			request.TagSpecifications = []*ec2.TagSpecification{
				{ResourceType: aws.String(ec2.ResourceTypeVolume)},
			}

			for k, v := range e.Tags {
				request.TagSpecifications[0].Tags = append(request.TagSpecifications[0].Tags, &ec2.Tag{
					Key:   aws.String(k),
					Value: aws.String(v),
				})
			}
		}

		response, err := t.Cloud.EC2().CreateVolume(request)
		if err != nil {
			return fmt.Errorf("error creating PersistentVolume: %v", err)
		}

		e.ID = response.VolumeId
	}

	if err := t.AddAWSTags(*e.ID, e.Tags); err != nil {
		return fmt.Errorf("error adding AWS Tags to EBS Volume: %v", err)
	}

	if a != nil && len(a.Tags) > 0 {
		tagsToDelete := e.getEBSVolumeTagsToDelete(a.Tags)
		if len(tagsToDelete) > 0 {
			return t.DeleteTags(*e.ID, tagsToDelete)
		}
	}

	return nil
}

// getEBSVolumeTagsToDelete loops through the currently set tags and builds
// a list of tags to be deleted from the EBS Volume
func (e *EBSVolume) getEBSVolumeTagsToDelete(currentTags map[string]string) map[string]string {
	tagsToDelete := map[string]string{}
	for k, v := range currentTags {
		if _, ok := e.Tags[k]; !ok {
			tagsToDelete[k] = v
		}
	}

	return tagsToDelete
}

type terraformVolume struct {
	AvailabilityZone *string           `json:"availability_zone,omitempty" cty:"availability_zone"`
	Size             *int64            `json:"size,omitempty" cty:"size"`
	Type             *string           `json:"type,omitempty" cty:"type"`
	Iops             *int64            `json:"iops,omitempty" cty:"iops"`
	KmsKeyId         *string           `json:"kms_key_id,omitempty" cty:"kms_key_id"`
	Encrypted        *bool             `json:"encrypted,omitempty" cty:"encrypted"`
	Tags             map[string]string `json:"tags,omitempty" cty:"tags"`
}

func (_ *EBSVolume) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *EBSVolume) error {
	tf := &terraformVolume{
		AvailabilityZone: e.AvailabilityZone,
		Size:             e.SizeGB,
		Type:             e.VolumeType,
		Iops:             e.VolumeIops,
		KmsKeyId:         e.KmsKeyId,
		Encrypted:        e.Encrypted,
		Tags:             e.Tags,
	}

	return t.RenderResource("aws_ebs_volume", *e.Name, tf)
}

func (e *EBSVolume) TerraformLink() *terraform.Literal {
	return terraform.LiteralSelfLink("aws_ebs_volume", *e.Name)
}

type cloudformationVolume struct {
	AvailabilityZone *string             `json:"AvailabilityZone,omitempty"`
	Size             *int64              `json:"Size,omitempty"`
	Type             *string             `json:"VolumeType,omitempty"`
	Iops             *int64              `json:"Iops,omitempty"`
	KmsKeyId         *string             `json:"KmsKeyId,omitempty"`
	Encrypted        *bool               `json:"Encrypted,omitempty"`
	Tags             []cloudformationTag `json:"Tags,omitempty"`
}

func (_ *EBSVolume) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *EBSVolume) error {
	cf := &cloudformationVolume{
		AvailabilityZone: e.AvailabilityZone,
		Size:             e.SizeGB,
		Type:             e.VolumeType,
		Iops:             e.VolumeIops,
		KmsKeyId:         e.KmsKeyId,
		Encrypted:        e.Encrypted,
		Tags:             buildCloudformationTags(e.Tags),
	}

	return t.RenderResource("AWS::EC2::Volume", *e.Name, cf)
}

func (e *EBSVolume) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::EC2::Volume", *e.Name)
}
