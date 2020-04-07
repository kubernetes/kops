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
	"encoding/base64"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"k8s.io/kops/pkg/apis/kops"

	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
)

// defaultRetainLaunchConfigurationCount is the number of launch configurations (matching the name prefix) that we should
// keep, we delete older ones
var defaultRetainLaunchConfigurationCount = 3

// RetainLaunchConfigurationCount returns the number of launch configurations to keep
func RetainLaunchConfigurationCount() int {
	if featureflag.KeepLaunchConfigurations.Enabled() {
		return math.MaxInt32
	}
	return defaultRetainLaunchConfigurationCount
}

// LaunchConfiguration is the specification for a launch configuration
type LaunchConfiguration struct {
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
	// InstanceType is the machine type to use
	InstanceType *string
	// RootVolumeDeleteOnTermination states if the root volume will be deleted after instance termination
	RootVolumeDeleteOnTermination *bool
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

var _ fi.CompareWithID = &LaunchConfiguration{}

var _ fi.ProducesDeletions = &LaunchConfiguration{}

func (e *LaunchConfiguration) CompareWithID() *string {
	return e.ID
}

// findLaunchConfigurations returns matching LaunchConfigurations, sorted by CreatedTime (ascending)
func (e *LaunchConfiguration) findLaunchConfigurations(c *fi.Context) ([]*autoscaling.LaunchConfiguration, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &autoscaling.DescribeLaunchConfigurationsInput{}

	prefix := *e.Name + "-"

	var configurations []*autoscaling.LaunchConfiguration
	err := cloud.Autoscaling().DescribeLaunchConfigurationsPages(request, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		for _, l := range page.LaunchConfigurations {
			name := aws.StringValue(l.LaunchConfigurationName)
			if strings.HasPrefix(name, prefix) {
				configurations = append(configurations, l)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing AutoscalingLaunchConfigurations: %v", err)
	}

	sort.Slice(configurations, func(i, j int) bool {
		ti := configurations[i].CreatedTime
		tj := configurations[j].CreatedTime
		if tj == nil {
			return true
		}
		if ti == nil {
			return false
		}
		return ti.UnixNano() < tj.UnixNano()
	})

	return configurations, nil
}

// Find is responsible for finding the launch configuration
func (e *LaunchConfiguration) Find(c *fi.Context) (*LaunchConfiguration, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	configurations, err := e.findLaunchConfigurations(c)
	if err != nil {
		return nil, err
	}

	if len(configurations) == 0 {
		return nil, nil
	}

	// We pick up the latest launch configuration
	// (TODO: this might not actually be attached to the AutoScalingGroup, if something went wrong previously)
	lc := configurations[len(configurations)-1]

	klog.V(2).Infof("found existing AutoscalingLaunchConfiguration: %q", *lc.LaunchConfigurationName)

	actual := &LaunchConfiguration{
		Name:                   e.Name,
		AssociatePublicIP:      lc.AssociatePublicIpAddress,
		ID:                     lc.LaunchConfigurationName,
		ImageID:                lc.ImageId,
		InstanceMonitoring:     lc.InstanceMonitoring.Enabled,
		InstanceType:           lc.InstanceType,
		RootVolumeOptimization: lc.EbsOptimized,
		SpotPrice:              aws.StringValue(lc.SpotPrice),
		Tenancy:                lc.PlacementTenancy,
	}

	// Only assign keyName if the existing launch config has one
	// lc.KeyName comes back as an empty string when there is no key assigned
	if lc.KeyName != nil && *lc.KeyName != "" {
		actual.SSHKey = &SSHKey{Name: lc.KeyName}
	}

	if lc.IamInstanceProfile != nil {
		actual.IAMInstanceProfile = &IAMInstanceProfile{Name: lc.IamInstanceProfile}
	}

	securityGroups := []*SecurityGroup{}
	for _, sgID := range lc.SecurityGroups {
		securityGroups = append(securityGroups, &SecurityGroup{ID: sgID})
	}
	sort.Sort(OrderSecurityGroupsById(securityGroups))

	actual.SecurityGroups = securityGroups

	// @step: get the image is order to find out the root device name as using the index
	// is not variable, under conditions they move
	image, err := cloud.ResolveImage(fi.StringValue(e.ImageID))
	if err != nil {
		return nil, err
	}

	// Find the root volume
	for _, b := range lc.BlockDeviceMappings {
		if b.Ebs == nil {
			continue
		}
		if b.DeviceName != nil && fi.StringValue(b.DeviceName) == fi.StringValue(image.RootDeviceName) {
			actual.RootVolumeSize = b.Ebs.VolumeSize
			actual.RootVolumeType = b.Ebs.VolumeType
			actual.RootVolumeIops = b.Ebs.Iops
			actual.RootVolumeDeleteOnTermination = b.Ebs.DeleteOnTermination
		} else {
			_, d := BlockDeviceMappingFromAutoscaling(b)
			actual.BlockDeviceMappings = append(actual.BlockDeviceMappings, d)
		}
	}

	if lc.UserData != nil {
		userData, err := base64.StdEncoding.DecodeString(aws.StringValue(lc.UserData))
		if err != nil {
			return nil, fmt.Errorf("error decoding UserData: %v", err)
		}
		actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))
	}

	// Avoid spurious changes on ImageId
	if e.ImageID != nil && actual.ImageID != nil && *actual.ImageID != *e.ImageID {
		image, err := cloud.ResolveImage(*e.ImageID)
		if err != nil {
			klog.Warningf("unable to resolve image: %q: %v", *e.ImageID, err)
		} else if image == nil {
			klog.Warningf("unable to resolve image: %q: not found", *e.ImageID)
		} else if aws.StringValue(image.ImageId) == *actual.ImageID {
			klog.V(4).Infof("Returning matching ImageId as expected name: %q -> %q", *actual.ImageID, *e.ImageID)
			actual.ImageID = e.ImageID
		}
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (e *LaunchConfiguration) Run(c *fi.Context) error {
	// TODO: Make Normalize a standard method
	e.Normalize()

	if e.SSHKey == nil && !useSSHKey(c.Cluster) {
		e.SSHKey = &SSHKey{}
	}

	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *LaunchConfiguration) Normalize() {
	// We need to sort our arrays consistently, so we don't get spurious changes
	sort.Stable(OrderSecurityGroupsById(e.SecurityGroups))
}

func (s *LaunchConfiguration) CheckChanges(a, e, changes *LaunchConfiguration) error {
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

// RenderAWS is responsible for creating the launchconfiguration via api
func (_ *LaunchConfiguration) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LaunchConfiguration) error {
	launchConfigurationName := *e.Name + "-" + fi.BuildTimestampString()

	klog.V(2).Infof("Creating AutoscalingLaunchConfiguration with Name:%q", launchConfigurationName)

	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}

	image, err := t.Cloud.ResolveImage(*e.ImageID)
	if err != nil {
		return err
	}

	request := &autoscaling.CreateLaunchConfigurationInput{
		AssociatePublicIpAddress: e.AssociatePublicIP,
		EbsOptimized:             e.RootVolumeOptimization,
		ImageId:                  image.ImageId,
		InstanceType:             e.InstanceType,
		LaunchConfigurationName:  &launchConfigurationName,
	}

	if e.SSHKey != nil {
		request.KeyName = e.SSHKey.Name
	}

	if e.Tenancy != nil {
		request.PlacementTenancy = e.Tenancy
	}

	securityGroupIDs := []*string{}
	for _, sg := range e.SecurityGroups {
		securityGroupIDs = append(securityGroupIDs, sg.ID)
	}

	request.SecurityGroups = securityGroupIDs
	request.AssociatePublicIpAddress = e.AssociatePublicIP
	if e.SpotPrice != "" {
		request.SpotPrice = aws.String(e.SpotPrice)
	}

	// Build up the actual block device mappings
	{
		rootDevices, err := e.buildRootDevice(t.Cloud)
		if err != nil {
			return err
		}
		ephemeralDevices, err := buildEphemeralDevices(t.Cloud, fi.StringValue(e.InstanceType))
		if err != nil {
			return err
		}
		additionalDevices, err := buildAdditionalDevices(e.BlockDeviceMappings)
		if err != nil {
			return err
		}

		// @step: add all the devices to the block device mappings
		for _, x := range []map[string]*BlockDeviceMapping{rootDevices, ephemeralDevices, additionalDevices} {
			for name, device := range x {
				request.BlockDeviceMappings = append(request.BlockDeviceMappings, device.ToAutoscaling(name))
			}
		}
	}

	if e.UserData != nil {
		d, err := e.UserData.AsBytes()
		if err != nil {
			return fmt.Errorf("error rendering AutoScalingLaunchConfiguration UserData: %v", err)
		}
		request.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
	}
	if e.IAMInstanceProfile != nil {
		request.IamInstanceProfile = e.IAMInstanceProfile.Name
	}
	if e.InstanceMonitoring != nil {
		request.InstanceMonitoring = &autoscaling.InstanceMonitoring{Enabled: e.InstanceMonitoring}
	} else {
		request.InstanceMonitoring = &autoscaling.InstanceMonitoring{Enabled: fi.Bool(false)}
	}

	attempt := 0
	maxAttempts := 10
	for {
		attempt++

		klog.V(8).Infof("AWS CreateLaunchConfiguration %s", aws.StringValue(request.LaunchConfigurationName))
		_, err = t.Cloud.Autoscaling().CreateLaunchConfiguration(request)
		if err == nil {
			break
		}

		if awsup.AWSErrorCode(err) == "ValidationError" {
			message := awsup.AWSErrorMessage(err)
			if strings.Contains(message, "not authorized") || strings.Contains(message, "Invalid IamInstance") {
				if attempt > maxAttempts {
					return fmt.Errorf("IAM instance profile not yet created/propagated (original error: %v)", message)
				}
				klog.V(4).Infof("got an error indicating that the IAM instance profile %q is not ready: %q", fi.StringValue(e.IAMInstanceProfile.Name), message)
				klog.Infof("waiting for IAM instance profile %q to be ready", fi.StringValue(e.IAMInstanceProfile.Name))
				time.Sleep(10 * time.Second)
				continue
			}
			klog.V(4).Infof("ErrorCode=%q, Message=%q", awsup.AWSErrorCode(err), awsup.AWSErrorMessage(err))
		}

		return fmt.Errorf("error creating AutoscalingLaunchConfiguration: %v", err)
	}

	e.ID = fi.String(launchConfigurationName)

	return nil // No tags on a launch configuration
}

// buildRootDevice is responsible for retrieving a boot device mapping from the image name
func (t *LaunchConfiguration) buildRootDevice(cloud awsup.AWSCloud) (map[string]*BlockDeviceMapping, error) {
	image := fi.StringValue(t.ImageID)

	// @step: resolve the image ami
	img, err := cloud.ResolveImage(image)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve image: %q: %v", image, err)
	} else if img == nil {
		return nil, fmt.Errorf("unable to resolve image: %q: not found", image)
	}

	bm := make(map[string]*BlockDeviceMapping)

	bm[aws.StringValue(img.RootDeviceName)] = &BlockDeviceMapping{
		EbsDeleteOnTermination: t.RootVolumeDeleteOnTermination,
		EbsVolumeSize:          t.RootVolumeSize,
		EbsVolumeType:          t.RootVolumeType,
		EbsVolumeIops:          t.RootVolumeIops,
	}

	return bm, nil
}

type terraformLaunchConfiguration struct {
	NamePrefix               *string                 `json:"name_prefix,omitempty" cty:"name_prefix"`
	ImageID                  *string                 `json:"image_id,omitempty" cty:"image_id"`
	InstanceType             *string                 `json:"instance_type,omitempty" cty:"instance_type"`
	KeyName                  *terraform.Literal      `json:"key_name,omitempty" cty:"key_name"`
	IAMInstanceProfile       *terraform.Literal      `json:"iam_instance_profile,omitempty" cty:"iam_instance_profile"`
	SecurityGroups           []*terraform.Literal    `json:"security_groups,omitempty" cty:"security_groups"`
	AssociatePublicIpAddress *bool                   `json:"associate_public_ip_address,omitempty" cty:"associate_public_ip_address"`
	UserData                 *terraform.Literal      `json:"user_data,omitempty" cty:"user_data"`
	RootBlockDevice          *terraformBlockDevice   `json:"root_block_device,omitempty" cty:"root_block_device"`
	EBSOptimized             *bool                   `json:"ebs_optimized,omitempty" cty:"ebs_optimized"`
	EBSBlockDevice           []*terraformBlockDevice `json:"ebs_block_device,omitempty" cty:"ebs_block_device"`
	EphemeralBlockDevice     []*terraformBlockDevice `json:"ephemeral_block_device,omitempty" cty:"ephemeral_block_device"`
	Lifecycle                *terraform.Lifecycle    `json:"lifecycle,omitempty" cty:"lifecycle"`
	SpotPrice                *string                 `json:"spot_price,omitempty" cty:"spot_price"`
	PlacementTenancy         *string                 `json:"placement_tenancy,omitempty" cty:"placement_tenancy"`
	InstanceMonitoring       *bool                   `json:"enable_monitoring,omitempty" cty:"enable_monitoring"`
}

type terraformBlockDevice struct {
	// For ephemeral devices
	DeviceName  *string `json:"device_name,omitempty" cty:"device_name"`
	VirtualName *string `json:"virtual_name,omitempty" cty:"virtual_name"`

	// For root
	VolumeType *string `json:"volume_type,omitempty" cty:"volume_type"`
	VolumeSize *int64  `json:"volume_size,omitempty" cty:"volume_size"`
	Iops       *int64  `json:"iops,omitempty" cty:"iops"`
	// Encryption
	Encrypted *bool `json:"encrypted,omitempty" cty:"encrypted"`
	// Termination
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty" cty:"delete_on_termination"`
}

// RenderTerraform is responsible for rendering the terraform json
func (_ *LaunchConfiguration) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LaunchConfiguration) error {
	cloud := t.Cloud.(awsup.AWSCloud)

	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}
	image, err := cloud.ResolveImage(*e.ImageID)
	if err != nil {
		return err
	}

	tf := &terraformLaunchConfiguration{
		NamePrefix:   fi.String(*e.Name + "-"),
		ImageID:      image.ImageId,
		InstanceType: e.InstanceType,
	}

	if e.SpotPrice != "" {
		tf.SpotPrice = aws.String(e.SpotPrice)
	}

	if e.SSHKey != nil {
		tf.KeyName = e.SSHKey.TerraformLink()
	}

	if e.Tenancy != nil {
		tf.PlacementTenancy = e.Tenancy
	}

	for _, sg := range e.SecurityGroups {
		tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
	}

	tf.AssociatePublicIpAddress = e.AssociatePublicIP
	tf.EBSOptimized = e.RootVolumeOptimization

	{
		rootDevices, err := e.buildRootDevice(cloud)
		if err != nil {
			return err
		}
		ephemeralDevices, err := buildEphemeralDevices(cloud, fi.StringValue(e.InstanceType))
		if err != nil {
			return err
		}
		additionalDevices, err := buildAdditionalDevices(e.BlockDeviceMappings)
		if err != nil {
			return err
		}

		if len(rootDevices) != 0 {
			if len(rootDevices) != 1 {
				return fmt.Errorf("unexpectedly found multiple root devices")
			}

			for _, bdm := range rootDevices {
				tf.RootBlockDevice = &terraformBlockDevice{
					VolumeType:          bdm.EbsVolumeType,
					VolumeSize:          bdm.EbsVolumeSize,
					Iops:                bdm.EbsVolumeIops,
					DeleteOnTermination: bdm.EbsDeleteOnTermination,
				}
			}
		}

		if len(ephemeralDevices) != 0 {
			tf.EphemeralBlockDevice = []*terraformBlockDevice{}
			for _, deviceName := range sets.StringKeySet(ephemeralDevices).List() {
				bdm := ephemeralDevices[deviceName]
				tf.EphemeralBlockDevice = append(tf.EphemeralBlockDevice, &terraformBlockDevice{
					VirtualName: bdm.VirtualName,
					DeviceName:  fi.String(deviceName),
				})
			}
		}

		if len(additionalDevices) != 0 {
			tf.EBSBlockDevice = []*terraformBlockDevice{}
			for _, deviceName := range sets.StringKeySet(additionalDevices).List() {
				bdm := additionalDevices[deviceName]
				tf.EBSBlockDevice = append(tf.EBSBlockDevice, &terraformBlockDevice{
					DeleteOnTermination: bdm.EbsDeleteOnTermination,
					DeviceName:          fi.String(deviceName),
					Encrypted:           bdm.EbsEncrypted,
					VolumeSize:          bdm.EbsVolumeSize,
					VolumeType:          bdm.EbsVolumeType,
				})
			}
		}
	}

	if e.UserData != nil {
		userData, err := fi.ResourceAsString(e.UserData)
		if err != nil {
			return err
		}
		if userData != "" {
			tf.UserData, err = t.AddFile("aws_launch_configuration", *e.Name, "user_data", e.UserData)
			if err != nil {
				return err
			}
		}
	}
	if e.IAMInstanceProfile != nil {
		tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
	}
	if e.InstanceMonitoring != nil {
		tf.InstanceMonitoring = e.InstanceMonitoring
	} else {
		tf.InstanceMonitoring = fi.Bool(false)
	}
	// So that we can update configurations
	tf.Lifecycle = &terraform.Lifecycle{CreateBeforeDestroy: fi.Bool(true)}

	return t.RenderResource("aws_launch_configuration", fi.StringValue(e.Name), tf)
}

