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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) DescribeLoadBalancers(ctx context.Context, request *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeLoadBalancers v2 %v", request)

	if request.PageSize != nil {
		klog.Warningf("PageSize not implemented")
	}
	if request.Marker != nil {
		klog.Fatalf("Marker not implemented")
	}

	var elbs []elbv2types.LoadBalancer
	for _, elb := range m.LoadBalancers {
		match := false

		if len(request.LoadBalancerArns) > 0 {
			for _, name := range request.LoadBalancerArns {
				if aws.ToString(elb.description.LoadBalancerArn) == name {
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

	return &elbv2.DescribeLoadBalancersOutput{
		LoadBalancers: elbs,
	}, nil
}

func (m *MockELBV2) CreateLoadBalancer(ctx context.Context, request *elbv2.CreateLoadBalancerInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateLoadBalancer v2 %v", request)

	lb := elbv2types.LoadBalancer{
		LoadBalancerName:      request.Name,
		Scheme:                request.Scheme,
		SecurityGroups:        request.SecurityGroups,
		Type:                  request.Type,
		IpAddressType:         request.IpAddressType,
		DNSName:               aws.String(fmt.Sprintf("%v.amazonaws.com", aws.ToString(request.Name))),
		CanonicalHostedZoneId: aws.String("HZ123456"),
	}
	zones := make([]elbv2types.AvailabilityZone, 0)
	vpc := "vpc-1"
	for _, subnet := range request.Subnets {
		zones = append(zones, elbv2types.AvailabilityZone{
			SubnetId: aws.String(subnet),
		})
		subnetsOutput, err := m.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(subnet)},
		})
		if err == nil {
			vpc = *subnetsOutput.Subnets[0].VpcId
		}
	}
	for _, subnetMapping := range request.SubnetMappings {
		var lbAddrs []elbv2types.LoadBalancerAddress
		if subnetMapping.PrivateIPv4Address != nil {
			lbAddrs = append(lbAddrs, elbv2types.LoadBalancerAddress{PrivateIPv4Address: subnetMapping.PrivateIPv4Address})
		}
		if subnetMapping.AllocationId != nil {
			lbAddrs = append(lbAddrs, elbv2types.LoadBalancerAddress{AllocationId: subnetMapping.AllocationId})
		}
		zones = append(zones, elbv2types.AvailabilityZone{
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
	arn := fmt.Sprintf("arn:aws-test:elasticloadbalancing:us-test-1:000000000000:loadbalancer/net/%v/%v", aws.ToString(request.Name), m.lbCount)

	lb.LoadBalancerArn = aws.String(arn)

	if m.LoadBalancers == nil {
		m.LoadBalancers = make(map[string]*loadBalancer)
	}
	if m.LBAttributes == nil {
		m.LBAttributes = make(map[string][]elbv2types.LoadBalancerAttribute)
	}
	if m.Tags == nil {
		m.Tags = make(map[string]elbv2types.TagDescription)
	}

	m.LoadBalancers[arn] = &loadBalancer{description: lb}
	m.LBAttributes[arn] = make([]elbv2types.LoadBalancerAttribute, 0)
	m.Tags[arn] = elbv2types.TagDescription{
		ResourceArn: aws.String(arn),
		Tags:        request.Tags,
	}

	return &elbv2.CreateLoadBalancerOutput{LoadBalancers: []elbv2types.LoadBalancer{lb}}, nil
}

func (m *MockELBV2) DescribeLoadBalancerAttributes(ctx context.Context, request *elbv2.DescribeLoadBalancerAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancerAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeLoadBalancerAttributes v2 %v", request)

	if attr, ok := m.LBAttributes[aws.ToString(request.LoadBalancerArn)]; ok {
		return &elbv2.DescribeLoadBalancerAttributesOutput{
			Attributes: attr,
		}, nil
	}
	return nil, fmt.Errorf("LoadBalancerNotFound: %v", aws.ToString(request.LoadBalancerArn))
}

func (m *MockELBV2) ModifyLoadBalancerAttributes(ctx context.Context, request *elbv2.ModifyLoadBalancerAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyLoadBalancerAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ModifyLoadBalancerAttributes v2 %v", request)

	if m.LBAttributes == nil {
		m.LBAttributes = make(map[string][]elbv2types.LoadBalancerAttribute)
	}

	arn := aws.ToString(request.LoadBalancerArn)
	if _, ok := m.LBAttributes[arn]; ok {
		for _, reqAttr := range request.Attributes {
			found := false
			for _, lbAttr := range m.LBAttributes[arn] {
				if aws.ToString(reqAttr.Key) == aws.ToString(lbAttr.Key) {
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
	return nil, fmt.Errorf("LoadBalancerNotFound: %v", aws.ToString(request.LoadBalancerArn))
}

func (m *MockELBV2) SetSecurityGroups(ctx context.Context, request *elbv2.SetSecurityGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.SetSecurityGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	arn := aws.ToString(request.LoadBalancerArn)
	if lb, ok := m.LoadBalancers[arn]; ok {
		lb.description.SecurityGroups = request.SecurityGroups
		return &elbv2.SetSecurityGroupsOutput{
			SecurityGroupIds: request.SecurityGroups,
		}, nil
	}
	return nil, fmt.Errorf("LoadBalancerNotFound: %v", aws.ToString(request.LoadBalancerArn))
}

func (m *MockELBV2) SetSubnets(ctx context.Context, request *elbv2.SetSubnetsInput, optFns ...func(*elbv2.Options)) (*elbv2.SetSubnetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	klog.Fatalf("elbv2.SetSubnets() not implemented")
	return nil, nil
}

func (m *MockELBV2) DeleteLoadBalancer(ctx context.Context, request *elbv2.DeleteLoadBalancerInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteLoadBalancerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteLoadBalancer %v", request)

	arn := aws.ToString(request.LoadBalancerArn)
	delete(m.LoadBalancers, arn)
	for listenerARN, listener := range m.Listeners {
		if aws.ToString(listener.description.LoadBalancerArn) == arn {
			delete(m.Listeners, listenerARN)
		}
	}
	return &elbv2.DeleteLoadBalancerOutput{}, nil
}
