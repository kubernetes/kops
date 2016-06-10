package kutil

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
)

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
		response, err := cloud.Autoscaling.DescribeTags(request)
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling cluster tags: %v", err)
		}

		for _, t := range response.Tags {
			switch *t.ResourceType {
			case "auto-scaling-group":
				asgNames = append(asgNames, t.ResourceId)
			default:
				glog.Warningf("Unknown resource type: %v", *t.ResourceType)

			}
		}
	}

	if len(asgNames) != 0 {
		request := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: asgNames,
		}
		response, err := cloud.Autoscaling.DescribeAutoScalingGroups(request)
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling groups: %v", err)
		}

		for _, asg := range response.AutoScalingGroups {
			if !matchesAsgTags(tags, asg.Tags) {
				// We used an inexact filter above
				continue
			}
			asgs = append(asgs, asg)
		}
	}

	return asgs, nil
}
