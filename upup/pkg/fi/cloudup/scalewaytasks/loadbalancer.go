/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaytasks

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// +kops:fitask
type LoadBalancer struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Region              *string
	LoadBalancerId      *string
	LoadBalancerAddress *string
	Tags                []string
	ForAPIServer        bool
}

var _ fi.CompareWithID = &LoadBalancer{}
var _ fi.HasAddress = &LoadBalancer{}

func (l *LoadBalancer) CompareWithID() *string {
	return l.LoadBalancerId
}

func (l *LoadBalancer) IsForAPIServer() bool {
	return l.ForAPIServer
}

func (l *LoadBalancer) Find(context *fi.CloudupContext) (*LoadBalancer, error) {
	if fi.ValueOf(l.LoadBalancerId) == "" {
		return nil, nil
	}

	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	loadBalancer, err := lbService.GetLB(&lb.GetLBRequest{
		Region: scw.Region(cloud.Region()),
		LBID:   fi.ValueOf(l.LoadBalancerId),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting load-balancer %s: %s", fi.ValueOf(l.LoadBalancerId), err)
	}

	lbIP := loadBalancer.IP[0].IPAddress
	if len(loadBalancer.IP) > 1 {
		klog.V(4).Infof("multiple IPs found for load-balancer, using %s", lbIP)
	}

	return &LoadBalancer{
		Name:                &loadBalancer.Name,
		LoadBalancerId:      &loadBalancer.ID,
		LoadBalancerAddress: &lbIP,
		Tags:                loadBalancer.Tags,
		Lifecycle:           l.Lifecycle,
		ForAPIServer:        l.ForAPIServer,
	}, nil
}

func (l *LoadBalancer) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	if l.LoadBalancerId == nil {
		return nil, nil
	}

	loadBalancer, err := lbService.GetLB(&lb.GetLBRequest{
		Region: scw.Region(cloud.Region()),
		LBID:   fi.ValueOf(l.LoadBalancerId),
	})
	if err != nil {
		return nil, err
	}

	addresses := []string(nil)
	for _, address := range loadBalancer.IP {
		addresses = append(addresses, address.IPAddress)
	}

	return addresses, nil
}

func (l *LoadBalancer) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(l, context)
}

func (_ *LoadBalancer) CheckChanges(actual, expected, changes *LoadBalancer) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.LoadBalancerId != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
	}
	return nil
}

func (l *LoadBalancer) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *LoadBalancer) error {
	lbService := t.Cloud.LBService()

	// We check if the load-balancer already exists
	lbs, err := lbService.ListLBs(&lb.ListLBsRequest{
		Region: scw.Region(fi.ValueOf(expected.Region)),
		Name:   expected.Name,
	}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing existing load-balancers: %w", err)
	}

	if lbs.TotalCount > 0 {
		loadBalancer := lbs.LBs[0]
		lbIP := loadBalancer.IP[0].IPAddress
		if len(loadBalancer.IP) > 1 {
			klog.V(4).Infof("multiple IPs found for load-balancer, using %s", lbIP)
		}
		expected.LoadBalancerId = &loadBalancer.ID
		expected.LoadBalancerAddress = &lbIP
		return nil
	}

	loadBalancer, err := lbService.CreateLB(&lb.CreateLBRequest{
		Region: scw.Region(fi.ValueOf(expected.Region)),
		Name:   fi.ValueOf(expected.Name),
		IPID:   nil,
		Tags:   expected.Tags,
	})
	if err != nil {
		return err
	}

	_, err = lbService.WaitForLb(&lb.WaitForLBRequest{
		LBID:   loadBalancer.ID,
		Region: scw.Region(fi.ValueOf(expected.Region)),
	})
	if err != nil {
		return fmt.Errorf("error waiting for load-balancer %s: %w", loadBalancer.ID, err)
	}

	expected.LoadBalancerId = &loadBalancer.ID

	if len(loadBalancer.IP) > 1 {
		klog.V(8).Infof("got more more than 1 IP for LB (got %d)", len(loadBalancer.IP))
	}
	ip := (*loadBalancer.IP[0]).IPAddress
	expected.LoadBalancerAddress = &ip

	// We create the load-balancer's backend
	backEnd, err := lbService.CreateBackend(&lb.CreateBackendRequest{
		Region:               scw.Region(fi.ValueOf(expected.Region)),
		LBID:                 loadBalancer.ID,
		Name:                 "lb-backend",
		ForwardProtocol:      "tcp",
		ForwardPort:          443,
		ForwardPortAlgorithm: "roundrobin",
		StickySessions:       "none",
		HealthCheck: &lb.HealthCheck{
			CheckMaxRetries: 5,
			TCPConfig:       &lb.HealthCheckTCPConfig{},
			Port:            443,
			CheckTimeout:    scw.TimeDurationPtr(3000),
			CheckDelay:      scw.TimeDurationPtr(1001),
		},
		ProxyProtocol: "proxy_protocol_none",
	})
	if err != nil {
		return fmt.Errorf("error creating back-end for load-balancer %s: %w", loadBalancer.ID, err)
	}

	_, err = lbService.WaitForLb(&lb.WaitForLBRequest{
		LBID:   loadBalancer.ID,
		Region: scw.Region(fi.ValueOf(expected.Region)),
	})
	if err != nil {
		return fmt.Errorf("error waiting for load-balancer %s: %w", loadBalancer.ID, err)
	}

	// We create the load-balancer's front-end
	_, err = lbService.CreateFrontend(&lb.CreateFrontendRequest{
		Region:      scw.Region(fi.ValueOf(expected.Region)),
		LBID:        loadBalancer.ID,
		Name:        "lb-frontend",
		InboundPort: 443,
		BackendID:   backEnd.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating front-end for load-balancer %s: %w", loadBalancer.ID, err)
	}
	_, err = lbService.WaitForLb(&lb.WaitForLBRequest{
		LBID:   loadBalancer.ID,
		Region: scw.Region(fi.ValueOf(expected.Region)),
	})
	if err != nil {
		return fmt.Errorf("error waiting for load-balancer %s: %w", loadBalancer.ID, err)
	}

	return nil
}
