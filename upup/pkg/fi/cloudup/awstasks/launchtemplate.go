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

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// LaunchTemplate defines the specification for a launch template.
// +kops:fitask
type LaunchTemplate struct {
	// ID is the launch configuration name
	ID *string
	// Name is the name of the configuration
	Name *string
	// Lifecycle is the resource lifecycle
	Lifecycle fi.Lifecycle

	// AssociatePublicIP indicates if a public ip address is assigned to instances
	AssociatePublicIP *bool
	// BlockDeviceMappings is a block device mappings
	BlockDeviceMappings []*BlockDeviceMapping
	// CPUCredits is the credit option for CPU Usage on some instance types
	CPUCredits *string
	// HTTPPutResponseHopLimit is the desired HTTP PUT response hop limit for instance metadata requests.
	HTTPPutResponseHopLimit *int64
	// HTTPTokens is the state of token usage for your instance metadata requests.
	HTTPTokens *string
	// HTTPProtocolIPv6 enables the IPv6 instance metadata endpoint
	HTTPProtocolIPv6 *string
	// IAMInstanceProfile is the IAM profile to assign to the nodes
	IAMInstanceProfile *IAMInstanceProfile
	// ImageID is the AMI to use for the instances
	ImageID *string
	// InstanceInterruptionBehavior defines if a spot instance should be terminated, hibernated,
	// or stopped after interruption
	InstanceInterruptionBehavior *string
	// InstanceMonitoring indicates if monitoring is enabled
	InstanceMonitoring *bool
	// InstanceType is the type of instance we are using
	InstanceType *string
	// Ipv6AddressCount is the number of IPv6 addresses to assign with the primary network interface.
	IPv6AddressCount *int64
	// RootVolumeIops is the provisioned IOPS when the volume type is io1, io2 or gp3
	RootVolumeIops *int64
	// RootVolumeOptimization enables EBS optimization for an instance
	RootVolumeOptimization *bool
	// RootVolumeSize is the size of the EBS root volume to use, in GB
	RootVolumeSize *int64
	// RootVolumeThroughput is the volume throughput in MBps when the volume type is gp3
	RootVolumeThroughput *int64
	// RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
	RootVolumeType *string
	// RootVolumeEncryption enables EBS root volume encryption for an instance
	RootVolumeEncryption *bool
	// RootVolumeKmsKey is the encryption key identifier for EBS root volume encryption
	RootVolumeKmsKey *string
	// SSHKey is the ssh key for the instances
	SSHKey *SSHKey
	// SecurityGroups is a list of security group associated
	SecurityGroups []*SecurityGroup
	// SpotPrice is set to the spot-price bid if this is a spot pricing request
	SpotPrice *string
	// SpotDurationInMinutes is set for requesting spot blocks
	SpotDurationInMinutes *int64
	// Tags are the keypairs to apply to the instance and volume on launch as well as the launch template itself.
	Tags map[string]string
	// Tenancy. Can be either default or dedicated.
	Tenancy *string
	// UserData is the user data configuration
	UserData fi.Resource
}

var (
	_ fi.CompareWithID     = &LaunchTemplate{}
	_ fi.ProducesDeletions = &LaunchTemplate{}
	_ fi.TaskNormalize     = &LaunchTemplate{}
	_ fi.Deletion          = &deleteLaunchTemplate{}
)

// CompareWithID implements the comparable interface
func (t *LaunchTemplate) CompareWithID() *string {
	return t.ID
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

	b := &BlockDeviceMapping{
		EbsDeleteOnTermination: aws.Bool(true),
		EbsVolumeSize:          t.RootVolumeSize,
		EbsVolumeType:          t.RootVolumeType,
		EbsVolumeIops:          t.RootVolumeIops,
		EbsVolumeThroughput:    t.RootVolumeThroughput,
		EbsEncrypted:           t.RootVolumeEncryption,
	}
	if aws.BoolValue(t.RootVolumeEncryption) && aws.StringValue(t.RootVolumeKmsKey) != "" {
		b.EbsKmsKey = t.RootVolumeKmsKey
	}

	bm := map[string]*BlockDeviceMapping{
		aws.StringValue(img.RootDeviceName): b,
	}

	return bm, nil
}

func (t *LaunchTemplate) Normalize(c *fi.Context) error {
	sort.Stable(OrderSecurityGroupsById(t.SecurityGroups))
	return nil
}

// Run is responsible for
func (t *LaunchTemplate) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(t, c)
}

// CheckChanges is responsible for ensuring certains fields
func (t *LaunchTemplate) CheckChanges(a, e, changes *LaunchTemplate) error {
	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
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

	list, err := t.findAllLaunchTemplates(c)
	if err != nil {
		return nil, err
	}

	for _, lt := range list {
		if aws.StringValue(lt.LaunchTemplateName) != aws.StringValue(t.Name) {
			removals = append(removals, &deleteLaunchTemplate{lc: lt})
		}
	}

	return removals, nil
}
