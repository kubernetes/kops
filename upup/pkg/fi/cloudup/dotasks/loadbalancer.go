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
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/util/pkg/vfs"
)

// +kops:fitask
type LoadBalancer struct {
	Name      *string
	ID        *string
	Lifecycle fi.Lifecycle

	Region       *string
	DropletTag   *string
	IPAddress    *string
	VPCUUID      *string
	VPCName      *string
	NetworkCIDR  *string
	ForAPIServer bool
}

var readBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   2,
	Jitter:   0.1,
	Steps:    10,
}

var (
	_ fi.CompareWithID = &LoadBalancer{}
	_ fi.HasAddress    = &LoadBalancer{}
)

func (lb *LoadBalancer) CompareWithID() *string {
	return lb.ID
}

func (lb *LoadBalancer) Find(c *fi.Context) (*LoadBalancer, error) {
	klog.V(10).Infof("load balancer FIND - ID=%s, name=%s", fi.StringValue(lb.ID), fi.StringValue(lb.Name))
	if fi.StringValue(lb.ID) == "" {
		// Loadbalancer = nil if not found
		return nil, nil
	}

	cloud := c.Cloud.(do.DOCloud)
	lbService := cloud.LoadBalancersService()
	loadbalancer, _, err := lbService.Get(context.TODO(), fi.StringValue(lb.ID))
	if err != nil {
		return nil, fmt.Errorf("load balancer service get request returned error %v", err)
	}

	return &LoadBalancer{
		Name:    fi.String(loadbalancer.Name),
		ID:      fi.String(loadbalancer.ID),
		Region:  fi.String(loadbalancer.Region.Slug),
		VPCUUID: fi.String(loadbalancer.VPCUUID),

		// Ignore system fields
		Lifecycle:    lb.Lifecycle,
		ForAPIServer: lb.ForAPIServer,
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

	// check if load balancer exist.
	loadBalancers, err := t.Cloud.GetAllLoadBalancers()
	if err != nil {
		return fmt.Errorf("LoadBalancers.List returned error: %v", err)
	}

	for _, loadbalancer := range loadBalancers {
		klog.V(10).Infof("load balancer retrieved=%s, e.Name=%s", loadbalancer.Name, fi.StringValue(e.Name))
		if strings.Contains(loadbalancer.Name, fi.StringValue(e.Name)) {
			// load balancer already exists.
			e.ID = fi.String(loadbalancer.ID)
			e.IPAddress = fi.String(loadbalancer.IP) // This will be empty on create, but will be filled later on FindIPAddress invokation.
			return nil
		}
	}

	// associate vpcuuid to the loadbalancer if set
	vpcUUID := ""
	if fi.StringValue(e.NetworkCIDR) != "" {
		vpcUUID, err = t.Cloud.GetVPCUUID(fi.StringValue(e.NetworkCIDR), fi.StringValue(e.VPCName))
		if err != nil {
			return fmt.Errorf("Error fetching vpcUUID from network cidr=%s", fi.StringValue(e.NetworkCIDR))
		}
	} else if fi.StringValue(e.VPCUUID) != "" {
		vpcUUID = fi.StringValue(e.VPCUUID)
	}

	loadBalancerService := t.Cloud.LoadBalancersService()
	loadbalancer, _, err := loadBalancerService.Create(context.TODO(), &godo.LoadBalancerRequest{
		Name:            fi.StringValue(e.Name),
		Region:          fi.StringValue(e.Region),
		Tag:             fi.StringValue(e.DropletTag),
		VPCUUID:         vpcUUID,
		ForwardingRules: Rules,
		HealthCheck:     HealthCheck,
	})
	if err != nil {
		return fmt.Errorf("Error creating load balancer with Name=%s, Error=%v", fi.StringValue(e.Name), err)
	}

	e.ID = fi.String(loadbalancer.ID)
	e.IPAddress = fi.String(loadbalancer.IP) // This will be empty on create, but will be filled later on FindIPAddress invokation.

	klog.V(2).Infof("load balancer for DO created with id: %s", loadbalancer.ID)
	return nil
}

func (lb *LoadBalancer) IsForAPIServer() bool {
	return lb.ForAPIServer
}

func (lb *LoadBalancer) FindIPAddress(c *fi.Context) (*string, error) {
	cloud := c.Cloud.(do.DOCloud)
	loadBalancerService := cloud.LoadBalancersService()
	address := ""

	if len(fi.StringValue(lb.ID)) > 0 {
		// able to retrieve ID.
		done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
			klog.V(2).Infof("Finding IP address for load balancer ID=%s", fi.StringValue(lb.ID))
			loadBalancer, _, err := loadBalancerService.Get(context.TODO(), fi.StringValue(lb.ID))
			if err != nil {
				klog.Errorf("Error fetching load balancer with Name=%s", fi.StringValue(lb.Name))
				return false, err
			}

			address = loadBalancer.IP

			if isIPv4(address) {
				klog.Infof("retrieved load balancer address=%s", address)
				return true, nil
			}
			return false, nil
		})
		if done {
			return &address, nil
		} else {
			if err == nil {
				err = wait.ErrWaitTimeout
			}
			return nil, err
		}
	}

	return nil, nil
}

func isIPv4(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.To4() != nil
}
