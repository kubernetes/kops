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

package spotinstmodel

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks"
)

const (
	// InstanceGroupLabelHybrid is the metadata label used on the instance group
	// to specify that the Spotinst provider should be used to upon creation.
	InstanceGroupLabelHybrid  = "spotinst.io/hybrid"
	InstanceGroupLabelManaged = "spotinst.io/managed" // for backward compatibility

	// InstanceGroupLabelSpotPercentage is the metadata label used on the
	// instance group to specify the percentage of Spot instances that
	// should spin up from the target capacity.
	InstanceGroupLabelSpotPercentage = "spotinst.io/spot-percentage"

	// InstanceGroupLabelOrientation is the metadata label used on the
	// instance group to specify which orientation should be used.
	InstanceGroupLabelOrientation = "spotinst.io/orientation"

	// InstanceGroupLabelUtilizeReservedInstances is the metadata label used
	// on the instance group to specify whether reserved instances should be
	// utilized.
	InstanceGroupLabelUtilizeReservedInstances = "spotinst.io/utilize-reserved-instances"

	// InstanceGroupLabelFallbackToOnDemand is the metadata label used on the
	// instance group to specify whether fallback to on-demand instances should
	// be enabled.
	InstanceGroupLabelFallbackToOnDemand = "spotinst.io/fallback-to-ondemand"

	// InstanceGroupLabelGracePeriod is the metadata label used on the
	// instance group to specify a period of time, in seconds, that Ocean
	// should wait before applying instance health checks.
	InstanceGroupLabelGracePeriod = "spotinst.io/grace-period"

	// InstanceGroupLabelHealthCheckType is the metadata label used on the
	// instance group to specify the type of the health check that should be used.
	InstanceGroupLabelHealthCheckType = "spotinst.io/health-check-type"

	// InstanceGroupLabelOceanDefaultLaunchSpec is the metadata label used on the
	// instance group to specify whether to use the InstanceGroup's spec as the default
	// Launch Spec for the Ocean cluster.
	InstanceGroupLabelOceanDefaultLaunchSpec = "spotinst.io/ocean-default-launchspec"

	// InstanceGroupLabelOceanInstanceTypes[White|Black]list are the metadata labels
	// used on the instance group to specify whether to whitelist or blacklist
	// specific instance types.
	InstanceGroupLabelOceanInstanceTypesWhitelist = "spotinst.io/ocean-instance-types-whitelist"
	InstanceGroupLabelOceanInstanceTypesBlacklist = "spotinst.io/ocean-instance-types-blacklist"

	// InstanceGroupLabelAutoScalerDisabled is the metadata label used on the
	// instance group to specify whether the auto scaler should be enabled.
	InstanceGroupLabelAutoScalerDisabled = "spotinst.io/autoscaler-disabled"

	// InstanceGroupLabelAutoScalerDefaultNodeLabels is the metadata label used on the
	// instance group to specify whether default node labels should be set for
	// the auto scaler.
	InstanceGroupLabelAutoScalerDefaultNodeLabels = "spotinst.io/autoscaler-default-node-labels"

	// InstanceGroupLabelAutoScalerHeadroom* are the metadata labels used on the
	// instance group to specify the headroom configuration used by the auto scaler.
	InstanceGroupLabelAutoScalerHeadroomCPUPerUnit = "spotinst.io/autoscaler-headroom-cpu-per-unit"
	InstanceGroupLabelAutoScalerHeadroomGPUPerUnit = "spotinst.io/autoscaler-headroom-gpu-per-unit"
	InstanceGroupLabelAutoScalerHeadroomMemPerUnit = "spotinst.io/autoscaler-headroom-mem-per-unit"
	InstanceGroupLabelAutoScalerHeadroomNumOfUnits = "spotinst.io/autoscaler-headroom-num-of-units"

	// InstanceGroupLabelAutoScalerCooldown is the metadata label used on the
	// instance group to specify the cooldown period (in seconds) for scaling actions.
	InstanceGroupLabelAutoScalerCooldown = "spotinst.io/autoscaler-cooldown"

	// InstanceGroupLabelAutoScalerScaleDown* are the metadata labels used on the
	// instance group to specify the scale down configuration used by the auto scaler.
	InstanceGroupLabelAutoScalerScaleDownMaxPercentage     = "spotinst.io/autoscaler-scale-down-max-percentage"
	InstanceGroupLabelAutoScalerScaleDownEvaluationPeriods = "spotinst.io/autoscaler-scale-down-evaluation-periods"
)

