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
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/util/pkg/maps"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

// CloudTagInstanceGroupRolePrefix is a cloud tag that defines the instance role
const CloudTagInstanceGroupRolePrefix = "k8s.io/role/"

// AutoscalingGroup provdes the definition for a autoscaling group in aws
type AutoscalingGroup struct {
	// Name is the name of the ASG
	Name *string
	// Lifecycle is the resource lifecycle
	Lifecycle *fi.Lifecycle

	// Granularity specifys the granularity of the metrics
	Granularity *string
	// InstanceProtection makes new instances in an autoscaling group protected from scale in
	InstanceProtection *bool
	// LaunchConfiguration is the launch configuration for the autoscaling group
	LaunchConfiguration *LaunchConfiguration
	// LaunchTemplate is the launch template for the asg
	LaunchTemplate *LaunchTemplate
	// MaxSize is the max number of nodes in asg
	MaxSize *int64
	// Metrics is a collection of metrics to monitor
	Metrics []string
	// MinSize is the smallest number of nodes in the asg
	MinSize *int64
	// MixedInstanceOverrides is a collection of instance types to use with fleet policy
	MixedInstanceOverrides []string
	// MixedOnDemandAllocationStrategy is allocation strategy to use for on-demand instances
	MixedOnDemandAllocationStrategy *string
	// MixedOnDemandBase is percentage split of On-Demand Instances and Spot Instances for your
	// additional capacity beyond the base portion
	MixedOnDemandBase *int64
	// MixedOnDemandAboveBase is the percentage split of On-Demand Instances and Spot Instances
	// for your additional capacity beyond the base portion.
	MixedOnDemandAboveBase *int64
	// MixedSpotAllocationStrategy diversifies your Spot capacity across multiple instance types to
	// find the best pricing. Higher Spot availability may result from a larger number of
	// instance types to choose from.
	MixedSpotAllocationStrategy *string
	// MixedSpotInstancePools is the number of Spot pools to use to allocate your Spot capacity (defaults to 2)
	// pools are determined from the different instance types in the Overrides array of LaunchTemplate
	MixedSpotInstancePools *int64
	// MixedSpotMaxPrice is the maximum price per unit hour you are willing to pay for a Spot Instance
	MixedSpotMaxPrice *string
	// Subnets is a collection of subnets to attach the nodes to
	Subnets []*Subnet
	// SuspendProcesses
	SuspendProcesses *[]string
	// Tags is a collection of keypairs to apply to the node on launch
	Tags map[string]string
}

var _ fi.CompareWithID = &AutoscalingGroup{}

// CompareWithID returns the ID of the ASG
func (e *AutoscalingGroup) CompareWithID() *string {
	return e.Name
}

// Find is used to discover the ASG in the cloud provider
func (e *AutoscalingGroup) Find(c *fi.Context) (*AutoscalingGroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	g, err := findAutoscalingGroup(cloud, fi.StringValue(e.Name))
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}

	actual := &AutoscalingGroup{
		Name:    g.AutoScalingGroupName,
		MaxSize: g.MaxSize,
		MinSize: g.MinSize,
	}

	if g.VPCZoneIdentifier != nil {
		subnets := strings.Split(*g.VPCZoneIdentifier, ",")
		for _, subnet := range subnets {
			actual.Subnets = append(actual.Subnets, &Subnet{ID: aws.String(subnet)})
		}
	}

	for _, enabledMetric := range g.EnabledMetrics {
		actual.Metrics = append(actual.Metrics, aws.StringValue(enabledMetric.Metric))
		actual.Granularity = enabledMetric.Granularity
	}
	sort.Strings(actual.Metrics)

	if len(g.Tags) != 0 {
		actual.Tags = make(map[string]string)
		for _, tag := range g.Tags {
			actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
		}
	}

	if g.LaunchConfigurationName != nil {
		actual.LaunchConfiguration = &LaunchConfiguration{ID: g.LaunchConfigurationName}
	}
	if g.LaunchTemplate != nil {
		actual.LaunchTemplate = &LaunchTemplate{ID: g.LaunchTemplate.LaunchTemplateName}
	}

	if g.MixedInstancesPolicy != nil {
		mp := g.MixedInstancesPolicy
		if mp.InstancesDistribution != nil {
			mpd := mp.InstancesDistribution
			actual.MixedOnDemandAboveBase = mpd.OnDemandPercentageAboveBaseCapacity
			actual.MixedOnDemandAllocationStrategy = mpd.OnDemandAllocationStrategy
			actual.MixedOnDemandBase = mpd.OnDemandBaseCapacity
			actual.MixedSpotAllocationStrategy = mpd.SpotAllocationStrategy
			actual.MixedSpotInstancePools = mpd.SpotInstancePools
			actual.MixedSpotMaxPrice = mpd.SpotMaxPrice
		}

		if g.MixedInstancesPolicy.LaunchTemplate != nil {
			if g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
				actual.LaunchTemplate = &LaunchTemplate{ID: g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName}
			}

			for _, n := range g.MixedInstancesPolicy.LaunchTemplate.Overrides {
				actual.MixedInstanceOverrides = append(actual.MixedInstanceOverrides, fi.StringValue(n.InstanceType))
			}
		}
	}

	if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
		actual.Subnets = e.Subnets
	}

	processes := []string{}
	for _, p := range g.SuspendedProcesses {
		processes = append(processes, *p.ProcessName)
	}

	actual.SuspendProcesses = &processes

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	if g.NewInstancesProtectedFromScaleIn != nil {
		actual.InstanceProtection = g.NewInstancesProtectedFromScaleIn
	}

	return actual, nil
}

