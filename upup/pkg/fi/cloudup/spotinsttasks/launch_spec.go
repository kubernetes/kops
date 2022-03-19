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

package spotinsttasks

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type LaunchSpec struct {
	Name      *string
	Lifecycle fi.Lifecycle

	SpotPercentage           *int64
	UserData                 fi.Resource
	SecurityGroups           []*awstasks.SecurityGroup
	Subnets                  []*awstasks.Subnet
	IAMInstanceProfile       *awstasks.IAMInstanceProfile
	ImageID                  *string
	InstanceTypes            []string
	Tags                     map[string]string
	RootVolumeOpts           *RootVolumeOpts
	AutoScalerOpts           *AutoScalerOpts
	RestrictScaleDown        *bool
	AssociatePublicIPAddress *bool
	MinSize                  *int64
	MaxSize                  *int64

	Ocean *Ocean
}

var (
	_ fi.Task            = &LaunchSpec{}
	_ fi.CompareWithID   = &LaunchSpec{}
	_ fi.HasDependencies = &LaunchSpec{}
)

func (o *LaunchSpec) CompareWithID() *string {
	return o.Name
}

func (o *LaunchSpec) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	if o.IAMInstanceProfile != nil {
		deps = append(deps, o.IAMInstanceProfile)
	}

	if o.SecurityGroups != nil {
		for _, sg := range o.SecurityGroups {
			deps = append(deps, sg)
		}
	}

	if o.Subnets != nil {
		for _, subnet := range o.Subnets {
			deps = append(deps, subnet)
		}
	}

	if o.Ocean != nil {
		deps = append(deps, o.Ocean)
	}

	if o.UserData != nil {
		deps = append(deps, fi.FindDependencies(tasks, o.UserData)...)
	}

	return deps
}

func (o *LaunchSpec) find(svc spotinst.LaunchSpecService, oceanID string) (*aws.LaunchSpec, error) {
	klog.V(4).Infof("Attempting to find LaunchSpec: %q", fi.StringValue(o.Name))

	specs, err := svc.List(context.Background(), oceanID)
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to find launch spec %q: %v", fi.StringValue(o.Name), err)
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("spotinst: no launch specs associated with ocean %q", oceanID)
	}

	var out *aws.LaunchSpec
	for _, spec := range specs {
		if spec.Name() == fi.StringValue(o.Name) {
			out = spec.Obj().(*aws.LaunchSpec)
			break
		}
	}
	if out == nil {
		return nil, fmt.Errorf("spotinst: failed to find launch spec %q", fi.StringValue(o.Name))
	}

	klog.V(4).Infof("LaunchSpec/%s: %s", fi.StringValue(o.Name), stringutil.Stringify(out))
	return out, nil
}

var _ fi.HasCheckExisting = &LaunchSpec{}

