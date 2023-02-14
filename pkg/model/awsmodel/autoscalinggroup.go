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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/api/resource"
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
	Lifecycle              fi.Lifecycle
	SecurityLifecycle      fi.Lifecycle
	Cluster                *kops.Cluster
}

var _ fi.CloudupModelBuilder = &AutoscalingGroupModelBuilder{}

// Build is responsible for constructing the aws autoscaling group from the kops spec
func (b *AutoscalingGroupModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
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
		if ig.Spec.Manager != "Karpenter" {
			tsk, err := b.buildAutoScalingGroupTask(c, name, ig)
			if err != nil {
				return err
			}
			tsk.LaunchTemplate = task
			c.AddTask(tsk)

			warmPool := b.Cluster.Spec.CloudProvider.AWS.WarmPool.ResolveDefaults(ig)

			enabled := fi.PtrTo(warmPool.IsEnabled())
			warmPoolTask := &awstasks.WarmPool{
				Name:      &name,
				Lifecycle: b.Lifecycle,
				Enabled:   enabled,
			}
			if warmPool.IsEnabled() {
				warmPoolTask.MinSize = warmPool.MinSize
				warmPoolTask.MaxSize = warmPool.MaxSize
				tsk.WarmPool = warmPoolTask
			} else {
				tsk.WarmPool = nil
			}
			c.AddTask(warmPoolTask)

			hookName := "kops-warmpool"
			name := fmt.Sprintf("%s-%s", hookName, ig.GetName())
			enableHook := warmPool.IsEnabled() && warmPool.EnableLifecycleHook

			lifecyleTask := &awstasks.AutoscalingLifecycleHook{
				ID:               aws.String(name),
				Name:             aws.String(name),
				HookName:         aws.String(hookName),
				AutoscalingGroup: b.LinkToAutoscalingGroup(ig),
				Lifecycle:        b.Lifecycle,
				DefaultResult:    aws.String("ABANDON"),
				// We let nodeup have 10 min to complete. Normally this should happen much faster,
				// but CP nodes need 5 min or so to start on new clusters, and we need to wait for that.
				HeartbeatTimeout:    aws.Int64(600),
				LifecycleTransition: aws.String("autoscaling:EC2_INSTANCE_LAUNCHING"),
				Enabled:             &enableHook,
			}

			c.AddTask(lifecyleTask)

		}
	}

	return nil
}