// TerraformLink returns the terraform reference
func (e *LaunchConfiguration) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_launch_configuration", fi.StringValue(e.Name), "id")
}

type cloudformationLaunchConfiguration struct {
	AssociatePublicIpAddress *bool                        `json:"AssociatePublicIpAddress,omitempty"`
	BlockDeviceMappings      []*cloudformationBlockDevice `json:"BlockDeviceMappings,omitempty"`
	EBSOptimized             *bool                        `json:"EbsOptimized,omitempty"`
	IAMInstanceProfile       *cloudformation.Literal      `json:"IamInstanceProfile,omitempty"`
	ImageID                  *string                      `json:"ImageId,omitempty"`
	InstanceType             *string                      `json:"InstanceType,omitempty"`
	KeyName                  *string                      `json:"KeyName,omitempty"`
	SecurityGroups           []*cloudformation.Literal    `json:"SecurityGroups,omitempty"`
	SpotPrice                *string                      `json:"SpotPrice,omitempty"`
	UserData                 *string                      `json:"UserData,omitempty"`
	PlacementTenancy         *string                      `json:"PlacementTenancy,omitempty"`
	InstanceMonitoring       *bool                        `json:"InstanceMonitoring,omitempty"`
}

type cloudformationBlockDevice struct {
	// For ephemeral devices
	DeviceName  *string `json:"DeviceName,omitempty"`
	VirtualName *string `json:"VirtualName,omitempty"`

	// For root
	Ebs *cloudformationBlockDeviceEBS `json:"Ebs,omitempty"`
}

