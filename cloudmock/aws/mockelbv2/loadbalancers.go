/*
Copyright 2020 The Kubernetes Authors.

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

package mockelbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) DescribeLoadBalancers(request *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeLoadBalancers v2 %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var elbs []*elbv2.LoadBalancer
	for _, elb := range m.LoadBalancers {
		match := false

		if len(request.LoadBalancerArns) > 0 {
			for _, name := range request.LoadBalancerArns {
				if aws.StringValue(elb.description.LoadBalancerArn) == aws.StringValue(name) {
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

	return &elbv2.DescribeLoadBalancersOutput{
		LoadBalancers: elbs,
	}, nil
}

func (m *MockELBV2) DescribeLoadBalancersPages(request *elbv2.DescribeLoadBalancersInput, callback func(p *elbv2.DescribeLoadBalancersOutput, lastPage bool) (shouldContinue bool)) error {
	// For the mock, we just send everything in one page
	page, err := m.DescribeLoadBalancers(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockELBV2) CreateLoadBalancer(request *elbv2.CreateLoadBalancerInput) (*elbv2.CreateLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateLoadBalancer v2 %v", request)

	lb := elbv2.LoadBalancer{
		LoadBalancerName:      request.Name,
		Scheme:                request.Scheme,
		Type:                  request.Type,
		IpAddressType:         request.IpAddressType,
		DNSName:               aws.String(fmt.Sprintf("%v.amazonaws.com", aws.StringValue(request.Name))),
		CanonicalHostedZoneId: aws.String("HZ123456"),
	}
	zones := make([]*elbv2.AvailabilityZone, 0)
	vpc := "vpc-1"
	for _, subnet := range request.Subnets {
		zones = append(zones, &elbv2.AvailabilityZone{
			SubnetId: subnet,
		})
		subnetsOutput, err := m.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{subnet},
		})
		if err == nil {
			vpc = *subnetsOutput.Subnets[0].VpcId
		}
	}
	for _, subnetMapping := range request.SubnetMappings {
		var lbAddrs []*elbv2.LoadBalancerAddress
		if subnetMapping.PrivateIPv4Address != nil {
			lbAddrs = append(lbAddrs, &elbv2.LoadBalancerAddress{PrivateIPv4Address: subnetMapping.PrivateIPv4Address})
		}
		if subnetMapping.AllocationId != nil {
			lbAddrs = append(lbAddrs, &elbv2.LoadBalancerAddress{AllocationId: subnetMapping.AllocationId})
		}
		zones = append(zones, &elbv2.AvailabilityZone{
			SubnetId:              subnetMapping.SubnetId,
			LoadBalancerAddresses: lbAddrs,
		})
		subnetsOutput, err := m.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{subnetMapping.SubnetId},
		})
		if err == nil {
			vpc = *subnetsOutput.Subnets[0].VpcId
		}
	}
	lb.AvailabilityZones = zones

	lb.VpcId = aws.String(vpc)

	m.lbCount++
	arn := fmt.Sprintf("arn:aws-test:elasticloadbalancing:us-test-1:000000000000:loadbalancer/net/%v/%v", aws.StringValue(request.Name), m.lbCount)

	lb.LoadBalancerArn = aws.String(arn)

	if m.LoadBalancers == nil {
		m.LoadBalancers = make(map[string]*loadBalancer)
	}
	if m.LBAttributes == nil {
		m.LBAttributes = make(map[string][]*elbv2.LoadBalancerAttribute)
	}
	if m.Tags == nil {
		m.Tags = make(map[string]*elbv2.TagDescription)
	}

	m.LoadBalancers[arn] = &loadBalancer{description: lb}
	m.LBAttributes[arn] = make([]*elbv2.LoadBalancerAttribute, 0)
	m.Tags[arn] = &elbv2.TagDescription{
		ResourceArn: aws.String(arn),
		Tags:        request.Tags,
	}

	return &elbv2.CreateLoadBalancerOutput{LoadBalancers: []*elbv2.LoadBalancer{&lb}}, nil
}

func (m *MockELBV2) DescribeLoadBalancerAttributes(request *elbv2.DescribeLoadBalancerAttributesInput) (*elbv2.DescribeLoadBalancerAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeLoadBalancerAttributes v2 %v", request)

	if attr, ok := m.LBAttributes[aws.StringValue(request.LoadBalancerArn)]; ok {
		return &elbv2.DescribeLoadBalancerAttributesOutput{
			Attributes: attr,
		}, nil
	}
	return nil, fmt.Errorf("LoadBalancerNotFound: %v", aws.StringValue(request.LoadBalancerArn))
}

func (m *MockELBV2) ModifyLoadBalancerAttributes(request *elbv2.ModifyLoadBalancerAttributesInput) (*elbv2.ModifyLoadBalancerAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ModifyLoadBalancerAttributes v2 %v", request)

	if m.LBAttributes == nil {
		m.LBAttributes = make(map[string][]*elbv2.LoadBalancerAttribute)
	}

	arn := aws.StringValue(request.LoadBalancerArn)
	if _, ok := m.LBAttributes[arn]; ok {
		for _, reqAttr := range request.Attributes {
			found := false
			for _, lbAttr := range m.LBAttributes[arn] {
				if aws.StringValue(reqAttr.Key) == aws.StringValue(lbAttr.Key) {
					lbAttr.Value = reqAttr.Value
					found = true
				}
			}
			if !found {
				m.LBAttributes[arn] = append(m.LBAttributes[arn], reqAttr)
			}
		}
		return &elbv2.ModifyLoadBalancerAttributesOutput{
			Attributes: m.LBAttributes[arn],
		}, nil
	}
	return nil, fmt.Errorf("LoadBalancerNotFound: %v", aws.StringValue(request.LoadBalancerArn))
}

func (m *MockELBV2) SetSubnets(request *elbv2.SetSubnetsInput) (*elbv2.SetSubnetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	klog.Fatalf("elbv2.SetSubnets() not implemented")
	return nil, nil
}

func (m *MockELBV2) DeleteLoadBalancer(request *elbv2.DeleteLoadBalancerInput) (*elbv2.DeleteLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteLoadBalancer %v", request)

	arn := aws.StringValue(request.LoadBalancerArn)
	delete(m.LoadBalancers, arn)
	for listenerARN, listener := range m.Listeners {
		if aws.StringValue(listener.description.LoadBalancerArn) == arn {
			delete(m.Listeners, listenerARN)
		}
	}
	return &elbv2.DeleteLoadBalancerOutput{}, nil
}