// buildLaunchTemplateTask is responsible for creating the template task into the aws model
func (b *AutoscalingGroupModelBuilder) buildLaunchTemplateTask(c *fi.CloudupModelBuilderContext, name string, ig *kops.InstanceGroup) (*awstasks.LaunchTemplate, error) {
	// @step: add the iam instance profile
	link, err := b.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return nil, fmt.Errorf("unable to find IAM profile link for instance group %q: %w", ig.ObjectMeta.Name, err)
	}

	rootVolumeSize, err := defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
	if err != nil {
		return nil, err
	}
	var rootVolumeType string
	rootVolumeEncryption := DefaultVolumeEncryption
	rootVolumeKmsKey := ""

	if ig.Spec.RootVolume != nil {
		if fi.ValueOf(ig.Spec.RootVolume.Size) > 0 {
			rootVolumeSize = fi.ValueOf(ig.Spec.RootVolume.Size)
		}

		rootVolumeType = fi.ValueOf(ig.Spec.RootVolume.Type)

		if ig.Spec.RootVolume.Encryption != nil {
			rootVolumeEncryption = fi.ValueOf(ig.Spec.RootVolume.Encryption)
		}

		if fi.ValueOf(ig.Spec.RootVolume.Encryption) && ig.Spec.RootVolume.EncryptionKey != nil {
			rootVolumeKmsKey = *ig.Spec.RootVolume.EncryptionKey
		}
	}
	if rootVolumeType == "" {
		rootVolumeType = DefaultVolumeType
	}

	securityGroups, err := b.buildSecurityGroups(c, ig)
	if err != nil {
		return nil, err
	}

	tags, err := b.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, fmt.Errorf("error building cloud tags: %v", err)
	}

	userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
	if err != nil {
		return nil, err
	}

	lt := &awstasks.LaunchTemplate{
		Name:                         fi.PtrTo(name),
		Lifecycle:                    b.Lifecycle,
		CPUCredits:                   fi.PtrTo(fi.ValueOf(ig.Spec.CPUCredits)),
		HTTPPutResponseHopLimit:      fi.PtrTo(int64(1)),
		HTTPTokens:                   fi.PtrTo(ec2.LaunchTemplateHttpTokensStateRequired),
		HTTPProtocolIPv6:             fi.PtrTo(ec2.LaunchTemplateInstanceMetadataProtocolIpv6Disabled),
		IAMInstanceProfile:           link,
		ImageID:                      fi.PtrTo(ig.Spec.Image),
		InstanceInterruptionBehavior: ig.Spec.InstanceInterruptionBehavior,
		InstanceMonitoring:           fi.PtrTo(false),
		IPv6AddressCount:             fi.PtrTo(int64(0)),
		RootVolumeIops:               fi.PtrTo(int64(0)),
		RootVolumeSize:               fi.PtrTo(int64(rootVolumeSize)),
		RootVolumeType:               fi.PtrTo(rootVolumeType),
		RootVolumeEncryption:         fi.PtrTo(rootVolumeEncryption),
		RootVolumeKmsKey:             fi.PtrTo(rootVolumeKmsKey),
		SecurityGroups:               securityGroups,
		Tags:                         tags,
		UserData:                     userData,
	}
	if ig.Spec.RootVolume != nil {
		lt.RootVolumeIops = fi.PtrTo(int64(fi.ValueOf(ig.Spec.RootVolume.IOPS)))
		lt.RootVolumeOptimization = ig.Spec.RootVolume.Optimization
	}

	if ig.Spec.Manager == kops.InstanceManagerCloudGroup {
		lt.InstanceType = fi.PtrTo(strings.Split(ig.Spec.MachineType, ",")[0])
	}

	{
		// @step: check the subnets are ok and pull together an array for us
		subnets, err := b.GatherSubnets(ig)
		if err != nil {
			return nil, err
		}

		// @step: check if we can add a public ip to this subnet
		switch subnets[0].Type {
		case kops.SubnetTypePublic, kops.SubnetTypeUtility:
			lt.AssociatePublicIP = fi.PtrTo(true)
			if ig.Spec.AssociatePublicIP != nil {
				lt.AssociatePublicIP = ig.Spec.AssociatePublicIP
			}
		case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
			lt.AssociatePublicIP = fi.PtrTo(false)
		}

		// @step: add an IPv6 address
		for _, clusterSubnet := range b.Cluster.Spec.Networking.Subnets {
			for _, igSubnet := range ig.Spec.Subnets {
				if clusterSubnet.Name != igSubnet {
					continue
				}
				if clusterSubnet.IPv6CIDR != "" {
					lt.IPv6AddressCount = fi.PtrTo(int64(1))
					lt.HTTPProtocolIPv6 = fi.PtrTo(ec2.LaunchTemplateInstanceMetadataProtocolIpv6Enabled)
				}
			}
		}
	}

	// @step: add any additional block devices
	for i := range ig.Spec.Volumes {
		x := &ig.Spec.Volumes[i]
		if x.Type == "" {
			x.Type = DefaultVolumeType
		}
		if x.Type == ec2.VolumeTypeIo1 || x.Type == ec2.VolumeTypeIo2 {
			if x.IOPS == nil {
				x.IOPS = fi.PtrTo(int64(DefaultVolumeIonIops))
			}
		} else if x.Type == ec2.VolumeTypeGp3 {
			if x.IOPS == nil {
				x.IOPS = fi.PtrTo(int64(DefaultVolumeGp3Iops))
			}
			if x.Throughput == nil {
				x.Throughput = fi.PtrTo(int64(DefaultVolumeGp3Throughput))
			}
		} else {
			x.IOPS = nil
		}
		deleteOnTermination := DefaultVolumeDeleteOnTermination
		if x.DeleteOnTermination != nil {
			deleteOnTermination = fi.ValueOf(x.DeleteOnTermination)
		}
		encryption := DefaultVolumeEncryption
		if x.Encrypted != nil {
			encryption = fi.ValueOf(x.Encrypted)
		}
		lt.BlockDeviceMappings = append(lt.BlockDeviceMappings, &awstasks.BlockDeviceMapping{
			DeviceName:             fi.PtrTo(x.Device),
			EbsDeleteOnTermination: fi.PtrTo(deleteOnTermination),
			EbsEncrypted:           fi.PtrTo(encryption),
			EbsKmsKey:              x.Key,
			EbsVolumeIops:          x.IOPS,
			EbsVolumeSize:          fi.PtrTo(x.Size),
			EbsVolumeThroughput:    x.Throughput,
			EbsVolumeType:          fi.PtrTo(x.Type),
		})
	}

	if ig.Spec.DetailedInstanceMonitoring != nil {
		lt.InstanceMonitoring = ig.Spec.DetailedInstanceMonitoring
	}

	if ig.Spec.InstanceMetadata != nil && ig.Spec.InstanceMetadata.HTTPPutResponseHopLimit != nil {
		lt.HTTPPutResponseHopLimit = ig.Spec.InstanceMetadata.HTTPPutResponseHopLimit
	}

	if ig.Spec.InstanceMetadata != nil && ig.Spec.InstanceMetadata.HTTPTokens != nil {
		lt.HTTPTokens = ig.Spec.InstanceMetadata.HTTPTokens
	} else if b.IsKubernetesLT("1.27") {
		lt.HTTPTokens = fi.PtrTo(ec2.LaunchTemplateHttpTokensStateOptional)
	}

	if rootVolumeType == ec2.VolumeTypeIo1 || rootVolumeType == ec2.VolumeTypeIo2 {
		if ig.Spec.RootVolume == nil || fi.ValueOf(ig.Spec.RootVolume.IOPS) < 100 {
			lt.RootVolumeIops = fi.PtrTo(int64(DefaultVolumeIonIops))
		}
	} else if rootVolumeType == ec2.VolumeTypeGp3 {
		if ig.Spec.RootVolume == nil || fi.ValueOf(ig.Spec.RootVolume.IOPS) < 3000 {
			lt.RootVolumeIops = fi.PtrTo(int64(DefaultVolumeGp3Iops))
		}
		if ig.Spec.RootVolume == nil || fi.ValueOf(ig.Spec.RootVolume.Throughput) < 125 {
			lt.RootVolumeThroughput = fi.PtrTo(int64(DefaultVolumeGp3Throughput))
		} else {
			lt.RootVolumeThroughput = fi.PtrTo(int64(fi.ValueOf(ig.Spec.RootVolume.Throughput)))
		}
	} else {
		lt.RootVolumeIops = nil
	}

	if b.AWSModelContext.UseSSHKey() {
		if lt.SSHKey, err = b.LinkToSSHKey(); err != nil {
			return nil, err
		}
	}

	// When using a MixedInstances ASG, AWS requires the SpotPrice be defined on the ASG
	// rather than the LaunchTemplate or else it returns this error:
	//   You cannot use a launch template that is set to request Spot Instances (InstanceMarketOptions)
	//   when you configure an Auto Scaling group with a mixed instances policy.
	if ig.Spec.MixedInstancesPolicy == nil && ig.Spec.MaxPrice != nil {
		lt.SpotPrice = ig.Spec.MaxPrice
	} else {
		lt.SpotPrice = fi.PtrTo("")
	}
	if ig.Spec.SpotDurationInMinutes != nil {
		lt.SpotDurationInMinutes = ig.Spec.SpotDurationInMinutes
	}

	if ig.Spec.Tenancy != "" {
		lt.Tenancy = fi.PtrTo(ig.Spec.Tenancy)
	}

	return lt, nil
}