func (o *LaunchSpec) Find(c *fi.Context) (*LaunchSpec, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	ocean, err := o.Ocean.find(cloud.Spotinst().Ocean())
	if err != nil {
		return nil, err
	}

	spec, err := o.find(cloud.Spotinst().LaunchSpec(), *ocean.ID)
	if err != nil {
		return nil, err
	}

	actual := &LaunchSpec{}
	actual.Name = spec.Name
	actual.Ocean = &Ocean{
		Name: ocean.Name,
	}
	actual.RestrictScaleDown = spec.RestrictScaleDown

	// Capacity.
	{
		if spec.ResourceLimits != nil {
			actual.MinSize = fi.Int64(int64(fi.IntValue(spec.ResourceLimits.MinInstanceCount)))
			actual.MaxSize = fi.Int64(int64(fi.IntValue(spec.ResourceLimits.MaxInstanceCount)))
		}
	}

	// Image.
	{
		actual.ImageID = spec.ImageID

		if o.ImageID != nil && actual.ImageID != nil &&
			fi.StringValue(actual.ImageID) != fi.StringValue(o.ImageID) {
			image, err := resolveImage(cloud, fi.StringValue(o.ImageID))
			if err != nil {
				return nil, err
			}
			if fi.StringValue(image.ImageId) == fi.StringValue(spec.ImageID) {
				actual.ImageID = o.ImageID
			}
		}
	}

	// User data.
	{
		if spec.UserData != nil {
			userData, err := base64.StdEncoding.DecodeString(fi.StringValue(spec.UserData))
			if err != nil {
				return nil, err
			}
			actual.UserData = fi.NewStringResource(string(userData))
		}
	}

	// IAM instance profile.
	{
		if spec.IAMInstanceProfile != nil {
			actual.IAMInstanceProfile = &awstasks.IAMInstanceProfile{Name: spec.IAMInstanceProfile.Name}
		}
	}

	// Root volume options.
	{
		// Volume size.
		{
			if spec.RootVolumeSize != nil {
				actual.RootVolumeOpts = new(RootVolumeOpts)
				actual.RootVolumeOpts.Size = fi.Int64(int64(*spec.RootVolumeSize))
			}
		}

		// Block device mappings.
		{
			if spec.BlockDeviceMappings != nil {
				for _, b := range spec.BlockDeviceMappings {
					if b.EBS == nil || b.EBS.SnapshotID != nil {
						continue // not the root
					}
					if actual.RootVolumeOpts == nil {
						actual.RootVolumeOpts = new(RootVolumeOpts)
					}
					if b.EBS.VolumeType != nil {
						actual.RootVolumeOpts.Type = fi.String(strings.ToLower(fi.StringValue(b.EBS.VolumeType)))
					}
					if b.EBS.VolumeSize != nil {
						actual.RootVolumeOpts.Size = fi.Int64(int64(fi.IntValue(b.EBS.VolumeSize)))
					}
					if b.EBS.IOPS != nil {
						actual.RootVolumeOpts.IOPS = fi.Int64(int64(fi.IntValue(b.EBS.IOPS)))
					}
					if b.EBS.Throughput != nil {
						actual.RootVolumeOpts.Throughput = fi.Int64(int64(fi.IntValue(b.EBS.Throughput)))
					}
				}
			}
		}
	}

	// Security groups.
	{
		if spec.SecurityGroupIDs != nil {
			for _, sgID := range spec.SecurityGroupIDs {
				actual.SecurityGroups = append(actual.SecurityGroups,
					&awstasks.SecurityGroup{ID: fi.String(sgID)})
			}
		}
	}

	// Subnets.
	{
		if spec.SubnetIDs != nil {
			for _, subnetID := range spec.SubnetIDs {
				actual.Subnets = append(actual.Subnets,
					&awstasks.Subnet{ID: fi.String(subnetID)})
			}
			if subnetSlicesEqualIgnoreOrder(actual.Subnets, o.Subnets) {
				actual.Subnets = o.Subnets
			}
		}
	}

	// Public IP.
	if spec.AssociatePublicIPAddress != nil {
		actual.AssociatePublicIPAddress = spec.AssociatePublicIPAddress
	}

	// Instance types.
	{
		if itypes := spec.InstanceTypes; itypes != nil {
			actual.InstanceTypes = itypes
		}
	}

	// Tags.
	{
		if len(spec.Tags) > 0 {
			actual.Tags = make(map[string]string)
			for _, tag := range spec.Tags {
				actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
			}
		}
	}

	// Auto Scaler.
	{
		if spec.AutoScale != nil {
			actual.AutoScalerOpts = new(AutoScalerOpts)

			// Headroom.
			if headrooms := spec.AutoScale.Headrooms; len(headrooms) > 0 {
				actual.AutoScalerOpts.Headroom = &AutoScalerHeadroomOpts{
					CPUPerUnit: headrooms[0].CPUPerUnit,
					GPUPerUnit: headrooms[0].GPUPerUnit,
					MemPerUnit: headrooms[0].MemoryPerUnit,
					NumOfUnits: headrooms[0].NumOfUnits,
				}
			}
		}
	}

	// Labels.
	if labels := spec.Labels; labels != nil {
		if actual.AutoScalerOpts == nil {
			actual.AutoScalerOpts = new(AutoScalerOpts)
		}

		actual.AutoScalerOpts.Labels = make(map[string]string)
		for _, label := range spec.Labels {
			actual.AutoScalerOpts.Labels[fi.StringValue(label.Key)] = fi.StringValue(label.Value)
		}
	}

	// Taints.
	if spec.Taints != nil {
		if actual.AutoScalerOpts == nil {
			actual.AutoScalerOpts = new(AutoScalerOpts)
		}

		actual.AutoScalerOpts.Taints = make([]*corev1.Taint, len(spec.Taints))
		for i, taint := range spec.Taints {
			actual.AutoScalerOpts.Taints[i] = &corev1.Taint{
				Key:    fi.StringValue(taint.Key),
				Value:  fi.StringValue(taint.Value),
				Effect: corev1.TaintEffect(fi.StringValue(taint.Effect)),
			}
		}
	}

	// Strategy.
	{
		if strategy := spec.Strategy; strategy != nil {
			if strategy.SpotPercentage != nil {
				actual.SpotPercentage = fi.Int64(int64(fi.IntValue(strategy.SpotPercentage)))
			}
		}
	}

	// Avoid spurious changes.
	actual.Lifecycle = o.Lifecycle

	return actual, nil
}

