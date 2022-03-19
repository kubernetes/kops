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
	"time"

	"k8s.io/klog"

	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type Ocean struct {
	Name      *string
	Lifecycle fi.Lifecycle

	MinSize                  *int64
	MaxSize                  *int64
	UtilizeReservedInstances *bool
	UtilizeCommitments       *bool
	FallbackToOnDemand       *bool
	DrainingTimeout          *int64
	GracePeriod              *int64
	InstanceTypesWhitelist   []string
	InstanceTypesBlacklist   []string
	Tags                     map[string]string
	UserData                 fi.Resource
	ImageID                  *string
	IAMInstanceProfile       *awstasks.IAMInstanceProfile
	SSHKey                   *awstasks.SSHKey
	Subnets                  []*awstasks.Subnet
	SecurityGroups           []*awstasks.SecurityGroup
	Monitoring               *bool
	AssociatePublicIPAddress *bool
	UseAsTemplateOnly        *bool
	RootVolumeOpts           *RootVolumeOpts
	AutoScalerOpts           *AutoScalerOpts
}

var (
	_ fi.Task            = &Ocean{}
	_ fi.CompareWithID   = &Ocean{}
	_ fi.HasDependencies = &Ocean{}
)

func (o *Ocean) CompareWithID() *string {
	return o.Name
}

func (o *Ocean) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	if o.IAMInstanceProfile != nil {
		deps = append(deps, o.IAMInstanceProfile)
	}

	if o.SSHKey != nil {
		deps = append(deps, o.SSHKey)
	}

	if o.Subnets != nil {
		for _, subnet := range o.Subnets {
			deps = append(deps, subnet)
		}
	}

	if o.SecurityGroups != nil {
		for _, sg := range o.SecurityGroups {
			deps = append(deps, sg)
		}
	}

	if o.UserData != nil {
		deps = append(deps, fi.FindDependencies(tasks, o.UserData)...)
	}

	return deps
}

func (o *Ocean) find(svc spotinst.InstanceGroupService) (*aws.Cluster, error) {
	klog.V(4).Infof("Attempting to find Ocean: %q", fi.StringValue(o.Name))

	oceans, err := svc.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to find ocean %q: %v", fi.StringValue(o.Name), err)
	}

	var out *aws.Cluster
	for _, ocean := range oceans {
		if ocean.Name() == fi.StringValue(o.Name) {
			out = ocean.Obj().(*aws.Cluster)
			break
		}
	}
	if out == nil {
		return nil, fmt.Errorf("spotinst: failed to find ocean %q", fi.StringValue(o.Name))
	}

	klog.V(4).Infof("Ocean/%s: %s", fi.StringValue(o.Name), stringutil.Stringify(out))
	return out, nil
}

var _ fi.HasCheckExisting = &Ocean{}

