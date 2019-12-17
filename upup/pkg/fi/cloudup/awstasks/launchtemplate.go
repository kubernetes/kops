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
	"sort"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/klog"
)

// LaunchTemplate defines the specificate for a template
type LaunchTemplate struct {
	// Name is the name of the configuration
	Name *string
	// Lifecycle is the resource lifecycle
	Lifecycle *fi.Lifecycle

	// AssociatePublicIP indicates if a public ip address is assigned to instabces
	AssociatePublicIP *bool
	// BlockDeviceMappings is a block device mappings
	BlockDeviceMappings []*BlockDeviceMapping
	// IAMInstanceProfile is the IAM profile to assign to the nodes
	IAMInstanceProfile *IAMInstanceProfile
	// ID is the launch configuration name
	ID *string
	// ImageID is the AMI to use for the instances
	ImageID *string
	// InstanceMonitoring indicates if monitoring is enabled
	InstanceMonitoring *bool
	// InstanceType is the type of instance we are using
	InstanceType *string
	// If volume type is io1, then we need to specify the number of Iops.
	RootVolumeIops *int64
	// RootVolumeOptimization enables EBS optimization for an instance
	RootVolumeOptimization *bool
	// RootVolumeSize is the size of the EBS root volume to use, in GB
	RootVolumeSize *int64
	// RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
	RootVolumeType *string
	// SSHKey is the ssh key for the instances
	SSHKey *SSHKey
	// SecurityGroups is a list of security group associated
	SecurityGroups []*SecurityGroup
	// SpotPrice is set to the spot-price bid if this is a spot pricing request
	SpotPrice string
	// Tenancy. Can be either default or dedicated.
	Tenancy *string
	// UserData is the user data configuration
	UserData *fi.ResourceHolder
}

var (
	_ fi.CompareWithID     = &LaunchTemplate{}
	_ fi.ProducesDeletions = &LaunchTemplate{}
	_ fi.Deletion          = &deleteLaunchTemplate{}
)

// CompareWithID implements the comparable interface
func (t *LaunchTemplate) CompareWithID() *string {
	return t.ID
}

// LaunchTemplateName returns the lanuch template name
func (t *LaunchTemplate) LaunchTemplateName() string {
	return fmt.Sprintf("%s-%s", fi.StringValue(t.Name), fi.BuildTimestampString())
}

// buildRootDevice is responsible for retrieving a boot device mapping from the image name
func (t *LaunchTemplate) buildRootDevice(cloud awsup.AWSCloud) (map[string]*BlockDeviceMapping, error) {
	image := fi.StringValue(t.ImageID)
	if image == "" {
		return map[string]*BlockDeviceMapping{}, nil
	}

	// @step: resolve the image ami
	img, err := cloud.ResolveImage(image)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve image: %q: %v", image, err)
	} else if img == nil {
		return nil, fmt.Errorf("unable to resolve image: %q: not found", image)
	}

	bm := make(map[string]*BlockDeviceMapping)
	bm[aws.StringValue(img.RootDeviceName)] = &BlockDeviceMapping{
		EbsDeleteOnTermination: aws.Bool(true),
		EbsVolumeSize:          t.RootVolumeSize,
		EbsVolumeType:          t.RootVolumeType,
		EbsVolumeIops:          t.RootVolumeIops,
	}

	return bm, nil
}

// Run is responsible for
func (t *LaunchTemplate) Run(c *fi.Context) error {
	t.Normalize()

	return fi.DefaultDeltaRunMethod(t, c)
}

// Normalize is responsible for normalizing any data within the resource
func (t *LaunchTemplate) Normalize() {
	sort.Stable(OrderSecurityGroupsById(t.SecurityGroups))
}

// CheckChanges is responsible for ensuring certains fields
func (t *LaunchTemplate) CheckChanges(a, e, changes *LaunchTemplate) error {
	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}
	if e.InstanceType == nil {
		return fi.RequiredField("InstanceType")
	}

	if a != nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

// FindDeletions is responsible for finding launch templates which can be deleted
func (t *LaunchTemplate) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	var removals []fi.Deletion

	configurations, err := t.findLaunchTemplates(c)
	if err != nil {
		return nil, err
	}

	if len(configurations) <= RetainLaunchConfigurationCount() {
		return nil, nil
	}

	configurations = configurations[:len(configurations)-RetainLaunchConfigurationCount()]

	for _, configuration := range configurations {
		removals = append(removals, &deleteLaunchTemplate{lc: configuration})
	}

	klog.V(2).Infof("will delete launch template: %v", removals)

	return removals, nil
}
