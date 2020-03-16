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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

type terraformLaunchTemplateNetworkInterface struct {
	// AssociatePublicIPAddress associates a public ip address with the network interface. Boolean value.
	AssociatePublicIPAddress *bool `json:"associate_public_ip_address,omitempty"`
	// DeleteOnTermination indicates whether the network interface should be destroyed on instance termination.
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// SecurityGroups is a list of security group ids.
	SecurityGroups []*terraform.Literal `json:"security_groups,omitempty"`
}

type terraformLaunchTemplateMonitoring struct {
	// Enabled indicates that monitoring is enabled
	Enabled *bool `json:"enabled,omitempty"`
}

type terraformLaunchTemplatePlacement struct {
	// Affinity is he affinity setting for an instance on a Dedicated Host.
	Affinity *string `json:"affinity,omitempty"`
	// AvailabilityZone is the Availability Zone for the instance.
	AvailabilityZone *string `json:"availability_zone,omitempty"`
	// GroupName is the name of the placement group for the instance.
	GroupName *string `json:"group_name,omitempty"`
	// HostID is the ID of the Dedicated Host for the instance.
	HostID *string `json:"host_id,omitempty"`
	// SpreadDomain are reserved for future use.
	SpreadDomain *string `json:"spread_domain,omitempty"`
	// Tenancy ist he tenancy of the instance. Can be default, dedicated, or host.
	Tenancy *string `json:"tenancy,omitempty"`
}

type terraformLaunchTemplateIAMProfile struct {
	// Name is the name of the profile
	Name *terraform.Literal `json:"name,omitempty"`
}

type terraformLaunchTemplateMarketOptionsSpotOptions struct {
	// BlockDurationMinutes is required duration in minutes. This value must be a multiple of 60.
	BlockDurationMinutes *int64 `json:"block_duration_minutes,omitempty"`
	// InstancesInterruptionBehavior is the behavior when a Spot Instance is interrupted. Can be hibernate, stop, or terminate
	InstancesInterruptionBehavior *string `json:"instances_interruption_behavior,omitempty"`
	// MaxPrice is the maximum hourly price you're willing to pay for the Spot Instances
	MaxPrice *string `json:"max_price,omitempty"`
	// SpotInstanceType is the Spot Instance request type. Can be one-time, or persistent
	SpotInstanceType *string `json:"spot_instance_type,omitempty"`
	// ValidUntil is the end date of the request
	ValidUntil *string `json:"valid_until,omitempty"`
}

type terraformLaunchTemplateMarketOptions struct {
	// MarketType is the option type
	MarketType *string `json:"market_type,omitempty"`
	// SpotOptions are the set of options
	SpotOptions []*terraformLaunchTemplateMarketOptionsSpotOptions `json:"spot_options,omitempty"`
}

type terraformLaunchTemplateBlockDeviceEBS struct {
	// VolumeType is the ebs type to use
	VolumeType *string `json:"volume_type,omitempty"`
	// VolumeSize is the volume size
	VolumeSize *int64 `json:"volume_size,omitempty"`
	// IOPS is the provisioned iops
	IOPS *int64 `json:"iops,omitempty"`
	// DeleteOnTermination indicates the volume should die with the instance
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// Encrypted indicates the device should be encrypted
	Encrypted *bool `json:"encrypted,omitempty"`
}

type terraformLaunchTemplateBlockDevice struct {
	// DeviceName is the name of the device
	DeviceName *string `json:"device_name,omitempty"`
	// VirtualName is used for the ephemeral devices
	VirtualName *string `json:"virtual_name,omitempty"`
	// EBS defines the ebs spec
	EBS []*terraformLaunchTemplateBlockDeviceEBS `json:"ebs,omitempty"`
}

type terraformLaunchTemplateTagSpecification struct {
	// ResourceType is the type of resource to tag.
	ResourceType *string `json:"resource_type,omitempty"`
	// Tags are the tags to apply to the resource.
	Tags map[string]string `json:"tags,omitempty"`
}

type terraformLaunchTemplate struct {
	// NamePrefix is the name of the launch template
	NamePrefix *string `json:"name_prefix,omitempty"`
	// Lifecycle is the terraform lifecycle
	Lifecycle *terraform.Lifecycle `json:"lifecycle,omitempty"`

	// BlockDeviceMappings is the device mappings
	BlockDeviceMappings []*terraformLaunchTemplateBlockDevice `json:"block_device_mappings,omitempty"`
	// EBSOptimized indicates if the root device is ebs optimized
	EBSOptimized *bool `json:"ebs_optimized,omitempty"`
	// IAMInstanceProfile is the IAM profile to assign to the nodes
	IAMInstanceProfile []*terraformLaunchTemplateIAMProfile `json:"iam_instance_profile,omitempty"`
	// ImageID is the ami to use for the instances
	ImageID *string `json:"image_id,omitempty"`
	// InstanceType is the type of instance
	InstanceType *string `json:"instance_type,omitempty"`
	// KeyName is the ssh key to use
	KeyName *terraform.Literal `json:"key_name,omitempty"`
	// MarketOptions are the spot pricing options
	MarketOptions []*terraformLaunchTemplateMarketOptions `json:"instance_market_options,omitempty"`
	// Monitoring are the instance monitoring options
	Monitoring []*terraformLaunchTemplateMonitoring `json:"monitoring,omitempty"`
	// NetworkInterfaces are the networking options
	NetworkInterfaces []*terraformLaunchTemplateNetworkInterface `json:"network_interfaces,omitempty"`
	// Placement are the tenancy options
	Placement []*terraformLaunchTemplatePlacement `json:"placement,omitempty"`
	// TagSpecifications are the tags to apply to a resource when it is created.
	TagSpecifications []*terraformLaunchTemplateTagSpecification `json:"tag_specifications,omitempty"`
	// UserData is the user data for the instances
	UserData *terraform.Literal `json:"user_data,omitempty"`
}

