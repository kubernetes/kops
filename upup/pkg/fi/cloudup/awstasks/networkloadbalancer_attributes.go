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
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type NetworkLoadBalancerAccessLog struct {
	Enabled        *bool
	S3BucketName   *string
	S3BucketPrefix *string
}

func (_ *NetworkLoadBalancerAccessLog) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type terraformNetworkLoadBalancerAccessLog struct {
	Enabled        *bool   `cty:"enabled"`
	S3BucketName   *string `cty:"bucket"`
	S3BucketPrefix *string `cty:"prefix"`
}

func findNetworkLoadBalancerAttributes(cloud awsup.AWSCloud, LoadBalancerArn string) ([]*elbv2.LoadBalancerAttribute, error) {
	request := &elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(LoadBalancerArn),
	}

	response, err := cloud.ELBV2().DescribeLoadBalancerAttributes(request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}

	// we get back an array of attributes

	/*
	   Key *string `type:"string"`
	   Value *string `type:"string"`
	*/

	return response.Attributes, nil
}

func (_ *NetworkLoadBalancer) modifyLoadBalancerAttributes(t *awsup.AWSAPITarget, a, e, changes *NetworkLoadBalancer, loadBalancerArn string) error {
	if changes.CrossZoneLoadBalancing == nil && changes.AccessLog == nil {
		klog.V(4).Infof("No LoadBalancerAttribute changes; skipping update")
		return nil
	}

	loadBalancerName := fi.ValueOf(e.LoadBalancerName)

	request := &elbv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(loadBalancerArn),
	}

	var attributes []*elbv2.LoadBalancerAttribute

	attribute := &elbv2.LoadBalancerAttribute{}
	attribute.Key = aws.String("load_balancing.cross_zone.enabled")
	if e.CrossZoneLoadBalancing == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.CrossZoneLoadBalancing)))
	}
	attributes = append(attributes, attribute)

	if e.AccessLog != nil {
		attr := &elbv2.LoadBalancerAttribute{
			Key:   aws.String("access_logs.s3.enabled"),
			Value: aws.String(strconv.FormatBool(aws.BoolValue(e.AccessLog.Enabled))),
		}
		attributes = append(attributes, attr)
	}
	if e.AccessLog != nil && e.AccessLog.S3BucketName != nil {
		attr := &elbv2.LoadBalancerAttribute{
			Key:   aws.String("access_logs.s3.bucket"),
			Value: e.AccessLog.S3BucketName,
		}
		attributes = append(attributes, attr)
	}
	if e.AccessLog != nil && e.AccessLog.S3BucketPrefix != nil {
		attr := &elbv2.LoadBalancerAttribute{
			Key:   aws.String("access_logs.s3.prefix"),
			Value: e.AccessLog.S3BucketPrefix,
		}
		attributes = append(attributes, attr)
	}

	request.Attributes = attributes

	klog.V(2).Infof("Configuring NLB attributes for NLB %q", loadBalancerName)

	response, err := t.Cloud.ELBV2().ModifyLoadBalancerAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring NLB attributes for NLB %q: %v", loadBalancerName, err)
	}

	klog.V(4).Infof("modified NLB attributes for NLB %q, response %+v", loadBalancerName, response)

	return nil
}
