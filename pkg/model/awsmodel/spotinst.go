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

package awsmodel

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks"
)

const (
	// SpotInstanceGroupLabelHybrid is the metadata label used on the instance group
	// to specify that the Spotinst provider should be used to upon creation.
	SpotInstanceGroupLabelHybrid  = "spotinst.io/hybrid"
	SpotInstanceGroupLabelManaged = "spotinst.io/managed" // for backward compatibility

	// SpotInstanceGroupLabelSpotPercentage is the metadata label used on the
	// instance group to specify the percentage of Spot instances that
	// should spin up from the target capacity.
	SpotInstanceGroupLabelSpotPercentage = "spotinst.io/spot-percentage"

	// SpotInstanceGroupLabelOrientation is the metadata label used on the
	// instance group to specify which orientation should be used.
	SpotInstanceGroupLabelOrientation = "spotinst.io/orientation"

	// SpotInstanceGroupLabelUtilizeReservedInstances is the metadata label used
	// on the instance group to specify whether reserved instances should be
	// utilized.
	SpotInstanceGroupLabelUtilizeReservedInstances = "spotinst.io/utilize-reserved-instances"

	// SpotInstanceGroupLabelUtilizeCommitments is the metadata label used
	// on the instance group to specify whether commitments should be utilized.
	SpotInstanceGroupLabelUtilizeCommitments = "spotinst.io/utilize-commitments"

	// SpotInstanceGroupLabelFallbackToOnDemand is the metadata label used on the
	// instance group to specify whether fallback to on-demand instances should
	// be enabled.
	SpotInstanceGroupLabelFallbackToOnDemand = "spotinst.io/fallback-to-ondemand"

	// SpotInstanceGroupLabelDrainingTimeout is the metadata label used on the
	// instance group to specify a period of time, in seconds, after a node
	// is marked for termination during which on running pods remains active.
	SpotInstanceGroupLabelDrainingTimeout = "spotinst.io/draining-timeout"

	// SpotInstanceGroupLabelGracePeriod is the metadata label used on the
	// instance group to specify a period of time, in seconds, that Ocean
	// should wait before applying instance health checks.
	SpotInstanceGroupLabelGracePeriod = "spotinst.io/grace-period"

	// SpotInstanceGroupLabelHealthCheckType is the metadata label used on the
	// instance group to specify the type of the health check that should be used.
	SpotInstanceGroupLabelHealthCheckType = "spotinst.io/health-check-type"

	// SpotInstanceGroupLabelOceanDefaultLaunchSpec is the metadata label used on the
	// instance group to specify whether to use the SpotInstanceGroup's spec as the default
	// Launch Spec for the Ocean cluster.
	SpotInstanceGroupLabelOceanDefaultLaunchSpec = "spotinst.io/ocean-default-launchspec"

	// SpotInstanceGroupLabelOceanInstanceTypes[White|Black]list are the metadata labels
	// used on the instance group to specify whether to whitelist or blacklist
	// specific instance types.
	SpotInstanceGroupLabelOceanInstanceTypesWhitelist = "spotinst.io/ocean-instance-types-whitelist"
	SpotInstanceGroupLabelOceanInstanceTypesBlacklist = "spotinst.io/ocean-instance-types-blacklist"
	SpotInstanceGroupLabelOceanInstanceTypes          = "spotinst.io/ocean-instance-types" // launchspec

	// SpotInstanceGroupLabelAutoScalerDisabled is the metadata label used on the
	// instance group to specify whether the auto scaler should be enabled.
	SpotInstanceGroupLabelAutoScalerDisabled = "spotinst.io/autoscaler-disabled"

	// SpotInstanceGroupLabelAutoScalerDefaultNodeLabels is the metadata label used on the
	// instance group to specify whether default node labels should be set for
	// the auto scaler.
	SpotInstanceGroupLabelAutoScalerDefaultNodeLabels = "spotinst.io/autoscaler-default-node-labels"

	// SpotInstanceGroupLabelAutoScalerAuto* are the metadata labels used on the
	// instance group to specify whether headroom resources should be
	// automatically configured and optimized.
	SpotInstanceGroupLabelAutoScalerAutoConfig             = "spotinst.io/autoscaler-auto-config"
	SpotInstanceGroupLabelAutoScalerAutoHeadroomPercentage = "spotinst.io/autoscaler-auto-headroom-percentage"

	// SpotInstanceGroupLabelAutoScalerHeadroom* are the metadata labels used on the
	// instance group to specify the headroom configuration used by the auto scaler.
	SpotInstanceGroupLabelAutoScalerHeadroomCPUPerUnit = "spotinst.io/autoscaler-headroom-cpu-per-unit"
	SpotInstanceGroupLabelAutoScalerHeadroomGPUPerUnit = "spotinst.io/autoscaler-headroom-gpu-per-unit"
	SpotInstanceGroupLabelAutoScalerHeadroomMemPerUnit = "spotinst.io/autoscaler-headroom-mem-per-unit"
	SpotInstanceGroupLabelAutoScalerHeadroomNumOfUnits = "spotinst.io/autoscaler-headroom-num-of-units"

	// SpotInstanceGroupLabelAutoScalerCooldown is the metadata label used on the
	// instance group to specify the cooldown period (in seconds) for scaling actions.
	SpotInstanceGroupLabelAutoScalerCooldown = "spotinst.io/autoscaler-cooldown"

	// SpotInstanceGroupLabelAutoScalerScaleDown* are the metadata labels used on the
	// instance group to specify the scale down configuration used by the auto scaler.
	SpotInstanceGroupLabelAutoScalerScaleDownMaxPercentage     = "spotinst.io/autoscaler-scale-down-max-percentage"
	SpotInstanceGroupLabelAutoScalerScaleDownEvaluationPeriods = "spotinst.io/autoscaler-scale-down-evaluation-periods"

	// SpotInstanceGroupLabelAutoScalerResourceLimits* are the metadata labels used on the
	// instance group to specify the resource limits configuration used by the auto scaler.
	SpotInstanceGroupLabelAutoScalerResourceLimitsMaxVCPU   = "spotinst.io/autoscaler-resource-limits-max-vcpu"
	SpotInstanceGroupLabelAutoScalerResourceLimitsMaxMemory = "spotinst.io/autoscaler-resource-limits-max-memory"

	// InstanceGroupLabelRestrictScaleDown is the metadata label used on the
	// instance group to specify whether the scale-down activities should be restricted.
	SpotInstanceGroupLabelRestrictScaleDown = "spotinst.io/restrict-scale-down"
)

