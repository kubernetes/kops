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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

//go:generate fitask -type=LoadBalancerConnectionSettings
type LoadBalancerConnectionSettings struct {
	Name         *string
	LoadBalancer *LoadBalancer

	IdleTimeout *int64
}

func (e *LoadBalancerConnectionSettings) Find(c *fi.Context) (*LoadBalancerConnectionSettings, error) {
	fmt.Println("Finding ELB")
	cloud := c.Cloud.(awsup.AWSCloud)
	elbName := fi.StringValue(e.LoadBalancer.ID)
	fmt.Println("Name=", elbName)

	lb, err := findELB(cloud, elbName)
	if err != nil {
		return nil, err
	}
	fmt.Println("elb got=", lb)
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

	actual := &LoadBalancerConnectionSettings{}
	actual.Name = e.Name
	actual.LoadBalancer = e.LoadBalancer

	if lbAttributes != nil {
		actual.IdleTimeout = lbAttributes.ConnectionSettings.IdleTimeout
	}

	return actual, nil
}

func (e *LoadBalancerConnectionSettings) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *LoadBalancerConnectionSettings) CheckChanges(a, e, changes *LoadBalancerConnectionSettings) error {
	if a == nil {
		if e.LoadBalancer == nil {
			return fi.RequiredField("LoadBalancer")
		}
		if e.IdleTimeout == nil {
			return fi.RequiredField("IdleTimeout")
		}
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *LoadBalancerConnectionSettings) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LoadBalancerConnectionSettings) error {
	return nil
}