// findAutoscalingGroup is responsible for finding all the autoscaling groups for us
func findAutoscalingGroup(cloud awsup.AWSCloud, name string) (*autoscaling.Group, error) {
	request := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&name},
	}

	var found []*autoscaling.Group
	err := cloud.Autoscaling().DescribeAutoScalingGroupsPages(request, func(p *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) (shouldContinue bool) {
		for _, g := range p.AutoScalingGroups {
			// Check for "Delete in progress" (the only use .Status). We won't be able to update or create while
			// this is true, but filtering it out here makes the messages slightly clearer.
			if g.Status != nil {
				klog.Warningf("Skipping AutoScalingGroup %v: %v", fi.StringValue(g.AutoScalingGroupName), fi.StringValue(g.Status))
				continue
			}

			if aws.StringValue(g.AutoScalingGroupName) == name {
				found = append(found, g)
			} else {
				klog.Warningf("Got ASG with unexpected name %q", fi.StringValue(g.AutoScalingGroupName))
			}
		}

		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing AutoscalingGroups: %v", err)
	}

	switch len(found) {
	case 0:
		return nil, nil
	case 1:
		return found[0], nil
	}

	return nil, fmt.Errorf("found multiple AutoscalingGroups with name: %q", name)
}

func (e *AutoscalingGroup) normalize(c *fi.Context) error {
	sort.Strings(e.Metrics)

	return nil
}

// Run is responsible for running the task
func (e *AutoscalingGroup) Run(c *fi.Context) error {
	err := e.normalize(c)
	if err != nil {
		return err
	}
	c.Cloud.(awsup.AWSCloud).AddTags(e.Name, e.Tags)

	return fi.DefaultDeltaRunMethod(e, c)
}

// CheckChanges is responsible for checking for changes??
func (e *AutoscalingGroup) CheckChanges(a, ex, changes *AutoscalingGroup) error {
	if a != nil {
		if ex.Name == nil {
			return fi.RequiredField("Name")
		}
	}

	return nil
}

