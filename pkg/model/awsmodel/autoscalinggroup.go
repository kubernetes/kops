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
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"

	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// DefaultVolumeType is the default volume type
	DefaultVolumeType = ec2.VolumeTypeGp2
	// DefaultVolumeIops is the default volume iops
	DefaultVolumeIops = 100
	// DefaultVolumeDeleteOnTermination is the default volume behavior after instance termination
	DefaultVolumeDeleteOnTermination = true
)

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*AWSModelContext

	BootstrapScript   *model.BootstrapScript
	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

// Build is responsible for constructing the aws autoscaling group from the kops spec
func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		name := b.AutoscalingGroupName(ig)

		// @check if his instancegroup is backed by a fleet and override with a launch template
		task, err := func() (fi.Task, error) {
			switch UseLaunchTemplate(ig) {
			case true:
				return b.buildLaunchTemplateTask(c, name, ig)
			default:
				return b.buildLaunchConfigurationTask(c, name, ig)
			}
		}()
		if err != nil {
			return err
		}
		c.AddTask(task)

		// @step: now lets build the autoscaling group task
		tsk, err := b.buildAutoScalingGroupTask(c, name, ig)
		if err != nil {
			return err
		}
		switch UseLaunchTemplate(ig) {
		case true:
			tsk.LaunchTemplate = task.(*awstasks.LaunchTemplate)
		default:
			tsk.LaunchConfiguration = task.(*awstasks.LaunchConfiguration)
		}
		c.AddTask(tsk)

		// @step: add any external load balancer attachments
		if err := b.buildExternalLoadBalancerTasks(c, ig); err != nil {
			return err
		}
	}

	return nil
}

// buildLaunchTemplateTask is responsible for creating the template task into the aws model
func (b *AutoscalingGroupModelBuilder) buildLaunchTemplateTask(c *fi.ModelBuilderContext, name string, ig *kops.InstanceGroup) (*awstasks.LaunchTemplate, error) {
	lc, err := b.buildLaunchConfigurationTask(c, name, ig)
	if err != nil {
		return nil, err
	}

	tags, err := b.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, fmt.Errorf("error building cloud tags: %v", err)
	}

	// @TODO check if there any a better way of doing this .. initially I had a type LaunchTemplate which included
	// LaunchConfiguration as an anonymous field, bit given up the task dependency walker works this caused issues, due
	// to the creation of a implicit dependency
	lt := &awstasks.LaunchTemplate{
		Name:                   fi.String(name),
		Lifecycle:              b.Lifecycle,
		AssociatePublicIP:      lc.AssociatePublicIP,
		BlockDeviceMappings:    lc.BlockDeviceMappings,
		IAMInstanceProfile:     lc.IAMInstanceProfile,
		ImageID:                lc.ImageID,
		InstanceMonitoring:     lc.InstanceMonitoring,
		InstanceType:           lc.InstanceType,
		RootVolumeOptimization: lc.RootVolumeOptimization,
		RootVolumeSize:         lc.RootVolumeSize,
		RootVolumeIops:         lc.RootVolumeIops,
		RootVolumeType:         lc.RootVolumeType,
		SSHKey:                 lc.SSHKey,
		SecurityGroups:         lc.SecurityGroups,
		Tags:                   tags,
		Tenancy:                lc.Tenancy,
		UserData:               lc.UserData,
	}
	// When using a MixedInstances ASG, AWS requires the SpotPrice be defined on the ASG
	// rather than the LaunchTemplate or else it returns this error:
	//   You cannot use a launch template that is set to request Spot Instances (InstanceMarketOptions)
	//   when you configure an Auto Scaling group with a mixed instances policy.
	if ig.Spec.MixedInstancesPolicy == nil {
		lt.SpotPrice = lc.SpotPrice
	}
	return lt, nil
}

