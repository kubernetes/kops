package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"reflect"
	"strings"
)

//go:generate fitask -type=AutoscalingGroup
type AutoscalingGroup struct {
	Name *string

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

func findAutoscalingGroup(cloud *awsup.AWSCloud, name string) (*autoscaling.Group, error) {
	request := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&name},
	}

	var found []*autoscaling.Group
	err := cloud.Autoscaling.DescribeAutoScalingGroupsPages(request, func(p *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) (shouldContinue bool) {
		for _, g := range p.AutoScalingGroups {
			if aws.StringValue(g.AutoScalingGroupName) == name {
				found = append(found, g)
			} else {
				glog.Warningf("Got ASG with unexpected name")
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
	cloud := c.Cloud.(*awsup.AWSCloud)

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

	if g.LaunchConfigurationName == nil {
		return nil, fmt.Errorf("autoscaling Group %q had no LaunchConfiguration", *actual.Name)
	}
	actual.LaunchConfiguration = &LaunchConfiguration{ID: g.LaunchConfigurationName}

	if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
		actual.Subnets = e.Subnets
	}

	return actual, nil
}

func (e *AutoscalingGroup) Run(c *fi.Context) error {
	c.Cloud.(*awsup.AWSCloud).AddTags(e.Name, e.Tags)
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

		tags := []*autoscaling.Tag{}
		for k, v := range e.buildTags(t.Cloud) {
			tags = append(tags, &autoscaling.Tag{
				Key:          aws.String(k),
				Value:        aws.String(v),
				ResourceId:   e.Name,
				ResourceType: aws.String("auto-scaling-group"),
			})
		}
		request.Tags = tags

		_, err := t.Cloud.Autoscaling.CreateAutoScalingGroup(request)
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

		empty := &AutoscalingGroup{}
		if !reflect.DeepEqual(empty, changes) {
			glog.Warningf("cannot apply changes to AutoScalingGroup: %v", changes)
		}

		glog.V(2).Infof("Updating autoscaling group %s", *e.Name)

		_, err := t.Cloud.Autoscaling.UpdateAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error updating AutoscalingGroup: %v", err)
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

	for k, v := range e.buildTags(t.Cloud) {
		tf.Tags = append(tf.Tags, &terraformASGTag{
			Key:               fi.String(k),
			Value:             fi.String(v),
			PropagateAtLaunch: fi.Bool(true),
		})
	}

	return t.RenderResource("aws_autoscaling_group", *e.Name, tf)
}

func (e *AutoscalingGroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_autoscaling_group", *e.Name, "id")
}