// RenderAWS is responsible for building the autoscaling group via AWS API
func (v *AutoscalingGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *AutoscalingGroup) error {
	// @step: did we find an autoscaling group?
	if a == nil {
		klog.V(2).Infof("Creating autoscaling group with name: %s", fi.StringValue(e.Name))

		request := &autoscaling.CreateAutoScalingGroupInput{
			AutoScalingGroupName: e.Name,
			MinSize:              e.MinSize,
			MaxSize:              e.MaxSize,
			Tags:                 v.AutoscalingGroupTags(),
			VPCZoneIdentifier:    fi.String(strings.Join(e.AutoscalingGroupSubnets(), ",")),
		}

		// @check are we using a launch configuration, mixed instances policy, or launch template
		if e.LaunchConfiguration != nil {
			request.LaunchConfigurationName = e.LaunchConfiguration.ID
		} else if e.UseMixedInstancesPolicy() {
			request.MixedInstancesPolicy = &autoscaling.MixedInstancesPolicy{
				InstancesDistribution: &autoscaling.InstancesDistribution{
					OnDemandPercentageAboveBaseCapacity: e.MixedOnDemandAboveBase,
					OnDemandBaseCapacity:                e.MixedOnDemandBase,
					SpotAllocationStrategy:              e.MixedSpotAllocationStrategy,
					SpotInstancePools:                   e.MixedSpotInstancePools,
					SpotMaxPrice:                        e.MixedSpotMaxPrice,
				},
				LaunchTemplate: &autoscaling.LaunchTemplate{
					LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
						LaunchTemplateName: e.LaunchTemplate.ID,
						Version:            aws.String("1"),
					},
				},
			}
			p := request.MixedInstancesPolicy.LaunchTemplate
			for _, x := range e.MixedInstanceOverrides {
				p.Overrides = append(p.Overrides, &autoscaling.LaunchTemplateOverrides{
					InstanceType: fi.String(x)},
				)
			}
		} else if e.LaunchTemplate != nil {
			request.LaunchTemplate = &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: e.LaunchTemplate.ID,
				Version:            aws.String("1"),
			}
		} else {
			return fmt.Errorf("could not find one of launch configuration, mixed instances policy, or launch template")
		}

		// @step: attempt to create the autoscaling group for us
		if _, err := t.Cloud.Autoscaling().CreateAutoScalingGroup(request); err != nil {
			return fmt.Errorf("error creating AutoscalingGroup: %v", err)
		}

		// @step: attempt to enable the metrics for us
		if _, err := t.Cloud.Autoscaling().EnableMetricsCollection(&autoscaling.EnableMetricsCollectionInput{
			AutoScalingGroupName: e.Name,
			Granularity:          e.Granularity,
			Metrics:              aws.StringSlice(e.Metrics),
		}); err != nil {
			return fmt.Errorf("error enabling metrics collection for AutoscalingGroup: %v", err)
		}

		if len(*e.SuspendProcesses) > 0 {
			toSuspend := []*string{}
			for _, p := range *e.SuspendProcesses {
				toSuspend = append(toSuspend, &p)
			}

			processQuery := &autoscaling.ScalingProcessQuery{}
			processQuery.AutoScalingGroupName = e.Name
			processQuery.ScalingProcesses = toSuspend

			if _, err := t.Cloud.Autoscaling().SuspendProcesses(processQuery); err != nil {
				return fmt.Errorf("error suspending processes: %v", err)
			}
		}

		if e.InstanceProtection != nil {
			request.NewInstancesProtectedFromScaleIn = e.InstanceProtection
		}
	} else {
		// @logic: else we have found a autoscaling group and we need to evaluate the difference
		request := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: e.Name,
		}

		if changes.LaunchConfiguration != nil {
			request.LaunchConfigurationName = e.LaunchConfiguration.ID
			changes.LaunchConfiguration = nil
		}
		setup := func(req *autoscaling.UpdateAutoScalingGroupInput) *autoscaling.MixedInstancesPolicy {
			if req.MixedInstancesPolicy == nil {
				req.MixedInstancesPolicy = &autoscaling.MixedInstancesPolicy{
					InstancesDistribution: &autoscaling.InstancesDistribution{},
				}
			}

			return req.MixedInstancesPolicy
		}

		launchTemplateVersion := "1"
		if e.LaunchTemplate != nil {
			dltRequest := &ec2.DescribeLaunchTemplatesInput{
				LaunchTemplateNames: []*string{e.LaunchTemplate.ID},
			}
			dltResponse, err := t.Cloud.EC2().DescribeLaunchTemplates(dltRequest)
			if err != nil {
				klog.Warningf("could not find existing LaunchTemplate: %v", err)
			} else {
				launchTemplateVersion = strconv.FormatInt(*dltResponse.LaunchTemplates[0].LatestVersionNumber, 10)
			}
		}

		if changes.LaunchTemplate != nil {
			// @note: at the moment we are only using launch templates when using mixed instance policies,
			// but this might change
			setup(request).LaunchTemplate = &autoscaling.LaunchTemplate{
				LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
					LaunchTemplateName: changes.LaunchTemplate.ID,
					Version:            &launchTemplateVersion,
				},
			}
			changes.LaunchTemplate = nil
		}

		if changes.MixedOnDemandAboveBase != nil {
			setup(request).InstancesDistribution.OnDemandPercentageAboveBaseCapacity = e.MixedOnDemandAboveBase
			changes.MixedOnDemandAboveBase = nil
		}
		if changes.MixedOnDemandBase != nil {
			setup(request).InstancesDistribution.OnDemandBaseCapacity = e.MixedOnDemandBase
			changes.MixedOnDemandBase = nil
		}
		if changes.MixedSpotAllocationStrategy != nil {
			setup(request).InstancesDistribution.SpotAllocationStrategy = e.MixedSpotAllocationStrategy
			changes.MixedSpotAllocationStrategy = nil
		}
		if changes.MixedSpotInstancePools != nil {
			setup(request).InstancesDistribution.SpotInstancePools = e.MixedSpotInstancePools
			changes.MixedSpotInstancePools = nil
		}
		if changes.MixedSpotMaxPrice != nil {
			setup(request).InstancesDistribution.SpotMaxPrice = e.MixedSpotMaxPrice
			changes.MixedSpotMaxPrice = nil
		}
		if changes.MixedInstanceOverrides != nil {
			if setup(request).LaunchTemplate == nil {
				setup(request).LaunchTemplate = &autoscaling.LaunchTemplate{
					LaunchTemplateSpecification: &autoscaling.LaunchTemplateSpecification{
						LaunchTemplateName: e.LaunchTemplate.ID,
						Version:            &launchTemplateVersion,
					},
				}
			}

			p := request.MixedInstancesPolicy.LaunchTemplate
			for _, x := range changes.MixedInstanceOverrides {
				p.Overrides = append(p.Overrides, &autoscaling.LaunchTemplateOverrides{InstanceType: fi.String(x)})
			}
			changes.MixedInstanceOverrides = nil
		}

		if changes.MinSize != nil {
			request.MinSize = e.MinSize
			changes.MinSize = nil
		}
		if changes.MaxSize != nil {
			request.MaxSize = e.MaxSize
			changes.MaxSize = nil
		}
		if changes.Subnets != nil {
			request.VPCZoneIdentifier = aws.String(strings.Join(e.AutoscalingGroupSubnets(), ","))
			changes.Subnets = nil
		}

		var updateTagsRequest *autoscaling.CreateOrUpdateTagsInput
		var deleteTagsRequest *autoscaling.DeleteTagsInput
		if changes.Tags != nil {
			updateTagsRequest = &autoscaling.CreateOrUpdateTagsInput{Tags: e.AutoscalingGroupTags()}

			if a != nil && len(a.Tags) > 0 {
				deleteTagsRequest = &autoscaling.DeleteTagsInput{}
				deleteTagsRequest.Tags = e.getASGTagsToDelete(a.Tags)
			}

			changes.Tags = nil
		}

		if changes.Metrics != nil || changes.Granularity != nil {
			// TODO: Support disabling metrics?
			if len(e.Metrics) != 0 {
				_, err := t.Cloud.Autoscaling().EnableMetricsCollection(&autoscaling.EnableMetricsCollectionInput{
					AutoScalingGroupName: e.Name,
					Granularity:          e.Granularity,
					Metrics:              aws.StringSlice(e.Metrics),
				})
				if err != nil {
					return fmt.Errorf("error enabling metrics collection for AutoscalingGroup: %v", err)
				}
				changes.Metrics = nil
				changes.Granularity = nil
			}
		}

		if changes.SuspendProcesses != nil {
			toSuspend := processCompare(e.SuspendProcesses, a.SuspendProcesses)
			toResume := processCompare(a.SuspendProcesses, e.SuspendProcesses)

			if len(toSuspend) > 0 {
				suspendProcessQuery := &autoscaling.ScalingProcessQuery{}
				suspendProcessQuery.AutoScalingGroupName = e.Name
				suspendProcessQuery.ScalingProcesses = toSuspend

				_, err := t.Cloud.Autoscaling().SuspendProcesses(suspendProcessQuery)
				if err != nil {
					return fmt.Errorf("error suspending processes: %v", err)
				}
			}
			if len(toResume) > 0 {
				resumeProcessQuery := &autoscaling.ScalingProcessQuery{}
				resumeProcessQuery.AutoScalingGroupName = e.Name
				resumeProcessQuery.ScalingProcesses = toResume

				_, err := t.Cloud.Autoscaling().ResumeProcesses(resumeProcessQuery)
				if err != nil {
					return fmt.Errorf("error resuming processes: %v", err)
				}
			}
			changes.SuspendProcesses = nil
		}

		if changes.InstanceProtection != nil {
			request.NewInstancesProtectedFromScaleIn = e.InstanceProtection
			changes.InstanceProtection = nil
		}

		empty := &AutoscalingGroup{}
		if !reflect.DeepEqual(empty, changes) {
			klog.Warningf("cannot apply changes to AutoScalingGroup: %v", changes)
		}

		klog.V(2).Infof("Updating autoscaling group %s", fi.StringValue(e.Name))

		if _, err := t.Cloud.Autoscaling().UpdateAutoScalingGroup(request); err != nil {
			return fmt.Errorf("error updating AutoscalingGroup: %v", err)
		}

		if deleteTagsRequest != nil && len(deleteTagsRequest.Tags) > 0 {
			if _, err := t.Cloud.Autoscaling().DeleteTags(deleteTagsRequest); err != nil {
				return fmt.Errorf("error deleting old AutoscalingGroup tags: %v", err)
			}
		}
		if updateTagsRequest != nil {
			if _, err := t.Cloud.Autoscaling().CreateOrUpdateTags(updateTagsRequest); err != nil {
				return fmt.Errorf("error updating AutoscalingGroup tags: %v", err)
			}
		}
	}

	return nil
}

