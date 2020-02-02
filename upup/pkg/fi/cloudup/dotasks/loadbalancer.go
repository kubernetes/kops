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

package dotasks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/digitalocean/godo"

	"k8s.io/klog"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
)

//go:generate fitask -type=LoadBalancer
type LoadBalancer struct {
	Name      *string
	ID        *string
	Lifecycle *fi.Lifecycle

	Region     *string
	DropletTag *string
	IPAddress  *string
}

var _ fi.CompareWithID = &LoadBalancer{}

func (lb *LoadBalancer) CompareWithID() *string {
	return lb.ID
}

func (lb *LoadBalancer) Find(c *fi.Context) (*LoadBalancer, error) {
	if fi.StringValue(lb.ID) == "" {
		// Loadbalancer = nil if not found
		return nil, nil
	}

	cloud := c.Cloud.(*digitalocean.Cloud)
	lbService := cloud.LoadBalancers()
	loadbalancer, _, err := lbService.Get(context.TODO(), fi.StringValue(lb.ID))

	if err != nil {
		return nil, fmt.Errorf("load balancer service get request returned error %v", err)
	}

	return &LoadBalancer{
		Name:      fi.String(loadbalancer.Name),
		ID:        fi.String(loadbalancer.ID),
		Lifecycle: lb.Lifecycle,
		Region:    fi.String(loadbalancer.Region.Slug),
	}, nil
}

func (lb *LoadBalancer) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(lb, c)
}

func (_ *LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Region == nil {
			return fi.RequiredField("Region")
		}
	}
	return nil
}

func (_ *LoadBalancer) RenderDO(t *do.DOAPITarget, a, e, changes *LoadBalancer) error {

	Rules := []godo.ForwardingRule{
		{
			EntryProtocol:  "https",
			EntryPort:      443,
			TargetProtocol: "https",
			TargetPort:     443,
			TlsPassthrough: true,
		},
		{
			EntryProtocol:  "http",
			EntryPort:      80,
			TargetProtocol: "http",
			TargetPort:     80,
		},
	}

	HealthCheck := &godo.HealthCheck{
		Protocol:               "tcp",
		Port:                   443,
		Path:                   "",
		CheckIntervalSeconds:   60,
		ResponseTimeoutSeconds: 5,
		UnhealthyThreshold:     3,
		HealthyThreshold:       5,
	}

	klog.V(10).Infof("Creating load balancer for DO")

	loadBalancerService := t.Cloud.LoadBalancers()
	loadbalancer, _, err := loadBalancerService.Create(context.TODO(), &godo.LoadBalancerRequest{
		Name:            fi.StringValue(e.Name),
		Region:          fi.StringValue(e.Region),
		Tag:             fi.StringValue(e.DropletTag),
		ForwardingRules: Rules,
		HealthCheck:     HealthCheck,
	})

	if err != nil {
		klog.Errorf("Error creating load balancer with Name=%s, Error=%v", fi.StringValue(e.Name), err)
		return err
	}

	e.ID = fi.String(loadbalancer.ID)
	e.IPAddress = fi.String(loadbalancer.IP) // This will be empty on create, but will be filled later on FindIPAddress invokation.

	return nil
}

func (lb *LoadBalancer) FindIPAddress(c *fi.Context) (*string, error) {
	cloud := c.Cloud.(*digitalocean.Cloud)
	loadBalancerService := cloud.LoadBalancers()

	klog.V(10).Infof("Find IP address for load balancer ID=%s", fi.StringValue(lb.ID))
	loadBalancer, _, err := loadBalancerService.Get(context.TODO(), fi.StringValue(lb.ID))
	if err != nil {
		klog.Errorf("Error fetching load balancer with Name=%s", fi.StringValue(lb.Name))
		return nil, err
	}

	address := loadBalancer.IP

	if isIPv4(address) {
		klog.V(10).Infof("load balancer address=%s", address)
		return &address, nil
	}

	const lbWaitTime = 10 * time.Second
	klog.Warningf("IP address for LB %s not yet available -- sleeping %s", fi.StringValue(lb.Name), lbWaitTime)
	time.Sleep(lbWaitTime)

	return nil, errors.New("IP Address is still empty.")
}

func isIPv4(host string) bool {

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.To4() != nil
}