// InstanceGroupModelBuilder configures InstanceGroup objects
type InstanceGroupModelBuilder struct {
	*model.KopsModelContext

	BootstrapScript   *model.BootstrapScript
	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &InstanceGroupModelBuilder{}

func (b *InstanceGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var nodeInstanceGroups []*kops.InstanceGroup
	var err error

	for _, ig := range b.InstanceGroups {
		name := b.AutoscalingGroupName(ig)

		if featureflag.SpotinstHybrid.Enabled() {
			if !HybridInstanceGroup(ig) {
				klog.V(2).Infof("Skipping instance group: %q", name)
				continue
			}
		}

		klog.V(2).Infof("Building instance group: %q", name)
		switch ig.Spec.Role {

		// Create both Master and Bastion instance groups as Elastigroups.
		case kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleBastion:
			err = b.buildElastigroup(c, ig)

		// Create Node instance groups as Elastigroups or a single Ocean with
		// multiple LaunchSpecs.
		case kops.InstanceGroupRoleNode:
			if featureflag.SpotinstOcean.Enabled() {
				nodeInstanceGroups = append(nodeInstanceGroups, ig)
			} else {
				err = b.buildElastigroup(c, ig)
			}

		default:
			err = fmt.Errorf("spotinst: unexpected instance group role: %s", ig.Spec.Role)
		}

		if err != nil {
			return fmt.Errorf("spotinst: error building elastigroup: %v", err)
		}
	}

	if len(nodeInstanceGroups) > 0 {
		if err = b.buildOcean(c, nodeInstanceGroups...); err != nil {
			return fmt.Errorf("spotinst: error building ocean: %v", err)
		}
	}

	return nil
}

func (b *InstanceGroupModelBuilder) buildElastigroup(c *fi.ModelBuilderContext, ig *kops.InstanceGroup) (err error) {
	klog.V(4).Infof("Building instance group as Elastigroup: %q", b.AutoscalingGroupName(ig))
	group := &spotinsttasks.Elastigroup{
		Lifecycle:            b.Lifecycle,
		Name:                 fi.String(b.AutoscalingGroupName(ig)),
		ImageID:              fi.String(ig.Spec.Image),
		OnDemandInstanceType: fi.String(strings.Split(ig.Spec.MachineType, ",")[0]),
		SpotInstanceTypes:    strings.Split(ig.Spec.MachineType, ","),
	}

	// Cloud config.
	if cfg := b.Cluster.Spec.CloudConfig; cfg != nil {
		group.Product = cfg.SpotinstProduct
		group.Orientation = cfg.SpotinstOrientation
	}

	// Strategy.
	for k, v := range ig.ObjectMeta.Labels {
		switch k {
		case InstanceGroupLabelSpotPercentage:
			group.SpotPercentage, err = parseFloat(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelOrientation:
			group.Orientation = fi.String(v)

		case InstanceGroupLabelUtilizeReservedInstances:
			group.UtilizeReservedInstances, err = parseBool(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelFallbackToOnDemand:
			group.FallbackToOnDemand, err = parseBool(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelHealthCheckType:
			group.HealthCheckType = fi.String(strings.ToUpper(v))
		}
	}

	// Spot percentage.
	if group.SpotPercentage == nil {
		group.SpotPercentage = defaultSpotPercentage(ig)
	}

	// Instance profile.
	group.IAMInstanceProfile, err = b.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return fmt.Errorf("error building iam instance profile: %v", err)
	}

	// Root volume.
	group.RootVolumeOpts, err = b.buildRootVolumeOpts(ig)
	if err != nil {
		return fmt.Errorf("error building root volume options: %v", err)
	}

	// Tenancy.
	if ig.Spec.Tenancy != "" {
		group.Tenancy = fi.String(ig.Spec.Tenancy)
	}

	// Security groups.
	group.SecurityGroups, err = b.buildSecurityGroups(c, ig)
	if err != nil {
		return fmt.Errorf("error building security groups: %v", err)
	}

	// SSH key.
	group.SSHKey, err = b.LinkToSSHKey()
	if err != nil {
		return fmt.Errorf("error building ssh key: %v", err)
	}

	// Load balancer.
	var lb *awstasks.LoadBalancer
	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		if b.UseLoadBalancerForAPI() {
			lb = b.LinkToELB("api")
		}
	case kops.InstanceGroupRoleBastion:
		lb = b.LinkToELB(model.BastionELBSecurityGroupPrefix)
	}
	if lb != nil {
		group.LoadBalancer = lb
	}

	// User data.
	group.UserData, err = b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
	if err != nil {
		return fmt.Errorf("error building user data: %v", err)
	}

	// Public IP.
	group.AssociatePublicIP, err = b.buildPublicIpOpts(ig)
	if err != nil {
		return fmt.Errorf("error building public ip options: %v", err)
	}

	// Subnets.
	group.Subnets, err = b.buildSubnets(ig)
	if err != nil {
		return fmt.Errorf("error building subnets: %v", err)
	}

	// Capacity.
	group.MinSize, group.MaxSize = b.buildCapacity(ig)

	// Monitoring.
	group.Monitoring = ig.Spec.DetailedInstanceMonitoring

	// Tags.
	group.Tags, err = b.buildTags(ig)
	if err != nil {
		return fmt.Errorf("error building cloud tags: %v", err)
	}

	// Auto Scaler.
	group.AutoScalerOpts, err = b.buildAutoScalerOpts(b.ClusterName(), ig)
	if err != nil {
		return fmt.Errorf("error building auto scaler options: %v", err)
	}
	if group.AutoScalerOpts != nil { // remove unsupported options
		group.AutoScalerOpts.Taints = nil
	}

	klog.V(4).Infof("Adding task: Elastigroup/%s", fi.StringValue(group.Name))
	c.AddTask(group)

	return nil
}

func (b *InstanceGroupModelBuilder) buildOcean(c *fi.ModelBuilderContext, igs ...*kops.InstanceGroup) (err error) {
	klog.V(4).Infof("Building instance group as Ocean: %q", "nodes."+b.ClusterName())
	ocean := &spotinsttasks.Ocean{
		Lifecycle: b.Lifecycle,
		Name:      fi.String("nodes." + b.ClusterName()),
	}

	// Attempt to find the default LaunchSpec.
	var ig *kops.InstanceGroup
	{
		// Single instance group.
		if len(igs) == 1 {
			ig = igs[0]
		}

		// Multiple instance groups.
		if len(igs) > 1 {
			for _, g := range igs {
				for k, v := range g.ObjectMeta.Labels {
					if k == InstanceGroupLabelOceanDefaultLaunchSpec {
						defaultLaunchSpec, err := parseBool(v)
						if err != nil {
							continue
						}

						if fi.BoolValue(defaultLaunchSpec) {
							if ig != nil {
								return fmt.Errorf("unable to detect default launch spec: "+
									"multiple instance groups labeled with `%s: \"true\"`",
									InstanceGroupLabelOceanDefaultLaunchSpec)
							}

							ig = g
							break
						}
					}
				}
			}
			if ig == nil {
				return fmt.Errorf("unable to detect default launch spec: "+
					"please label the desired default instance group with `%s: \"true\"`",
					InstanceGroupLabelOceanDefaultLaunchSpec)
			}
		}

		klog.V(4).Infof("Detected default launch spec: %q", b.AutoscalingGroupName(ig))
	}

	// Image.
	ocean.ImageID = fi.String(ig.Spec.Image)

	// Strategy and instance types.
	for k, v := range ig.ObjectMeta.Labels {
		switch k {
		case InstanceGroupLabelSpotPercentage:
			ocean.SpotPercentage, err = parseFloat(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelUtilizeReservedInstances:
			ocean.UtilizeReservedInstances, err = parseBool(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelFallbackToOnDemand:
			ocean.FallbackToOnDemand, err = parseBool(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelGracePeriod:
			ocean.GracePeriod, err = parseInt(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelOceanInstanceTypesWhitelist:
			ocean.InstanceTypesWhitelist, err = parseStringSlice(v)
			if err != nil {
				return err
			}

		case InstanceGroupLabelOceanInstanceTypesBlacklist:
			ocean.InstanceTypesBlacklist, err = parseStringSlice(v)
			if err != nil {
				return err
			}
		}
	}

	// Spot percentage.
	if ocean.SpotPercentage == nil {
		ocean.SpotPercentage = defaultSpotPercentage(ig)
	}

	// Capacity.
	ocean.MinSize = fi.Int64(0)
	ocean.MaxSize = fi.Int64(0)

	// Monitoring.
	ocean.Monitoring = ig.Spec.DetailedInstanceMonitoring

	// User data.
	ocean.UserData, err = b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
	if err != nil {
		return fmt.Errorf("error building user data: %v", err)
	}

	// Instance profile.
	ocean.IAMInstanceProfile, err = b.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return fmt.Errorf("error building iam instance profile: %v", err)
	}

	// Root volume.
	rootVolumeOpts, err := b.buildRootVolumeOpts(ig)
	if err != nil {
		return fmt.Errorf("error building root volume options: %v", err)
	}
	if rootVolumeOpts != nil {
		ocean.RootVolumeOpts = rootVolumeOpts
		ocean.RootVolumeOpts.Type = nil // not supported in Ocean
	}

	// Security groups.
	ocean.SecurityGroups, err = b.buildSecurityGroups(c, ig)
	if err != nil {
		return fmt.Errorf("error building security groups: %v", err)
	}

	// SSH key.
	ocean.SSHKey, err = b.LinkToSSHKey()
	if err != nil {
		return fmt.Errorf("error building ssh key: %v", err)
	}

	// Public IP.
	ocean.AssociatePublicIP, err = b.buildPublicIpOpts(ig)
	if err != nil {
		return fmt.Errorf("error building public ip options: %v", err)
	}

	// Subnets.
	ocean.Subnets, err = b.buildSubnets(ig)
	if err != nil {
		return fmt.Errorf("error building subnets: %v", err)
	}

	// Tags.
	ocean.Tags, err = b.buildTags(ig)
	if err != nil {
		return fmt.Errorf("error building cloud tags: %v", err)
	}

	// Auto Scaler.
	ocean.AutoScalerOpts, err = b.buildAutoScalerOpts(b.ClusterName(), ig)
	if err != nil {
		return fmt.Errorf("error building auto scaler options: %v", err)
	}
	if ocean.AutoScalerOpts != nil { // remove unsupported options
		ocean.AutoScalerOpts.Labels = nil
		ocean.AutoScalerOpts.Taints = nil
	}

	// Create a Launch Spec for each instance group.
	for _, ig := range igs {
		if err := b.buildLaunchSpec(c, ig, ocean); err != nil {
			return fmt.Errorf("error building launch spec: %v", err)
		}
	}

	klog.V(4).Infof("Adding task: Ocean/%s", fi.StringValue(ocean.Name))
	c.AddTask(ocean)

	return nil
}

func (b *InstanceGroupModelBuilder) buildLaunchSpec(c *fi.ModelBuilderContext,
	ig *kops.InstanceGroup, ocean *spotinsttasks.Ocean) (err error) {

	klog.V(4).Infof("Building instance group as LaunchSpec: %q", b.AutoscalingGroupName(ig))
	launchSpec := &spotinsttasks.LaunchSpec{
		Name:    fi.String(b.AutoscalingGroupName(ig)),
		ImageID: fi.String(ig.Spec.Image),
		Ocean:   ocean, // link to Ocean
	}

	// Capacity.
	minSize, maxSize := b.buildCapacity(ig)
	ocean.MinSize = fi.Int64(fi.Int64Value(ocean.MinSize) + fi.Int64Value(minSize))
	ocean.MaxSize = fi.Int64(fi.Int64Value(ocean.MaxSize) + fi.Int64Value(maxSize))

	// User data.
	launchSpec.UserData, err = b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
	if err != nil {
		return fmt.Errorf("error building user data: %v", err)
	}

	// Instance profile.
	launchSpec.IAMInstanceProfile, err = b.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return fmt.Errorf("error building iam instance profile: %v", err)
	}

	// Security groups.
	launchSpec.SecurityGroups, err = b.buildSecurityGroups(c, ig)
	if err != nil {
		return fmt.Errorf("error building security groups: %v", err)
	}

	// Subnets.
	launchSpec.Subnets, err = b.buildSubnets(ig)
	if err != nil {
		return fmt.Errorf("error building subnets: %v", err)
	}

	// Tags.
	launchSpec.Tags, err = b.buildTags(ig)
	if err != nil {
		return fmt.Errorf("error building cloud tags: %v", err)
	}

	// Auto Scaler.
	autoScalerOpts, err := b.buildAutoScalerOpts(b.ClusterName(), ig)
	if err != nil {
		return fmt.Errorf("error building auto scaler options: %v", err)
	}
	if autoScalerOpts != nil { // remove unsupported options
		autoScalerOpts.Enabled = nil
		autoScalerOpts.ClusterID = nil
		autoScalerOpts.Cooldown = nil
		autoScalerOpts.Down = nil

		if autoScalerOpts.Labels != nil || autoScalerOpts.Taints != nil || autoScalerOpts.Headroom != nil {
			launchSpec.AutoScalerOpts = autoScalerOpts
		}
	}

	klog.V(4).Infof("Adding task: LaunchSpec/%s", fi.StringValue(launchSpec.Name))
	c.AddTask(launchSpec)

	return nil
}

func (b *InstanceGroupModelBuilder) buildSecurityGroups(c *fi.ModelBuilderContext,
	ig *kops.InstanceGroup) ([]*awstasks.SecurityGroup, error) {

	securityGroups := []*awstasks.SecurityGroup{
		b.LinkToSecurityGroup(ig.Spec.Role),
	}

	for _, id := range ig.Spec.AdditionalSecurityGroups {
		sg := &awstasks.SecurityGroup{
			Name:   fi.String(id),
			ID:     fi.String(id),
			Shared: fi.Bool(true),
		}
		if err := c.EnsureTask(sg); err != nil {
			return nil, err
		}
		securityGroups = append(securityGroups, sg)
	}

	return securityGroups, nil
}

func (b *InstanceGroupModelBuilder) buildSubnets(ig *kops.InstanceGroup) ([]*awstasks.Subnet, error) {
	subnets, err := b.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("could not determine any subnets for InstanceGroup %q; subnets was %s", ig.ObjectMeta.Name, ig.Spec.Subnets)
	}

	out := make([]*awstasks.Subnet, len(subnets))
	for i, subnet := range subnets {
		out[i] = b.LinkToSubnet(subnet)
	}

	return out, nil
}

func (b *InstanceGroupModelBuilder) buildPublicIpOpts(ig *kops.InstanceGroup) (*bool, error) {
	subnetMap := make(map[string]*kops.ClusterSubnetSpec)
	for i := range b.Cluster.Spec.Subnets {
		subnet := &b.Cluster.Spec.Subnets[i]
		subnetMap[subnet.Name] = subnet
	}

	var subnetType kops.SubnetType
	for _, subnetName := range ig.Spec.Subnets {
		subnet := subnetMap[subnetName]
		if subnet == nil {
			return nil, fmt.Errorf("InstanceGroup %q uses subnet %q that does not exist", ig.ObjectMeta.Name, subnetName)
		}
		if subnetType != "" && subnetType != subnet.Type {
			return nil, fmt.Errorf("InstanceGroup %q cannot be in subnets of different Type", ig.ObjectMeta.Name)
		}
		subnetType = subnet.Type
	}

	associatePublicIP := true
	switch subnetType {
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		associatePublicIP = true
		if ig.Spec.AssociatePublicIP != nil {
			associatePublicIP = *ig.Spec.AssociatePublicIP
		}
	case kops.SubnetTypePrivate:
		associatePublicIP = false
		if ig.Spec.AssociatePublicIP != nil {
			if *ig.Spec.AssociatePublicIP {
				klog.Warningf("Ignoring AssociatePublicIP=true for private InstanceGroup %q", ig.ObjectMeta.Name)
			}
		}
	default:
		return nil, fmt.Errorf("unknown subnet type %q", subnetType)
	}

	return fi.Bool(associatePublicIP), nil
}

func (b *InstanceGroupModelBuilder) buildRootVolumeOpts(ig *kops.InstanceGroup) (*spotinsttasks.RootVolumeOpts, error) {
	opts := &spotinsttasks.RootVolumeOpts{
		IOPS: ig.Spec.RootVolumeIops,
	}

	// Optimization.
	{
		if fi.BoolValue(ig.Spec.RootVolumeOptimization) {
			opts.Optimization = ig.Spec.RootVolumeOptimization
		}
	}

	// Size.
	{
		size := fi.Int32Value(ig.Spec.RootVolumeSize)
		if size == 0 {
			var err error
			size, err = defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
			if err != nil {
				return nil, err
			}
		}
		opts.Size = fi.Int32(size)
	}

	// Type.
	{
		typ := fi.StringValue(ig.Spec.RootVolumeType)
		if typ == "" {
			typ = "gp2"
		}
		opts.Type = fi.String(typ)
	}

	return opts, nil
}

func (b *InstanceGroupModelBuilder) buildCapacity(ig *kops.InstanceGroup) (*int64, *int64) {
	minSize := int32(1)
	if ig.Spec.MinSize != nil {
		minSize = fi.Int32Value(ig.Spec.MinSize)
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		minSize = 2
	}

	maxSize := int32(1)
	if ig.Spec.MaxSize != nil {
		maxSize = *ig.Spec.MaxSize
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		maxSize = 2
	}

	return fi.Int64(int64(minSize)), fi.Int64(int64(maxSize))
}

func (b *InstanceGroupModelBuilder) buildTags(ig *kops.InstanceGroup) (map[string]string, error) {
	tags, err := b.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (b *InstanceGroupModelBuilder) buildAutoScalerOpts(clusterID string, ig *kops.InstanceGroup) (*spotinsttasks.AutoScalerOpts, error) {
	opts := &spotinsttasks.AutoScalerOpts{
		ClusterID: fi.String(clusterID),
	}

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		return opts, nil

	case kops.InstanceGroupRoleBastion:
		return nil, nil
	}

	// Enable the auto scaler for Node instance groups.
	opts.Enabled = fi.Bool(true)

	// Parse instance group labels.
	var defaultNodeLabels bool
	for k, v := range ig.ObjectMeta.Labels {
		switch k {
		case InstanceGroupLabelAutoScalerDisabled:
			{
				v, err := parseBool(v)
				if err != nil {
					return nil, err
				}
				opts.Enabled = fi.Bool(!fi.BoolValue(v))
			}

		case InstanceGroupLabelAutoScalerDefaultNodeLabels:
			{
				v, err := parseBool(v)
				if err != nil {
					return nil, err
				}
				defaultNodeLabels = fi.BoolValue(v)
			}

		case InstanceGroupLabelAutoScalerCooldown:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				opts.Cooldown = fi.Int(int(fi.Int64Value(v)))
			}

		case InstanceGroupLabelAutoScalerHeadroomCPUPerUnit:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.Headroom == nil {
					opts.Headroom = new(spotinsttasks.AutoScalerHeadroomOpts)
				}
				opts.Headroom.CPUPerUnit = fi.Int(int(fi.Int64Value(v)))
			}

		case InstanceGroupLabelAutoScalerHeadroomGPUPerUnit:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.Headroom == nil {
					opts.Headroom = new(spotinsttasks.AutoScalerHeadroomOpts)
				}
				opts.Headroom.GPUPerUnit = fi.Int(int(fi.Int64Value(v)))
			}

		case InstanceGroupLabelAutoScalerHeadroomMemPerUnit:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.Headroom == nil {
					opts.Headroom = new(spotinsttasks.AutoScalerHeadroomOpts)
				}
				opts.Headroom.MemPerUnit = fi.Int(int(fi.Int64Value(v)))
			}

		case InstanceGroupLabelAutoScalerHeadroomNumOfUnits:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.Headroom == nil {
					opts.Headroom = new(spotinsttasks.AutoScalerHeadroomOpts)
				}
				opts.Headroom.NumOfUnits = fi.Int(int(fi.Int64Value(v)))
			}

		case InstanceGroupLabelAutoScalerScaleDownMaxPercentage:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.Down == nil {
					opts.Down = new(spotinsttasks.AutoScalerDownOpts)
				}
				opts.Down.MaxPercentage = fi.Int(int(fi.Int64Value(v)))
			}

		case InstanceGroupLabelAutoScalerScaleDownEvaluationPeriods:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.Down == nil {
					opts.Down = new(spotinsttasks.AutoScalerDownOpts)
				}
				opts.Down.EvaluationPeriods = fi.Int(int(fi.Int64Value(v)))
			}
		}
	}

	// Configure Elastigroup defaults to avoid state drifts.
	if !featureflag.SpotinstOcean.Enabled() {
		if opts.Cooldown == nil {
			opts.Cooldown = fi.Int(300)
		}
		if opts.Down != nil && opts.Down.EvaluationPeriods == nil {
			opts.Down.EvaluationPeriods = fi.Int(5)
		}
	}

	// Configure node labels.
	labels := make(map[string]string)
	for k, v := range ig.Spec.NodeLabels {
		if strings.HasPrefix(k, kops.NodeLabelInstanceGroup) && !defaultNodeLabels {
			continue
		}
		labels[k] = v
	}
	if len(labels) > 0 {
		opts.Labels = labels
	}

	// Configure node taints.
	taints, err := parseTaints(ig.Spec.Taints)
	if err != nil {
		return nil, err
	}
	if len(taints) > 0 {
		opts.Taints = taints
	}

	return opts, nil
}