type cloudformationBlockDeviceEBS struct {
	VolumeType          *string `json:"VolumeType,omitempty"`
	VolumeSize          *int64  `json:"VolumeSize,omitempty"`
	Iops                *int64  `json:"Iops,omitempty"`
	DeleteOnTermination *bool   `json:"DeleteOnTermination,omitempty"`
	Encrypted           *bool   `json:"Encrypted,omitempty"`
}

// RenderCloudformation is responsible for rendering the cloudformation template
func (_ *LaunchConfiguration) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *LaunchConfiguration) error {
	cloud := t.Cloud.(awsup.AWSCloud)

	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}
	image, err := cloud.ResolveImage(*e.ImageID)
	if err != nil {
		return err
	}

	cf := &cloudformationLaunchConfiguration{
		//NamePrefix:   fi.String(*e.Name + "-"),
		ImageID:      image.ImageId,
		InstanceType: e.InstanceType,
	}

	if e.SpotPrice != "" {
		cf.SpotPrice = aws.String(e.SpotPrice)
	}

	if e.SSHKey != nil && !e.SSHKey.NoSSHKey() {
		if e.SSHKey.Name == nil {
			return fmt.Errorf("SSHKey Name not set")
		}
		cf.KeyName = e.SSHKey.Name
	}

	if e.Tenancy != nil {
		cf.PlacementTenancy = e.Tenancy
	}

	for _, sg := range e.SecurityGroups {
		cf.SecurityGroups = append(cf.SecurityGroups, sg.CloudformationLink())
	}
	cf.AssociatePublicIpAddress = e.AssociatePublicIP

	cf.EBSOptimized = e.RootVolumeOptimization

	{
		rootDevices, err := e.buildRootDevice(cloud)
		if err != nil {
			return err
		}
		ephemeralDevices, err := buildEphemeralDevices(cloud, fi.StringValue(e.InstanceType))
		if err != nil {
			return err
		}
		additionalDevices, err := buildAdditionalDevices(e.BlockDeviceMappings)
		if err != nil {
			return err
		}

		if len(rootDevices) != 0 {
			if len(rootDevices) != 1 {
				return fmt.Errorf("unexpectedly found multiple root devices")
			}

			for deviceName, bdm := range rootDevices {
				d := &cloudformationBlockDevice{
					DeviceName: fi.String(deviceName),
					Ebs: &cloudformationBlockDeviceEBS{
						VolumeType:          bdm.EbsVolumeType,
						VolumeSize:          bdm.EbsVolumeSize,
						Iops:                bdm.EbsVolumeIops,
						DeleteOnTermination: bdm.EbsDeleteOnTermination,
					},
				}
				cf.BlockDeviceMappings = append(cf.BlockDeviceMappings, d)
			}
		}

		if len(ephemeralDevices) != 0 {
			for deviceName, bdm := range ephemeralDevices {
				cf.BlockDeviceMappings = append(cf.BlockDeviceMappings, &cloudformationBlockDevice{
					VirtualName: bdm.VirtualName,
					DeviceName:  fi.String(deviceName),
				})
			}
		}

		if len(additionalDevices) != 0 {
			for deviceName, bdm := range additionalDevices {
				d := &cloudformationBlockDevice{
					DeviceName: fi.String(deviceName),
					Ebs: &cloudformationBlockDeviceEBS{
						VolumeType:          bdm.EbsVolumeType,
						VolumeSize:          bdm.EbsVolumeSize,
						DeleteOnTermination: bdm.EbsDeleteOnTermination,
						Encrypted:           bdm.EbsEncrypted,
					},
				}
				cf.BlockDeviceMappings = append(cf.BlockDeviceMappings, d)
			}
		}
	}

	if e.UserData != nil {
		d, err := e.UserData.AsBytes()
		if err != nil {
			return fmt.Errorf("error rendering AutoScalingLaunchConfiguration UserData: %v", err)
		}
		cf.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
	}

	if e.IAMInstanceProfile != nil {
		cf.IAMInstanceProfile = e.IAMInstanceProfile.CloudformationLink()
	}

	if e.InstanceMonitoring != nil {
		cf.InstanceMonitoring = e.InstanceMonitoring
	} else {
		cf.InstanceMonitoring = fi.Bool(false)
	}
	// So that we can update configurations
	//tf.Lifecycle = &cloudformation.Lifecycle{CreateBeforeDestroy: fi.Bool(true)}

	return t.RenderResource("AWS::AutoScaling::LaunchConfiguration", *e.Name, cf)
}