func (o *Ocean) Find(c *fi.Context) (*Ocean, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	ocean, err := o.find(cloud.Spotinst().Ocean())
	if err != nil {
		return nil, err
	}

	actual := &Ocean{}
	actual.Name = ocean.Name

	// Capacity.
	{
		if !fi.BoolValue(ocean.Compute.LaunchSpecification.UseAsTemplateOnly) {
			actual.MinSize = fi.Int64(int64(fi.IntValue(ocean.Capacity.Minimum)))
			actual.MaxSize = fi.Int64(int64(fi.IntValue(ocean.Capacity.Maximum)))
		}
	}

	// Strategy.
	{
		if strategy := ocean.Strategy; strategy != nil {
			actual.FallbackToOnDemand = strategy.FallbackToOnDemand
			actual.UtilizeReservedInstances = strategy.UtilizeReservedInstances
			actual.UtilizeCommitments = strategy.UtilizeCommitments

			if strategy.DrainingTimeout != nil {
				actual.DrainingTimeout = fi.Int64(int64(fi.IntValue(strategy.DrainingTimeout)))
			}

			if strategy.GracePeriod != nil {
				actual.GracePeriod = fi.Int64(int64(fi.IntValue(strategy.GracePeriod)))
			}
		}
	}

	// Compute.
	{
		compute := ocean.Compute

		// Subnets.
		{
			if subnets := compute.SubnetIDs; subnets != nil {
				for _, subnetID := range subnets {
					actual.Subnets = append(actual.Subnets,
						&awstasks.Subnet{ID: fi.String(subnetID)})
				}
				if subnetSlicesEqualIgnoreOrder(actual.Subnets, o.Subnets) {
					actual.Subnets = o.Subnets
				}
			}
		}

		// Instance types.
		{
			if itypes := compute.InstanceTypes; itypes != nil {
				// Whitelist.
				if len(itypes.Whitelist) > 0 {
					actual.InstanceTypesWhitelist = itypes.Whitelist
				}

				// Blacklist.
				if len(itypes.Blacklist) > 0 {
					actual.InstanceTypesBlacklist = itypes.Blacklist
				}
			}
		}
	}

	// Launch specification.
	{
		lc := ocean.Compute.LaunchSpecification

		// Image.
		{
			actual.ImageID = lc.ImageID

			if o.ImageID != nil && actual.ImageID != nil &&
				fi.StringValue(actual.ImageID) != fi.StringValue(o.ImageID) {
				image, err := resolveImage(cloud, fi.StringValue(o.ImageID))
				if err != nil {
					return nil, err
				}
				if fi.StringValue(image.ImageId) == fi.StringValue(lc.ImageID) {
					actual.ImageID = o.ImageID
				}
			}
		}

		// Tags.
		{
			if lc.Tags != nil && len(lc.Tags) > 0 {
				actual.Tags = make(map[string]string)
				for _, tag := range lc.Tags {
					actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
				}
			}
		}

		// Security groups.
		{
			if lc.SecurityGroupIDs != nil {
				for _, sgID := range lc.SecurityGroupIDs {
					actual.SecurityGroups = append(actual.SecurityGroups,
						&awstasks.SecurityGroup{ID: fi.String(sgID)})
				}
			}
		}

		// User data.
		{
			if lc.UserData != nil {
				userData, err := base64.StdEncoding.DecodeString(fi.StringValue(lc.UserData))
				if err != nil {
					return nil, err
				}
				actual.UserData = fi.NewStringResource(string(userData))
			}
		}

		// EBS optimization.
		{
			if fi.BoolValue(lc.EBSOptimized) {
				if actual.RootVolumeOpts == nil {
					actual.RootVolumeOpts = new(RootVolumeOpts)
				}
				actual.RootVolumeOpts.Optimization = lc.EBSOptimized
			}
		}

		// IAM instance profile.
		if lc.IAMInstanceProfile != nil {
			actual.IAMInstanceProfile = &awstasks.IAMInstanceProfile{Name: lc.IAMInstanceProfile.Name}
		}

		// SSH key.
		if lc.KeyPair != nil {
			actual.SSHKey = &awstasks.SSHKey{Name: lc.KeyPair}
		}

		// Public IP.
		if lc.AssociatePublicIPAddress != nil {
			actual.AssociatePublicIPAddress = lc.AssociatePublicIPAddress
		}

		// Root volume options.
		if lc.RootVolumeSize != nil {
			actual.RootVolumeOpts = new(RootVolumeOpts)
			actual.RootVolumeOpts.Size = fi.Int64(int64(*lc.RootVolumeSize))
		}

		// Monitoring.
		if lc.Monitoring != nil {
			actual.Monitoring = lc.Monitoring
		}

		// Template.
		if lc.UseAsTemplateOnly != nil {
			actual.UseAsTemplateOnly = lc.UseAsTemplateOnly
		}
	}

	// Auto Scaler.
	{
		if ocean.AutoScaler != nil {
			actual.AutoScalerOpts = new(AutoScalerOpts)
			actual.AutoScalerOpts.ClusterID = ocean.ControllerClusterID
			actual.AutoScalerOpts.Enabled = ocean.AutoScaler.IsEnabled
			actual.AutoScalerOpts.AutoConfig = ocean.AutoScaler.IsAutoConfig
			actual.AutoScalerOpts.AutoHeadroomPercentage = ocean.AutoScaler.AutoHeadroomPercentage
			actual.AutoScalerOpts.Cooldown = ocean.AutoScaler.Cooldown

			// Scale down.
			if down := ocean.AutoScaler.Down; down != nil {
				actual.AutoScalerOpts.Down = &AutoScalerDownOpts{
					MaxPercentage:     down.MaxScaleDownPercentage,
					EvaluationPeriods: down.EvaluationPeriods,
				}
			}

			// Resource limits.
			if limits := ocean.AutoScaler.ResourceLimits; limits != nil {
				actual.AutoScalerOpts.ResourceLimits = &AutoScalerResourceLimitsOpts{
					MaxVCPU:   limits.MaxVCPU,
					MaxMemory: limits.MaxMemoryGiB,
				}
			}
		}
	}

	// Avoid spurious changes.
	actual.Lifecycle = o.Lifecycle

	return actual, nil
}