// TerraformLink returns the terraform reference
func (t *LaunchTemplate) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_launch_template", fi.StringValue(t.Name), "id")
}

// VersionLink returns the terraform version reference
func (t *LaunchTemplate) VersionLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_launch_template", fi.StringValue(t.Name), "latest_version")
}

// RenderTerraform is responsible for rendering the terraform json
func (t *LaunchTemplate) RenderTerraform(target *terraform.TerraformTarget, a, e, changes *LaunchTemplate) error {
	var err error

	cloud := target.Cloud.(awsup.AWSCloud)

	var image *string
	if e.ImageID != nil {
		im, err := cloud.ResolveImage(fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}
		image = im.ImageId
	}

	tf := terraformLaunchTemplate{
		NamePrefix:   fi.String(fi.StringValue(e.Name) + "-"),
		EBSOptimized: e.RootVolumeOptimization,
		ImageID:      image,
		InstanceType: e.InstanceType,
		Lifecycle:    &terraform.Lifecycle{CreateBeforeDestroy: fi.Bool(true)},
		NetworkInterfaces: []*terraformLaunchTemplateNetworkInterface{
			{
				AssociatePublicIPAddress: e.AssociatePublicIP,
				DeleteOnTermination:      fi.Bool(true),
			},
		},
	}

	if e.SpotPrice != "" {
		tf.MarketOptions = []*terraformLaunchTemplateMarketOptions{
			{
				MarketType:  fi.String("spot"),
				SpotOptions: []*terraformLaunchTemplateMarketOptionsSpotOptions{{MaxPrice: fi.String(e.SpotPrice)}},
			},
		}
	}
	for _, x := range e.SecurityGroups {
		tf.NetworkInterfaces[0].SecurityGroups = append(tf.NetworkInterfaces[0].SecurityGroups, x.TerraformLink())
	}
	if e.SSHKey != nil {
		tf.KeyName = e.SSHKey.TerraformLink()
	}
	if e.Tenancy != nil {
		tf.Placement = []*terraformLaunchTemplatePlacement{{Tenancy: e.Tenancy}}
	}
	if e.IAMInstanceProfile != nil {
		tf.IAMInstanceProfile = []*terraformLaunchTemplateIAMProfile{
			{Name: e.IAMInstanceProfile.TerraformLink()},
		}
	}
	if e.UserData != nil {
		d, err := e.UserData.AsBytes()
		if err != nil {
			return err
		}
		b64d := base64.StdEncoding.EncodeToString(d)
		b64UserDataResource := fi.WrapResource(fi.NewStringResource(b64d))

		tf.UserData, err = target.AddFile("aws_launch_template", fi.StringValue(e.Name), "user_data", b64UserDataResource)
		if err != nil {
			return err
		}
	}
	devices, err := e.buildRootDevice(cloud)
	if err != nil {
		return err
	}
	for n, x := range devices {
		tf.BlockDeviceMappings = append(tf.BlockDeviceMappings, &terraformLaunchTemplateBlockDevice{
			DeviceName: fi.String(n),
			EBS: []*terraformLaunchTemplateBlockDeviceEBS{
				{
					DeleteOnTermination: fi.Bool(true),
					IOPS:                x.EbsVolumeIops,
					VolumeSize:          x.EbsVolumeSize,
					VolumeType:          x.EbsVolumeType,
				},
			},
		})
	}
	additionals, err := buildAdditionalDevices(e.BlockDeviceMappings)
	if err != nil {
		return err
	}
	for n, x := range additionals {
		tf.BlockDeviceMappings = append(tf.BlockDeviceMappings, &terraformLaunchTemplateBlockDevice{
			DeviceName: fi.String(n),
			EBS: []*terraformLaunchTemplateBlockDeviceEBS{
				{
					DeleteOnTermination: fi.Bool(true),
					Encrypted:           x.EbsEncrypted,
					IOPS:                x.EbsVolumeIops,
					VolumeSize:          x.EbsVolumeSize,
					VolumeType:          x.EbsVolumeType,
				},
			},
		})
	}

	devices, err = buildEphemeralDevices(cloud, fi.StringValue(e.InstanceType))
	if err != nil {
		return err
	}
	for n, x := range devices {
		tf.BlockDeviceMappings = append(tf.BlockDeviceMappings, &terraformLaunchTemplateBlockDevice{
			VirtualName: x.VirtualName,
			DeviceName:  fi.String(n),
		})
	}

	if e.Tags != nil {
		tf.TagSpecifications = append(tf.TagSpecifications, &terraformLaunchTemplateTagSpecification{
			ResourceType: fi.String("instance"),
			Tags:         e.Tags,
		})
		tf.TagSpecifications = append(tf.TagSpecifications, &terraformLaunchTemplateTagSpecification{
			ResourceType: fi.String("volume"),
			Tags:         e.Tags,
		})
	}

	return target.RenderResource("aws_launch_template", fi.StringValue(e.Name), tf)
}