func (o *LaunchSpec) CheckExisting(c *fi.Context) bool {
	spec, err := o.Find(c)
	return err == nil && spec != nil
}

func (o *LaunchSpec) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(o, c)
}

func (s *LaunchSpec) CheckChanges(a, e, changes *LaunchSpec) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	return nil
}

func (o *LaunchSpec) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LaunchSpec) error {
	return o.createOrUpdate(t.Cloud, a, e, changes)
}

func (o *LaunchSpec) createOrUpdate(cloud awsup.AWSCloud, a, e, changes *LaunchSpec) error {
	if a == nil {
		return o.create(cloud, a, e, changes)
	} else {
		return o.update(cloud, a, e, changes)
	}
}

func (_ *LaunchSpec) create(cloud awsup.AWSCloud, a, e, changes *LaunchSpec) error {
	ocean, err := e.Ocean.find(cloud.Spotinst().Ocean())
	if err != nil {
		return err
	}

	klog.V(2).Infof("Creating Launch Spec for Ocean %q", *ocean.ID)

	spec := &aws.LaunchSpec{
		Strategy: new(aws.LaunchSpecStrategy),
	}

	spec.SetName(e.Name)
	spec.SetOceanId(ocean.ID)

	// Capacity.
	{
		if e.MinSize != nil || e.MaxSize != nil {
			spec.ResourceLimits = new(aws.ResourceLimits)
			spec.ResourceLimits.SetMinInstanceCount(fi.Int(int(*e.MinSize)))
			spec.ResourceLimits.SetMaxInstanceCount(fi.Int(int(*e.MaxSize)))
		}
	}

	// Image.
	{
		if e.ImageID != nil {
			image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
			if err != nil {
				return err
			}
			spec.SetImageId(image.ImageId)
		}
	}

	// User data.
	{
		if e.UserData != nil {
			userData, err := fi.ResourceAsString(e.UserData)
			if err != nil {
				return err
			}

			if len(userData) > 0 {
				encoded := base64.StdEncoding.EncodeToString([]byte(userData))
				spec.SetUserData(fi.String(encoded))
			}
		}
	}

	// IAM instance profile.
	{
		if e.IAMInstanceProfile != nil {
			iprof := new(aws.IAMInstanceProfile)
			iprof.SetName(e.IAMInstanceProfile.GetName())
			spec.SetIAMInstanceProfile(iprof)
		}
	}

	// Root volume options.
	{
		if opts := e.RootVolumeOpts; opts != nil {
			rootDevice, err := buildRootDevice(cloud, e.RootVolumeOpts, e.ImageID)
			if err != nil {
				return err
			}

			spec.SetBlockDeviceMappings([]*aws.BlockDeviceMapping{
				e.convertBlockDeviceMapping(rootDevice),
			})
		}
	}

	// Security groups.
	{
		if e.SecurityGroups != nil {
			securityGroupIDs := make([]string, len(e.SecurityGroups))
			for i, sg := range e.SecurityGroups {
				securityGroupIDs[i] = *sg.ID
			}
			spec.SetSecurityGroupIDs(securityGroupIDs)
		}
	}

	// Subnets.
	{
		if e.Subnets != nil {
			subnetIDs := make([]string, len(e.Subnets))
			for i, subnet := range e.Subnets {
				subnetIDs[i] = fi.StringValue(subnet.ID)
			}
			spec.SetSubnetIDs(subnetIDs)
		}
	}

	// Public IP.
	{
		if e.AssociatePublicIPAddress != nil {
			spec.SetAssociatePublicIPAddress(e.AssociatePublicIPAddress)
		}
	}

	// Instance types.
	{
		if e.InstanceTypes != nil {
			spec.SetInstanceTypes(e.InstanceTypes)
		}
	}

	// Tags.
	{
		if e.Tags != nil {
			spec.SetTags(e.buildTags())
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			// Headroom.
			if headroom := opts.Headroom; headroom != nil {
				autoScale := new(aws.AutoScale)
				autoScale.Headrooms = []*aws.AutoScaleHeadroom{
					{
						CPUPerUnit:    headroom.CPUPerUnit,
						GPUPerUnit:    headroom.GPUPerUnit,
						MemoryPerUnit: headroom.MemPerUnit,
						NumOfUnits:    headroom.NumOfUnits,
					},
				}
				spec.SetAutoScale(autoScale)
			}

			// Labels.
			if len(opts.Labels) > 0 {
				var labels []*aws.Label
				for k, v := range opts.Labels {
					labels = append(labels, &aws.Label{
						Key:   fi.String(k),
						Value: fi.String(v),
					})
				}
				spec.SetLabels(labels)
			}

			// Taints.
			if len(opts.Taints) > 0 {
				taints := make([]*aws.Taint, len(opts.Taints))
				for i, taint := range opts.Taints {
					taints[i] = &aws.Taint{
						Key:    fi.String(taint.Key),
						Value:  fi.String(taint.Value),
						Effect: fi.String(string(taint.Effect)),
					}
				}
				spec.SetTaints(taints)
			}
		}
	}

	// Strategy.
	{
		if e.SpotPercentage != nil {
			spec.Strategy.SetSpotPercentage(fi.Int(int(*e.SpotPercentage)))
		}
	}

	// Restrictions.
	{
		if fi.BoolValue(e.RestrictScaleDown) {
			spec.SetRestrictScaleDown(e.RestrictScaleDown)
		}
	}

	// Wrap the raw object as a LaunchSpec.
	sp, err := spotinst.NewLaunchSpec(cloud.ProviderID(), spec)
	if err != nil {
		return err
	}

	// Create a new LaunchSpec.
	_, err = cloud.Spotinst().LaunchSpec().Create(context.Background(), sp)
	if err != nil {
		return fmt.Errorf("spotinst: failed to create launch spec: %v", err)
	}

	return nil
}

