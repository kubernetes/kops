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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type LoadBalancerAccessLog struct {
	EmitInterval   *int64
	Enabled        *bool
	S3BucketName   *string
	S3BucketPrefix *string
}

func (_ *LoadBalancerAccessLog) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type terraformLoadBalancerAccessLog struct {
	EmitInterval   *int64  `json:"interval,omitempty" cty:"interval"`
	Enabled        *bool   `json:"enabled,omitempty" cty:"enabled"`
	S3BucketName   *string `json:"bucket,omitempty" cty:"bucket"`
	S3BucketPrefix *string `json:"bucket_prefix,omitempty" cty:"bucket_prefix"`
}

type cloudformationLoadBalancerAccessLog struct {
	EmitInterval   *int64  `json:"EmitInterval,omitempty"`
	Enabled        *bool   `json:"Enabled,omitempty"`
	S3BucketName   *string `json:"S3BucketName,omitempty"`
	S3BucketPrefix *string `json:"S3BucketPrefix,omitempty"`
}

//type LoadBalancerAdditionalAttribute struct {
//	Key   *string
//	Value *string
//}
//
//func (_ *LoadBalancerAdditionalAttribute) GetDependencies(tasks map[string]fi.Task) []fi.Task {
//	return nil
//}

type LoadBalancerConnectionDraining struct {
	Enabled *bool
	Timeout *int64
}

func (_ *LoadBalancerConnectionDraining) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type LoadBalancerCrossZoneLoadBalancing struct {
	Enabled *bool
}

func (_ *LoadBalancerCrossZoneLoadBalancing) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type LoadBalancerConnectionSettings struct {
	IdleTimeout *int64
}

func (_ *LoadBalancerConnectionSettings) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

func findELBAttributes(cloud awsup.AWSCloud, name string) (*elb.LoadBalancerAttributes, error) {
	request := &elb.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(name),
	}

	response, err := cloud.ELB().DescribeLoadBalancerAttributes(request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}

	return response.LoadBalancerAttributes, nil
}

func (_ *LoadBalancer) modifyLoadBalancerAttributes(t *awsup.AWSAPITarget, a, e, changes *LoadBalancer) error {
	if changes.AccessLog == nil &&
		changes.ConnectionDraining == nil &&
		changes.ConnectionSettings == nil &&
		changes.CrossZoneLoadBalancing == nil {
		klog.V(4).Infof("No LoadBalancerAttribute changes; skipping update")
		return nil
	}

	loadBalancerName := fi.StringValue(e.LoadBalancerName)

	request := &elb.ModifyLoadBalancerAttributesInput{}
	request.LoadBalancerName = e.LoadBalancerName
	request.LoadBalancerAttributes = &elb.LoadBalancerAttributes{}

	// Setting mandatory attributes to default values if empty
	request.LoadBalancerAttributes.AccessLog = &elb.AccessLog{}
	if e.AccessLog == nil || e.AccessLog.Enabled == nil {
		request.LoadBalancerAttributes.AccessLog.Enabled = fi.Bool(false)
	}
	request.LoadBalancerAttributes.ConnectionDraining = &elb.ConnectionDraining{}
	if e.ConnectionDraining == nil || e.ConnectionDraining.Enabled == nil {
		request.LoadBalancerAttributes.ConnectionDraining.Enabled = fi.Bool(false)
	}
	if e.ConnectionDraining == nil || e.ConnectionDraining.Timeout == nil {
		request.LoadBalancerAttributes.ConnectionDraining.Timeout = fi.Int64(300)
	}
	request.LoadBalancerAttributes.ConnectionSettings = &elb.ConnectionSettings{}
	if e.ConnectionSettings == nil || e.ConnectionSettings.IdleTimeout == nil {
		request.LoadBalancerAttributes.ConnectionSettings.IdleTimeout = fi.Int64(60)
	}
	request.LoadBalancerAttributes.CrossZoneLoadBalancing = &elb.CrossZoneLoadBalancing{}
	if e.CrossZoneLoadBalancing == nil || e.CrossZoneLoadBalancing.Enabled == nil {
		request.LoadBalancerAttributes.CrossZoneLoadBalancing.Enabled = fi.Bool(false)
	} else {
		request.LoadBalancerAttributes.CrossZoneLoadBalancing.Enabled = e.CrossZoneLoadBalancing.Enabled
	}

	// Setting non mandatory values only if not empty

	// We don't map AdditionalAttributes (yet)
	//if len(e.AdditionalAttributes) != 0 {
	//	var additionalAttributes []*elb.AdditionalAttribute
	//	for index, additionalAttribute := range e.AdditionalAttributes {
	//		additionalAttributes[index] = &elb.AdditionalAttribute{
	//			Key:   additionalAttribute.Key,
	//			Value: additionalAttribute.Value,
	//		}
	//	}
	//	request.LoadBalancerAttributes.AdditionalAttributes = additionalAttributes
	//}

	if e.AccessLog != nil && e.AccessLog.EmitInterval != nil {
		request.LoadBalancerAttributes.AccessLog.EmitInterval = e.AccessLog.EmitInterval
	}
	if e.AccessLog != nil && e.AccessLog.S3BucketName != nil {
		request.LoadBalancerAttributes.AccessLog.S3BucketName = e.AccessLog.S3BucketName
	}
	if e.AccessLog != nil && e.AccessLog.S3BucketPrefix != nil {
		request.LoadBalancerAttributes.AccessLog.S3BucketPrefix = e.AccessLog.S3BucketPrefix
	}
	if e.ConnectionDraining != nil && e.ConnectionDraining.Timeout != nil {
		request.LoadBalancerAttributes.ConnectionDraining.Timeout = e.ConnectionDraining.Timeout
	}
	if e.ConnectionSettings != nil && e.ConnectionSettings.IdleTimeout != nil {
		request.LoadBalancerAttributes.ConnectionSettings.IdleTimeout = e.ConnectionSettings.IdleTimeout
	}

	klog.V(2).Infof("Configuring ELB attributes for ELB %q", loadBalancerName)

	response, err := t.Cloud.ELB().ModifyLoadBalancerAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring ELB attributes for ELB %q: %v", loadBalancerName, err)
	}

	klog.V(4).Infof("modified ELB attributes for ELB %q, response %+v", loadBalancerName, response)

	return nil
}