// UseMixedInstancesPolicy checks if we should add a mixed instances policy to the asg
func (e *AutoscalingGroup) UseMixedInstancesPolicy() bool {
	if e.LaunchTemplate == nil {
		return false
	}
	// @check if any of the mixed instance policies settings are toggled
	if e.MixedOnDemandAboveBase != nil {
		return true
	}
	if e.MixedOnDemandBase != nil {
		return true
	}
	if e.MixedSpotAllocationStrategy != nil {
		return true
	}
	if e.MixedSpotInstancePools != nil {
		return true
	}
	if len(e.MixedInstanceOverrides) > 0 {
		return true
	}
	if e.MixedSpotMaxPrice != nil {
		return true
	}

	return false
}

// AutoscalingGroupTags is responsible for generating the tagging for the asg
func (e *AutoscalingGroup) AutoscalingGroupTags() []*autoscaling.Tag {
	var list []*autoscaling.Tag

	for k, v := range e.Tags {
		list = append(list, &autoscaling.Tag{
			Key:               aws.String(k),
			Value:             aws.String(v),
			ResourceId:        e.Name,
			ResourceType:      aws.String("auto-scaling-group"),
			PropagateAtLaunch: aws.Bool(true),
		})
	}

	return list
}

// AutoscalingGroupSubnets returns the subnets list
func (e *AutoscalingGroup) AutoscalingGroupSubnets() []string {
	var list []string

	for _, x := range e.Subnets {
		list = append(list, fi.StringValue(x.ID))
	}

	return list
}

