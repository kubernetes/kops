package kutil

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// findAutoscalingGroups finds autoscaling groups matching the specified tags
// This isn't entirely trivial because autoscaling doesn't let us filter with as much precision as we wouldlike
func findAutoscalingGroups(cloud *awsup.AWSCloud, tags map[string]string) ([]*autoscaling.Group, error) {
	var asgs []*autoscaling.Group

	glog.V(2).Infof("Listing all Autoscaling groups matching cluster tags")
	var asgNames []*string
	{
		var asFilters []*autoscaling.Filter
		for _, v := range tags {
			// Not an exact match, but likely the best we can do
			asFilters = append(asFilters, &autoscaling.Filter{
				Name:   aws.String("value"),
				Values: []*string{aws.String(v)},
			})
		}
		request := &autoscaling.DescribeTagsInput{
			Filters: asFilters,
		}

		err := cloud.Autoscaling.DescribeTagsPages(request, func(p *autoscaling.DescribeTagsOutput, lastPage bool) bool {
			for _, t := range p.Tags {
				switch *t.ResourceType {
				case "auto-scaling-group":
					asgNames = append(asgNames, t.ResourceId)
				default:
					glog.Warningf("Unknown resource type: %v", *t.ResourceType)

				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling cluster tags: %v", err)
		}
	}

	if len(asgNames) != 0 {
		request := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: asgNames,
		}
		err := cloud.Autoscaling.DescribeAutoScalingGroupsPages(request, func(p *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			for _, asg := range p.AutoScalingGroups {
				if !matchesAsgTags(tags, asg.Tags) {
					// We used an inexact filter above
					continue
				}
				asgs = append(asgs, asg)
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling groups: %v", err)
		}

	}

	return asgs, nil
}

func findAutoscalingLaunchConfiguration(cloud *awsup.AWSCloud, name string) (*autoscaling.LaunchConfiguration, error) {
	glog.V(2).Infof("Retrieving Autoscaling LaunchConfigurations %q", name)

	var results []*autoscaling.LaunchConfiguration

	request := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{&name},
	}
	err := cloud.Autoscaling.DescribeLaunchConfigurationsPages(request, func(p *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		for _, t := range p.LaunchConfigurations {
			results = append(results, t)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing autoscaling LaunchConfigurations: %v", err)
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("Found multiple LaunchConfigurations with name %q", name)
	}
	return results[0], nil
}
