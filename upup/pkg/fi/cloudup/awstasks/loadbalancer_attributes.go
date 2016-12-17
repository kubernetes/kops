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
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

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

type LoadBalancerAccessLog struct {
	EmitInterval   *int64
	Enabled        *bool
	S3BucketName   *string
	S3BucketPrefix *string
}

func (_ *LoadBalancerAccessLog) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type LoadBalancerAdditionalAttribute struct {
	Key   *string
	Value *string
}

func (_ *LoadBalancerAdditionalAttribute) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

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

func (e *LoadBalancerAttributes) Find(c *fi.Context) (*LoadBalancerAttributes, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	elbName := fi.StringValue(e.LoadBalancer.ID)

	lb, err := findLoadBalancer(cloud, elbName)
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
		actual.AccessLog = &LoadBalancerAccessLog{}
		if lbAttributes.AccessLog.EmitInterval != nil {
			actual.AccessLog.EmitInterval = lbAttributes.AccessLog.EmitInterval
		}
		if lbAttributes.AccessLog.Enabled != nil {
			actual.AccessLog.Enabled = lbAttributes.AccessLog.Enabled
		}
		if lbAttributes.AccessLog.S3BucketName != nil {
			actual.AccessLog.S3BucketName = lbAttributes.AccessLog.S3BucketName
		}
		if lbAttributes.AccessLog.S3BucketPrefix != nil {
			actual.AccessLog.S3BucketPrefix = lbAttributes.AccessLog.S3BucketPrefix
		}

		var additionalAttributes []*LoadBalancerAdditionalAttribute
		for index, additionalAttribute := range lbAttributes.AdditionalAttributes {
			additionalAttributes[index] = &LoadBalancerAdditionalAttribute{
				Key:   additionalAttribute.Key,
				Value: additionalAttribute.Value,
			}
		}
		actual.AdditionalAttributes = additionalAttributes

		actual.ConnectionDraining = &LoadBalancerConnectionDraining{}
		if lbAttributes.ConnectionDraining.Enabled != nil {
			actual.ConnectionDraining.Enabled = lbAttributes.ConnectionDraining.Enabled
		}
		if lbAttributes.ConnectionDraining.Timeout != nil {
			actual.ConnectionDraining.Timeout = lbAttributes.ConnectionDraining.Timeout
		}

		actual.ConnectionSettings = &LoadBalancerConnectionSettings{}
		if lbAttributes.ConnectionSettings.IdleTimeout != nil {
			actual.ConnectionSettings.IdleTimeout = lbAttributes.ConnectionSettings.IdleTimeout
		}

		actual.CrossZoneLoadBalancing = &LoadBalancerCrossZoneLoadBalancing{}
		if lbAttributes.CrossZoneLoadBalancing.Enabled != nil {
			actual.CrossZoneLoadBalancing.Enabled = lbAttributes.CrossZoneLoadBalancing.Enabled
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
		//if e.ConnectionSettings != nil {
		//	if e.ConnectionSettings.IdleTimeout == nil {
		//		return fi.RequiredField("ConnectionSettings.IdleTimeout")
		//	}
		//}
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
	request := &elb.ModifyLoadBalancerAttributesInput{}
	request.LoadBalancerName = e.LoadBalancer.ID
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
	}

	// Setting non mandatory values only if not empty
	if len(additionalAttributes) != 0 {
		request.LoadBalancerAttributes.AdditionalAttributes = additionalAttributes
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
	if e.ConnectionDraining != nil && e.ConnectionDraining.Timeout != nil {
		request.LoadBalancerAttributes.ConnectionDraining.Timeout = e.ConnectionDraining.Timeout
	}
	if e.ConnectionSettings != nil && e.ConnectionSettings.IdleTimeout != nil {
		request.LoadBalancerAttributes.ConnectionSettings.IdleTimeout = e.ConnectionSettings.IdleTimeout
	}

	glog.V(2).Infof("Configuring ELB attributes for ELB %q", *e.LoadBalancer.ID)

	_, err := t.Cloud.ELB().ModifyLoadBalancerAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring ELB attributes for ELB: %v", err)
	}

	return nil
}

func (_ *LoadBalancerAttributes) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancerAttributes) error {
	glog.Warning("LoadBalancerAttributes RenderTerraform not implemented")
	return nil
}
