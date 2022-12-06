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
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	elbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/truncate"
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

		sess, err := session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return fmt.Errorf("error starting a new AWS session: %v", err)
		}

		client := ec2.New(sess, config)

		response, err := client.DescribeRegions(request)
		if err != nil {
			return fmt.Errorf("got an error while querying for valid regions (verify your AWS credentials?): %v", err)
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
	for _, subnet := range cluster.Spec.Networking.Subnets {
		if len(subnet.Zone) <= 2 {
			return "", fmt.Errorf("invalid AWS zone: %q in subnet %q", subnet.Zone, subnet.Name)
		}

		nodeZones[subnet.Zone] = true

		zoneRegion := subnet.Zone[:len(subnet.Zone)-1]
		if region != "" && zoneRegion != region {
			return "", fmt.Errorf("error Clusters cannot span multiple regions (found zone %q, but region is %q)", subnet.Zone, region)
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

// EC2TagSpecification converts a map of tags to an EC2 TagSpecification
func EC2TagSpecification(resourceType string, tags map[string]string) []*ec2.TagSpecification {
	if len(tags) == 0 {
		return nil
	}
	specification := &ec2.TagSpecification{
		ResourceType: aws.String(resourceType),
		Tags:         make([]*ec2.Tag, 0),
	}
	for k, v := range tags {
		specification.Tags = append(specification.Tags, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return []*ec2.TagSpecification{specification}
}

// ELBv2Tags converts a map of tags to ELBv2 Tags
func ELBv2Tags(tags map[string]string) []*elbv2.Tag {
	if len(tags) == 0 {
		return nil
	}
	elbv2Tags := make([]*elbv2.Tag, 0)
	for k, v := range tags {
		elbv2Tags = append(elbv2Tags, &elbv2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return elbv2Tags
}

// GetClusterName40 will attempt to calculate a meaningful cluster name with a max length of 40
func GetClusterName40(cluster string) string {
	return truncate.TruncateString(cluster, truncate.TruncateStringOptions{
		MaxLength: 40,
	})
}

// GetResourceName32 will attempt to calculate a meaningful name for a resource given a prefix
// Will never return a string longer than 32 chars
func GetResourceName32(cluster string, prefix string) string {
	s := prefix + "-" + strings.Replace(cluster, ".", "-", -1)

	// We always compute the hash and add it, lest we trick users into assuming that we never do this
	opt := truncate.TruncateStringOptions{
		MaxLength:     32,
		AlwaysAddHash: true,
		HashLength:    6,
	}
	return truncate.TruncateString(s, opt)
}

// GetTargetGroupNameFromARN will attempt to parse a target group ARN and return its name
func GetTargetGroupNameFromARN(targetGroupARN string) (string, error) {
	parsed, err := arn.Parse(targetGroupARN)
	if err != nil {
		return "", fmt.Errorf("error parsing target group ARN: %v", err)
	}
	resource := strings.Split(parsed.Resource, "/")
	if len(resource) != 3 || resource[0] != "targetgroup" {
		return "", fmt.Errorf("error parsing target group ARN resource: %q", parsed.Resource)
	}
	return resource[1], nil
}
