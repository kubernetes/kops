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
	"os"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
)

// +kops:fitask
type EBSVolume struct {
	Name      *string
	Lifecycle fi.Lifecycle

	AvailabilityZone *string
	Encrypted        *bool
	ID               *string
	KmsKeyId         *string
	SizeGB           *int64
	Tags             map[string]string
	VolumeIops       *int64
	VolumeThroughput *int64
	VolumeType       *string
}

var _ fi.CompareWithID = &EBSVolume{}
var _ fi.CloudupTaskNormalize = &EBSVolume{}

func (e *EBSVolume) CompareWithID() *string {
	return e.ID
}

func (e *EBSVolume) Find(context *fi.CloudupContext) (*EBSVolume, error) {
	cloud := context.T.Cloud.(awsup.AWSCloud)

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
		VolumeThroughput: v.Throughput,
	}

	actual.Tags = mapEC2TagsToMap(v.Tags)

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	e.ID = actual.ID

	return actual, nil
}

func (e *EBSVolume) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(awsup.AWSCloud).AddTags(e.Name, e.Tags)
	return nil
}

func (e *EBSVolume) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
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
		if changes.AvailabilityZone != nil {
			return fi.CannotChangeField("AvailabilityZone")
		}
		if changes.Encrypted != nil {
			return fi.CannotChangeField("Encrypted")
		}
		if changes.KmsKeyId != nil {
			return fi.CannotChangeField("KmsKeyID")
		}
	}
	return nil
}

func (_ *EBSVolume) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *EBSVolume) error {
	if a == nil {
		klog.V(2).Infof("Creating PersistentVolume with Name:%q", *e.Name)

		request := &ec2.CreateVolumeInput{
			Size:              e.SizeGB,
			AvailabilityZone:  e.AvailabilityZone,
			VolumeType:        e.VolumeType,
			KmsKeyId:          e.KmsKeyId,
			Encrypted:         e.Encrypted,
			Iops:              e.VolumeIops,
			Throughput:        e.VolumeThroughput,
			TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeVolume, e.Tags),
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

	if a != nil {
		if len(changes.Tags) > 0 {
			tagsToDelete := e.getEBSVolumeTagsToDelete(a.Tags)
			if len(tagsToDelete) > 0 {
				return t.DeleteTags(*e.ID, tagsToDelete)
			}
		}

		if changes.VolumeType != nil ||
			changes.VolumeIops != nil ||
			changes.VolumeThroughput != nil ||
			changes.SizeGB != nil {

			request := &ec2.ModifyVolumeInput{
				VolumeId:   a.ID,
				VolumeType: e.VolumeType,
				Iops:       e.VolumeIops,
				Throughput: e.VolumeThroughput,
				Size:       e.SizeGB,
			}

			_, err := t.Cloud.EC2().ModifyVolume(request)
			if err != nil {
				return fmt.Errorf("error modifying volume: %v", err)
			}
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
	AvailabilityZone *string           `cty:"availability_zone"`
	Size             *int64            `cty:"size"`
	Type             *string           `cty:"type"`
	Iops             *int64            `cty:"iops"`
	Throughput       *int64            `cty:"throughput"`
	KmsKeyId         *string           `cty:"kms_key_id"`
	Encrypted        *bool             `cty:"encrypted"`
	Tags             map[string]string `cty:"tags"`
}

func (_ *EBSVolume) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *EBSVolume) error {
	tf := &terraformVolume{
		AvailabilityZone: e.AvailabilityZone,
		Size:             e.SizeGB,
		Type:             e.VolumeType,
		Iops:             e.VolumeIops,
		Throughput:       e.VolumeThroughput,
		KmsKeyId:         e.KmsKeyId,
		Encrypted:        e.Encrypted,
		Tags:             e.Tags,
	}

	tfName, _ := e.TerraformName()
	return t.RenderResource("aws_ebs_volume", tfName, tf)
}

func (e *EBSVolume) TerraformLink() *terraformWriter.Literal {
	tfName, _ := e.TerraformName()
	return terraformWriter.LiteralSelfLink("aws_ebs_volume", tfName)
}

// TerraformName returns the terraform-safe name, along with a boolean indicating of whether name-prefixing was needed.
func (e *EBSVolume) TerraformName() (string, bool) {
	usedPrefix := false
	name := fi.ValueOf(e.Name)
	if name[0] >= '0' && name[0] <= '9' {
		usedPrefix = true
		return fmt.Sprintf("ebs-%v", name), usedPrefix
	}
	return name, usedPrefix
}

// PreRun is run before general task execution, and checks for terraform breaking changes.
func (e *EBSVolume) PreRun(c *fi.CloudupContext) error {
	if _, ok := c.Target.(*terraform.TerraformTarget); ok {
		_, usedPrefix := e.TerraformName()
		if usedPrefix {
			if os.Getenv("KOPS_TERRAFORM_0_12_RENAMED") == "" {
				fmt.Fprintf(os.Stderr, "Terraform 0.12 broke compatibility and disallowed names that begin with a number.\n")
				fmt.Fprintf(os.Stderr, "  To move an existing cluster to the new syntax, you must first move existing volumes to the new names.\n")
				fmt.Fprintf(os.Stderr, "  To indicate that you have already performed the rename, pass KOPS_TERRAFORM_0_12_RENAMED=ebs environment variable.\n")
				fmt.Fprintf(os.Stderr, "  Not doing so will result in data loss.\n")
				fmt.Fprintf(os.Stderr, "For detailed instructions: https://github.com/kubernetes/kops/blob/master/permalinks/terraform_renamed.md\n")
				return fmt.Errorf("must update terraform state for 0.12, and then pass KOPS_TERRAFORM_0_12_RENAMED=ebs")
			}
		}
	}

	return nil
}