// buildLaunchConfigurationTask is responsible for building a launch configuration task into the model
func (b *AutoscalingGroupModelBuilder) buildLaunchConfigurationTask(c *fi.ModelBuilderContext, name string, ig *kops.InstanceGroup) (*awstasks.LaunchConfiguration, error) {
	// @step: lets add the root volume settings
	volumeSize, err := defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
	if err != nil {
		return nil, err
	}
	if fi.Int32Value(ig.Spec.RootVolumeSize) > 0 {
		volumeSize = fi.Int32Value(ig.Spec.RootVolumeSize)
	}

	volumeType := fi.StringValue(ig.Spec.RootVolumeType)
	if volumeType == "" {
		volumeType = DefaultVolumeType
	}

	rootVolumeDeleteOnTermination := DefaultVolumeDeleteOnTermination
	if ig.Spec.RootVolumeDeleteOnTermination != nil {
		rootVolumeDeleteOnTermination = fi.BoolValue(ig.Spec.RootVolumeDeleteOnTermination)
	}

	// @step: if required we add the override for the security group for this instancegroup
	sgLink := b.LinkToSecurityGroup(ig.Spec.Role)
	if ig.Spec.SecurityGroupOverride != nil {
		sgName := fmt.Sprintf("%v-%v", fi.StringValue(ig.Spec.SecurityGroupOverride), ig.Spec.Role)
		sgLink = &awstasks.SecurityGroup{
			ID:     ig.Spec.SecurityGroupOverride,
			Name:   &sgName,
			Shared: fi.Bool(true),
		}
	}

	// @step: add the iam instance profile
	link, err := b.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return nil, fmt.Errorf("unable to find iam profile link for instance group %q: %v", ig.ObjectMeta.Name, err)
	}

	t := &awstasks.LaunchConfiguration{
		Name:                          fi.String(name),
		Lifecycle:                     b.Lifecycle,
		IAMInstanceProfile:            link,
		ImageID:                       fi.String(ig.Spec.Image),
		InstanceMonitoring:            ig.Spec.DetailedInstanceMonitoring,
		InstanceType:                  fi.String(strings.Split(ig.Spec.MachineType, ",")[0]),
		RootVolumeDeleteOnTermination: fi.Bool(rootVolumeDeleteOnTermination),
		RootVolumeOptimization:        ig.Spec.RootVolumeOptimization,
		RootVolumeSize:                fi.Int64(int64(volumeSize)),
		RootVolumeType:                fi.String(volumeType),
		SecurityGroups:                []*awstasks.SecurityGroup{sgLink},
	}

	if volumeType == ec2.VolumeTypeIo1 {
		if fi.Int32Value(ig.Spec.RootVolumeIops) <= 0 {
			t.RootVolumeIops = fi.Int64(int64(DefaultVolumeIops))
		} else {
			t.RootVolumeIops = fi.Int64(int64(fi.Int32Value(ig.Spec.RootVolumeIops)))
		}
	}

	if ig.Spec.Tenancy != "" {
		t.Tenancy = fi.String(ig.Spec.Tenancy)
	}

	// @step: add any additional security groups to the instancegroup
	for _, id := range ig.Spec.AdditionalSecurityGroups {
		sgTask := &awstasks.SecurityGroup{
			ID:        fi.String(id),
			Lifecycle: b.SecurityLifecycle,
			Name:      fi.String(id),
			Shared:    fi.Bool(true),
		}
		if err := c.EnsureTask(sgTask); err != nil {
			return nil, err
		}
		t.SecurityGroups = append(t.SecurityGroups, sgTask)
	}

	// @step: add any additional block devices to the launch configuration
	for _, x := range ig.Spec.Volumes {
		if x.Type == "" {
			x.Type = DefaultVolumeType
		}
		if x.Type == ec2.VolumeTypeIo1 {
			if x.Iops == nil {
				x.Iops = fi.Int64(DefaultVolumeIops)
			}
		} else {
			x.Iops = nil
		}
		deleteOnTermination := DefaultVolumeDeleteOnTermination
		if x.DeleteOnTermination != nil {
			deleteOnTermination = fi.BoolValue(x.DeleteOnTermination)
		}
		t.BlockDeviceMappings = append(t.BlockDeviceMappings, &awstasks.BlockDeviceMapping{
			DeviceName:             fi.String(x.Device),
			EbsDeleteOnTermination: fi.Bool(deleteOnTermination),
			EbsEncrypted:           x.Encrypted,
			EbsVolumeIops:          x.Iops,
			EbsVolumeSize:          fi.Int64(x.Size),
			EbsVolumeType:          fi.String(x.Type),
		})
	}

	if b.AWSModelContext.UseSSHKey() {
		if t.SSHKey, err = b.LinkToSSHKey(); err != nil {
			return nil, err
		}
	}

	// @step: add the instancegroup userdata
	if t.UserData, err = b.BootstrapScript.ResourceNodeUp(ig, b.Cluster); err != nil {
		return nil, err
	}

	// @step: set up instance spot pricing
	if fi.StringValue(ig.Spec.MaxPrice) != "" {
		spotPrice := fi.StringValue(ig.Spec.MaxPrice)
		t.SpotPrice = spotPrice
	}

	// @step: check the subnets are ok and pull together an array for us
	subnets, err := b.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}

	// @step: check if we can add an public ip to this subnet
	switch subnets[0].Type {
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		t.AssociatePublicIP = fi.Bool(true)
		if ig.Spec.AssociatePublicIP != nil {
			t.AssociatePublicIP = ig.Spec.AssociatePublicIP
		}
	case kops.SubnetTypePrivate:
		t.AssociatePublicIP = fi.Bool(false)
	}

	return t, nil
}