func (o *Ocean) CheckExisting(c *fi.Context) bool {
	cloud := c.Cloud.(awsup.AWSCloud)
	ocean, err := o.find(cloud.Spotinst().Ocean())
	return err == nil && ocean != nil
}

func (o *Ocean) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(o, c)
}

func (s *Ocean) CheckChanges(a, e, changes *Ocean) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	return nil
}

func (o *Ocean) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Ocean) error {
	return o.createOrUpdate(t.Cloud, a, e, changes)
}

func (o *Ocean) createOrUpdate(cloud awsup.AWSCloud, a, e, changes *Ocean) error {
	if a == nil {
		return o.create(cloud, a, e, changes)
	} else {
		return o.update(cloud, a, e, changes)
	}
}

func (_ *Ocean) create(cloud awsup.AWSCloud, a, e, changes *Ocean) error {
	klog.V(2).Infof("Creating Ocean %q", *e.Name)

	ocean := &aws.Cluster{
		Capacity: new(aws.Capacity),
		Strategy: new(aws.Strategy),
		Compute: &aws.Compute{
			LaunchSpecification: new(aws.LaunchSpecification),
		},
	}

	// General.
	{
		ocean.SetName(e.Name)
		ocean.SetRegion(fi.String(cloud.Region()))
	}

	// Capacity.
	{
		if !fi.BoolValue(e.UseAsTemplateOnly) {
			ocean.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
			ocean.Capacity.SetMinimum(fi.Int(int(*e.MinSize)))
			ocean.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
		}
	}

	// Strategy.
	{
		ocean.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
		ocean.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
		ocean.Strategy.SetUtilizeCommitments(e.UtilizeCommitments)

		if e.DrainingTimeout != nil {
			ocean.Strategy.SetDrainingTimeout(fi.Int(int(*e.DrainingTimeout)))
		}

		if e.GracePeriod != nil {
			ocean.Strategy.SetGracePeriod(fi.Int(int(*e.GracePeriod)))
		}
	}

	// Compute.
	{
		// Subnets.
		{
			if e.Subnets != nil {
				subnetIDs := make([]string, len(e.Subnets))
				for i, subnet := range e.Subnets {
					subnetIDs[i] = fi.StringValue(subnet.ID)
				}
				ocean.Compute.SetSubnetIDs(subnetIDs)
			}
		}

		// Instance types.
		{
			itypes := new(aws.InstanceTypes)

			// Whitelist.
			if e.InstanceTypesWhitelist != nil {
				itypes.SetWhitelist(e.InstanceTypesWhitelist)
			}

			// Blacklist.
			if e.InstanceTypesBlacklist != nil {
				itypes.SetBlacklist(e.InstanceTypesBlacklist)
			}

			if len(itypes.Whitelist) > 0 || len(itypes.Blacklist) > 0 {
				ocean.Compute.SetInstanceTypes(itypes)
			}
		}

		// Launch specification.
		{
			ocean.Compute.LaunchSpecification.SetUseAsTemplateOnly(e.UseAsTemplateOnly)
			ocean.Compute.LaunchSpecification.SetMonitoring(e.Monitoring)
			ocean.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)

			// Image.
			{
				if e.ImageID != nil {
					image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
					if err != nil {
						return err
					}
					ocean.Compute.LaunchSpecification.SetImageId(image.ImageId)
				}
			}

			// Security groups.
			{
				if e.SecurityGroups != nil {
					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}
					ocean.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
				}
			}

			if !fi.BoolValue(e.UseAsTemplateOnly) {
				// User data.
				{
					if e.UserData != nil {
						userData, err := fi.ResourceAsString(e.UserData)
						if err != nil {
							return err
						}

						if len(userData) > 0 {
							encoded := base64.StdEncoding.EncodeToString([]byte(userData))
							ocean.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
						}
					}
				}

				// IAM instance profile.
				{
					if e.IAMInstanceProfile != nil {
						iprof := new(aws.IAMInstanceProfile)
						iprof.SetName(e.IAMInstanceProfile.GetName())
						ocean.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
					}
				}

				// Root volume options.
				{
					if opts := e.RootVolumeOpts; opts != nil {

						// Volume size.
						if opts.Size != nil {
							ocean.Compute.LaunchSpecification.SetRootVolumeSize(fi.Int(int(*opts.Size)))
						}

						// EBS optimization.
						if opts.Optimization != nil {
							ocean.Compute.LaunchSpecification.SetEBSOptimized(opts.Optimization)
						}
					}
				}

				// Public IP.
				{
					if e.AssociatePublicIPAddress != nil {
						ocean.Compute.LaunchSpecification.SetAssociatePublicIPAddress(e.AssociatePublicIPAddress)
					}
				}

				// Tags.
				{
					if e.Tags != nil {
						ocean.Compute.LaunchSpecification.SetTags(e.buildTags())
					}
				}
			}

		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			ocean.SetControllerClusterId(opts.ClusterID)

			if opts.Enabled != nil {
				autoScaler := new(aws.AutoScaler)
				autoScaler.IsEnabled = opts.Enabled
				autoScaler.IsAutoConfig = opts.AutoConfig
				autoScaler.AutoHeadroomPercentage = opts.AutoHeadroomPercentage
				autoScaler.Cooldown = opts.Cooldown

				// Scale down.
				if down := opts.Down; down != nil {
					autoScaler.Down = &aws.AutoScalerDown{
						MaxScaleDownPercentage: down.MaxPercentage,
						EvaluationPeriods:      down.EvaluationPeriods,
					}
				}

				// Resource limits.
				if limits := opts.ResourceLimits; limits != nil {
					autoScaler.ResourceLimits = &aws.AutoScalerResourceLimits{
						MaxVCPU:      limits.MaxVCPU,
						MaxMemoryGiB: limits.MaxMemory,
					}
				}

				ocean.SetAutoScaler(autoScaler)
			}
		}
	}

	attempt := 0
	maxAttempts := 10