// processCompare returns processes that exist in a but not in b
func processCompare(a *[]string, b *[]string) []*string {
	notInB := []*string{}
	for _, ap := range *a {
		found := false
		for _, bp := range *b {
			if ap == bp {
				found = true
				break
			}
		}
		if !found {
			notFound := ap
			notInB = append(notInB, &notFound)
		}
	}
	return notInB
}

// getASGTagsToDelete loops through the currently set tags and builds a list of
// tags to be deleted from the Autoscaling Group
func (e *AutoscalingGroup) getASGTagsToDelete(currentTags map[string]string) []*autoscaling.Tag {
	tagsToDelete := []*autoscaling.Tag{}

	for k, v := range currentTags {
		if _, ok := e.Tags[k]; !ok {
			tagsToDelete = append(tagsToDelete, &autoscaling.Tag{
				Key:          aws.String(k),
				Value:        aws.String(v),
				ResourceId:   e.Name,
				ResourceType: aws.String("auto-scaling-group"),
			})
		}
	}
	return tagsToDelete
}

type terraformASGTag struct {
	Key               *string `json:"key"`
	Value             *string `json:"value"`
	PropagateAtLaunch *bool   `json:"propagate_at_launch"`
}

type terraformAutoscalingLaunchTemplateSpecification struct {
	// LaunchTemplateID is the ID of the template to use.
	LaunchTemplateID *terraform.Literal `json:"id,omitempty"`
	// Version is the version of the Launch Template to use.
	Version *terraform.Literal `json:"version,omitempty"`
}

type terraformAutoscalingMixedInstancesPolicyLaunchTemplateSpecification struct {
	// LaunchTemplateID is the ID of the template to use
	LaunchTemplateID *terraform.Literal `json:"launch_template_id,omitempty"`
	// Version is the version of the Launch Template to use
	Version *terraform.Literal `json:"version,omitempty"`
}

type terraformAutoscalingMixedInstancesPolicyLaunchTemplateOverride struct {
	// InstanceType is the instance to use
	InstanceType *string `json:"instance_type,omitempty"`
}

type terraformAutoscalingMixedInstancesPolicyLaunchTemplate struct {
	// LaunchTemplateSpecification is the definition for a LT
	LaunchTemplateSpecification []*terraformAutoscalingMixedInstancesPolicyLaunchTemplateSpecification `json:"launch_template_specification,omitempty"`
	// Override the is machine type override
	Override []*terraformAutoscalingMixedInstancesPolicyLaunchTemplateOverride `json:"override,omitempty"`
}

type terraformAutoscalingInstanceDistribution struct {
	// OnDemandAllocationStrategy
	OnDemandAllocationStrategy *string `json:"on_demand_allocation_strategy,omitempty"`
	// OnDemandBaseCapacity is the base ondemand requirement
	OnDemandBaseCapacity *int64 `json:"on_demand_base_capacity,omitempty"`
	// OnDemandPercentageAboveBaseCapacity is the percentage above base for on-demand instances
	OnDemandPercentageAboveBaseCapacity *int64 `json:"on_demand_percentage_above_base_capacity,omitempty"`
	// SpotAllocationStrategy is the spot allocation stratergy
	SpotAllocationStrategy *string `json:"spot_allocation_strategy,omitempty"`
	// SpotInstancePool is the number of pools
	SpotInstancePool *int64 `json:"spot_instance_pools,omitempty"`
	// SpotMaxPrice is the max bid on spot instance, defaults to demand value
	SpotMaxPrice *string `json:"spot_max_price,omitempty"`
}

