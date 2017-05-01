/*
Copyright 2016 The Kubernetes Authors.

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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

const CloudTagInstanceGroupRolePrefix = "k8s.io/role/"

//go:generate fitask -type=AutoscalingGroup
type AutoscalingGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	MinSize *int64
	MaxSize *int64
	Subnets []*Subnet
	Tags    map[string]string

	LaunchConfiguration *LaunchConfiguration
}

var _ fi.CompareWithID = &AutoscalingGroup{}

func (e *AutoscalingGroup) CompareWithID() *string {
	return e.Name
}

func findAutoscalingGroup(cloud awsup.AWSCloud, name string) (*autoscaling.Group, error) {
	request := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&name},
	}

	var found []*autoscaling.Group
	err := cloud.Autoscaling().DescribeAutoScalingGroupsPages(request, func(p *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) (shouldContinue bool) {
		for _, g := range p.AutoScalingGroups {
			// Check for "Delete in progress" (the only use of
			// .Status). We won't be able to update or create while
			// this is true, but filtering it out here makes the error
			// messages slightly clearer.
			if g.Status != nil {
				glog.Warningf("Skipping AutoScalingGroup %v: %v", *g.AutoScalingGroupName, *g.Status)
				continue
			}

			if aws.StringValue(g.AutoScalingGroupName) == name {
				found = append(found, g)
			} else {
				glog.Warningf("Got ASG with unexpected name %q", aws.StringValue(g.AutoScalingGroupName))
			}
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("error listing AutoscalingGroups: %v", err)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple AutoscalingGroups with name %q", name)
	}

	return found[0], nil
}

func (e *AutoscalingGroup) Find(c *fi.Context) (*AutoscalingGroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	g, err := findAutoscalingGroup(cloud, *e.Name)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}

	actual := &AutoscalingGroup{}
	actual.Name = g.AutoScalingGroupName
	actual.MinSize = g.MinSize
	actual.MaxSize = g.MaxSize

	if g.VPCZoneIdentifier != nil {
		subnets := strings.Split(*g.VPCZoneIdentifier, ",")
		for _, subnet := range subnets {
			actual.Subnets = append(actual.Subnets, &Subnet{ID: aws.String(subnet)})
		}
	}

	if len(g.Tags) != 0 {
		actual.Tags = make(map[string]string)
		for _, tag := range g.Tags {
			actual.Tags[*tag.Key] = *tag.Value
		}
	}

	if fi.StringValue(g.LaunchConfigurationName) == "" {
		glog.Warningf("autoscaling Group %q had no LaunchConfiguration", fi.StringValue(g.AutoScalingGroupName))
	} else {
		actual.LaunchConfiguration = &LaunchConfiguration{ID: g.LaunchConfigurationName}
	}

	if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
		actual.Subnets = e.Subnets
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *AutoscalingGroup) Run(c *fi.Context) error {
	c.Cloud.(awsup.AWSCloud).AddTags(e.Name, e.Tags)
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *AutoscalingGroup) CheckChanges(a, e, changes *AutoscalingGroup) error {
	if a != nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (e *AutoscalingGroup) buildTags(cloud fi.Cloud) map[string]string {
	tags := make(map[string]string)
	for k, v := range e.Tags {
		tags[k] = v
	}
	return tags
}

func (_ *AutoscalingGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *AutoscalingGroup) error {
	tags := []*autoscaling.Tag{}
	for k, v := range e.buildTags(t.Cloud) {
		tags = append(tags, &autoscaling.Tag{
			Key:               aws.String(k),
			Value:             aws.String(v),
			ResourceId:        e.Name,
			ResourceType:      aws.String("auto-scaling-group"),
			PropagateAtLaunch: aws.Bool(true),
		})
	}

	if a == nil {
		glog.V(2).Infof("Creating autoscaling Group with Name:%q", *e.Name)

		request := &autoscaling.CreateAutoScalingGroupInput{}
		request.AutoScalingGroupName = e.Name
		request.LaunchConfigurationName = e.LaunchConfiguration.ID
		request.MinSize = e.MinSize
		request.MaxSize = e.MaxSize

		var subnetIDs []string
		for _, s := range e.Subnets {
			subnetIDs = append(subnetIDs, *s.ID)
		}
		request.VPCZoneIdentifier = aws.String(strings.Join(subnetIDs, ","))

		request.Tags = tags

		_, err := t.Cloud.Autoscaling().CreateAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error creating AutoscalingGroup: %v", err)
		}
	} else {
		request := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: e.Name,
		}

		if changes.LaunchConfiguration != nil {
			request.LaunchConfigurationName = e.LaunchConfiguration.ID
			changes.LaunchConfiguration = nil
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
			var subnetIDs []string
			for _, s := range e.Subnets {
				subnetIDs = append(subnetIDs, *s.ID)
			}
			request.VPCZoneIdentifier = aws.String(strings.Join(subnetIDs, ","))
			changes.Subnets = nil
		}

		var tagsRequest *autoscaling.CreateOrUpdateTagsInput
		if changes.Tags != nil {
			tagsRequest = &autoscaling.CreateOrUpdateTagsInput{}
			tagsRequest.Tags = tags
			changes.Tags = nil
		}

		empty := &AutoscalingGroup{}
		if !reflect.DeepEqual(empty, changes) {
			glog.Warningf("cannot apply changes to AutoScalingGroup: %v", changes)
		}

		glog.V(2).Infof("Updating autoscaling group %s", *e.Name)

		_, err := t.Cloud.Autoscaling().UpdateAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error updating AutoscalingGroup: %v", err)
		}

		if tagsRequest != nil {
			_, err := t.Cloud.Autoscaling().CreateOrUpdateTags(tagsRequest)
			if err != nil {
				return fmt.Errorf("error updating AutoscalingGroup tags: %v", err)
			}
		}
	}

	// TODO: Use PropagateAtLaunch = false for tagging?

	return nil // We have
}

type terraformASGTag struct {
	Key               *string `json:"key"`
	Value             *string `json:"value"`
	PropagateAtLaunch *bool   `json:"propagate_at_launch"`
}
type terraformAutoscalingGroup struct {
	Name                    *string              `json:"name,omitempty"`
	LaunchConfigurationName *terraform.Literal   `json:"launch_configuration,omitempty"`
	MaxSize                 *int64               `json:"max_size,omitempty"`
	MinSize                 *int64               `json:"min_size,omitempty"`
	VPCZoneIdentifier       []*terraform.Literal `json:"vpc_zone_identifier,omitempty"`
	Tags                    []*terraformASGTag   `json:"tag,omitempty"`
}

func (_ *AutoscalingGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *AutoscalingGroup) error {
	tf := &terraformAutoscalingGroup{
		Name:                    e.Name,
		MinSize:                 e.MinSize,
		MaxSize:                 e.MaxSize,
		LaunchConfigurationName: e.LaunchConfiguration.TerraformLink(),
	}

	for _, s := range e.Subnets {
		tf.VPCZoneIdentifier = append(tf.VPCZoneIdentifier, s.TerraformLink())
	}

	tags := e.buildTags(t.Cloud)
	// Make sure we output in a stable order
	var tagKeys []string
	for k := range tags {
		tagKeys = append(tagKeys, k)
	}
	sort.Strings(tagKeys)
	for _, k := range tagKeys {
		v := tags[k]
		tf.Tags = append(tf.Tags, &terraformASGTag{
			Key:               fi.String(k),
			Value:             fi.String(v),
			PropagateAtLaunch: fi.Bool(true),
		})
	}

	if e.LaunchConfiguration != nil {
		// Create TF output variable with security group ids
		// This is in the launch configuration, but the ASG has the information about the instance group type

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

		if role != "" {
			for _, sg := range e.LaunchConfiguration.SecurityGroups {
				t.AddOutputVariableArray(role+"_security_group_ids", sg.TerraformLink())
			}
		}

		if role == "node" {
			for _, s := range e.Subnets {
				t.AddOutputVariableArray(role+"_subnet_ids", s.TerraformLink())
			}
		}
	}

	return t.RenderResource("aws_autoscaling_group", *e.Name, tf)
}

func (e *AutoscalingGroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_autoscaling_group", *e.Name, "id")
}

type cloudformationASGTag struct {
	Key               *string `json:"Key"`
	Value             *string `json:"Value"`
	PropagateAtLaunch *bool   `json:"PropagateAtLaunch"`
}
type cloudformationAutoscalingGroup struct {
	//Name                    *string              `json:"name,omitempty"`
	LaunchConfigurationName *cloudformation.Literal   `json:"LaunchConfigurationName,omitempty"`
	MaxSize                 *int64                    `json:"MaxSize,omitempty"`
	MinSize                 *int64                    `json:"MinSize,omitempty"`
	VPCZoneIdentifier       []*cloudformation.Literal `json:"VPCZoneIdentifier,omitempty"`
	Tags                    []*cloudformationASGTag   `json:"Tags,omitempty"`

	LoadBalancerNames []*cloudformation.Literal `json:"LoadBalancerNames,omitempty"`
}

func (_ *AutoscalingGroup) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *AutoscalingGroup) error {
	tf := &cloudformationAutoscalingGroup{
		//Name:                    e.Name,
		MinSize:                 e.MinSize,
		MaxSize:                 e.MaxSize,
		LaunchConfigurationName: e.LaunchConfiguration.CloudformationLink(),
	}

	for _, s := range e.Subnets {
		tf.VPCZoneIdentifier = append(tf.VPCZoneIdentifier, s.CloudformationLink())
	}

	tags := e.buildTags(t.Cloud)
	// Make sure we output in a stable order
	var tagKeys []string
	for k := range tags {
		tagKeys = append(tagKeys, k)
	}
	sort.Strings(tagKeys)
	for _, k := range tagKeys {
		v := tags[k]
		tf.Tags = append(tf.Tags, &cloudformationASGTag{
			Key:               fi.String(k),
			Value:             fi.String(v),
			PropagateAtLaunch: fi.Bool(true),
		})
	}

	return t.RenderResource("AWS::AutoScaling::AutoScalingGroup", *e.Name, tf)
}

func (e *AutoscalingGroup) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::AutoScaling::AutoScalingGroup", *e.Name)
}
