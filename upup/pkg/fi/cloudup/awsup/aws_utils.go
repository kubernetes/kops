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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/truncate"
)

// allRegions is the list of all regions; tests will set the values
var allRegions []ec2types.Region
var allRegionsMutex sync.Mutex

// ValidateRegion checks that an AWS region name is valid
func ValidateRegion(ctx context.Context, region string) error {
	allRegionsMutex.Lock()
	defer allRegionsMutex.Unlock()

	if allRegions == nil {
		klog.V(2).Infof("Querying EC2 for all valid regions")

		request := &ec2.DescribeRegionsInput{}
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-east-1"
		}
		cfg, err := loadAWSConfig(ctx, awsRegion)
		if err != nil {
			return fmt.Errorf("error loading AWS config: %v", err)
		}

		if err != nil {
			return fmt.Errorf("error starting a new AWS session: %v", err)
		}

		client := ec2.NewFromConfig(cfg)

		response, err := client.DescribeRegions(ctx, request)
		if err != nil {
			return fmt.Errorf("got an error while querying for valid regions (verify your AWS credentials?): %v", err)
		}
		allRegions = response.Regions
	}

	for _, r := range allRegions {
		name := aws.ToString(r.RegionName)
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
func FindEC2Tag(tags []ec2types.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.ToString(tag.Key) {
			return aws.ToString(tag.Value), true
		}
	}
	return "", false
}

// FindASGTag find the value of the tag with the specified key
func FindASGTag(tags []autoscalingtypes.TagDescription, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.ToString(tag.Key) {
			return aws.ToString(tag.Value), true
		}
	}
	return "", false
}

// FindELBTag find the value of the tag with the specified key
func FindELBTag(tags []elbtypes.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.ToString(tag.Key) {
			return aws.ToString(tag.Value), true
		}
	}
	return "", false
}

// FindELBV2Tag find the value of the tag with the specified key
func FindELBV2Tag(tags []elbv2types.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if key == aws.ToString(tag.Key) {
			return aws.ToString(tag.Value), true
		}
	}
	return "", false
}

// AWSErrorCode returns the aws error code, if it is an awserr.Error or smithy.APIError, otherwise ""
func AWSErrorCode(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode()
	}
	return ""
}

// AWSErrorMessage returns the aws error message, if it is an awserr.Error or smithy.APIError, otherwise ""
func AWSErrorMessage(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorMessage()
	}
	return ""
}

// EC2TagSpecification converts a map of tags to an EC2 TagSpecification
func EC2TagSpecification(resourceType ec2types.ResourceType, tags map[string]string) []ec2types.TagSpecification {
	if len(tags) == 0 {
		return nil
	}
	specification := ec2types.TagSpecification{
		ResourceType: resourceType,
		Tags:         make([]ec2types.Tag, 0),
	}
	for k, v := range tags {
		specification.Tags = append(specification.Tags, ec2types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return []ec2types.TagSpecification{specification}
}

// ELBv2Tags converts a map of tags to ELBv2 Tags
func ELBv2Tags(tags map[string]string) []elbv2types.Tag {
	if len(tags) == 0 {
		return nil
	}
	elbv2Tags := make([]elbv2types.Tag, 0)
	for k, v := range tags {
		elbv2Tags = append(elbv2Tags, elbv2types.Tag{
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

// NameForExternalTargetGroup will attempt to calculate a meaningful name for a target group given an ARN.
func NameForExternalTargetGroup(targetGroupARN string) (string, error) {
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

func IsIAMNoSuchEntityException(err error) bool {
	if err == nil {
		return false
	}
	var nse *iamtypes.NoSuchEntityException
	return errors.As(err, &nse)
}
