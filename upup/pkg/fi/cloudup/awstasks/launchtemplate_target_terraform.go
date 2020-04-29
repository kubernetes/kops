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
	AssociatePublicIPAddress *bool `json:"associate_public_ip_address,omitempty" cty:"associate_public_ip_address"`
	// DeleteOnTermination indicates whether the network interface should be destroyed on instance termination.
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty" cty:"delete_on_termination"`
	// SecurityGroups is a list of security group ids.
	SecurityGroups []*terraform.Literal `json:"security_groups,omitempty" cty:"security_groups"`
}

type terraformLaunchTemplateMonitoring struct {
	// Enabled indicates that monitoring is enabled
	Enabled *bool `json:"enabled,omitempty" cty:"enabled"`
}

type terraformLaunchTemplatePlacement struct {
	// Affinity is he affinity setting for an instance on a Dedicated Host.
	Affinity *string `json:"affinity,omitempty" cty:"affinity"`
	// AvailabilityZone is the Availability Zone for the instance.
	AvailabilityZone *string `json:"availability_zone,omitempty" cty:"availability_zone"`
	// GroupName is the name of the placement group for the instance.
	GroupName *string `json:"group_name,omitempty" cty:"group_name"`
	// HostID is the ID of the Dedicated Host for the instance.
	HostID *string `json:"host_id,omitempty" cty:"host_id"`
	// SpreadDomain are reserved for future use.
	SpreadDomain *string `json:"spread_domain,omitempty" cty:"spread_domain"`
	// Tenancy ist he tenancy of the instance. Can be default, dedicated, or host.
	Tenancy *string `json:"tenancy,omitempty" cty:"tenancy"`
}

type terraformLaunchTemplateIAMProfile struct {
	// Name is the name of the profile
	Name *terraform.Literal `json:"name,omitempty" cty:"name"`
}

type terraformLaunchTemplateMarketOptionsSpotOptions struct {
	// BlockDurationMinutes is required duration in minutes. This value must be a multiple of 60.
	BlockDurationMinutes *int64 `json:"block_duration_minutes,omitempty" cty:"block_duration_minutes"`
	// InstanceInterruptionBehavior is the behavior when a Spot Instance is interrupted. Can be hibernate, stop, or terminate
	InstanceInterruptionBehavior *string `json:"instance_interruption_behavior,omitempty" cty:"instance_interruption_behavior"`
	// MaxPrice is the maximum hourly price you're willing to pay for the Spot Instances
	MaxPrice *string `json:"max_price,omitempty" cty:"max_price"`
	// SpotInstanceType is the Spot Instance request type. Can be one-time, or persistent
	SpotInstanceType *string `json:"spot_instance_type,omitempty" cty:"spot_instance_type"`
	// ValidUntil is the end date of the request
	ValidUntil *string `json:"valid_until,omitempty" cty:"valid_until"`
}

type terraformLaunchTemplateMarketOptions struct {
	// MarketType is the option type
	MarketType *string `json:"market_type,omitempty" cty:"market_type"`
	// SpotOptions are the set of options
	SpotOptions []*terraformLaunchTemplateMarketOptionsSpotOptions `json:"spot_options,omitempty" cty:"spot_options"`
}

type terraformLaunchTemplateBlockDeviceEBS struct {
	// VolumeType is the ebs type to use
	VolumeType *string `json:"volume_type,omitempty" cty:"volume_type"`
	// VolumeSize is the volume size
	VolumeSize *int64 `json:"volume_size,omitempty" cty:"volume_size"`
	// IOPS is the provisioned iops
	IOPS *int64 `json:"iops,omitempty" cty:"iops"`
	// DeleteOnTermination indicates the volume should die with the instance
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty" cty:"delete_on_termination"`
	// Encrypted indicates the device should be encrypted
	Encrypted *bool `json:"encrypted,omitempty" cty:"encrypted"`
}

