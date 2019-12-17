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

package mockelb

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"k8s.io/klog"
)

const elbZoneID = "FAKEZONE-CLOUDMOCK-ELB"

type MockELB struct {
	elbiface.ELBAPI

	mutex sync.Mutex

	LoadBalancers map[string]*loadBalancer
}

type loadBalancer struct {
	description elb.LoadBalancerDescription
	attributes  elb.LoadBalancerAttributes
	tags        map[string]string
}

func (m *MockELB) DescribeLoadBalancers(request *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("DescribeLoadBalancers %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var elbs []*elb.LoadBalancerDescription
	for _, elb := range m.LoadBalancers {
		match := false

		if len(request.LoadBalancerNames) > 0 {
			for _, name := range request.LoadBalancerNames {
				if aws.StringValue(elb.description.LoadBalancerName) == aws.StringValue(name) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			elbs = append(elbs, &elb.description)
		}
	}

	return &elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: elbs,
	}, nil
}

func (m *MockELB) DescribeLoadBalancersPages(request *elb.DescribeLoadBalancersInput, callback func(p *elb.DescribeLoadBalancersOutput, lastPage bool) (shouldContinue bool)) error {
	// For the mock, we just send everything in one page
	page, err := m.DescribeLoadBalancers(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockELB) CreateLoadBalancer(request *elb.CreateLoadBalancerInput) (*elb.CreateLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("CreateLoadBalancer %v", request)
	createdTime := time.Now().UTC()

	dnsName := *request.LoadBalancerName + ".elb.cloudmock.com"

	lb := &loadBalancer{
		description: elb.LoadBalancerDescription{
			AvailabilityZones: request.AvailabilityZones,
			CreatedTime:       &createdTime,
			LoadBalancerName:  request.LoadBalancerName,
			Scheme:            request.Scheme,
			SecurityGroups:    request.SecurityGroups,
			Subnets:           request.Subnets,
			DNSName:           aws.String(dnsName),

			CanonicalHostedZoneNameID: aws.String(elbZoneID),
		},
		tags: make(map[string]string),
	}

	for _, listener := range request.Listeners {
		lb.description.ListenerDescriptions = append(lb.description.ListenerDescriptions, &elb.ListenerDescription{
			Listener: listener,
		})
	}

	// for _, tag := range input.Tags {
	// 	g.Tags = append(g.Tags, &autoscaling.TagDescription{
	// 		Key:               tag.Key,
	// 		PropagateAtLaunch: tag.PropagateAtLaunch,
	// 		ResourceId:        tag.ResourceId,
	// 		ResourceType:      tag.ResourceType,
	// 		Value:             tag.Value,
	// 	})
	// }

	if m.LoadBalancers == nil {
		m.LoadBalancers = make(map[string]*loadBalancer)
	}
	m.LoadBalancers[*request.LoadBalancerName] = lb

	return &elb.CreateLoadBalancerOutput{
		DNSName: aws.String(dnsName),
	}, nil
}

func (m *MockELB) DeleteLoadBalancer(request *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteLoadBalancer: %v", request)

	id := aws.StringValue(request.LoadBalancerName)
	o := m.LoadBalancers[id]
	if o == nil {
		return nil, fmt.Errorf("LoadBalancer %q not found", id)
	}
	delete(m.LoadBalancers, id)

	return &elb.DeleteLoadBalancerOutput{}, nil
}
