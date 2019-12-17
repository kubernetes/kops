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

package awsup

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	elbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
)

// allRegions is the list of all regions; tests will set the values
var allRegions []*ec2.Region
var allRegionsMutex sync.Mutex

// isRegionCompiledInToAWSSDK checks if the specified region is in the AWS SDK
func isRegionCompiledInToAWSSDK(region string) bool {
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
	for _, p := range partitions {
		for _, r := range p.Regions() {
			if r.ID() == region {
				return true
			}
		}
	}
	return false
}

// ValidateRegion checks that an AWS region name is valid
func ValidateRegion(region string) error {
	if isRegionCompiledInToAWSSDK(region) {
		return nil
	}

	allRegionsMutex.Lock()
	defer allRegionsMutex.Unlock()

	if allRegions == nil {
		klog.V(2).Infof("Querying EC2 for all valid regions")

		request := &ec2.DescribeRegionsInput{}
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-east-1"
		}
		config := aws.NewConfig().WithRegion(awsRegion)
		config = config.WithCredentialsChainVerboseErrors(true)

		sess, err := session.NewSession(config)
		if err != nil {
			return fmt.Errorf("Error starting a new AWS session: %v", err)
		}

		client := ec2.New(sess, config)

		response, err := client.DescribeRegions(request)
		if err != nil {
			return fmt.Errorf("Got an error while querying for valid regions (verify your AWS credentials?): %v", err)
		}
		allRegions = response.Regions
	}

	for _, r := range allRegions {
		name := aws.StringValue(r.RegionName)
		if name == region {
			return nil
		}
	}

	if os.Getenv("SKIP_REGION_CHECK") != "" {
		klog.Infof("AWS region does not appear to be valid, but skipping because SKIP_REGION_CHECK is set")
		return nil
	}

	return fmt.Errorf("Region is not a recognized EC2 region: %q (check you have specified valid zones?)", region)
}

// FindRegion determines the region from the zones specified in the cluster
func FindRegion(cluster *kops.Cluster) (string, error) {
	region := ""

	nodeZones := make(map[string]bool)
	for _, subnet := range cluster.Spec.Subnets {
		if len(subnet.Zone) <= 2 {
			return "", fmt.Errorf("invalid AWS zone: %q in subnet %q", subnet.Zone, subnet.Name)
		}

		nodeZones[subnet.Zone] = true

		zoneRegion := subnet.Zone[:len(subnet.Zone)-1]
		if region != "" && zoneRegion != region {
			return "", fmt.Errorf("Clusters cannot span multiple regions (found zone %q, but region is %q)", subnet.Zone, region)
		}

		region = zoneRegion
	}

	return region, nil
}

// FindEC2Tag find the value of the tag with the specified key
func FindEC2Tag(tags []*ec2.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.StringValue(tag.Key) {
			return aws.StringValue(tag.Value), true
		}
	}
	return "", false
}

// FindASGTag find the value of the tag with the specified key
func FindASGTag(tags []*autoscaling.TagDescription, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.StringValue(tag.Key) {
			return aws.StringValue(tag.Value), true
		}
	}
	return "", false
}

// FindELBTag find the value of the tag with the specified key
func FindELBTag(tags []*elb.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.StringValue(tag.Key) {
			return aws.StringValue(tag.Value), true
		}
	}
	return "", false
}

// FindELBV2Tag find the value of the tag with the specified key
func FindELBV2Tag(tags []*elbv2.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.StringValue(tag.Key) {
			return aws.StringValue(tag.Value), true
		}
	}
	return "", false
}

// AWSErrorCode returns the aws error code, if it is an awserr.Error, otherwise ""
func AWSErrorCode(err error) string {
	if awsError, ok := err.(awserr.Error); ok {
		return awsError.Code()
	}
	return ""
}

// AWSErrorMessage returns the aws error message, if it is an awserr.Error, otherwise ""
func AWSErrorMessage(err error) string {
	if awsError, ok := err.(awserr.Error); ok {
		return awsError.Message()
	}
	return ""
}
