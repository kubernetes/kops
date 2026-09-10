/*
Copyright 2026 The Kubernetes Authors.

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
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cloudmock/aws/mockelbv2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

type networkLoadBalancerTestCloud struct {
	*awsup.MockAWSCloud
	findNetworkInterfaces func(vpcID, loadBalancerName string) ([]ec2types.NetworkInterface, error)
}

func (c *networkLoadBalancerTestCloud) FindELBV2NetworkInterfacesByName(vpcID, loadBalancerName string) ([]ec2types.NetworkInterface, error) {
	return c.findNetworkInterfaces(vpcID, loadBalancerName)
}

func TestNetworkLoadBalancerFindAddressesWithoutVPCID(t *testing.T) {
	ctx := context.Background()
	ec2Mock := &mockec2.MockEC2{}
	_, err := ec2Mock.CreateSubnetWithId(&ec2.CreateSubnetInput{
		VpcId:     aws.String("vpc-existing"),
		CidrBlock: aws.String("172.20.0.0/24"),
	}, "subnet-existing")
	if err != nil {
		t.Fatalf("error creating test subnet: %v", err)
	}
	elbv2Mock := &mockelbv2.MockELBV2{EC2: ec2Mock}
	_, err = elbv2Mock.CreateLoadBalancer(ctx, &elbv2.CreateLoadBalancerInput{
		Name:    aws.String("api-existing"),
		Type:    elbv2types.LoadBalancerTypeEnumNetwork,
		Scheme:  elbv2types.LoadBalancerSchemeEnumInternal,
		Subnets: []string{"subnet-existing"},
		Tags:    []elbv2types.Tag{{Key: aws.String("Name"), Value: aws.String("api.cluster.k8s.local")}},
	})
	if err != nil {
		t.Fatalf("error creating test NLB: %v", err)
	}

	cloud := &networkLoadBalancerTestCloud{
		MockAWSCloud: awsup.BuildMockAWSCloud("us-east-1", "a"),
		findNetworkInterfaces: func(vpcID, loadBalancerName string) ([]ec2types.NetworkInterface, error) {
			if vpcID != "vpc-existing" || loadBalancerName != "api-existing" {
				t.Fatalf("unexpected NLB interface lookup: VPC %q, load balancer %q", vpcID, loadBalancerName)
			}
			return []ec2types.NetworkInterface{{
				PrivateIpAddress: aws.String("172.20.0.10"),
				Ipv6Addresses:    []ec2types.NetworkInterfaceIpv6Address{{Ipv6Address: aws.String("2001:db8::10")}},
			}}, nil
		},
	}
	cloud.MockELBV2 = elbv2Mock
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{AWS: &kops.AWSSpec{}},
			Networking: kops.NetworkingSpec{
				Topology: &kops.TopologySpec{DNS: kops.DNSTypeNone},
			},
		},
	}
	target := terraform.NewTerraformTarget(cloud, "", t.TempDir(), nil)
	c, err := fi.NewCloudupContext(ctx, fi.DeletionProcessingModeIgnore, target, cluster, cloud, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("error building context: %v", err)
	}

	nlb := &NetworkLoadBalancer{
		Name: aws.String("api.cluster.k8s.local"),
		// Terraform does not call VPC.Find, so the VPC task has no ID.
		VPC: &VPC{Name: aws.String("cluster.k8s.local")},
	}
	addresses, err := nlb.FindAddresses(c)
	if err != nil {
		t.Fatalf("FindAddresses failed: %v", err)
	}
	want := []string{"172.20.0.10", "2001:db8::10", "api-existing.amazonaws.com"}
	if !reflect.DeepEqual(addresses, want) {
		t.Fatalf("FindAddresses = %v, want %v", addresses, want)
	}
}