readyLoop:
	for {
		attempt++
		klog.V(2).Infof("(%d/%d) Attempting to create Ocean: %q, config: %s",
			attempt, maxAttempts, *e.Name, stringutil.Stringify(ocean))

		// Wait for IAM instance profile to be ready.
		time.Sleep(10 * time.Second)

		// Wrap the raw object as an Ocean.
		oc, err := spotinst.NewOcean(cloud.ProviderID(), ocean)
		if err != nil {
			return err
		}

		// Create a new Ocean.
		_, err = cloud.Spotinst().Ocean().Create(context.Background(), oc)
		if err == nil {
			break
		}

		if errs, ok := err.(client.Errors); ok {
			for _, err := range errs {
				if strings.Contains(err.Message, "Invalid IAM Instance Profile name") {
					if attempt > maxAttempts {
						return fmt.Errorf("IAM instance profile not yet created/propagated (original error: %v)", err)
					}

					klog.V(4).Infof("Got an error indicating that the IAM instance profile %q is not ready %q", fi.StringValue(e.IAMInstanceProfile.Name), err)
					klog.Infof("Waiting for IAM instance profile %q to be ready", fi.StringValue(e.IAMInstanceProfile.Name))
					goto readyLoop
				}
			}

			return fmt.Errorf("spotinst: failed to create ocean: %v", err)
		}
	}

	return nil
}