// buildAutoscalingGroupTask is responsible for building the autoscaling task into the model
func (b *AutoscalingGroupModelBuilder) buildAutoScalingGroupTask(c *fi.ModelBuilderContext, name string, ig *kops.InstanceGroup) (*awstasks.AutoscalingGroup, error) {

	t := &awstasks.AutoscalingGroup{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,

		Granularity: fi.String("1Minute"),
		Metrics: []string{
			"GroupDesiredCapacity",
			"GroupInServiceInstances",
			"GroupMaxSize",
			"GroupMinSize",
			"GroupPendingInstances",
			"GroupStandbyInstances",
			"GroupTerminatingInstances",
			"GroupTotalInstances",
		},
	}

	minSize := int32(1)
	maxSize := int32(1)
	if ig.Spec.MinSize != nil {
		minSize = fi.Int32Value(ig.Spec.MinSize)
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		minSize = 2
	}
	if ig.Spec.MaxSize != nil {
		maxSize = *ig.Spec.MaxSize
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		maxSize = 2
	}

	t.MinSize = fi.Int64(int64(minSize))
	t.MaxSize = fi.Int64(int64(maxSize))

	subnets, err := b.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("could not determine any subnets for InstanceGroup %q; subnets was %s", ig.ObjectMeta.Name, ig.Spec.Subnets)
	}
	for _, subnet := range subnets {
		t.Subnets = append(t.Subnets, b.LinkToSubnet(subnet))
	}

	tags, err := b.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, fmt.Errorf("error building cloud tags: %v", err)
	}
	t.Tags = tags

	processes := []string{}
	processes = append(processes, ig.Spec.SuspendProcesses...)
	t.SuspendProcesses = &processes

	t.InstanceProtection = ig.Spec.InstanceProtection

	// @step: are we using a mixed instance policy
	if ig.Spec.MixedInstancesPolicy != nil {
		spec := ig.Spec.MixedInstancesPolicy

		t.MixedInstanceOverrides = spec.Instances
		t.MixedOnDemandAboveBase = spec.OnDemandAboveBase
		t.MixedOnDemandAllocationStrategy = spec.OnDemandAllocationStrategy
		t.MixedOnDemandBase = spec.OnDemandBase
		t.MixedSpotAllocationStrategy = spec.SpotAllocationStrategy
		t.MixedSpotInstancePools = spec.SpotInstancePools
		t.MixedSpotMaxPrice = ig.Spec.MaxPrice
	}

	return t, nil
}

// buildExternlLoadBalancerTasks is responsible for adding any ELB attachment tasks to the model
func (b *AutoscalingGroupModelBuilder) buildExternalLoadBalancerTasks(c *fi.ModelBuilderContext, ig *kops.InstanceGroup) error {
	for _, x := range ig.Spec.ExternalLoadBalancers {
		if x.LoadBalancerName != nil {
			c.AddTask(&awstasks.ExternalLoadBalancerAttachment{
				Name:             fi.String("extlb-" + *x.LoadBalancerName + "-" + ig.Name),
				Lifecycle:        b.Lifecycle,
				LoadBalancerName: *x.LoadBalancerName,
				AutoscalingGroup: b.LinkToAutoscalingGroup(ig),
			})
		}

		if x.TargetGroupARN != nil {
			c.AddTask(&awstasks.ExternalTargetGroupAttachment{
				Name:             fi.String("exttg-" + *x.TargetGroupARN + "-" + ig.Name),
				Lifecycle:        b.Lifecycle,
				TargetGroupARN:   *x.TargetGroupARN,
				AutoscalingGroup: b.LinkToAutoscalingGroup(ig),
			})
		}
	}

	return nil
}