func parseBool(str string) (*bool, error) {
	v, err := strconv.ParseBool(str)
	if err != nil {
		return nil, fmt.Errorf("unexpected boolean value: %q", str)
	}
	return &v, nil
}

func parseFloat(str string) (*float64, error) {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, fmt.Errorf("unexpected float value: %q", str)
	}
	return &v, nil
}

func parseInt(str string) (*int64, error) {
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unexpected integer value: %q", str)
	}
	return &v, nil
}

func parseTaints(taintSpecs []string) ([]*corev1.Taint, error) {
	var taints []*corev1.Taint

	for _, taintSpec := range taintSpecs {
		taint, err := parseTaint(taintSpec)
		if err != nil {
			return nil, err
		}
		taints = append(taints, taint)
	}

	return taints, nil
}

func parseTaint(taintSpec string) (*corev1.Taint, error) {
	var taint corev1.Taint

	parts := strings.Split(taintSpec, ":")
	switch len(parts) {
	case 1:
		taint.Key = parts[0]
	case 2:
		taint.Effect = corev1.TaintEffect(parts[1])
		partsKV := strings.Split(parts[0], "=")
		if len(partsKV) > 2 {
			return nil, fmt.Errorf("invalid taint spec: %v", taintSpec)
		}
		taint.Key = partsKV[0]
		if len(partsKV) == 2 {
			taint.Value = partsKV[1]
		}
	default:
		return nil, fmt.Errorf("invalid taint spec: %v", taintSpec)
	}

	return &taint, nil
}

func parseStringSlice(str string) ([]string, error) {
	v := strings.Split(str, ",")
	for i, s := range v {
		v[i] = strings.TrimSpace(s)
	}
	return v, nil
}

func defaultSpotPercentage(ig *kops.InstanceGroup) *float64 {
	var percentage float64

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleBastion:
		percentage = 0
	case kops.InstanceGroupRoleNode:
		percentage = 100
	}

	return &percentage
}

// HybridInstanceGroup indicates whether the instance group labeled with
// a metadata label `spotinst.io/hybrid` which means the Spotinst provider
// should be used to upon creation if the `SpotinstHybrid` feature flag is on.
func HybridInstanceGroup(ig *kops.InstanceGroup) bool {
	v, ok := ig.ObjectMeta.Labels[InstanceGroupLabelHybrid]
	if !ok {
		v = ig.ObjectMeta.Labels[InstanceGroupLabelManaged]
	}

	hybrid, _ := strconv.ParseBool(v)
	return hybrid
}