type terraformMixedInstancesPolicy struct {
	// LaunchTemplate is the launch template spec
	LaunchTemplate []*terraformAutoscalingMixedInstancesPolicyLaunchTemplate `json:"launch_template,omitempty"`
	// InstanceDistribution is the distribution strategy
	InstanceDistribution []*terraformAutoscalingInstanceDistribution `json:"instances_distribution,omitempty"`
}

type terraformAutoscalingGroup struct {
	Name                    *string                                          `json:"name,omitempty"`
	LaunchConfigurationName *terraform.Literal                               `json:"launch_configuration,omitempty"`
	LaunchTemplate          *terraformAutoscalingLaunchTemplateSpecification `json:"launch_template,omitempty"`
	MaxSize                 *int64                                           `json:"max_size,omitempty"`
	MinSize                 *int64                                           `json:"min_size,omitempty"`
	MixedInstancesPolicy    []*terraformMixedInstancesPolicy                 `json:"mixed_instances_policy,omitempty"`
	VPCZoneIdentifier       []*terraform.Literal                             `json:"vpc_zone_identifier,omitempty"`
	Tags                    []*terraformASGTag                               `json:"tag,omitempty"`
	MetricsGranularity      *string                                          `json:"metrics_granularity,omitempty"`
	EnabledMetrics          []*string                                        `json:"enabled_metrics,omitempty"`
	SuspendedProcesses      []*string                                        `json:"suspended_processes,omitempty"`
	InstanceProtection      *bool                                            `json:"protect_from_scale_in,omitempty"`
}

// RenderTerraform is responsible for rendering the terraform codebase
func (_ *AutoscalingGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *AutoscalingGroup) error {
	tf := &terraformAutoscalingGroup{
		Name:               e.Name,
		MinSize:            e.MinSize,
		MaxSize:            e.MaxSize,
		MetricsGranularity: e.Granularity,
		EnabledMetrics:     aws.StringSlice(e.Metrics),
		InstanceProtection: e.InstanceProtection,
	}

	for _, s := range e.Subnets {
		tf.VPCZoneIdentifier = append(tf.VPCZoneIdentifier, s.TerraformLink())
	}

	for _, k := range maps.SortedKeys(e.Tags) {
		v := e.Tags[k]
		tf.Tags = append(tf.Tags, &terraformASGTag{
			Key:               fi.String(k),
			Value:             fi.String(v),
			PropagateAtLaunch: fi.Bool(true),
		})
	}

	if e.LaunchConfiguration != nil {
		tf.LaunchConfigurationName = e.LaunchConfiguration.TerraformLink()
	} else if e.UseMixedInstancesPolicy() {
		// Temporary warning until https://github.com/terraform-providers/terraform-provider-aws/issues/9750 is resolved
		if e.MixedSpotAllocationStrategy == fi.String("capacity-optimized") {
			fmt.Print("Terraform does not currently support a capacity optimized strategy - please see https://github.com/terraform-providers/terraform-provider-aws/issues/9750")
		}

		tf.MixedInstancesPolicy = []*terraformMixedInstancesPolicy{
			{
				LaunchTemplate: []*terraformAutoscalingMixedInstancesPolicyLaunchTemplate{
					{
						LaunchTemplateSpecification: []*terraformAutoscalingMixedInstancesPolicyLaunchTemplateSpecification{
							{
								LaunchTemplateID: e.LaunchTemplate.TerraformLink(),
								Version:          e.LaunchTemplate.VersionLink(),
							},
						},
					},
				},
				InstanceDistribution: []*terraformAutoscalingInstanceDistribution{
					{
						OnDemandAllocationStrategy:          e.MixedOnDemandAllocationStrategy,
						OnDemandBaseCapacity:                e.MixedOnDemandBase,
						OnDemandPercentageAboveBaseCapacity: e.MixedOnDemandAboveBase,
						SpotAllocationStrategy:              e.MixedSpotAllocationStrategy,
						SpotInstancePool:                    e.MixedSpotInstancePools,
						SpotMaxPrice:                        e.MixedSpotMaxPrice,
					},
				},
			},
		}

		for _, x := range e.MixedInstanceOverrides {
			tf.MixedInstancesPolicy[0].LaunchTemplate[0].Override = append(tf.MixedInstancesPolicy[0].LaunchTemplate[0].Override, &terraformAutoscalingMixedInstancesPolicyLaunchTemplateOverride{InstanceType: fi.String(x)})
		}
	} else if e.LaunchTemplate != nil {
		tf.LaunchTemplate = &terraformAutoscalingLaunchTemplateSpecification{
			LaunchTemplateID: e.LaunchTemplate.TerraformLink(),
			Version:          e.LaunchTemplate.VersionLink(),
		}
	} else {
		return fmt.Errorf("could not find one of launch configuration, mixed instances policy, or launch template")
	}

	role := ""
	for k := range e.Tags {
		if strings.HasPrefix(k, CloudTagInstanceGroupRolePrefix) {
			suffix := strings.TrimPrefix(k, CloudTagInstanceGroupRolePrefix)
			if role != "" && role != suffix {
				return fmt.Errorf("Found multiple role tags: %q vs %q", role, suffix)
			}
			role = suffix
		}
	}

	if e.LaunchConfiguration != nil {
		// Create TF output variable with security group ids
		// This is in the launch configuration, but the ASG has the information about the instance group type
		if role != "" {
			for _, sg := range e.LaunchConfiguration.SecurityGroups {
				if err := t.AddOutputVariableArray(role+"_security_group_ids", sg.TerraformLink()); err != nil {
					return err
				}
			}
		}
	} else if e.LaunchTemplate != nil && role != "" {
		for _, sg := range e.LaunchTemplate.SecurityGroups {
			if err := t.AddOutputVariableArray(role+"_security_group_ids", sg.TerraformLink()); err != nil {
				return err
			}
		}
	}
	if role != "" {
		if err := t.AddOutputVariableArray(role+"_autoscaling_group_ids", e.TerraformLink()); err != nil {
			return err
		}
	}
	if role == "node" {
		for _, s := range e.Subnets {
			if err := t.AddOutputVariableArray(role+"_subnet_ids", s.TerraformLink()); err != nil {
				return err
			}
		}
	}

	var processes []*string
	if e.SuspendProcesses != nil {
		for _, p := range *e.SuspendProcesses {
			processes = append(processes, fi.String(p))
		}
	}
	tf.SuspendedProcesses = processes

	return t.RenderResource("aws_autoscaling_group", *e.Name, tf)
}