func (_ *Ocean) update(cloud awsup.AWSCloud, a, e, changes *Ocean) error {
	klog.V(2).Infof("Updating Ocean %q", *e.Name)

	actual, err := e.find(cloud.Spotinst().Ocean())
	if err != nil {
		klog.Errorf("Unable to resolve Ocean %q, error: %s", *e.Name, err)
		return err
	}

	var changed bool
	ocean := new(aws.Cluster)
	ocean.SetId(actual.ID)

	// Strategy.
	{
		// Fallback to on-demand.
		if changes.FallbackToOnDemand != nil {
			if ocean.Strategy == nil {
				ocean.Strategy = new(aws.Strategy)
			}

			ocean.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
			changes.FallbackToOnDemand = nil
			changed = true
		}

		// Utilize reserved instances.
		if changes.UtilizeReservedInstances != nil {
			if ocean.Strategy == nil {
				ocean.Strategy = new(aws.Strategy)
			}

			ocean.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
			changes.UtilizeReservedInstances = nil
			changed = true
		}

		// Utilize commitments.
		if changes.UtilizeCommitments != nil {
			if ocean.Strategy == nil {
				ocean.Strategy = new(aws.Strategy)
			}

			ocean.Strategy.SetUtilizeCommitments(e.UtilizeCommitments)
			changes.UtilizeCommitments = nil
			changed = true
		}

		// Draining timeout.
		if changes.DrainingTimeout != nil {
			if ocean.Strategy == nil {
				ocean.Strategy = new(aws.Strategy)
			}

			ocean.Strategy.SetDrainingTimeout(fi.Int(int(*e.DrainingTimeout)))
			changes.DrainingTimeout = nil
			changed = true
		}

		// Grace period.
		if changes.GracePeriod != nil {
			if ocean.Strategy == nil {
				ocean.Strategy = new(aws.Strategy)
			}

			ocean.Strategy.SetGracePeriod(fi.Int(int(*e.GracePeriod)))
			changes.GracePeriod = nil
			changed = true
		}
	}

	// Compute.
	{
		// Subnets.
		{
			if changes.Subnets != nil {
				if ocean.Compute == nil {
					ocean.Compute = new(aws.Compute)
				}

				subnetIDs := make([]string, len(e.Subnets))
				for i, subnet := range e.Subnets {
					subnetIDs[i] = fi.StringValue(subnet.ID)
				}

				ocean.Compute.SetSubnetIDs(subnetIDs)
				changes.Subnets = nil
				changed = true
			}
		}

		// Instance types.
		{
			// Whitelist.
			{
				if changes.InstanceTypesWhitelist != nil {
					if ocean.Compute == nil {
						ocean.Compute = new(aws.Compute)
					}
					if ocean.Compute.InstanceTypes == nil {
						ocean.Compute.InstanceTypes = new(aws.InstanceTypes)
					}

					ocean.Compute.InstanceTypes.SetWhitelist(e.InstanceTypesWhitelist)
					changes.InstanceTypesWhitelist = nil
					changed = true
				}
			}

			// Blacklist.
			{
				if changes.InstanceTypesBlacklist != nil {
					if ocean.Compute == nil {
						ocean.Compute = new(aws.Compute)
					}
					if ocean.Compute.InstanceTypes == nil {
						ocean.Compute.InstanceTypes = new(aws.InstanceTypes)
					}

					ocean.Compute.InstanceTypes.SetBlacklist(e.InstanceTypesBlacklist)
					changes.InstanceTypesBlacklist = nil
					changed = true
				}
			}
		}

		// Launch specification.
		{
			// Security groups.
			{
				if changes.SecurityGroups != nil {
					if ocean.Compute == nil {
						ocean.Compute = new(aws.Compute)
					}
					if ocean.Compute.LaunchSpecification == nil {
						ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}

					ocean.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
					changes.SecurityGroups = nil
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

					if *actual.Compute.LaunchSpecification.ImageID != *image.ImageId {
						if ocean.Compute == nil {
							ocean.Compute = new(aws.Compute)
						}
						if ocean.Compute.LaunchSpecification == nil {
							ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						ocean.Compute.LaunchSpecification.SetImageId(image.ImageId)
						changed = true
					}

					changes.ImageID = nil
				}
			}

			// Monitoring.
			{
				if changes.Monitoring != nil {
					if ocean.Compute == nil {
						ocean.Compute = new(aws.Compute)
					}
					if ocean.Compute.LaunchSpecification == nil {
						ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					ocean.Compute.LaunchSpecification.SetMonitoring(e.Monitoring)
					changes.Monitoring = nil
					changed = true
				}
			}

			// Template.
			{
				if changes.UseAsTemplateOnly != nil {
					if ocean.Compute == nil {
						ocean.Compute = new(aws.Compute)
					}
					if ocean.Compute.LaunchSpecification == nil {
						ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					ocean.Compute.LaunchSpecification.SetUseAsTemplateOnly(e.UseAsTemplateOnly)
					changes.UseAsTemplateOnly = nil
					changed = true
				}
			}

			// SSH key.
			{
				if changes.SSHKey != nil {
					if ocean.Compute == nil {
						ocean.Compute = new(aws.Compute)
					}
					if ocean.Compute.LaunchSpecification == nil {
						ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					ocean.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)
					changes.SSHKey = nil
					changed = true
				}
			}

			if !fi.BoolValue(e.UseAsTemplateOnly) {
				// User data.
				{
					if changes.UserData != nil {
						userData, err := fi.ResourceAsString(e.UserData)
						if err != nil {
							return err
						}

						if len(userData) > 0 {
							if ocean.Compute == nil {
								ocean.Compute = new(aws.Compute)
							}
							if ocean.Compute.LaunchSpecification == nil {
								ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							encoded := base64.StdEncoding.EncodeToString([]byte(userData))
							ocean.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
							changed = true
						}

						changes.UserData = nil
					}
				}

				// Tags.
				{
					if changes.Tags != nil {
						if ocean.Compute == nil {
							ocean.Compute = new(aws.Compute)
						}
						if ocean.Compute.LaunchSpecification == nil {
							ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						ocean.Compute.LaunchSpecification.SetTags(e.buildTags())
						changes.Tags = nil
						changed = true
					}
				}

				// IAM instance profile.
				{
					if changes.IAMInstanceProfile != nil {
						if ocean.Compute == nil {
							ocean.Compute = new(aws.Compute)
						}
						if ocean.Compute.LaunchSpecification == nil {
							ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						iprof := new(aws.IAMInstanceProfile)
						iprof.SetName(e.IAMInstanceProfile.GetName())

						ocean.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
						changes.IAMInstanceProfile = nil
						changed = true
					}
				}

				// Public IP.
				{
					if changes.AssociatePublicIPAddress != nil {
						if ocean.Compute == nil {
							ocean.Compute = new(aws.Compute)
						}
						if ocean.Compute.LaunchSpecification == nil {
							ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						ocean.Compute.LaunchSpecification.SetAssociatePublicIPAddress(e.AssociatePublicIPAddress)
						changes.AssociatePublicIPAddress = nil
						changed = true
					}
				}

				// Root volume options.
				{
					if opts := changes.RootVolumeOpts; opts != nil {

						// Volume size.
						if opts.Size != nil {
							if ocean.Compute == nil {
								ocean.Compute = new(aws.Compute)
							}
							if ocean.Compute.LaunchSpecification == nil {
								ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							ocean.Compute.LaunchSpecification.SetRootVolumeSize(fi.Int(int(*opts.Size)))
							changed = true
						}

						// EBS optimization.
						if opts.Optimization != nil {
							if ocean.Compute == nil {
								ocean.Compute = new(aws.Compute)
							}
							if ocean.Compute.LaunchSpecification == nil {
								ocean.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							ocean.Compute.LaunchSpecification.SetEBSOptimized(e.RootVolumeOpts.Optimization)
							changed = true
						}

						changes.RootVolumeOpts = nil
					}
				}
			}
		}
	}

	// Capacity.
	{
		if !fi.BoolValue(e.UseAsTemplateOnly) {
			if changes.MinSize != nil {
				if ocean.Capacity == nil {
					ocean.Capacity = new(aws.Capacity)
				}

				ocean.Capacity.SetMinimum(fi.Int(int(*e.MinSize)))
				changes.MinSize = nil
				changed = true

				// Scale up the target capacity, if needed.
				if int64(*actual.Capacity.Target) < *e.MinSize {
					ocean.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
				}
			}
			if changes.MaxSize != nil {
				if ocean.Capacity == nil {
					ocean.Capacity = new(aws.Capacity)
				}

				ocean.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
				changes.MaxSize = nil
				changed = true
			}
		}
	}

	// Auto Scaler.
	{
		if opts := changes.AutoScalerOpts; opts != nil {
			if opts.Enabled != nil {
				autoScaler := new(aws.AutoScaler)
				autoScaler.IsEnabled = e.AutoScalerOpts.Enabled
				autoScaler.IsAutoConfig = e.AutoScalerOpts.AutoConfig
				autoScaler.AutoHeadroomPercentage = e.AutoScalerOpts.AutoHeadroomPercentage
				autoScaler.Cooldown = e.AutoScalerOpts.Cooldown

				// Scale down.
				if down := opts.Down; down != nil {
					autoScaler.Down = &aws.AutoScalerDown{
						MaxScaleDownPercentage: down.MaxPercentage,
						EvaluationPeriods:      down.EvaluationPeriods,
					}
				} else if a.AutoScalerOpts.Down != nil {
					autoScaler.SetDown(nil)
				}

				// Resource limits.
				if limits := opts.ResourceLimits; limits != nil {
					autoScaler.ResourceLimits = &aws.AutoScalerResourceLimits{
						MaxVCPU:      limits.MaxVCPU,
						MaxMemoryGiB: limits.MaxMemory,
					}
				} else if a.AutoScalerOpts.ResourceLimits != nil {
					autoScaler.SetResourceLimits(nil)
				}

				ocean.SetAutoScaler(autoScaler)
				changed = true
			}

			changes.AutoScalerOpts = nil
		}
	}

	empty := &Ocean{}
	if !reflect.DeepEqual(empty, changes) {
		klog.Warningf("Not all changes applied to Ocean %q: %v", *e.Name, changes)
	}

	if !changed {
		klog.V(2).Infof("No changes detected in Ocean %q", *e.Name)
		return nil
	}

	klog.V(2).Infof("Updating Ocean %q (config: %s)", *e.Name, stringutil.Stringify(ocean))

	// Wrap the raw object as an Ocean.
	oc, err := spotinst.NewOcean(cloud.ProviderID(), ocean)
	if err != nil {
		return err
	}

	// Update an existing Ocean.
	if err := cloud.Spotinst().Ocean().Update(context.Background(), oc); err != nil {
		return fmt.Errorf("spotinst: failed to update ocean: %v", err)
	}

	return nil
}

type terraformOcean struct {
	Name                   *string                    `cty:"name"`
	ControllerClusterID    *string                    `cty:"controller_id"`
	Region                 *string                    `cty:"region"`
	InstanceTypesWhitelist []string                   `cty:"whitelist"`
	InstanceTypesBlacklist []string                   `cty:"blacklist"`
	SubnetIDs              []*terraformWriter.Literal `cty:"subnet_ids"`
	AutoScaler             *terraformAutoScaler       `cty:"autoscaler"`
	Tags                   []*terraformKV             `cty:"tags"`

	MinSize         *int64 `cty:"min_size"`
	MaxSize         *int64 `cty:"max_size"`
	DesiredCapacity *int64 `cty:"desired_capacity"`

	FallbackToOnDemand       *bool  `cty:"fallback_to_ondemand"`
	UtilizeReservedInstances *bool  `cty:"utilize_reserved_instances"`
	UtilizeCommitments       *bool  `cty:"utilize_commitments"`
	DrainingTimeout          *int64 `cty:"draining_timeout"`
	GracePeriod              *int64 `cty:"grace_period"`

	UseAsTemplateOnly        *bool                      `cty:"use_as_template_only"`
	Monitoring               *bool                      `cty:"monitoring"`
	EBSOptimized             *bool                      `cty:"ebs_optimized"`
	ImageID                  *string                    `cty:"image_id"`
	AssociatePublicIPAddress *bool                      `cty:"associate_public_ip_address"`
	RootVolumeSize           *int64                     `cty:"root_volume_size"`
	UserData                 *terraformWriter.Literal   `cty:"user_data"`
	IAMInstanceProfile       *terraformWriter.Literal   `cty:"iam_instance_profile"`
	KeyName                  *terraformWriter.Literal   `cty:"key_name"`
	SecurityGroups           []*terraformWriter.Literal `cty:"security_groups"`
}

func (_ *Ocean) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Ocean) error {
	cloud := t.Cloud.(awsup.AWSCloud)

	tf := &terraformOcean{
		Name:                     e.Name,
		Region:                   fi.String(cloud.Region()),
		UseAsTemplateOnly:        e.UseAsTemplateOnly,
		FallbackToOnDemand:       e.FallbackToOnDemand,
		UtilizeReservedInstances: e.UtilizeReservedInstances,
		UtilizeCommitments:       e.UtilizeCommitments,
		DrainingTimeout:          e.DrainingTimeout,
		GracePeriod:              e.GracePeriod,
	}

	// Image.
	if e.ImageID != nil {
		image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}
		tf.ImageID = image.ImageId
	}

	var role string
	for key := range e.Tags {
		if strings.HasPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix) {
			suffix := strings.TrimPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix)
			if role != "" && role != suffix {
				return fmt.Errorf("spotinst: found multiple role tags %q vs %q", role, suffix)
			}
			role = suffix
		}
	}

	// Security groups.
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

	// Monitoring.
	if e.Monitoring != nil {
		tf.Monitoring = e.Monitoring
	}

	// SSH key.
	if e.SSHKey != nil {
		tf.KeyName = e.SSHKey.TerraformLink()
	}

	// Subnets.
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

	// Instance types.
	{
		// Whitelist.
		if e.InstanceTypesWhitelist != nil {
			tf.InstanceTypesWhitelist = e.InstanceTypesWhitelist
		}

		// Blacklist.
		if e.InstanceTypesBlacklist != nil {
			tf.InstanceTypesBlacklist = e.InstanceTypesBlacklist
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			tf.ControllerClusterID = opts.ClusterID

			if opts.Enabled != nil {
				tf.AutoScaler = &terraformAutoScaler{
					Enabled:                opts.Enabled,
					AutoConfig:             opts.AutoConfig,
					AutoHeadroomPercentage: opts.AutoHeadroomPercentage,
					Cooldown:               opts.Cooldown,
				}

				// Scale down.
				if down := opts.Down; down != nil {
					tf.AutoScaler.Down = &terraformAutoScalerDown{
						MaxPercentage:     down.MaxPercentage,
						EvaluationPeriods: down.EvaluationPeriods,
					}
				}

				// Resource limits.
				if limits := opts.ResourceLimits; limits != nil {
					tf.AutoScaler.ResourceLimits = &terraformAutoScalerResourceLimits{
						MaxVCPU:   limits.MaxVCPU,
						MaxMemory: limits.MaxVCPU,
					}
				}
			}
		}
	}

	if !fi.BoolValue(tf.UseAsTemplateOnly) {
		// Capacity.
		tf.DesiredCapacity = e.MinSize
		tf.MinSize = e.MinSize
		tf.MaxSize = e.MaxSize

		// Root volume options.
		if opts := e.RootVolumeOpts; opts != nil {

			// Volume size.
			if opts.Size != nil {
				tf.RootVolumeSize = opts.Size
			}

			// EBS optimization.
			if opts.Optimization != nil {
				tf.EBSOptimized = opts.Optimization
			}
		}

		// IAM instance profile.
		if e.IAMInstanceProfile != nil {
			tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
		}

		// User data.
		if e.UserData != nil {
			var err error
			tf.UserData, err = t.AddFileResource("spotinst_ocean_aws", *e.Name, "user_data", e.UserData, false)
			if err != nil {
				return err
			}
		}

		// Public IP.
		if e.AssociatePublicIPAddress != nil {
			tf.AssociatePublicIPAddress = e.AssociatePublicIPAddress
		}

		// Tags.
		if e.Tags != nil {
			for _, tag := range e.buildTags() {
				tf.Tags = append(tf.Tags, &terraformKV{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
		}
	}

	return t.RenderResource("spotinst_ocean_aws", *e.Name, tf)
}

func (o *Ocean) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("spotinst_ocean_aws", *o.Name, "id")
}

func (o *Ocean) buildTags() []*aws.Tag {
	tags := make([]*aws.Tag, 0, len(o.Tags))

	for key, value := range o.Tags {
		tags = append(tags, &aws.Tag{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return tags
}