func (_ *LaunchSpec) update(cloud awsup.AWSCloud, a, e, changes *LaunchSpec) error {
	klog.V(2).Infof("Updating Launch Spec for Ocean %q", *a.Ocean.Name)

	ocean, err := a.Ocean.find(cloud.Spotinst().Ocean())
	if err != nil {
		klog.Errorf("Unable to resolve Ocean %q, error: %v", *a.Ocean.Name, err)
		return err
	}

	actual, err := e.find(cloud.Spotinst().LaunchSpec(), *ocean.ID)
	if err != nil {
		klog.Errorf("Unable to resolve Launch Spec %q, error: %v", *e.Name, err)
		return err
	}

	var changed bool
	spec := new(aws.LaunchSpec)
	spec.SetId(actual.ID)

	// Capacity.
	{
		if changes.MinSize != nil {
			if spec.ResourceLimits == nil {
				spec.ResourceLimits = new(aws.ResourceLimits)
			}

			spec.ResourceLimits.SetMinInstanceCount(fi.Int(int(*e.MinSize)))
			changes.MinSize = nil
			changed = true
		}
		if changes.MaxSize != nil {
			if spec.ResourceLimits == nil {
				spec.ResourceLimits = new(aws.ResourceLimits)
			}

			spec.ResourceLimits.SetMaxInstanceCount(fi.Int(int(*e.MaxSize)))
			changes.MaxSize = nil
			changed = true
		}
	}

	// Image.
	{
		if changes.ImageID != nil {
			image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
			if err != nil {
				return err
			}

			if fi.StringValue(actual.ImageID) != fi.StringValue(image.ImageId) {
				spec.SetImageId(image.ImageId)
			}

			changes.ImageID = nil
			changed = true
		}
	}

	// User data.
	{
		if changes.UserData != nil {
			userData, err := fi.ResourceAsString(e.UserData)
			if err != nil {
				return err
			}

			if len(userData) > 0 {
				encoded := base64.StdEncoding.EncodeToString([]byte(userData))
				spec.SetUserData(fi.String(encoded))
				changed = true
			}

			changes.UserData = nil
		}
	}

	// IAM instance profile.
	{
		if changes.IAMInstanceProfile != nil {
			iprof := new(aws.IAMInstanceProfile)
			iprof.SetName(e.IAMInstanceProfile.GetName())

			spec.SetIAMInstanceProfile(iprof)
			changes.IAMInstanceProfile = nil
			changed = true
		}
	}

	// Root volume options.
	{
		if opts := changes.RootVolumeOpts; opts != nil {
			rootDevice, err := buildRootDevice(cloud, opts, e.ImageID)
			if err != nil {
				return err
			}

			spec.SetRootVolumeSize(nil) // mutually exclusive
			spec.SetBlockDeviceMappings([]*aws.BlockDeviceMapping{
				e.convertBlockDeviceMapping(rootDevice),
			})
			changes.RootVolumeOpts = nil
			changed = true
		}
	}

	// Security groups.
	{
		if changes.SecurityGroups != nil {
			securityGroupIDs := make([]string, len(e.SecurityGroups))
			for i, sg := range e.SecurityGroups {
				securityGroupIDs[i] = *sg.ID
			}

			spec.SetSecurityGroupIDs(securityGroupIDs)
			changes.SecurityGroups = nil
			changed = true
		}
	}

	// Subnets.
	{
		if changes.Subnets != nil {
			subnetIDs := make([]string, len(e.Subnets))
			for i, subnet := range e.Subnets {
				subnetIDs[i] = fi.StringValue(subnet.ID)
			}

			spec.SetSubnetIDs(subnetIDs)
			changes.Subnets = nil
			changed = true
		}
	}

	// Public IP.
	{
		if changes.AssociatePublicIPAddress != nil {
			spec.SetAssociatePublicIPAddress(e.AssociatePublicIPAddress)
			changes.AssociatePublicIPAddress = nil
			changed = true
		}
	}

	// Instance types.
	{
		if changes.InstanceTypes != nil {
			spec.SetInstanceTypes(e.InstanceTypes)
			changes.InstanceTypes = nil
			changed = true
		}
	}

	// Tags.
	{
		if changes.Tags != nil {
			spec.SetTags(e.buildTags())
			changes.Tags = nil
			changed = true
		}
	}

	// Auto Scaler.
	{
		if opts := changes.AutoScalerOpts; opts != nil {
			// Headroom.
			if headroom := opts.Headroom; headroom != nil {
				autoScale := new(aws.AutoScale)
				autoScale.Headrooms = []*aws.AutoScaleHeadroom{
					{
						CPUPerUnit:    e.AutoScalerOpts.Headroom.CPUPerUnit,
						GPUPerUnit:    e.AutoScalerOpts.Headroom.GPUPerUnit,
						MemoryPerUnit: e.AutoScalerOpts.Headroom.MemPerUnit,
						NumOfUnits:    e.AutoScalerOpts.Headroom.NumOfUnits,
					},
				}

				spec.SetAutoScale(autoScale)
				opts.Headroom = nil
				changed = true
			}

			// Labels.
			if opts.Labels != nil {
				labels := make([]*aws.Label, 0, len(e.AutoScalerOpts.Labels))
				for k, v := range e.AutoScalerOpts.Labels {
					labels = append(labels, &aws.Label{
						Key:   fi.String(k),
						Value: fi.String(v),
					})
				}

				spec.SetLabels(labels)
				opts.Labels = nil
				changed = true
			}

			// Taints.
			if opts.Taints != nil {
				taints := make([]*aws.Taint, 0, len(e.AutoScalerOpts.Taints))
				for _, taint := range e.AutoScalerOpts.Taints {
					taints = append(taints, &aws.Taint{
						Key:    fi.String(taint.Key),
						Value:  fi.String(taint.Value),
						Effect: fi.String(string(taint.Effect)),
					})
				}

				spec.SetTaints(taints)
				opts.Taints = nil
				changed = true
			}

			changes.AutoScalerOpts = nil
		}
	}

	// Strategy.
	{
		// Spot percentage.
		if changes.SpotPercentage != nil {
			if spec.Strategy == nil {
				spec.Strategy = new(aws.LaunchSpecStrategy)
			}

			spec.Strategy.SetSpotPercentage(fi.Int(int(fi.Int64Value(e.SpotPercentage))))
			changes.SpotPercentage = nil
			changed = true
		}
	}

	// Restrictions.
	{
		if changes.RestrictScaleDown != nil {
			spec.SetRestrictScaleDown(e.RestrictScaleDown)
			changes.RestrictScaleDown = nil
			changed = true
		}
	}

	empty := &LaunchSpec{}
	if !reflect.DeepEqual(empty, changes) {
		klog.Warningf("Not all changes applied to Launch Spec %q: %v", *e.Name, changes)
	}

	if !changed {
		klog.V(2).Infof("No changes detected in Launch Spec %q", *e.Name)
		return nil
	}

	klog.V(2).Infof("Updating Launch Spec %q (config: %s)", *e.Name, stringutil.Stringify(spec))
	ctx := context.Background()

	// Reset the Spot percentage on the Cluster level.
	if spec.Strategy != nil && spec.Strategy.SpotPercentage != nil &&
		ocean.Strategy != nil && ocean.Strategy.SpotPercentage != nil {
		c := &aws.Cluster{Strategy: new(aws.Strategy)}
		c.SetId(ocean.ID)
		c.Strategy.SetSpotPercentage(nil)

		// Wrap the raw object as a Cluster.
		o, err := spotinst.NewOcean(cloud.ProviderID(), c)
		if err != nil {
			return err
		}

		// Update the existing Cluster.
		if err = cloud.Spotinst().Ocean().Update(ctx, o); err != nil {
			return err
		}
	}

	// Wrap the raw object as a Launch Spec.
	sp, err := spotinst.NewLaunchSpec(cloud.ProviderID(), spec)
	if err != nil {
		return err
	}

	// Update the existing Launch Spec.
	if err = cloud.Spotinst().LaunchSpec().Update(ctx, sp); err != nil {
		return fmt.Errorf("spotinst: failed to update launch spec: %v", err)
	}

	return nil
}