// TerraformLink fills in the property
func (e *AutoscalingGroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_autoscaling_group", fi.StringValue(e.Name), "id")
}

type cloudformationASGTag struct {
	Key               *string `json:"Key"`
	Value             *string `json:"Value"`
	PropagateAtLaunch *bool   `json:"PropagateAtLaunch"`
}

type cloudformationASGMetricsCollection struct {
	Granularity *string   `json:"Granularity"`
	Metrics     []*string `json:"Metrics"`
}

type cloudformationAutoscalingLaunchTemplateSpecification struct {
	// LaunchTemplateId is the IDx of the template to use.
	LaunchTemplateId *cloudformation.Literal `json:"LaunchTemplateId,omitempty"`
	// Version is the version number of the template to use.
	Version *cloudformation.Literal `json:"Version,omitempty"`
}

type cloudformationAutoscalingLaunchTemplateOverride struct {
	// InstanceType is the instance to use
	InstanceType *string `json:"InstanceType,omitempty"`
}

type cloudformationAutoscalingLaunchTemplate struct {
	// LaunchTemplateSpecification is the definition for a LT
	LaunchTemplateSpecification *cloudformationAutoscalingLaunchTemplateSpecification `json:"LaunchTemplateSpecification,omitempty"`
	// Override the is machine type override
	Overrides []*cloudformationAutoscalingLaunchTemplateOverride `json:"Overrides,omitempty"`
}

type cloudformationAutoscalingInstanceDistribution struct {
	// OnDemandAllocationStrategy
	OnDemandAllocationStrategy *string `json:"InstancesDistribution,omitempty"`
	// OnDemandBaseCapacity is the base ondemand requirement
	OnDemandBaseCapacity *int64 `json:"OnDemandBaseCapacity,omitempty"`
	// OnDemandPercentageAboveBaseCapacity is the percentage above base for on-demand instances
	OnDemandPercentageAboveBaseCapacity *int64 `json:"OnDemandPercentageAboveBaseCapacity,omitempty"`
	// SpotAllocationStrategy is the spot allocation stratergy
	SpotAllocationStrategy *string `json:"SpotAllocationStrategy,omitempty"`
	// SpotInstancePool is the number of pools
	SpotInstancePool *int64 `json:"SpotInstancePool,omitempty"`
	// SpotMaxPrice is the max bid on spot instance, defaults to demand value
	SpotMaxPrice *string `json:"SpotMaxPrice,omitempty"`
}