type terraformLaunchTemplateBlockDevice struct {
	// DeviceName is the name of the device
	DeviceName *string `json:"device_name,omitempty" cty:"device_name"`
	// VirtualName is used for the ephemeral devices
	VirtualName *string `json:"virtual_name,omitempty" cty:"virtual_name"`
	// EBS defines the ebs spec
	EBS []*terraformLaunchTemplateBlockDeviceEBS `json:"ebs,omitempty" cty:"ebs"`
}

type terraformLaunchTemplateTagSpecification struct {
	// ResourceType is the type of resource to tag.
	ResourceType *string `json:"resource_type,omitempty" cty:"resource_type"`
	// Tags are the tags to apply to the resource.
	Tags map[string]string `json:"tags,omitempty" cty:"tags"`
}

type terraformLaunchTemplate struct {
	// NamePrefix is the name of the launch template
	NamePrefix *string `json:"name_prefix,omitempty" cty:"name_prefix"`
	// Lifecycle is the terraform lifecycle
	Lifecycle *terraform.Lifecycle `json:"lifecycle,omitempty" cty:"lifecycle"`

	// BlockDeviceMappings is the device mappings
	BlockDeviceMappings []*terraformLaunchTemplateBlockDevice `json:"block_device_mappings,omitempty" cty:"block_device_mappings"`
	// EBSOptimized indicates if the root device is ebs optimized
	EBSOptimized *bool `json:"ebs_optimized,omitempty" cty:"ebs_optimized"`
	// IAMInstanceProfile is the IAM profile to assign to the nodes
	IAMInstanceProfile []*terraformLaunchTemplateIAMProfile `json:"iam_instance_profile,omitempty" cty:"iam_instance_profile"`
	// ImageID is the ami to use for the instances
	ImageID *string `json:"image_id,omitempty" cty:"image_id"`
	// InstanceType is the type of instance
	InstanceType *string `json:"instance_type,omitempty" cty:"instance_type"`
	// KeyName is the ssh key to use
	KeyName *terraform.Literal `json:"key_name,omitempty" cty:"key_name"`
	// MarketOptions are the spot pricing options
	MarketOptions []*terraformLaunchTemplateMarketOptions `json:"instance_market_options,omitempty" cty:"instance_market_options"`
	// Monitoring are the instance monitoring options
	Monitoring []*terraformLaunchTemplateMonitoring `json:"monitoring,omitempty" cty:"monitoring"`
	// NetworkInterfaces are the networking options
	NetworkInterfaces []*terraformLaunchTemplateNetworkInterface `json:"network_interfaces,omitempty" cty:"network_interfaces"`
	// Placement are the tenancy options
	Placement []*terraformLaunchTemplatePlacement `json:"placement,omitempty" cty:"placement"`
	// TagSpecifications are the tags to apply to a resource when it is created.
	TagSpecifications []*terraformLaunchTemplateTagSpecification `json:"tag_specifications,omitempty" cty:"tag_specifications"`
	// UserData is the user data for the instances
	UserData *terraform.Literal `json:"user_data,omitempty" cty:"user_data"`
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
		marketSpotOptions := terraformLaunchTemplateMarketOptionsSpotOptions{MaxPrice: fi.String(e.SpotPrice)}
		if e.SpotDurationInMinutes != nil {
			marketSpotOptions.BlockDurationMinutes = e.SpotDurationInMinutes
		}
		if e.InstanceInterruptionBehavior != nil {
			marketSpotOptions.InstanceInterruptionBehavior = e.InstanceInterruptionBehavior
		}
		tf.MarketOptions = []*terraformLaunchTemplateMarketOptions{
			{
				MarketType:  fi.String("spot"),
				SpotOptions: []*terraformLaunchTemplateMarketOptionsSpotOptions{&marketSpotOptions},
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
		if b64d != "" {
			b64UserDataResource := fi.WrapResource(fi.NewStringResource(b64d))

			tf.UserData, err = target.AddFile("aws_launch_template", fi.StringValue(e.Name), "user_data", b64UserDataResource)
			if err != nil {
				return err
			}
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