// SpotInstanceGroupModelBuilder configures SpotInstanceGroup objects
type SpotInstanceGroupModelBuilder struct {
	*AWSModelContext
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
	SecurityLifecycle      fi.Lifecycle
}

var _ fi.ModelBuilder = &SpotInstanceGroupModelBuilder{}

func (b *SpotInstanceGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var nodeSpotInstanceGroups []*kops.InstanceGroup
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
				nodeSpotInstanceGroups = append(nodeSpotInstanceGroups, ig)
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

	if len(nodeSpotInstanceGroups) > 0 {
		if err = b.buildOcean(c, nodeSpotInstanceGroups...); err != nil {
			return fmt.Errorf("spotinst: error building ocean: %v", err)
		}
	}

	return nil
}

func (b *SpotInstanceGroupModelBuilder) buildElastigroup(c *fi.ModelBuilderContext, ig *kops.InstanceGroup) (err error) {
	klog.V(4).Infof("Building instance group as Elastigroup: %q", b.AutoscalingGroupName(ig))
	group := &spotinsttasks.Elastigroup{
		Lifecycle:            b.Lifecycle,
		Name:                 fi.String(b.AutoscalingGroupName(ig)),
		Region:               fi.String(b.Region),
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
		case SpotInstanceGroupLabelSpotPercentage:
			group.SpotPercentage, err = parseFloat(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelOrientation:
			group.Orientation = fi.String(v)

		case SpotInstanceGroupLabelUtilizeReservedInstances:
			group.UtilizeReservedInstances, err = parseBool(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelUtilizeCommitments:
			group.UtilizeCommitments, err = parseBool(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelFallbackToOnDemand:
			group.FallbackToOnDemand, err = parseBool(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelDrainingTimeout:
			group.DrainingTimeout, err = parseInt(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelHealthCheckType:
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

	// Load balancers.
	group.LoadBalancers, group.TargetGroups, err = b.buildLoadBalancers(c, ig)
	if err != nil {
		return fmt.Errorf("error building load balancers: %v", err)
	}

	// User data.
	group.UserData, err = b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
	if err != nil {
		return fmt.Errorf("error building user data: %v", err)
	}

	// Public IP.
	group.AssociatePublicIPAddress, err = b.buildPublicIPOpts(ig)
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

func (b *SpotInstanceGroupModelBuilder) buildOcean(c *fi.ModelBuilderContext, igs ...*kops.InstanceGroup) (err error) {
	klog.V(4).Infof("Building instance group as Ocean: %q", "nodes."+b.ClusterName())
	ocean := &spotinsttasks.Ocean{
		Lifecycle: b.Lifecycle,
		Name:      fi.String("nodes." + b.ClusterName()),
	}

	if featureflag.SpotinstOceanTemplate.Enabled() {
		ocean.UseAsTemplateOnly = fi.Bool(true)
	}

	ig := igs[0].DeepCopy()
	if len(igs) > 1 {
		for _, g := range igs {
			for k, v := range g.ObjectMeta.Labels {
				if k == SpotInstanceGroupLabelOceanDefaultLaunchSpec {
					defaultLaunchSpec, err := parseBool(v)
					if err != nil {
						continue
					}
					if fi.BoolValue(defaultLaunchSpec) {
						if ig != nil {
							return fmt.Errorf("unable to detect default launch spec: "+
								"multiple instance groups labeled with `%s: \"true\"`",
								SpotInstanceGroupLabelOceanDefaultLaunchSpec)
						}
						ig = g.DeepCopy()
						break
					}
				}
			}
		}

		klog.V(4).Infof("Detected default launch spec: %q", b.AutoscalingGroupName(ig))
	}

	// Image.
	ocean.ImageID = fi.String(ig.Spec.Image)

	// Strategy and instance types.
	for k, v := range ig.ObjectMeta.Labels {
		switch k {
		case SpotInstanceGroupLabelUtilizeReservedInstances:
			ocean.UtilizeReservedInstances, err = parseBool(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelUtilizeCommitments:
			ocean.UtilizeCommitments, err = parseBool(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelFallbackToOnDemand:
			ocean.FallbackToOnDemand, err = parseBool(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelGracePeriod:
			ocean.GracePeriod, err = parseInt(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelDrainingTimeout:
			ocean.DrainingTimeout, err = parseInt(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelOceanInstanceTypesWhitelist:
			ocean.InstanceTypesWhitelist, err = parseStringSlice(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelOceanInstanceTypesBlacklist:
			ocean.InstanceTypesBlacklist, err = parseStringSlice(v)
			if err != nil {
				return err
			}
		}
	}

	// Monitoring.
	ocean.Monitoring = ig.Spec.DetailedInstanceMonitoring

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

	// Subnets.
	ocean.Subnets, err = b.buildSubnets(ig)
	if err != nil {
		return fmt.Errorf("error building subnets: %v", err)
	}

	// Auto Scaler.
	ocean.AutoScalerOpts, err = b.buildAutoScalerOpts(b.ClusterName(), ig)
	if err != nil {
		return fmt.Errorf("error building auto scaler options: %v", err)
	}
	if ocean.AutoScalerOpts != nil { // remove unsupported options
		ocean.AutoScalerOpts.Labels = nil
		ocean.AutoScalerOpts.Taints = nil
		ocean.AutoScalerOpts.Headroom = nil
	}

	if !fi.BoolValue(ocean.UseAsTemplateOnly) {
		// Capacity.
		ocean.MinSize = fi.Int64(0)
		ocean.MaxSize = fi.Int64(0)

		// User data.
		ocean.UserData, err = b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
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

		// Public IP.
		ocean.AssociatePublicIPAddress, err = b.buildPublicIPOpts(ig)
		if err != nil {
			return fmt.Errorf("error building public ip options: %v", err)
		}

		// Tags.
		ocean.Tags, err = b.buildTags(ig)
		if err != nil {
			return fmt.Errorf("error building cloud tags: %v", err)
		}
	}

	// Create a Launch Spec for each instance group.
	for _, g := range igs {
		if err := b.buildLaunchSpec(c, g, ig, ocean); err != nil {
			return fmt.Errorf("error building launch spec: %v", err)
		}
	}

	klog.V(4).Infof("Adding task: Ocean/%s", fi.StringValue(ocean.Name))
	c.AddTask(ocean)

	return nil
}

func (b *SpotInstanceGroupModelBuilder) buildLaunchSpec(c *fi.ModelBuilderContext,
	ig, igOcean *kops.InstanceGroup, ocean *spotinsttasks.Ocean) (err error) {
	klog.V(4).Infof("Building instance group as LaunchSpec: %q", b.AutoscalingGroupName(ig))
	launchSpec := &spotinsttasks.LaunchSpec{
		Name:      fi.String(b.AutoscalingGroupName(ig)),
		Lifecycle: b.Lifecycle,
		ImageID:   fi.String(ig.Spec.Image),
		Ocean:     ocean, // link to Ocean
	}

	// Instance types and strategy.
	for k, v := range ig.ObjectMeta.Labels {
		switch k {
		case SpotInstanceGroupLabelOceanInstanceTypesWhitelist, SpotInstanceGroupLabelOceanInstanceTypes:
			launchSpec.InstanceTypes, err = parseStringSlice(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelSpotPercentage:
			launchSpec.SpotPercentage, err = parseInt(v)
			if err != nil {
				return err
			}

		case SpotInstanceGroupLabelRestrictScaleDown:
			launchSpec.RestrictScaleDown, err = parseBool(v)
			if err != nil {
				return err
			}
		}
	}

	policy := ig.Spec.MixedInstancesPolicy
	if len(launchSpec.InstanceTypes) == 0 && policy != nil && len(policy.Instances) > 0 {
		launchSpec.InstanceTypes = policy.Instances
	}

	// Capacity.
	minSize, maxSize := b.buildCapacity(ig)
	if fi.BoolValue(ocean.UseAsTemplateOnly) {
		launchSpec.MinSize = minSize
		launchSpec.MaxSize = maxSize
	} else {
		ocean.MinSize = fi.Int64(fi.Int64Value(ocean.MinSize) + fi.Int64Value(minSize))
		ocean.MaxSize = fi.Int64(fi.Int64Value(ocean.MaxSize) + fi.Int64Value(maxSize))
	}

	// User data.
	if ig.Name == igOcean.Name && !featureflag.SpotinstOceanTemplate.Enabled() {
		launchSpec.UserData = ocean.UserData
	} else {
		launchSpec.UserData, err = b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return fmt.Errorf("error building user data: %v", err)
		}
	}

	// Instance profile.
	launchSpec.IAMInstanceProfile, err = b.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return fmt.Errorf("error building iam instance profile: %v", err)
	}

	// Root volume.
	rootVolumeOpts, err := b.buildRootVolumeOpts(ig)
	if err != nil {
		return fmt.Errorf("error building root volume options: %v", err)
	}
	if rootVolumeOpts != nil { // remove unsupported options
		launchSpec.RootVolumeOpts = rootVolumeOpts
		launchSpec.RootVolumeOpts.Optimization = nil
	}

	// Public IP.
	launchSpec.AssociatePublicIPAddress, err = b.buildPublicIPOpts(ig)
	if err != nil {
		return fmt.Errorf("error building public ip options: %v", err)
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
		autoScalerOpts.AutoConfig = nil
		autoScalerOpts.AutoHeadroomPercentage = nil
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

func (b *SpotInstanceGroupModelBuilder) buildSecurityGroups(c *fi.ModelBuilderContext,
	ig *kops.InstanceGroup) ([]*awstasks.SecurityGroup, error) {
	securityGroups := []*awstasks.SecurityGroup{
		b.LinkToSecurityGroup(ig.Spec.Role),
	}

	for _, id := range ig.Spec.AdditionalSecurityGroups {
		sg := &awstasks.SecurityGroup{
			Lifecycle: b.SecurityLifecycle,
			ID:        fi.String(id),
			Name:      fi.String(id),
			Shared:    fi.Bool(true),
		}
		if err := c.EnsureTask(sg); err != nil {
			return nil, err
		}
		securityGroups = append(securityGroups, sg)
	}

	return securityGroups, nil
}

func (b *SpotInstanceGroupModelBuilder) buildSubnets(ig *kops.InstanceGroup) ([]*awstasks.Subnet, error) {
	subnets, err := b.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("could not determine any subnets for SpotInstanceGroup %q; subnets was %s", ig.ObjectMeta.Name, ig.Spec.Subnets)
	}

	out := make([]*awstasks.Subnet, len(subnets))
	for i, subnet := range subnets {
		out[i] = b.LinkToSubnet(subnet)
	}

	return out, nil
}

func (b *SpotInstanceGroupModelBuilder) buildPublicIPOpts(ig *kops.InstanceGroup) (*bool, error) {
	subnetMap := make(map[string]*kops.ClusterSubnetSpec)
	for i := range b.Cluster.Spec.Subnets {
		subnet := &b.Cluster.Spec.Subnets[i]
		subnetMap[subnet.Name] = subnet
	}

	var subnetType kops.SubnetType
	for _, subnetName := range ig.Spec.Subnets {
		subnet := subnetMap[subnetName]
		if subnet == nil {
			return nil, fmt.Errorf("SpotInstanceGroup %q uses subnet %q that does not exist", ig.ObjectMeta.Name, subnetName)
		}
		if subnetType != "" && subnetType != subnet.Type {
			return nil, fmt.Errorf("SpotInstanceGroup %q cannot be in subnets of different Type", ig.ObjectMeta.Name)
		}
		subnetType = subnet.Type
	}

	var associatePublicIP bool
	switch subnetType {
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		associatePublicIP = true
		if ig.Spec.AssociatePublicIP != nil {
			associatePublicIP = *ig.Spec.AssociatePublicIP
		}
	case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
		associatePublicIP = false
		if ig.Spec.AssociatePublicIP != nil {
			if *ig.Spec.AssociatePublicIP {
				klog.Warningf("Ignoring AssociatePublicIPAddress=true for private SpotInstanceGroup %q", ig.ObjectMeta.Name)
			}
		}
	default:
		return nil, fmt.Errorf("unknown subnet type %q", subnetType)
	}

	return fi.Bool(associatePublicIP), nil
}

func (b *SpotInstanceGroupModelBuilder) buildRootVolumeOpts(ig *kops.InstanceGroup) (*spotinsttasks.RootVolumeOpts, error) {
	opts := new(spotinsttasks.RootVolumeOpts)

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
		opts.Size = fi.Int64(int64(size))
	}

	// Type.
	{
		typ := fi.StringValue(ig.Spec.RootVolumeType)
		if typ == "" {
			typ = "gp2"
		}
		opts.Type = fi.String(typ)
	}

	// IOPS.
	{
		iops := fi.Int32Value(ig.Spec.RootVolumeIOPS)
		if iops > 0 {
			opts.IOPS = fi.Int64(int64(iops))
		}
	}

	// Throughput.
	{
		throughput := fi.Int32Value(ig.Spec.RootVolumeThroughput)
		if throughput > 0 {
			opts.Throughput = fi.Int64(int64(throughput))
		}
	}

	return opts, nil
}

func (b *SpotInstanceGroupModelBuilder) buildCapacity(ig *kops.InstanceGroup) (*int64, *int64) {
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

func (b *SpotInstanceGroupModelBuilder) buildLoadBalancers(c *fi.ModelBuilderContext,
	ig *kops.InstanceGroup) ([]*awstasks.ClassicLoadBalancer, []*awstasks.TargetGroup, error) {
	var loadBalancers []*awstasks.ClassicLoadBalancer
	var targetGroups []*awstasks.TargetGroup

	if b.UseLoadBalancerForAPI() && ig.HasAPIServer() {
		if b.UseNetworkLoadBalancer() {
			targetGroups = append(targetGroups, b.LinkToTargetGroup("tcp"))
			if b.Cluster.Spec.API.LoadBalancer.SSLCertificate != "" {
				targetGroups = append(targetGroups, b.LinkToTargetGroup("tls"))
			}
		} else {
			loadBalancers = append(loadBalancers, b.LinkToCLB("api"))
		}
	}

	if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		loadBalancers = append(loadBalancers, b.LinkToCLB("bastion"))
	}

	for _, extLB := range ig.Spec.ExternalLoadBalancers {
		if extLB.LoadBalancerName != nil {
			lb := &awstasks.ClassicLoadBalancer{
				Name:             extLB.LoadBalancerName,
				LoadBalancerName: extLB.LoadBalancerName,
				Shared:           fi.Bool(true),
			}
			loadBalancers = append(loadBalancers, lb)
			c.EnsureTask(lb)
		}
		if extLB.TargetGroupARN != nil {
			targetGroupName, err := awsup.GetTargetGroupNameFromARN(fi.StringValue(extLB.TargetGroupARN))
			if err != nil {
				return nil, nil, err
			}
			tg := &awstasks.TargetGroup{
				Name:   fi.String(ig.Name + "-" + targetGroupName),
				ARN:    extLB.TargetGroupARN,
				Shared: fi.Bool(true),
			}
			targetGroups = append(targetGroups, tg)
			c.AddTask(tg)
		}
	}

	return loadBalancers, targetGroups, nil
}

func (b *SpotInstanceGroupModelBuilder) buildTags(ig *kops.InstanceGroup) (map[string]string, error) {
	tags, err := b.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (b *SpotInstanceGroupModelBuilder) buildAutoScalerOpts(clusterID string, ig *kops.InstanceGroup) (*spotinsttasks.AutoScalerOpts, error) {
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
	opts.AutoConfig = fi.Bool(true)

	// Parse instance group labels.
	var defaultNodeLabels bool
	for k, v := range ig.ObjectMeta.Labels {
		switch k {
		case SpotInstanceGroupLabelAutoScalerDisabled:
			{
				v, err := parseBool(v)
				if err != nil {
					return nil, err
				}
				opts.Enabled = fi.Bool(!fi.BoolValue(v))
			}

		case SpotInstanceGroupLabelAutoScalerDefaultNodeLabels:
			{
				v, err := parseBool(v)
				if err != nil {
					return nil, err
				}
				defaultNodeLabels = fi.BoolValue(v)
			}

		case SpotInstanceGroupLabelAutoScalerCooldown:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				opts.Cooldown = fi.Int(int(fi.Int64Value(v)))
			}

		case SpotInstanceGroupLabelAutoScalerAutoConfig:
			{
				v, err := parseBool(v)
				if err != nil {
					return nil, err
				}
				opts.AutoConfig = v
			}

		case SpotInstanceGroupLabelAutoScalerAutoHeadroomPercentage:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				opts.AutoHeadroomPercentage = fi.Int(int(fi.Int64Value(v)))
			}

		case SpotInstanceGroupLabelAutoScalerHeadroomCPUPerUnit:
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

		case SpotInstanceGroupLabelAutoScalerHeadroomGPUPerUnit:
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

		case SpotInstanceGroupLabelAutoScalerHeadroomMemPerUnit:
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

		case SpotInstanceGroupLabelAutoScalerHeadroomNumOfUnits:
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

		case SpotInstanceGroupLabelAutoScalerScaleDownMaxPercentage:
			{
				v, err := parseFloat(v)
				if err != nil {
					return nil, err
				}
				if opts.Down == nil {
					opts.Down = new(spotinsttasks.AutoScalerDownOpts)
				}
				opts.Down.MaxPercentage = v
			}

		case SpotInstanceGroupLabelAutoScalerScaleDownEvaluationPeriods:
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

		case SpotInstanceGroupLabelAutoScalerResourceLimitsMaxVCPU:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.ResourceLimits == nil {
					opts.ResourceLimits = new(spotinsttasks.AutoScalerResourceLimitsOpts)
				}
				opts.ResourceLimits.MaxVCPU = fi.Int(int(fi.Int64Value(v)))
			}

		case SpotInstanceGroupLabelAutoScalerResourceLimitsMaxMemory:
			{
				v, err := parseInt(v)
				if err != nil {
					return nil, err
				}
				if opts.ResourceLimits == nil {
					opts.ResourceLimits = new(spotinsttasks.AutoScalerResourceLimitsOpts)
				}
				opts.ResourceLimits.MaxMemory = fi.Int(int(fi.Int64Value(v)))
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
	v, ok := ig.ObjectMeta.Labels[SpotInstanceGroupLabelHybrid]
	if !ok {
		v = ig.ObjectMeta.Labels[SpotInstanceGroupLabelManaged]
	}

	hybrid, _ := strconv.ParseBool(v)
	return hybrid
}