type terraformLaunchSpec struct {
	Name    *string                  `cty:"name"`
	OceanID *terraformWriter.Literal `cty:"ocean_id"`

	Monitoring               *bool                              `cty:"monitoring"`
	EBSOptimized             *bool                              `cty:"ebs_optimized"`
	ImageID                  *string                            `cty:"image_id"`
	AssociatePublicIPAddress *bool                              `cty:"associate_public_ip_address"`
	RestrictScaleDown        *bool                              `cty:"restrict_scale_down"`
	RootVolumeSize           *int64                             `cty:"root_volume_size"`
	UserData                 *terraformWriter.Literal           `cty:"user_data"`
	IAMInstanceProfile       *terraformWriter.Literal           `cty:"iam_instance_profile"`
	KeyName                  *terraformWriter.Literal           `cty:"key_name"`
	InstanceTypes            []string                           `cty:"instance_types"`
	SubnetIDs                []*terraformWriter.Literal         `cty:"subnet_ids"`
	SecurityGroups           []*terraformWriter.Literal         `cty:"security_groups"`
	Taints                   []*terraformTaint                  `cty:"taints"`
	Labels                   []*terraformKV                     `cty:"labels"`
	Tags                     []*terraformKV                     `cty:"tags"`
	Headrooms                []*terraformAutoScalerHeadroom     `cty:"autoscale_headrooms"`
	BlockDeviceMappings      []*terraformBlockDeviceMapping     `cty:"block_device_mappings"`
	Strategy                 *terraformLaunchSpecStrategy       `cty:"strategy"`
	ResourceLimits           *terraformLaunchSpecResourceLimits `cty:"resource_limits"`
}