// buildSecurityGroups is responsible for building security groups for a launch template.
func (b *AutoscalingGroupModelBuilder) buildSecurityGroups(c *fi.CloudupModelBuilderContext, ig *kops.InstanceGroup) ([]*awstasks.SecurityGroup, error) {
	// @step: if required we add the override for the security group for this instancegroup
	sgLink := b.LinkToSecurityGroup(ig.Spec.Role)
	if ig.Spec.SecurityGroupOverride != nil {
		sgName := fmt.Sprintf("%v-%v", fi.ValueOf(ig.Spec.SecurityGroupOverride), ig.Spec.Role)
		sgLink = &awstasks.SecurityGroup{
			ID:     ig.Spec.SecurityGroupOverride,
			Name:   &sgName,
			Shared: fi.PtrTo(true),
		}
	}

	securityGroups := []*awstasks.SecurityGroup{sgLink}

	if ig.HasAPIServer() &&
		b.APILoadBalancerClass() == kops.LoadBalancerClassNetwork {
		for _, id := range b.Cluster.Spec.API.LoadBalancer.AdditionalSecurityGroups {
			sgTask := &awstasks.SecurityGroup{
				ID:        fi.PtrTo(id),
				Lifecycle: b.SecurityLifecycle,
				Name:      fi.PtrTo("nlb-" + id),
				Shared:    fi.PtrTo(true),
			}
			c.EnsureTask(sgTask)
			securityGroups = append(securityGroups, sgTask)
		}
	}

	// @step: add any additional security groups to the instancegroup
	for _, id := range ig.Spec.AdditionalSecurityGroups {
		sgTask := &awstasks.SecurityGroup{
			ID:        fi.PtrTo(id),
			Lifecycle: b.SecurityLifecycle,
			Name:      fi.PtrTo(id),
			Shared:    fi.PtrTo(true),
		}
		c.EnsureTask(sgTask)
		securityGroups = append(securityGroups, sgTask)
	}

	return securityGroups, nil
}

