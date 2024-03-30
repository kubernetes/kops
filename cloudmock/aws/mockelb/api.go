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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

const elbZoneID = "FAKEZONE-CLOUDMOCK-ELB"

type MockELB struct {
	awsinterfaces.ELBAPI

	mutex sync.Mutex

	LoadBalancers map[string]*loadBalancer
}

type loadBalancer struct {
	description elbtypes.LoadBalancerDescription
	attributes  elbtypes.LoadBalancerAttributes
	tags        map[string]string
}

func (m *MockELB) DescribeLoadBalancers(ctx context.Context, request *elb.DescribeLoadBalancersInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("DescribeLoadBalancers %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var elbs []elbtypes.LoadBalancerDescription
	for _, elb := range m.LoadBalancers {
		match := false

		if len(request.LoadBalancerNames) > 0 {
			for _, name := range request.LoadBalancerNames {
				if aws.ToString(elb.description.LoadBalancerName) == name {
					match = true
				}
			}
		} else {
			match = true
		}

		if match {
			elbs = append(elbs, elb.description)
		}
	}

	return &elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: elbs,
	}, nil
}

func (m *MockELB) CreateLoadBalancer(ctx context.Context, request *elb.CreateLoadBalancerInput, optFns ...func(*elb.Options)) (*elb.CreateLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("CreateLoadBalancer %v", request)
	createdTime := time.Now().UTC()

	dnsName := *request.LoadBalancerName + ".elb.cloudmock.com"

	lb := &loadBalancer{
		description: elbtypes.LoadBalancerDescription{
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
		lb.description.ListenerDescriptions = append(lb.description.ListenerDescriptions, elbtypes.ListenerDescription{
			Listener: &listener,
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

func (m *MockELB) DeleteLoadBalancer(ctx context.Context, request *elb.DeleteLoadBalancerInput, optFns ...func(*elb.Options)) (*elb.DeleteLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteLoadBalancer: %v", request)

	id := aws.ToString(request.LoadBalancerName)
	o := m.LoadBalancers[id]
	if o == nil {
		return nil, fmt.Errorf("LoadBalancer %q not found", id)
	}
	delete(m.LoadBalancers, id)

	return &elb.DeleteLoadBalancerOutput{}, nil
}
