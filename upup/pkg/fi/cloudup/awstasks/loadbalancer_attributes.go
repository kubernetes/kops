/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type LoadBalancerAccessLog struct {
	EmitInterval   *int64
	Enabled        *bool
	S3BucketName   *string
	S3BucketPrefix *string
}

type LoadBalancerAdditionalAttribute struct {
	Key   *string
	Value *string
}

type LoadBalancerConnectionDraining struct {
	Enabled *bool
	Timeout *int64
}

type LoadBalancerCrossZoneLoadBalancing struct {
	Enabled *bool
}

//go:generate fitask -type=LoadBalancerAttributes
type LoadBalancerAttributes struct {
	Name         *string
	LoadBalancer *LoadBalancer

	AccessLog              *LoadBalancerAccessLog
	AdditionalAttributes   []*LoadBalancerAdditionalAttribute
	ConnectionDraining     *LoadBalancerConnectionDraining
	ConnectionSettings     *LoadBalancerConnectionSettings
	CrossZoneLoadBalancing *LoadBalancerCrossZoneLoadBalancing
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

func (e *LoadBalancerAttributes) Find(c *fi.Context) (*LoadBalancerAttributes, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	elbName := fi.StringValue(e.LoadBalancer.ID)

	lb, err := findELB(cloud, elbName)
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	lbAttributes, err := findELBAttributes(cloud, elbName)
	if err != nil {
		return nil, err
	}
	if lbAttributes == nil {
		return nil, nil
	}

	actual := &LoadBalancerAttributes{}
	actual.Name = e.Name
	actual.LoadBalancer = e.LoadBalancer

	if lbAttributes != nil {
		actual.AccessLog = &LoadBalancerAccessLog{
			EmitInterval:   lbAttributes.AccessLog.EmitInterval,
			Enabled:        lbAttributes.AccessLog.Enabled,
			S3BucketName:   lbAttributes.AccessLog.S3BucketName,
			S3BucketPrefix: lbAttributes.AccessLog.S3BucketPrefix,
		}
		var additionalAttributes []*LoadBalancerAdditionalAttribute
		for index, additionalAttribute := range lbAttributes.AdditionalAttributes {
			additionalAttributes[index] = &LoadBalancerAdditionalAttribute{
				Key:   additionalAttribute.Key,
				Value: additionalAttribute.Value,
			}
		}
		actual.AdditionalAttributes = additionalAttributes
		actual.ConnectionDraining = &LoadBalancerConnectionDraining{
			Enabled: lbAttributes.ConnectionDraining.Enabled,
			Timeout: lbAttributes.ConnectionDraining.Timeout,
		}
		actual.ConnectionSettings = &LoadBalancerConnectionSettings{
			Name:         e.Name,
			LoadBalancer: e.LoadBalancer,
			IdleTimeout:  lbAttributes.ConnectionSettings.IdleTimeout,
		}
		actual.CrossZoneLoadBalancing = &LoadBalancerCrossZoneLoadBalancing{
			Enabled: lbAttributes.CrossZoneLoadBalancing.Enabled,
		}
	}
	return actual, nil

}

func (e *LoadBalancerAttributes) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *LoadBalancerAttributes) CheckChanges(a, e, changes *LoadBalancerAttributes) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.LoadBalancer == nil {
			return fi.RequiredField("LoadBalancer")
		}
		if e.AccessLog != nil {
			if e.AccessLog.Enabled == nil {
				return fi.RequiredField("Acceslog.Enabled")
			}
			if *e.AccessLog.Enabled {
				if e.AccessLog.S3BucketName == nil {
					return fi.RequiredField("Acceslog.S3Bucket")
				}
			}
		}
		if e.ConnectionDraining != nil {
			if e.ConnectionDraining.Enabled == nil {
				return fi.RequiredField("ConnectionDraining.Enabled")
			}
		}
		if e.ConnectionSettings != nil {
			if e.ConnectionSettings.IdleTimeout == nil {
				return fi.RequiredField("ConnectionSettings.IdleTimeout")
			}
		}
		if e.CrossZoneLoadBalancing != nil {
			if e.CrossZoneLoadBalancing.Enabled == nil {
				return fi.RequiredField("CrossZoneLoadBalancing.Enabled")
			}
		}
	}
	return nil
}

func (_ *LoadBalancerAttributes) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LoadBalancerAttributes) error {
	var additionalAttributes []*elb.AdditionalAttribute
	for index, additionalAttribute := range e.AdditionalAttributes {
		additionalAttributes[index] = &elb.AdditionalAttribute{
			Key:   additionalAttribute.Key,
			Value: additionalAttribute.Value,
		}
	}

	request := &elb.ModifyLoadBalancerAttributesInput{
		LoadBalancerAttributes: &elb.LoadBalancerAttributes{
			AccessLog: &elb.AccessLog{
				EmitInterval:   e.AccessLog.EmitInterval,
				Enabled:        e.AccessLog.Enabled,
				S3BucketName:   e.AccessLog.S3BucketName,
				S3BucketPrefix: e.AccessLog.S3BucketPrefix,
			},
			AdditionalAttributes: additionalAttributes,
			ConnectionDraining: &elb.ConnectionDraining{
				Enabled: e.ConnectionDraining.Enabled,
				Timeout: e.ConnectionDraining.Timeout,
			},
			ConnectionSettings: &elb.ConnectionSettings{
				IdleTimeout: e.ConnectionSettings.IdleTimeout,
			},
			CrossZoneLoadBalancing: &elb.CrossZoneLoadBalancing{
				Enabled: e.CrossZoneLoadBalancing.Enabled,
			},
		},
		LoadBalancerName: e.LoadBalancer.Name,
	}

	glog.V(2).Infof("Configuring ELB attributes for ELB %q", *e.LoadBalancer.ID)

	_, err := t.Cloud.ELB().ModifyLoadBalancerAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring ELB attributes for ELB: %v", err)
	}

	return nil
}