// buildAutoscalingGroupTask is responsible for building the autoscaling task into the model
func (b *AutoscalingGroupModelBuilder) buildAutoScalingGroupTask(c *fi.CloudupModelBuilderContext, name string, ig *kops.InstanceGroup) (*awstasks.AutoscalingGroup, error) {
	t := &awstasks.AutoscalingGroup{
		Name:      fi.PtrTo(name),
		Lifecycle: b.Lifecycle,

		Granularity: fi.PtrTo("1Minute"),
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

		InstanceProtection: fi.PtrTo(false),
	}

	minSize := fi.PtrTo(int64(1))
	maxSize := fi.PtrTo(int64(1))
	if ig.Spec.MinSize != nil {
		minSize = fi.PtrTo(int64(*ig.Spec.MinSize))
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		minSize = fi.PtrTo(int64(2))
	}
	if ig.Spec.MaxSize != nil {
		maxSize = fi.PtrTo(int64(*ig.Spec.MaxSize))
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		maxSize = fi.PtrTo(int64(2))
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

	if ig.Spec.InstanceProtection != nil {
		t.InstanceProtection = ig.Spec.InstanceProtection
	}

	if ig.Spec.CapacityRebalance != nil {
		t.CapacityRebalance = ig.Spec.CapacityRebalance
	}

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
				if b.Cluster.UsesNoneDNS() && ig.IsControlPlane() {
					t.TargetGroups = append(t.TargetGroups, b.LinkToTargetGroup("kops-controller"))
				}
				if b.Cluster.Spec.API.LoadBalancer.SSLCertificate != "" {
					t.TargetGroups = append(t.TargetGroups, b.LinkToTargetGroup("tls"))
				}
			} else {
				t.LoadBalancers = append(t.LoadBalancers, b.LinkToCLB("api"))
			}
		}

		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			t.TargetGroups = append(t.TargetGroups, b.LinkToTargetGroup("bastion"))
		}
	}

	for _, extLB := range ig.Spec.ExternalLoadBalancers {
		if extLB.LoadBalancerName != nil {
			lb := &awstasks.ClassicLoadBalancer{
				Name:             extLB.LoadBalancerName,
				Lifecycle:        b.Lifecycle,
				LoadBalancerName: extLB.LoadBalancerName,
				Shared:           fi.PtrTo(true),
			}
			t.LoadBalancers = append(t.LoadBalancers, lb)
			c.EnsureTask(lb)
		}

		if extLB.TargetGroupARN != nil {
			targetGroupName, err := awsup.GetTargetGroupNameFromARN(fi.ValueOf(extLB.TargetGroupARN))
			if err != nil {
				return nil, err
			}
			tg := &awstasks.TargetGroup{
				Name:      fi.PtrTo(targetGroupName),
				Lifecycle: b.Lifecycle,
				ARN:       extLB.TargetGroupARN,
				Shared:    fi.PtrTo(true),
			}
			t.TargetGroups = append(t.TargetGroups, tg)
			c.EnsureTask(tg)
		}
	}
	sort.Stable(awstasks.OrderLoadBalancersByName(t.LoadBalancers))
	sort.Stable(awstasks.OrderTargetGroupsByName(t.TargetGroups))

	// @step: are we using a mixed instance policy
	if ig.Spec.MixedInstancesPolicy != nil && ig.Spec.Manager == kops.InstanceManagerCloudGroup {
		spec := ig.Spec.MixedInstancesPolicy

		if spec.InstanceRequirements != nil {

			ir := &awstasks.InstanceRequirements{}

			cpu := spec.InstanceRequirements.CPU
			if cpu != nil {
				if cpu.Max != nil {
					cpuMax, _ := spec.InstanceRequirements.CPU.Max.AsInt64()
					ir.CPUMax = &cpuMax
				}
				if cpu.Min != nil {
					cpuMin, _ := spec.InstanceRequirements.CPU.Min.AsInt64()
					ir.CPUMin = &cpuMin
				}
			} else {
				ir.CPUMin = fi.PtrTo(int64(0))
			}

			memory := spec.InstanceRequirements.Memory
			if memory != nil {
				if memory.Max != nil {
					memoryMax := spec.InstanceRequirements.Memory.Max.ScaledValue(resource.Mega)
					ir.MemoryMax = &memoryMax
				}
				if memory.Min != nil {
					memoryMin := spec.InstanceRequirements.Memory.Min.ScaledValue(resource.Mega)
					ir.MemoryMin = &memoryMin
				}
			} else {
				ir.MemoryMin = fi.PtrTo(int64(0))
			}
			t.InstanceRequirements = ir
		}

		t.MixedInstanceOverrides = spec.Instances
		t.MixedOnDemandAboveBase = spec.OnDemandAboveBase
		t.MixedOnDemandAllocationStrategy = spec.OnDemandAllocationStrategy
		t.MixedOnDemandBase = spec.OnDemandBase
		t.MixedSpotAllocationStrategy = spec.SpotAllocationStrategy
		t.MixedSpotInstancePools = spec.SpotInstancePools
		// In order to unset maxprice, the value needs to be ""
		if ig.Spec.MaxPrice == nil {
			t.MixedSpotMaxPrice = fi.PtrTo("")
		} else {
			t.MixedSpotMaxPrice = ig.Spec.MaxPrice
		}
	}

	if ig.Spec.MaxInstanceLifetime != nil {
		lifetimeSec := int64(ig.Spec.MaxInstanceLifetime.Seconds())
		t.MaxInstanceLifetime = fi.PtrTo(lifetimeSec)
	} else {
		t.MaxInstanceLifetime = fi.PtrTo(int64(0))
	}
	return t, nil
}