func (e *LaunchConfiguration) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::AutoScaling::LaunchConfiguration", *e.Name)
}

// deleteLaunchConfiguration tracks a LaunchConfiguration that we're going to delete
// It implements fi.Deletion
type deleteLaunchConfiguration struct {
	lc *autoscaling.LaunchConfiguration
}

var _ fi.Deletion = &deleteLaunchConfiguration{}

func (d *deleteLaunchConfiguration) TaskName() string {
	return "LaunchConfiguration"
}

func (d *deleteLaunchConfiguration) Item() string {
	return aws.StringValue(d.lc.LaunchConfigurationName)
}

func (d *deleteLaunchConfiguration) Delete(t fi.Target) error {
	klog.V(2).Infof("deleting launch configuration %v", d)

	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	request := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: d.lc.LaunchConfigurationName,
	}

	name := aws.StringValue(request.LaunchConfigurationName)
	klog.V(2).Infof("Calling autoscaling DeleteLaunchConfiguration for %s", name)
	_, err := awsTarget.Cloud.Autoscaling().DeleteLaunchConfiguration(request)
	if err != nil {
		return fmt.Errorf("error deleting autoscaling LaunchConfiguration %s: %v", name, err)
	}

	return nil
}

func (d *deleteLaunchConfiguration) String() string {
	return d.TaskName() + "-" + d.Item()
}

func (e *LaunchConfiguration) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	var removals []fi.Deletion

	configurations, err := e.findLaunchConfigurations(c)
	if err != nil {
		return nil, err
	}

	if len(configurations) <= RetainLaunchConfigurationCount() {
		return nil, nil
	}

	configurations = configurations[:len(configurations)-RetainLaunchConfigurationCount()]

	for _, configuration := range configurations {
		removals = append(removals, &deleteLaunchConfiguration{lc: configuration})
	}

	klog.V(2).Infof("will delete launch configurations: %v", removals)

	return removals, nil
}

func useSSHKey(c *kops.Cluster) bool {
	if c != nil {
		sshKeyName := c.Spec.SSHKeyName
		return sshKeyName != nil && *sshKeyName != ""
	}
	return true
}
