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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type ClassicLoadBalancerAccessLog struct {
	EmitInterval   *int32
	Enabled        *bool
	S3BucketName   *string
	S3BucketPrefix *string
}

func (_ *ClassicLoadBalancerAccessLog) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

type terraformLoadBalancerAccessLog struct {
	EmitInterval   *int32  `cty:"interval"`
	Enabled        *bool   `cty:"enabled"`
	S3BucketName   *string `cty:"bucket"`
	S3BucketPrefix *string `cty:"bucket_prefix"`
}

//type LoadBalancerAdditionalAttribute struct {
//	Key   *string
//	Value *string
//}
//
//func (_ *ClassicLoadBalancerAdditionalAttribute) GetDependencies(tasks map[string]fi.Task) []fi.Task {
//	return nil
//}

type ClassicLoadBalancerConnectionDraining struct {
	Enabled *bool
	Timeout *int32
}

func (_ *ClassicLoadBalancerConnectionDraining) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

type ClassicLoadBalancerCrossZoneLoadBalancing struct {
	Enabled *bool
}

func (_ *ClassicLoadBalancerCrossZoneLoadBalancing) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

type ClassicLoadBalancerConnectionSettings struct {
	IdleTimeout *int32
}

func (_ *ClassicLoadBalancerConnectionSettings) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

func findELBAttributes(ctx context.Context, cloud awsup.AWSCloud, name string) (*elbtypes.LoadBalancerAttributes, error) {
	request := &elb.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(name),
	}

	response, err := cloud.ELB().DescribeLoadBalancerAttributes(ctx, request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}

	return response.LoadBalancerAttributes, nil
}

func (_ *ClassicLoadBalancer) modifyLoadBalancerAttributes(t *awsup.AWSAPITarget, a, e, changes *ClassicLoadBalancer) error {
	if changes.AccessLog == nil &&
		changes.ConnectionDraining == nil &&
		changes.ConnectionSettings == nil &&
		changes.CrossZoneLoadBalancing == nil {
		klog.V(4).Infof("No LoadBalancerAttribute changes; skipping update")
		return nil
	}
	ctx := context.TODO()

	loadBalancerName := fi.ValueOf(e.LoadBalancerName)

	request := &elb.ModifyLoadBalancerAttributesInput{}
	request.LoadBalancerName = e.LoadBalancerName
	request.LoadBalancerAttributes = &elbtypes.LoadBalancerAttributes{}

	// Setting mandatory attributes to default values if empty
	request.LoadBalancerAttributes.AccessLog = &elbtypes.AccessLog{}
	if e.AccessLog == nil || e.AccessLog.Enabled == nil {
		request.LoadBalancerAttributes.AccessLog.Enabled = false
	}
	request.LoadBalancerAttributes.ConnectionDraining = &elbtypes.ConnectionDraining{}
	if e.ConnectionDraining == nil || e.ConnectionDraining.Enabled == nil {
		request.LoadBalancerAttributes.ConnectionDraining.Enabled = false
	}
	if e.ConnectionDraining == nil || e.ConnectionDraining.Timeout == nil {
		request.LoadBalancerAttributes.ConnectionDraining.Timeout = aws.Int32(300)
	}
	request.LoadBalancerAttributes.ConnectionSettings = &elbtypes.ConnectionSettings{}
	if e.ConnectionSettings == nil || e.ConnectionSettings.IdleTimeout == nil {
		request.LoadBalancerAttributes.ConnectionSettings.IdleTimeout = aws.Int32(60)
	}
	request.LoadBalancerAttributes.CrossZoneLoadBalancing = &elbtypes.CrossZoneLoadBalancing{}
	if e.CrossZoneLoadBalancing == nil || e.CrossZoneLoadBalancing.Enabled == nil {
		request.LoadBalancerAttributes.CrossZoneLoadBalancing.Enabled = false
	} else {
		request.LoadBalancerAttributes.CrossZoneLoadBalancing.Enabled = aws.ToBool(e.CrossZoneLoadBalancing.Enabled)
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

	if e.AccessLog != nil && e.AccessLog.Enabled != nil {
		request.LoadBalancerAttributes.AccessLog.Enabled = aws.ToBool(e.AccessLog.Enabled)
	}
	if e.AccessLog != nil && e.AccessLog.EmitInterval != nil {
		request.LoadBalancerAttributes.AccessLog.EmitInterval = e.AccessLog.EmitInterval
	}
	if e.AccessLog != nil && e.AccessLog.S3BucketName != nil {
		request.LoadBalancerAttributes.AccessLog.S3BucketName = e.AccessLog.S3BucketName
	}
	if e.AccessLog != nil && e.AccessLog.S3BucketPrefix != nil {
		request.LoadBalancerAttributes.AccessLog.S3BucketPrefix = e.AccessLog.S3BucketPrefix
	}
	if e.ConnectionDraining != nil && e.ConnectionDraining.Enabled != nil {
		request.LoadBalancerAttributes.ConnectionDraining.Enabled = aws.ToBool(e.ConnectionDraining.Enabled)
	}
	if e.ConnectionDraining != nil && e.ConnectionDraining.Timeout != nil {
		request.LoadBalancerAttributes.ConnectionDraining.Timeout = e.ConnectionDraining.Timeout
	}
	if e.ConnectionSettings != nil && e.ConnectionSettings.IdleTimeout != nil {
		request.LoadBalancerAttributes.ConnectionSettings.IdleTimeout = e.ConnectionSettings.IdleTimeout
	}

	klog.V(2).Infof("Configuring ELB attributes for ELB %q", loadBalancerName)

	response, err := t.Cloud.ELB().ModifyLoadBalancerAttributes(ctx, request)
	if err != nil {
		return fmt.Errorf("error configuring ELB attributes for ELB %q: %v", loadBalancerName, err)
	}

	klog.V(4).Infof("modified ELB attributes for ELB %q, response %+v", loadBalancerName, response)

	return nil
}
