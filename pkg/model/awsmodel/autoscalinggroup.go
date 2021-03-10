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
	"sort"

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

const (
	// DefaultHTTPTokens is the default state of token usage for the instance metadata requests
	DefaultHTTPTokens = ec2.LaunchTemplateHttpTokensStateOptional
	// DefaultHTTPPutResponseHopLimit is the default HTTP PUT response hop limit for instance metadata requests
	DefaultHTTPPutResponseHopLimit = 1
	// DefaultVolumeType is the default volume type
	DefaultVolumeType = ec2.VolumeTypeGp3
	// DefaultVolumeIonIops is the default volume IOPS when volume type is io1 or io2
	DefaultVolumeIonIops = 100
	// DefaultVolumeGp3Iops is the default volume IOPS when volume type is gp3
	DefaultVolumeGp3Iops = 3000
	// DefaultVolumeGp3Throughput is the default volume throughput when volume type is gp3
	DefaultVolumeGp3Throughput = 125
	// DefaultVolumeDeleteOnTermination is the default volume behavior after instance termination
	DefaultVolumeDeleteOnTermination = true
	// DefaultVolumeEncryption is the default volume encryption behavior
	DefaultVolumeEncryption = true
)

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*AWSModelContext

	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              *fi.Lifecycle
	SecurityLifecycle      *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

// Build is responsible for constructing the aws autoscaling group from the kops spec
func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		name := b.AutoscalingGroupName(ig)

		if featureflag.SpotinstHybrid.Enabled() {
			if HybridInstanceGroup(ig) {
				klog.V(2).Infof("Skipping instance group: %q", name)
				continue
			}
		}

		task, err := b.buildLaunchTemplateTask(c, name, ig)
		if err != nil {
			return err
		}
		c.AddTask(task)

		// @step: now lets build the autoscaling group task
		tsk, err := b.buildAutoScalingGroupTask(c, name, ig)
		if err != nil {
			return err
		}
		tsk.LaunchTemplate = task
		c.AddTask(tsk)
	}

	return nil
}