type cloudformationMixedInstancesPolicy struct {
	// LaunchTemplate is the launch template spec
	LaunchTemplate *cloudformationAutoscalingLaunchTemplate `json:"LaunchTemplate,omitempty"`
	// InstanceDistribution is the distribution strategy
	InstanceDistribution *cloudformationAutoscalingInstanceDistribution `json:"InstancesDistribution,omitempty"`
}

type cloudformationAutoscalingGroup struct {
	Name                    *string                                               `json:"AutoScalingGroupName,omitempty"`
	LaunchConfigurationName *cloudformation.Literal                               `json:"LaunchConfigurationName,omitempty"`
	LaunchTemplate          *cloudformationAutoscalingLaunchTemplateSpecification `json:"LaunchTemplate,omitempty"`
	MaxSize                 *int64                                                `json:"MaxSize,omitempty"`
	MinSize                 *int64                                                `json:"MinSize,omitempty"`
	VPCZoneIdentifier       []*cloudformation.Literal                             `json:"VPCZoneIdentifier,omitempty"`
	Tags                    []*cloudformationASGTag                               `json:"Tags,omitempty"`
	MetricsCollection       []*cloudformationASGMetricsCollection                 `json:"MetricsCollection,omitempty"`
	MixedInstancesPolicy    *cloudformationMixedInstancesPolicy                   `json:"MixedInstancesPolicy,omitempty"`
	LoadBalancerNames       []*cloudformation.Literal                             `json:"LoadBalancerNames,omitempty"`
	TargetGroupARNs         []*cloudformation.Literal                             `json:"TargetGroupARNs,omitempty"`
}

// RenderCloudformation is responsible for generating the cloudformation template
func (_ *AutoscalingGroup) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *AutoscalingGroup) error {
	cf := &cloudformationAutoscalingGroup{
		Name:    e.Name,
		MinSize: e.MinSize,
		MaxSize: e.MaxSize,
		MetricsCollection: []*cloudformationASGMetricsCollection{
			{
				Granularity: e.Granularity,
				Metrics:     aws.StringSlice(e.Metrics),
			},
		},
	}

	if e.LaunchConfiguration != nil {
		cf.LaunchConfigurationName = e.LaunchConfiguration.CloudformationLink()
	} else if e.UseMixedInstancesPolicy() {
		cf.MixedInstancesPolicy = &cloudformationMixedInstancesPolicy{
			LaunchTemplate: &cloudformationAutoscalingLaunchTemplate{
				LaunchTemplateSpecification: &cloudformationAutoscalingLaunchTemplateSpecification{
					LaunchTemplateId: e.LaunchTemplate.CloudformationLink(),
					Version:          e.LaunchTemplate.CloudformationVersion(),
				},
			},
			InstanceDistribution: &cloudformationAutoscalingInstanceDistribution{
				OnDemandAllocationStrategy:          e.MixedOnDemandAllocationStrategy,
				OnDemandBaseCapacity:                e.MixedOnDemandBase,
				OnDemandPercentageAboveBaseCapacity: e.MixedOnDemandAboveBase,
				SpotAllocationStrategy:              e.MixedSpotAllocationStrategy,
				SpotInstancePool:                    e.MixedSpotInstancePools,
				SpotMaxPrice:                        e.MixedSpotMaxPrice,
			},
		}

		for _, x := range e.MixedInstanceOverrides {
			cf.MixedInstancesPolicy.LaunchTemplate.Overrides = append(cf.MixedInstancesPolicy.LaunchTemplate.Overrides, &cloudformationAutoscalingLaunchTemplateOverride{InstanceType: fi.String(x)})
		}
	} else if e.LaunchTemplate != nil {
		cf.LaunchTemplate = &cloudformationAutoscalingLaunchTemplateSpecification{
			LaunchTemplateId: e.LaunchTemplate.CloudformationLink(),
			Version:          e.LaunchTemplate.CloudformationVersion(),
		}
	} else {
		return fmt.Errorf("could not find one of launch configuration, mixed instances policy, or launch template")
	}

	for _, s := range e.Subnets {
		cf.VPCZoneIdentifier = append(cf.VPCZoneIdentifier, s.CloudformationLink())
	}

	for _, k := range maps.SortedKeys(e.Tags) {
		v := e.Tags[k]
		cf.Tags = append(cf.Tags, &cloudformationASGTag{
			Key:               fi.String(k),
			Value:             fi.String(v),
			PropagateAtLaunch: fi.Bool(true),
		})
	}

	return t.RenderResource("AWS::AutoScaling::AutoScalingGroup", fi.StringValue(e.Name), cf)
}

// CloudformationLink is adds a reference
func (e *AutoscalingGroup) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::AutoScaling::AutoScalingGroup", fi.StringValue(e.Name))
}