type terraformLaunchSpecStrategy struct {
	SpotPercentage *int64 `cty:"spot_percentage"`
}

type terraformLaunchSpecResourceLimits struct {
	MinInstanceCount *int64 `cty:"min_instance_count"`
	MaxInstanceCount *int64 `cty:"max_instance_count"`
}

type terraformBlockDeviceMapping struct {
	DeviceName *string                         `cty:"device_name"`
	EBS        *terraformBlockDeviceMappingEBS `cty:"ebs"`
}

type terraformBlockDeviceMappingEBS struct {
	VirtualName         *string `cty:"virtual_name"`
	VolumeType          *string `cty:"volume_type"`
	VolumeSize          *int64  `cty:"volume_size"`
	VolumeIOPS          *int64  `cty:"iops"`
	VolumeThroughput    *int64  `cty:"throughput"`
	DeleteOnTermination *bool   `cty:"delete_on_termination"`
}

func (_ *LaunchSpec) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LaunchSpec) error {
	cloud := t.Cloud.(awsup.AWSCloud)

	tf := &terraformLaunchSpec{
		Name:                     e.Name,
		OceanID:                  e.Ocean.TerraformLink(),
		InstanceTypes:            e.InstanceTypes,
		AssociatePublicIPAddress: e.AssociatePublicIPAddress,
	}

	// Capacity.
	{
		if e.MinSize != nil || e.MaxSize != nil {
			tf.ResourceLimits = &terraformLaunchSpecResourceLimits{
				MinInstanceCount: e.MinSize,
				MaxInstanceCount: e.MaxSize,
			}
		}
	}

	// Image.
	{
		if e.ImageID != nil {
			image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
			if err != nil {
				return err
			}
			tf.ImageID = image.ImageId
		}
	}

	var role string
	for key := range e.Ocean.Tags {
		if strings.HasPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix) {
			suffix := strings.TrimPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix)
			if role != "" && role != suffix {
				return fmt.Errorf("spotinst: found multiple role tags %q vs %q", role, suffix)
			}
			role = suffix
		}
	}

	// Security groups.
	{
		if e.SecurityGroups != nil {
			for _, sg := range e.SecurityGroups {
				tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
				if role != "" {
					if err := t.AddOutputVariableArray(role+"_security_groups", sg.TerraformLink()); err != nil {
						return err
					}
				}
			}
		}
	}

	// Subnets.
	{
		if e.Subnets != nil {
			for _, subnet := range e.Subnets {
				tf.SubnetIDs = append(tf.SubnetIDs, subnet.TerraformLink())
				if role != "" {
					if err := t.AddOutputVariableArray(role+"_subnet_ids", subnet.TerraformLink()); err != nil {
						return err
					}
				}
			}
		}
	}

	// User data.
	{
		if e.UserData != nil {
			var err error
			tf.UserData, err = t.AddFileResource("spotinst_ocean_aws_launch_spec", *e.Name, "user_data", e.UserData, false)
			if err != nil {
				return err
			}
		}
	}

	// IAM instance profile.
	{
		if e.IAMInstanceProfile != nil {
			tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
		}
	}

	// Root volume options.
	if opts := e.RootVolumeOpts; opts != nil {
		rootDevice, err := buildRootDevice(t.Cloud.(awsup.AWSCloud), e.RootVolumeOpts, e.ImageID)
		if err != nil {
			return err
		}

		tf.BlockDeviceMappings = append(tf.BlockDeviceMappings, &terraformBlockDeviceMapping{
			DeviceName: rootDevice.DeviceName,
			EBS: &terraformBlockDeviceMappingEBS{
				VolumeType:          rootDevice.EbsVolumeType,
				VolumeSize:          rootDevice.EbsVolumeSize,
				VolumeIOPS:          rootDevice.EbsVolumeIops,
				VolumeThroughput:    rootDevice.EbsVolumeThroughput,
				DeleteOnTermination: fi.Bool(true),
			},
		})
	}

	// Tags.
	{
		if e.Tags != nil {
			for _, tag := range e.buildTags() {
				tf.Tags = append(tf.Tags, &terraformKV{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			// Headroom.
			if headroom := opts.Headroom; headroom != nil {
				tf.Headrooms = []*terraformAutoScalerHeadroom{
					{
						CPUPerUnit: headroom.CPUPerUnit,
						GPUPerUnit: headroom.GPUPerUnit,
						MemPerUnit: headroom.MemPerUnit,
						NumOfUnits: headroom.NumOfUnits,
					},
				}
			}

			// Labels.
			if len(opts.Labels) > 0 {
				tf.Labels = make([]*terraformKV, 0, len(opts.Labels))
				for k, v := range opts.Labels {
					tf.Labels = append(tf.Labels, &terraformKV{
						Key:   fi.String(k),
						Value: fi.String(v),
					})
				}
			}

			// Taints.
			if len(opts.Taints) > 0 {
				tf.Taints = make([]*terraformTaint, len(opts.Taints))
				for i, taint := range opts.Taints {
					t := &terraformTaint{
						Key:    fi.String(taint.Key),
						Effect: fi.String(string(taint.Effect)),
					}
					if taint.Value != "" {
						t.Value = fi.String(taint.Value)
					}
					tf.Taints[i] = t
				}
			}
		}
	}

	// Strategy.
	{
		if e.SpotPercentage != nil {
			tf.Strategy = &terraformLaunchSpecStrategy{
				SpotPercentage: e.SpotPercentage,
			}
		}
	}

	// Restrictions.
	{
		if fi.BoolValue(e.RestrictScaleDown) {
			tf.RestrictScaleDown = e.RestrictScaleDown
		}
	}

	return t.RenderResource("spotinst_ocean_aws_launch_spec", *e.Name, tf)
}

func (o *LaunchSpec) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("spotinst_ocean_aws_launch_spec", *o.Name, "id")
}

func (o *LaunchSpec) buildTags() []*aws.Tag {
	tags := make([]*aws.Tag, 0, len(o.Tags))

	for key, value := range o.Tags {
		tags = append(tags, &aws.Tag{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return tags
}

func (o *LaunchSpec) convertBlockDeviceMapping(in *awstasks.BlockDeviceMapping) *aws.BlockDeviceMapping {
	out := &aws.BlockDeviceMapping{
		DeviceName:  in.DeviceName,
		VirtualName: in.VirtualName,
	}

	if in.EbsDeleteOnTermination != nil || in.EbsVolumeSize != nil || in.EbsVolumeType != nil {
		out.EBS = &aws.EBS{
			VolumeType:          in.EbsVolumeType,
			VolumeSize:          fi.Int(int(fi.Int64Value(in.EbsVolumeSize))),
			DeleteOnTermination: in.EbsDeleteOnTermination,
		}

		// IOPS is not valid for gp2 volumes.
		if in.EbsVolumeIops != nil && fi.StringValue(in.EbsVolumeType) != "gp2" {
			out.EBS.IOPS = fi.Int(int(fi.Int64Value(in.EbsVolumeIops)))
		}

		// Throughput is only valid for gp3 volumes.
		if in.EbsVolumeThroughput != nil && fi.StringValue(in.EbsVolumeType) == "gp3" {
			out.EBS.Throughput = fi.Int(int(fi.Int64Value(in.EbsVolumeThroughput)))
		}
	}

	return out
}