// buildLaunchTemplateTask is responsible for creating the template task into the aws model
func (b *AutoscalingGroupModelBuilder) buildLaunchTemplateTask(c *fi.ModelBuilderContext, name string, ig *kops.InstanceGroup) (*awstasks.LaunchTemplate, error) {
	t := &awstasks.LaunchTemplate{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,

		CPUCredits:                   fi.String(""),
		HTTPPutResponseHopLimit:      fi.Int64(DefaultHTTPPutResponseHopLimit),
		HTTPTokens:                   fi.String(DefaultHTTPTokens),
		ImageID:                      fi.String(ig.Spec.Image),
		InstanceInterruptionBehavior: ig.Spec.InstanceInterruptionBehavior,
		InstanceMonitoring:           ig.Spec.DetailedInstanceMonitoring,
		InstanceType:                 fi.String(ig.Spec.MachineType),
		RootVolumeEncryption:         fi.Bool(DefaultVolumeEncryption),
		RootVolumeKmsKey:             fi.String(""),
		RootVolumeOptimization:       ig.Spec.RootVolumeOptimization,
		RootVolumeType:               fi.String(DefaultVolumeType),
		SecurityGroups:               []*awstasks.SecurityGroup{},
		SpotDurationInMinutes:        ig.Spec.SpotDurationInMinutes,
		SpotPrice:                    fi.String(""),
	}

	// Define generic error placeholder
	var err error

	// Set the link to the IAM instance profile
	if t.IAMInstanceProfile, err = b.LinkToIAMInstanceProfile(ig); err != nil {
		return nil, fmt.Errorf("unable to find IAM profile link for instance group %q: %w", ig.ObjectMeta.Name, err)
	}

	// Set the tags
	if t.Tags, err = b.CloudTagsForInstanceGroup(ig); err != nil {
		return nil, fmt.Errorf("error building cloud tags: %v", err)
	}

	// Set the user data
	if t.UserData, err = b.BootstrapScriptBuilder.ResourceNodeUp(c, ig); err != nil {
		return nil, err
	}

	// Set the state of token usage for the instance metadata requests
	if ig.Spec.InstanceMetadata != nil && ig.Spec.InstanceMetadata.HTTPTokens != nil {
		t.HTTPTokens = ig.Spec.InstanceMetadata.HTTPTokens
	}

	// Set the HTTP PUT response hop limit for instance metadata requests
	if ig.Spec.InstanceMetadata != nil && ig.Spec.InstanceMetadata.HTTPPutResponseHopLimit != nil {
		t.HTTPPutResponseHopLimit = ig.Spec.InstanceMetadata.HTTPPutResponseHopLimit
	}

	// Set the main security group task
	if ig.Spec.SecurityGroupOverride == nil {
		sgTask := b.LinkToSecurityGroup(ig.Spec.Role)
		klog.Infof("Main SG: %q", fi.StringValue(sgTask.Name))
		t.SecurityGroups = append(t.SecurityGroups, sgTask)
	} else {
		sgName := fmt.Sprintf("%v-%v", fi.StringValue(ig.Spec.SecurityGroupOverride), ig.Spec.Role)
		sgTask := &awstasks.SecurityGroup{
			ID:     ig.Spec.SecurityGroupOverride,
			Name:   &sgName,
			Shared: fi.Bool(true),
		}
		t.SecurityGroups = append(t.SecurityGroups, sgTask)
	}

	// Add the additional NLB API security groups tasks
	if ig.HasAPIServer() &&
		b.APILoadBalancerClass() == kops.LoadBalancerClassNetwork {
		for _, id := range b.Cluster.Spec.API.LoadBalancer.AdditionalSecurityGroups {
			sgTask := &awstasks.SecurityGroup{
				ID:        fi.String(id),
				Lifecycle: b.SecurityLifecycle,
				Name:      fi.String("nlb-" + id),
				Shared:    fi.Bool(true),
			}
			if err := c.EnsureTask(sgTask); err != nil {
				return nil, err
			}
			t.SecurityGroups = append(t.SecurityGroups, sgTask)
		}
	}

	// Add the the additional security groups tasks
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

	// Set the root volume
	if ig.Spec.RootVolumeEncryption != nil {
		t.RootVolumeEncryption = ig.Spec.RootVolumeEncryption
	}
	if fi.BoolValue(ig.Spec.RootVolumeEncryption) && ig.Spec.RootVolumeEncryptionKey != nil {
		t.RootVolumeKmsKey = ig.Spec.RootVolumeEncryptionKey
	}
	if fi.Int32Value(ig.Spec.RootVolumeSize) > 0 {
		t.RootVolumeSize = fi.Int64(int64(fi.Int32Value(ig.Spec.RootVolumeSize)))
	} else {
		defaultVolumeSize, err := defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
		if err != nil {
			return nil, err
		}
		t.RootVolumeSize = fi.Int64(int64(defaultVolumeSize))
	}
	if ig.Spec.RootVolumeType != nil {
		t.RootVolumeType = ig.Spec.RootVolumeType
	}
	switch fi.StringValue(t.RootVolumeType) {
	case ec2.VolumeTypeGp3:
		t.RootVolumeIops = fi.Int64(int64(DefaultVolumeGp3Iops))
		if fi.Int32Value(ig.Spec.RootVolumeIops) > DefaultVolumeGp3Iops {
			t.RootVolumeIops = fi.Int64(int64(fi.Int32Value(ig.Spec.RootVolumeIops)))
		}
		t.RootVolumeThroughput = fi.Int64(int64(DefaultVolumeGp3Throughput))
		if fi.Int32Value(ig.Spec.RootVolumeThroughput) > DefaultVolumeGp3Throughput {
			t.RootVolumeThroughput = fi.Int64(int64(fi.Int32Value(ig.Spec.RootVolumeThroughput)))
		}
	case ec2.VolumeTypeIo1, ec2.VolumeTypeIo2:
		t.RootVolumeIops = fi.Int64(int64(DefaultVolumeIonIops))
		if fi.Int32Value(ig.Spec.RootVolumeIops) > DefaultVolumeIonIops {
			t.RootVolumeIops = fi.Int64(int64(fi.Int32Value(ig.Spec.RootVolumeIops)))
		}
	default:
		t.RootVolumeIops = nil
	}

	// Add the additional volumes
	for i := range ig.Spec.Volumes {
		x := &ig.Spec.Volumes[i]
		if x.DeleteOnTermination == nil {
			x.DeleteOnTermination = fi.Bool(DefaultVolumeDeleteOnTermination)
		}
		if x.Encrypted == nil {
			x.Encrypted = fi.Bool(DefaultVolumeEncryption)
		}
		if x.Type == "" {
			x.Type = DefaultVolumeType
		}
		switch x.Type {
		case ec2.VolumeTypeIo1, ec2.VolumeTypeIo2:
			if fi.Int64Value(x.Iops) < DefaultVolumeIonIops {
				x.Iops = fi.Int64(DefaultVolumeIonIops)
			}
		case ec2.VolumeTypeGp3:
			if fi.Int64Value(x.Iops) < DefaultVolumeGp3Iops {
				x.Iops = fi.Int64(DefaultVolumeGp3Iops)
			}
			if fi.Int64Value(x.Throughput) < DefaultVolumeGp3Throughput {
				x.Throughput = fi.Int64(DefaultVolumeGp3Throughput)
			}
		default:
			x.Iops = nil
		}
		t.BlockDeviceMappings = append(t.BlockDeviceMappings, &awstasks.BlockDeviceMapping{
			DeviceName:             fi.String(x.Device),
			EbsDeleteOnTermination: x.DeleteOnTermination,
			EbsEncrypted:           x.Encrypted,
			EbsKmsKey:              x.Key,
			EbsVolumeIops:          x.Iops,
			EbsVolumeSize:          fi.Int64(x.Size),
			EbsVolumeThroughput:    x.Throughput,
			EbsVolumeType:          fi.String(x.Type),
		})
	}

	if ig.Spec.Tenancy != "" {
		t.Tenancy = fi.String(ig.Spec.Tenancy)
	}

	if b.AWSModelContext.UseSSHKey() {
		if t.SSHKey, err = b.LinkToSSHKey(); err != nil {
			return nil, err
		}
	}

	// Add public IP based on subnet type
	subnets, err := b.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	switch subnets[0].Type {
	case kops.SubnetTypePrivate:
		t.AssociatePublicIP = fi.Bool(false)
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		t.AssociatePublicIP = fi.Bool(true)
		if ig.Spec.AssociatePublicIP != nil {
			t.AssociatePublicIP = ig.Spec.AssociatePublicIP
		}
	}

	// When using a MixedInstances ASG, AWS requires the SpotPrice be defined on the ASG
	// rather than the LaunchTemplate or else it returns this error:
	//   You cannot use a launch template that is set to request Spot Instances (InstanceMarketOptions)
	//   when you configure an Auto Scaling group with a mixed instances policy.
	if ig.Spec.MixedInstancesPolicy == nil && ig.Spec.MaxPrice != nil {
		t.SpotPrice = ig.Spec.MaxPrice
	}

	if ig.Spec.CPUCredits != nil {
		t.CPUCredits = ig.Spec.CPUCredits
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

	minSize := fi.Int64(1)
	maxSize := fi.Int64(1)
	if ig.Spec.MinSize != nil {
		minSize = fi.Int64(int64(*ig.Spec.MinSize))
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		minSize = fi.Int64(2)
	}
	if ig.Spec.MaxSize != nil {
		maxSize = fi.Int64(int64(*ig.Spec.MaxSize))
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		maxSize = fi.Int64(2)
	}

	t.MinSize = minSize
	t.MaxSize = maxSize

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

	t.LoadBalancers = []*awstasks.ClassicLoadBalancer{}
	t.TargetGroups = []*awstasks.TargetGroup{}

	// Spotinst handles load balancer attachments internally, so there's no
	// need to create separate attachments for both managed (+Spotinst) and
	// hybrid (+SpotinstHybrid) instance groups.
	if !featureflag.Spotinst.Enabled() ||
		(featureflag.SpotinstHybrid.Enabled() && !HybridInstanceGroup(ig)) {
		if b.UseLoadBalancerForAPI() && ig.HasAPIServer() {
			if b.UseNetworkLoadBalancer() {
				t.TargetGroups = append(t.TargetGroups, b.LinkToTargetGroup("tcp"))
				if b.Cluster.Spec.API.LoadBalancer.SSLCertificate != "" {
					t.TargetGroups = append(t.TargetGroups, b.LinkToTargetGroup("tls"))
				}
			} else {
				t.LoadBalancers = append(t.LoadBalancers, b.LinkToCLB("api"))
			}
		}

		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			t.LoadBalancers = append(t.LoadBalancers, b.LinkToCLB("bastion"))
		}
	}

	for _, extLB := range ig.Spec.ExternalLoadBalancers {
		if extLB.LoadBalancerName != nil {
			lb := &awstasks.ClassicLoadBalancer{
				Name:             extLB.LoadBalancerName,
				LoadBalancerName: extLB.LoadBalancerName,
				Shared:           fi.Bool(true),
			}
			t.LoadBalancers = append(t.LoadBalancers, lb)
			c.EnsureTask(lb)
		}

		if extLB.TargetGroupARN != nil {
			targetGroupName, err := awsup.GetTargetGroupNameFromARN(fi.StringValue(extLB.TargetGroupARN))
			if err != nil {
				return nil, err
			}
			tg := &awstasks.TargetGroup{
				Name:   fi.String(name + "-" + targetGroupName),
				ARN:    extLB.TargetGroupARN,
				Shared: fi.Bool(true),
			}
			t.TargetGroups = append(t.TargetGroups, tg)
			c.AddTask(tg)
		}
	}
	sort.Stable(awstasks.OrderLoadBalancersByName(t.LoadBalancers))
	sort.Stable(awstasks.OrderTargetGroupsByName(t.TargetGroups))

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
