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
	EmitInterval   *int64
	Enabled        *bool   //TODO: change to S3Enabled
	S3BucketName   *string //TODO: change to S3Bucket
	S3BucketPrefix *string //TODO: change to S3Prefix
}

func (_ *NetworkLoadBalancerAccessLog) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

//type LoadBalancerAdditionalAttribute struct {
//	Key   *string
//	Value *string
//}
//
//func (_ *LoadBalancerAdditionalAttribute) GetDependencies(tasks map[string]fi.Task) []fi.Task {
//	return nil
//}

type TargetGroupProxyProtocolV2 struct {
	Enabled *bool
}

func (_ *TargetGroupProxyProtocolV2) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type TargetGroupStickiness struct {
	Enabled *bool
	Type    *string
}

func (_ *TargetGroupStickiness) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type TargetGroupDeregistrationDelay struct {
	TimeoutSeconds *int64
}

func (_ *TargetGroupDeregistrationDelay) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type NetworkLoadBalancerCrossZoneLoadBalancing struct {
	Enabled *bool
}

func (_ *NetworkLoadBalancerCrossZoneLoadBalancing) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type NetworkLoadBalancerDeletionProtection struct {
	Enabled *bool
}

func (_ *NetworkLoadBalancerDeletionProtection) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
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

	//we get back an array of attributes

	/*
	   Key *string `type:"string"`
	   Value *string `type:"string"`
	*/

	return response.Attributes, nil
}

func findTargetGroupAttributes(cloud awsup.AWSCloud, TargetGroupArn string) ([]*elbv2.TargetGroupAttribute, error) {

	request := &elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(TargetGroupArn),
	}

	response, err := cloud.ELBV2().DescribeTargetGroupAttributes(request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}

	//we get back an array of attributes

	/*
	   Key *string `type:"string"`
	   Value *string `type:"string"`
	*/

	return response.Attributes, nil
}

func (_ *NetworkLoadBalancer) modifyLoadBalancerAttributes(t *awsup.AWSAPITarget, a, e, changes *NetworkLoadBalancer, loadBalancerArn string) error {
	if changes.AccessLog == nil &&
		changes.DeletionProtection == nil &&
		changes.CrossZoneLoadBalancing == nil {
		klog.V(4).Infof("No LoadBalancerAttribute changes; skipping update")
		return nil
	}

	loadBalancerName := fi.StringValue(e.LoadBalancerName)

	request := &elbv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(loadBalancerArn),
	}

	var attributes []*elbv2.LoadBalancerAttribute

	attribute := &elbv2.LoadBalancerAttribute{}
	attribute.Key = aws.String("access_logs.s3.enabled")
	if e.AccessLog == nil || e.AccessLog.Enabled == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.AccessLog.Enabled)))
	}
	attributes = append(attributes, attribute)

	/*if *e.AccessLog.Enabled { //TODO: Should we capture these? -- These are not settable from spec.
		attribute = &elbv2.LoadBalancerAttribute{}
		attribute.Key = aws.String("access_logs.s3.bucket")
		if e.AccessLog == nil || e.AccessLog.S3BucketName == nil {
			attribute.Value = aws.String("") //TOOD: ValidationError: The value of 'access_logs.s3.bucket' cannot be empty
		} else {
			attribute.Value = e.AccessLog.S3BucketName
			attributes = append(attributes, attribute)
		}

		attribute = &elbv2.LoadBalancerAttribute{}
		attribute.Key = aws.String("access_logs.s3.prefix")
		if e.AccessLog == nil || e.AccessLog.S3BucketPrefix == nil {
			attribute.Value = aws.String("") //TODO: ValidationError: The value of 'access_logs.s3.bucket' cannot be empty
		} else {
			attribute.Value = e.AccessLog.S3BucketPrefix
			attributes = append(attributes, attribute)
		}
	}*/

	attribute = &elbv2.LoadBalancerAttribute{}
	attribute.Key = aws.String("deletion_protection.enabled")
	if e.DeletionProtection == nil || e.DeletionProtection.Enabled == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.DeletionProtection.Enabled)))
	}
	attributes = append(attributes, attribute)

	attribute = &elbv2.LoadBalancerAttribute{}
	attribute.Key = aws.String("load_balancing.cross_zone.enabled")
	if e.CrossZoneLoadBalancing == nil || e.CrossZoneLoadBalancing.Enabled == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.CrossZoneLoadBalancing.Enabled)))
	}
	attributes = append(attributes, attribute)

	request.Attributes = attributes

	klog.V(2).Infof("Configuring NLB attributes for NLB %q", loadBalancerName)

	response, err := t.Cloud.ELBV2().ModifyLoadBalancerAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring NLB attributes for NLB %q: %v", loadBalancerName, err)
	}

	klog.V(4).Infof("modified NLB attributes for NLB %q, response %+v", loadBalancerName, response)

	return nil
}

func (_ *NetworkLoadBalancer) modifyTargetGroupAttributes(t *awsup.AWSAPITarget, a, e, changes *NetworkLoadBalancer, targetGroupArn string) error {
	if changes.ProxyProtocolV2 == nil &&
		changes.Stickiness == nil &&
		changes.DeregistationDelay == nil {
		klog.V(4).Infof("No TargetGroup changes; skipping update")
		return nil
	}

	loadBalancerName := fi.StringValue(e.LoadBalancerName)
	request := &elbv2.ModifyTargetGroupAttributesInput{
		TargetGroupArn: aws.String(targetGroupArn),
	}

	var attributes []*elbv2.TargetGroupAttribute

	attribute := &elbv2.TargetGroupAttribute{}
	attribute.Key = aws.String("deregistration_delay.timeout_seconds")
	if e.DeregistationDelay == nil || e.DeregistationDelay.TimeoutSeconds == nil {
		attribute.Value = aws.String("300")
	} else {
		attribute.Value = aws.String(strconv.Itoa(int(*e.DeregistationDelay.TimeoutSeconds)))
	}
	attributes = append(attributes, attribute)

	attribute = &elbv2.TargetGroupAttribute{}
	attribute.Key = aws.String("stickiness.enabled")
	if e.Stickiness == nil || e.Stickiness.Enabled == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.Stickiness.Enabled)))
	}
	attributes = append(attributes, attribute)

	attribute = &elbv2.TargetGroupAttribute{}
	attribute.Key = aws.String("stickiness.type ")
	attribute.Value = aws.String("source_ip") //TODO: can we set this even if enabled = false?
	attributes = append(attributes, attribute)

	attribute = &elbv2.TargetGroupAttribute{}
	attribute.Key = aws.String("proxy_protocol_v2.enabled")
	if e.ProxyProtocolV2 == nil || e.ProxyProtocolV2.Enabled == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.ProxyProtocolV2.Enabled)))
	}
	attributes = append(attributes, attribute)

	request.Attributes = attributes

	responseTG, err := t.Cloud.ELBV2().ModifyTargetGroupAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring NLB target group attributes for NLB %q: %v", loadBalancerName, err)
	}

	klog.V(4).Infof("modified NLB target group attributes for NLB %q, response %+v", loadBalancerName, responseTG)

	return nil
}
