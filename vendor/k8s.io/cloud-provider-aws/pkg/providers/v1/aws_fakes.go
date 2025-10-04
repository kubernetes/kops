/*
Copyright 2017 The Kubernetes Authors.

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

package aws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-aws/pkg/providers/v1/config"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
)

// FakeAWSServices is an fake AWS session used for testing
type FakeAWSServices struct {
	region                      string
	instances                   []*ec2types.Instance
	selfInstance                *ec2types.Instance
	networkInterfacesMacs       []string
	networkInterfacesPrivateIPs [][]string
	networkInterfacesVpcIDs     []string

	ec2      FakeEC2
	elb      ELB
	elbv2    ELBV2
	asg      *FakeASG
	metadata *FakeMetadata
	kms      *FakeKMS

	callCounts map[string]int
}

// NewFakeAWSServices creates a new FakeAWSServices
func NewFakeAWSServices(clusterID string) *FakeAWSServices {
	s := &FakeAWSServices{}
	s.region = "us-west-2"
	s.ec2 = &FakeEC2Impl{aws: s}
	s.elb = &FakeELB{aws: s}
	s.elbv2 = &FakeELBV2{aws: s}
	s.asg = &FakeASG{aws: s}
	s.metadata = &FakeMetadata{aws: s}
	s.kms = &FakeKMS{aws: s}

	s.networkInterfacesMacs = []string{"aa:bb:cc:dd:ee:00", "aa:bb:cc:dd:ee:01"}
	s.networkInterfacesVpcIDs = []string{"vpc-mac0", "vpc-mac1"}

	selfInstance := &ec2types.Instance{}
	selfInstance.InstanceId = aws.String("i-self")
	selfInstance.Placement = &ec2types.Placement{
		AvailabilityZone: aws.String("us-west-2a"),
	}
	selfInstance.PrivateDnsName = aws.String("ip-172-20-0-100.ec2.internal")
	selfInstance.PrivateIpAddress = aws.String("192.168.0.1")
	selfInstance.PublicIpAddress = aws.String("1.2.3.4")
	s.selfInstance = selfInstance
	s.instances = []*ec2types.Instance{selfInstance}

	selfInstance.NetworkInterfaces = []ec2types.InstanceNetworkInterface{
		{
			Attachment: &ec2types.InstanceNetworkInterfaceAttachment{
				DeviceIndex: aws.Int32(1),
			},
			PrivateIpAddresses: []ec2types.InstancePrivateIpAddress{
				{
					Primary:          aws.Bool(true),
					PrivateDnsName:   aws.String("ip-172-20-1-100.ec2.internal"),
					PrivateIpAddress: aws.String("172.20.1.1"),
				},
				{
					Primary:          aws.Bool(false),
					PrivateDnsName:   aws.String("ip-172-20-1-101.ec2.internal"),
					PrivateIpAddress: aws.String("172.20.1.2"),
				},
			},
			Status: ec2types.NetworkInterfaceStatusInUse,
		},
		{
			Attachment: &ec2types.InstanceNetworkInterfaceAttachment{
				DeviceIndex: aws.Int32(0),
			},
			PrivateIpAddresses: []ec2types.InstancePrivateIpAddress{
				{
					Primary:          aws.Bool(true),
					PrivateDnsName:   aws.String("ip-172-20-0-100.ec2.internal"),
					PrivateIpAddress: aws.String("172.20.0.100"),
				},
				{
					Primary:          aws.Bool(false),
					PrivateDnsName:   aws.String("ip-172-20-0-101.ec2.internal"),
					PrivateIpAddress: aws.String("172.20.0.101"),
				},
			},
			Status: ec2types.NetworkInterfaceStatusInUse,
		},
	}

	var tag ec2types.Tag
	tag.Key = aws.String(TagNameKubernetesClusterLegacy)
	tag.Value = aws.String(clusterID)
	selfInstance.Tags = []ec2types.Tag{tag}
	s.callCounts = make(map[string]int)

	return s
}

// WithAz sets the ec2 placement availability zone
func (s *FakeAWSServices) WithAz(az string) *FakeAWSServices {
	if s.selfInstance.Placement == nil {
		s.selfInstance.Placement = &ec2types.Placement{}
	}
	s.selfInstance.Placement.AvailabilityZone = aws.String(az)
	return s
}

// WithRegion sets the AWS region
func (s *FakeAWSServices) WithRegion(region string) *FakeAWSServices {
	s.region = region
	return s
}

// countCall increments the counter for the given service, api, and resourceID and returns the resulting call count
func (s *FakeAWSServices) countCall(service string, api string, resourceID string) int {
	key := fmt.Sprintf("%s:%s:%s", service, api, resourceID)
	s.callCounts[key]++
	count := s.callCounts[key]
	klog.Warningf("call count: %s:%d", key, count)
	return count
}

// Compute returns a fake EC2 client
func (s *FakeAWSServices) Compute(ctx context.Context, region string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (iface.EC2, error) {
	return s.ec2, nil
}

// LoadBalancing returns a fake ELB client
func (s *FakeAWSServices) LoadBalancing(ctx context.Context, region string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (ELB, error) {
	return s.elb, nil
}

// LoadBalancingV2 returns a fake ELBV2 client
func (s *FakeAWSServices) LoadBalancingV2(ctx context.Context, region string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (ELBV2, error) {
	return s.elbv2, nil
}

// Metadata returns a fake EC2Metadata client
func (s *FakeAWSServices) Metadata(ctx context.Context) (config.EC2Metadata, error) {
	return s.metadata, nil
}

// KeyManagement returns a fake KMS client
func (s *FakeAWSServices) KeyManagement(ctx context.Context, regionName string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (KMS, error) {
	return s.kms, nil
}

// FakeEC2 is a fake EC2 client used for testing
type FakeEC2 interface {
	iface.EC2
	CreateSubnet(*ec2types.Subnet) (*ec2.CreateSubnetOutput, error)
	RemoveSubnets()
	CreateRouteTable(*ec2types.RouteTable) (*ec2.CreateRouteTableOutput, error)
	RemoveRouteTables()
}

// FakeEC2Impl is an implementation of the FakeEC2 interface used for testing
type FakeEC2Impl struct {
	aws                      *FakeAWSServices
	Subnets                  []ec2types.Subnet
	DescribeSubnetsInput     *ec2.DescribeSubnetsInput
	RouteTables              []ec2types.RouteTable
	DescribeRouteTablesInput *ec2.DescribeRouteTablesInput
}

// DescribeInstances returns fake instance descriptions
func (ec2i *FakeEC2Impl) DescribeInstances(ctx context.Context, request *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) ([]ec2types.Instance, error) {
	matches := []ec2types.Instance{}
	var matchedInstances []string
	for _, instance := range ec2i.aws.instances {
		if request.InstanceIds != nil {
			if instance.InstanceId == nil {
				klog.Warning("Instance with no instance id: ", instance)
				continue
			}

			found := false
			for _, instanceID := range request.InstanceIds {
				if instanceID == *instance.InstanceId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if request.Filters != nil {
			allMatch := true
			for _, filter := range request.Filters {
				if !instanceMatchesFilter(instance, filter) {
					allMatch = false
					break
				}
			}
			if !allMatch {
				continue
			}
		}
		matches = append(matches, *instance)
		matchedInstances = append(matchedInstances, *instance.InstanceId)
	}

	ec2i.aws.countCall("ec2", "DescribeInstances", strings.Join(matchedInstances, ","))
	return matches, nil
}

// DescribeInstanceTopology is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) DescribeInstanceTopology(ctx context.Context, request *ec2.DescribeInstanceTopologyInput, optFns ...func(*ec2.Options)) ([]ec2types.InstanceTopology, error) {
	panic("Not implemented")
}

// AttachVolume is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) AttachVolume(request *ec2.AttachVolumeInput) (resp *ec2types.VolumeAttachment, err error) {
	panic("Not implemented")
}

// DetachVolume is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) DetachVolume(request *ec2.DetachVolumeInput) (resp *ec2types.VolumeAttachment, err error) {
	panic("Not implemented")
}

// DescribeVolumes is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) DescribeVolumes(request *ec2.DescribeVolumesInput) ([]*ec2types.Volume, error) {
	panic("Not implemented")
}

// CreateVolume is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) CreateVolume(request *ec2.CreateVolumeInput) (resp *ec2types.Volume, err error) {
	panic("Not implemented")
}

// DeleteVolume is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) DeleteVolume(request *ec2.DeleteVolumeInput) (resp *ec2.DeleteVolumeOutput, err error) {
	panic("Not implemented")
}

// DescribeSecurityGroups is not implemented but is required for interface
// conformance
func (ec2i *FakeEC2Impl) DescribeSecurityGroups(ctx context.Context, request *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) ([]ec2types.SecurityGroup, error) {
	panic("Not implemented")
}

// CreateSecurityGroup is not implemented but is required for interface
// conformance
func (ec2i *FakeEC2Impl) CreateSecurityGroup(ctx context.Context, request *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
	panic("Not implemented")
}

// DeleteSecurityGroup is not implemented but is required for interface
// conformance
func (ec2i *FakeEC2Impl) DeleteSecurityGroup(ctx context.Context, request *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	panic("Not implemented")
}

// AuthorizeSecurityGroupIngress is not implemented but is required for
// interface conformance
func (ec2i *FakeEC2Impl) AuthorizeSecurityGroupIngress(ctx context.Context, request *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	panic("Not implemented")
}

// RevokeSecurityGroupIngress is not implemented but is required for interface
// conformance
func (ec2i *FakeEC2Impl) RevokeSecurityGroupIngress(ctx context.Context, request *ec2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	panic("Not implemented")
}

// DescribeVolumeModifications is not implemented but is required for interface
// conformance
func (ec2i *FakeEC2Impl) DescribeVolumeModifications(*ec2.DescribeVolumesModificationsInput) ([]*ec2types.VolumeModification, error) {
	panic("Not implemented")
}

// ModifyVolume is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) ModifyVolume(*ec2.ModifyVolumeInput) (*ec2.ModifyVolumeOutput, error) {
	panic("Not implemented")
}

// CreateSubnet creates fake subnets
func (ec2i *FakeEC2Impl) CreateSubnet(request *ec2types.Subnet) (*ec2.CreateSubnetOutput, error) {
	ec2i.Subnets = append(ec2i.Subnets, *request)
	response := &ec2.CreateSubnetOutput{
		Subnet: request,
	}
	return response, nil
}

// DescribeSubnets returns fake subnet descriptions
func (ec2i *FakeEC2Impl) DescribeSubnets(ctx context.Context, request *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) ([]ec2types.Subnet, error) {
	ec2i.DescribeSubnetsInput = request
	return ec2i.Subnets, nil
}

// RemoveSubnets clears subnets on client
func (ec2i *FakeEC2Impl) RemoveSubnets() {
	ec2i.Subnets = ec2i.Subnets[:0]
}

// DescribeAvailabilityZones returns fake availability zones
// For every input returns a hardcoded list of fake availability zones for the moment
func (ec2i *FakeEC2Impl) DescribeAvailabilityZones(ctx context.Context, request *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) ([]ec2types.AvailabilityZone, error) {
	return []ec2types.AvailabilityZone{
		{
			ZoneName: aws.String("us-west-2a"),
			ZoneType: aws.String("availability-zone"),
			ZoneId:   aws.String("az1"),
		},
		{
			ZoneName: aws.String("us-west-2b"),
			ZoneType: aws.String("availability-zone"),
			ZoneId:   aws.String("az2"),
		},
		{
			ZoneName: aws.String("us-west-2c"),
			ZoneType: aws.String("availability-zone"),
			ZoneId:   aws.String("az3"),
		},
		{
			ZoneName: aws.String("az-local"),
			ZoneType: aws.String("local-zone"),
			ZoneId:   aws.String("lz1"),
		},
		{
			ZoneName: aws.String("az-wavelength"),
			ZoneType: aws.String("wavelength"),
			ZoneId:   aws.String("wl1"),
		},
	}, nil
}

// CreateTags is a mock for CreateTags from EC2
func (ec2i *FakeEC2Impl) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	for _, id := range input.Resources {
		callCount := ec2i.aws.countCall("ec2", "CreateTags", id)
		if id == "i-error" {
			return nil, errors.New("Unable to tag")
		}

		if id == "i-not-found" {
			return nil, errors.New("InvalidInstanceID.NotFound: Instance not found")
		}
		// return an Instance not found error for the first `n` calls
		// instance ID should be of the format `i-not-found-count-$N-$SUFFIX`
		if strings.HasPrefix(id, "i-not-found-count-") {
			notFoundCount, err := strconv.Atoi(strings.Split(id, "-")[4])
			if err != nil {
				panic(err)
			}
			if callCount < notFoundCount {
				return nil, errors.New("InvalidInstanceID.NotFound: Instance not found")
			}
		}
	}
	return &ec2.CreateTagsOutput{}, nil
}

// DeleteTags is a mock for DeleteTags from EC2
func (ec2i *FakeEC2Impl) DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error) {
	for _, id := range input.Resources {
		if id == "i-error" {
			return nil, errors.New("Unable to remove tag")
		}

		if id == "i-not-found" {
			return nil, errors.New("InvalidInstanceID.NotFound: Instance not found")
		}
	}
	return &ec2.DeleteTagsOutput{}, nil
}

// DescribeRouteTables returns fake route table descriptions
func (ec2i *FakeEC2Impl) DescribeRouteTables(ctx context.Context, request *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) ([]ec2types.RouteTable, error) {
	ec2i.DescribeRouteTablesInput = request
	return ec2i.RouteTables, nil
}

// CreateRouteTable creates fake route tables
func (ec2i *FakeEC2Impl) CreateRouteTable(request *ec2types.RouteTable) (*ec2.CreateRouteTableOutput, error) {
	ec2i.RouteTables = append(ec2i.RouteTables, *request)
	response := &ec2.CreateRouteTableOutput{
		RouteTable: request,
	}
	return response, nil
}

// RemoveRouteTables clears route tables on client
func (ec2i *FakeEC2Impl) RemoveRouteTables() {
	ec2i.RouteTables = ec2i.RouteTables[:0]
}

// CreateRoute is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) CreateRoute(ctx context.Context, request *ec2.CreateRouteInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error) {
	panic("Not implemented")
}

// DeleteRoute is not implemented but is required for interface conformance
func (ec2i *FakeEC2Impl) DeleteRoute(ctx context.Context, request *ec2.DeleteRouteInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteOutput, error) {
	panic("Not implemented")
}

// ModifyInstanceAttribute is not implemented but is required for interface
// conformance
func (ec2i *FakeEC2Impl) ModifyInstanceAttribute(ctx context.Context, request *ec2.ModifyInstanceAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error) {
	panic("Not implemented")
}

// DescribeVpcs returns fake VPC descriptions
func (ec2i *FakeEC2Impl) DescribeVpcs(ctx context.Context, request *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return &ec2.DescribeVpcsOutput{Vpcs: []ec2types.Vpc{{CidrBlock: aws.String("172.20.0.0/16")}}}, nil
}

// FakeMetadata is a fake EC2 metadata service client used for testing
type FakeMetadata struct {
	aws *FakeAWSServices
}

// GetMetadata returns fake EC2 metadata for testing
func (m *FakeMetadata) GetMetadata(ctx context.Context, input *imds.GetMetadataInput, optFns ...func(*imds.Options)) (*imds.GetMetadataOutput, error) {
	key := input.Path
	networkInterfacesPrefix := "network/interfaces/macs/"
	i := m.aws.selfInstance
	if key == "placement/availability-zone" {
		az := ""
		if i.Placement != nil {
			az = aws.ToString(i.Placement.AvailabilityZone)
		}
		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(az))}, nil
	} else if key == "instance-id" {
		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(*i.InstanceId))}, nil
	} else if key == "local-hostname" {
		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(*i.PrivateDnsName))}, nil
	} else if key == "public-hostname" {
		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(*i.PublicDnsName))}, nil
	} else if key == "local-ipv4" {
		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(*i.PrivateIpAddress))}, nil
	} else if key == "public-ipv4" {
		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(*i.PublicIpAddress))}, nil
	} else if strings.HasPrefix(key, networkInterfacesPrefix) {
		if key == networkInterfacesPrefix {
			// Return the MACs sorted lexically rather than in device-number
			// order; this matches AWS's observed behavior and lets us test
			// that we fix up the ordering correctly in NodeAddresses().
			macs := make([]string, len(m.aws.networkInterfacesMacs))
			copy(macs, m.aws.networkInterfacesMacs)
			sort.Strings(macs)
			value := strings.Join(macs, "/\n") + "/\n"
			return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(value))}, nil
		}

		keySplit := strings.Split(key, "/")
		macParam := keySplit[3]
		if len(keySplit) == 5 && keySplit[4] == "vpc-id" {
			for i, macElem := range m.aws.networkInterfacesMacs {
				if macParam == macElem {
					return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(m.aws.networkInterfacesVpcIDs[i]))}, nil
				}
			}
		}
		if len(keySplit) == 5 && keySplit[4] == "device-number" {
			for i, macElem := range m.aws.networkInterfacesMacs {
				if macParam == macElem {
					n := i
					if n > 0 {
						// Introduce an artificial gap, just to test eg: [eth0, eth2]
						n++
					}
					value := fmt.Sprintf("%d\n", n)
					return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(value))}, nil
				}
			}
		}
		if len(keySplit) == 5 && keySplit[4] == "local-ipv4s" {
			for i, macElem := range m.aws.networkInterfacesMacs {
				if macParam == macElem {
					value := strings.Join(m.aws.networkInterfacesPrivateIPs[i], "/\n")
					return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(value))}, nil
				}
			}
		}

		return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(""))}, nil
	}

	return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(""))}, nil
}

// GetRegion returns AWS region
func (m *FakeMetadata) GetRegion(ctx context.Context, params *imds.GetRegionInput, optFns ...func(*imds.Options)) (*imds.GetRegionOutput, error) {
	return &imds.GetRegionOutput{Region: m.aws.region}, nil
}

// FakeELB is a fake ELB client used for testing
type FakeELB struct {
	aws *FakeAWSServices
}

// CreateLoadBalancer is not implemented but is required for interface
// conformance
func (e *FakeELB) CreateLoadBalancer(ctx context.Context, input *elb.CreateLoadBalancerInput, opts ...func(*elb.Options)) (*elb.CreateLoadBalancerOutput, error) {
	panic("Not implemented")
}

// DeleteLoadBalancer is not implemented but is required for interface
// conformance
func (e *FakeELB) DeleteLoadBalancer(ctx context.Context, input *elb.DeleteLoadBalancerInput, opts ...func(*elb.Options)) (*elb.DeleteLoadBalancerOutput, error) {
	return &elb.DeleteLoadBalancerOutput{}, nil
}

// DescribeLoadBalancers is not implemented but is required for interface
// conformance
func (e *FakeELB) DescribeLoadBalancers(ctx context.Context, input *elb.DescribeLoadBalancersInput, opts ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error) {
	panic("Not implemented")
}

// AddTags is not implemented but is required for interface conformance
func (e *FakeELB) AddTags(ctx context.Context, input *elb.AddTagsInput, opts ...func(*elb.Options)) (*elb.AddTagsOutput, error) {
	panic("Not implemented")
}

// RegisterInstancesWithLoadBalancer is not implemented but is required for
// interface conformance
func (e *FakeELB) RegisterInstancesWithLoadBalancer(ctx context.Context, input *elb.RegisterInstancesWithLoadBalancerInput, opts ...func(*elb.Options)) (*elb.RegisterInstancesWithLoadBalancerOutput, error) {
	panic("Not implemented")
}

// DeregisterInstancesFromLoadBalancer is not implemented but is required for
// interface conformance
func (e *FakeELB) DeregisterInstancesFromLoadBalancer(ctx context.Context, input *elb.DeregisterInstancesFromLoadBalancerInput, opts ...func(*elb.Options)) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
	panic("Not implemented")
}

// DetachLoadBalancerFromSubnets is not implemented but is required for
// interface conformance
func (e *FakeELB) DetachLoadBalancerFromSubnets(ctx context.Context, input *elb.DetachLoadBalancerFromSubnetsInput, opts ...func(*elb.Options)) (*elb.DetachLoadBalancerFromSubnetsOutput, error) {
	panic("Not implemented")
}

// AttachLoadBalancerToSubnets is not implemented but is required for interface
// conformance
func (e *FakeELB) AttachLoadBalancerToSubnets(ctx context.Context, input *elb.AttachLoadBalancerToSubnetsInput, opts ...func(*elb.Options)) (*elb.AttachLoadBalancerToSubnetsOutput, error) {
	panic("Not implemented")
}

// CreateLoadBalancerListeners is not implemented but is required for interface
// conformance
func (e *FakeELB) CreateLoadBalancerListeners(ctx context.Context, input *elb.CreateLoadBalancerListenersInput, opts ...func(*elb.Options)) (*elb.CreateLoadBalancerListenersOutput, error) {
	panic("Not implemented")
}

// DeleteLoadBalancerListeners is not implemented but is required for interface
// conformance
func (e *FakeELB) DeleteLoadBalancerListeners(ctx context.Context, input *elb.DeleteLoadBalancerListenersInput, opts ...func(*elb.Options)) (*elb.DeleteLoadBalancerListenersOutput, error) {
	panic("Not implemented")
}

// ApplySecurityGroupsToLoadBalancer is not implemented but is required for
// interface conformance
func (e *FakeELB) ApplySecurityGroupsToLoadBalancer(ctx context.Context, input *elb.ApplySecurityGroupsToLoadBalancerInput, opts ...func(*elb.Options)) (*elb.ApplySecurityGroupsToLoadBalancerOutput, error) {
	panic("Not implemented")
}

// ConfigureHealthCheck is not implemented but is required for interface
// conformance
func (e *FakeELB) ConfigureHealthCheck(ctx context.Context, input *elb.ConfigureHealthCheckInput, opts ...func(*elb.Options)) (*elb.ConfigureHealthCheckOutput, error) {
	panic("Not implemented")
}

// CreateLoadBalancerPolicy is not implemented but is required for interface
// conformance
func (e *FakeELB) CreateLoadBalancerPolicy(ctx context.Context, input *elb.CreateLoadBalancerPolicyInput, opts ...func(*elb.Options)) (*elb.CreateLoadBalancerPolicyOutput, error) {
	return &elb.CreateLoadBalancerPolicyOutput{}, nil
}

// SetLoadBalancerPoliciesForBackendServer is not implemented but is required
// for interface conformance
func (e *FakeELB) SetLoadBalancerPoliciesForBackendServer(ctx context.Context, input *elb.SetLoadBalancerPoliciesForBackendServerInput, opts ...func(*elb.Options)) (*elb.SetLoadBalancerPoliciesForBackendServerOutput, error) {
	panic("Not implemented")
}

// SetLoadBalancerPoliciesOfListener is not implemented but is required for
// interface conformance
func (e *FakeELB) SetLoadBalancerPoliciesOfListener(ctx context.Context, input *elb.SetLoadBalancerPoliciesOfListenerInput, opts ...func(*elb.Options)) (*elb.SetLoadBalancerPoliciesOfListenerOutput, error) {
	panic("Not implemented")
}

// DescribeLoadBalancerPolicies is not implemented but is required for
// interface conformance
func (e *FakeELB) DescribeLoadBalancerPolicies(ctx context.Context, input *elb.DescribeLoadBalancerPoliciesInput, opts ...func(*elb.Options)) (*elb.DescribeLoadBalancerPoliciesOutput, error) {
	if aws.ToString(input.LoadBalancerName) == "" {
		return nil, &elbtypes.LoadBalancerAttributeNotFoundException{}
	}
	if len(input.PolicyNames) == 0 || input.PolicyNames[0] == "k8s-SSLNegotiationPolicy-" {
		return nil, &elbtypes.PolicyNotFoundException{}
	}
	return &elb.DescribeLoadBalancerPoliciesOutput{}, nil
}

// DescribeLoadBalancerAttributes is not implemented but is required for
// interface conformance
func (e *FakeELB) DescribeLoadBalancerAttributes(ctx context.Context, input *elb.DescribeLoadBalancerAttributesInput, opts ...func(*elb.Options)) (*elb.DescribeLoadBalancerAttributesOutput, error) {
	panic("Not implemented")
}

// ModifyLoadBalancerAttributes is not implemented but is required for
// interface conformance
func (e *FakeELB) ModifyLoadBalancerAttributes(ctx context.Context, input *elb.ModifyLoadBalancerAttributesInput, opts ...func(*elb.Options)) (*elb.ModifyLoadBalancerAttributesOutput, error) {
	panic("Not implemented")
}

// FakeELBV2 is a fake ELBV2 client used for testing
type FakeELBV2 struct {
	aws *FakeAWSServices
}

// AddTags is not implemented but is required for interface conformance
func (elb *FakeELBV2) AddTags(ctx context.Context, input *elbv2.AddTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.AddTagsOutput, error) {
	panic("Not implemented")
}

// CreateLoadBalancer is not implemented but is required for interface conformance
func (elb *FakeELBV2) CreateLoadBalancer(ctx context.Context, input *elbv2.CreateLoadBalancerInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateLoadBalancerOutput, error) {
	panic("Not implemented")
}

// DescribeLoadBalancers is not implemented but is required for interface conformance
func (elb *FakeELBV2) DescribeLoadBalancers(ctx context.Context, input *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error) {
	panic("Not implemented")
}

// DeleteLoadBalancer is not implemented but is required for interface conformance
func (elb *FakeELBV2) DeleteLoadBalancer(ctx context.Context, input *elbv2.DeleteLoadBalancerInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteLoadBalancerOutput, error) {
	panic("Not implemented")
}

// ModifyLoadBalancerAttributes is not implemented but is required for interface conformance
func (elb *FakeELBV2) ModifyLoadBalancerAttributes(ctx context.Context, input *elbv2.ModifyLoadBalancerAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyLoadBalancerAttributesOutput, error) {
	panic("Not implemented")
}

// DescribeLoadBalancerAttributes is not implemented but is required for interface conformance
func (elb *FakeELBV2) DescribeLoadBalancerAttributes(ctx context.Context, input *elbv2.DescribeLoadBalancerAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancerAttributesOutput, error) {
	panic("Not implemented")
}

// CreateTargetGroup is not implemented but is required for interface conformance
func (elb *FakeELBV2) CreateTargetGroup(ctx context.Context, input *elbv2.CreateTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateTargetGroupOutput, error) {
	panic("Not implemented")
}

// DescribeTargetGroups is not implemented but is required for interface conformance
func (elb *FakeELBV2) DescribeTargetGroups(ctx context.Context, input *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error) {
	panic("Not implemented")
}

// ModifyTargetGroup is not implemented but is required for interface conformance
func (elb *FakeELBV2) ModifyTargetGroup(ctx context.Context, input *elbv2.ModifyTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyTargetGroupOutput, error) {
	panic("Not implemented")
}

// DeleteTargetGroup is not implemented but is required for interface conformance
func (elb *FakeELBV2) DeleteTargetGroup(ctx context.Context, input *elbv2.DeleteTargetGroupInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteTargetGroupOutput, error) {
	panic("Not implemented")
}

// DescribeTargetHealth is not implemented but is required for interface conformance
func (elb *FakeELBV2) DescribeTargetHealth(ctx context.Context, input *elbv2.DescribeTargetHealthInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetHealthOutput, error) {
	panic("Not implemented")
}

// DescribeTargetGroupAttributes is not implemented but is required for interface conformance
func (elb *FakeELBV2) DescribeTargetGroupAttributes(ctx context.Context, input *elbv2.DescribeTargetGroupAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupAttributesOutput, error) {
	panic("Not implemented")
}

// ModifyTargetGroupAttributes is not implemented but is required for interface conformance
func (elb *FakeELBV2) ModifyTargetGroupAttributes(ctx context.Context, input *elbv2.ModifyTargetGroupAttributesInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyTargetGroupAttributesOutput, error) {
	panic("Not implemented")
}

// RegisterTargets is not implemented but is required for interface conformance
func (elb *FakeELBV2) RegisterTargets(ctx context.Context, input *elbv2.RegisterTargetsInput, optFns ...func(*elbv2.Options)) (*elbv2.RegisterTargetsOutput, error) {
	panic("Not implemented")
}

// DeregisterTargets is not implemented but is required for interface conformance
func (elb *FakeELBV2) DeregisterTargets(ctx context.Context, input *elbv2.DeregisterTargetsInput, optFns ...func(*elbv2.Options)) (*elbv2.DeregisterTargetsOutput, error) {
	panic("Not implemented")
}

// CreateListener is not implemented but is required for interface conformance
func (elb *FakeELBV2) CreateListener(ctx context.Context, input *elbv2.CreateListenerInput, optFns ...func(*elbv2.Options)) (*elbv2.CreateListenerOutput, error) {
	panic("Not implemented")
}

// DescribeListeners is not implemented but is required for interface conformance
func (elb *FakeELBV2) DescribeListeners(ctx context.Context, input *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error) {
	panic("Not implemented")
}

// DeleteListener is not implemented but is required for interface conformance
func (elb *FakeELBV2) DeleteListener(ctx context.Context, input *elbv2.DeleteListenerInput, optFns ...func(*elbv2.Options)) (*elbv2.DeleteListenerOutput, error) {
	panic("Not implemented")
}

// ModifyListener is not implemented but is required for interface conformance
func (elb *FakeELBV2) ModifyListener(ctx context.Context, input *elbv2.ModifyListenerInput, optFns ...func(*elbv2.Options)) (*elbv2.ModifyListenerOutput, error) {
	panic("Not implemented")
}

// FakeASG is a fake Autoscaling client used for testing
type FakeASG struct {
	aws *FakeAWSServices
}

// UpdateAutoScalingGroup is not implemented but is required for interface
// conformance
func (a *FakeASG) UpdateAutoScalingGroup(*autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	panic("Not implemented")
}

// DescribeAutoScalingGroups is not implemented but is required for interface
// conformance
func (a *FakeASG) DescribeAutoScalingGroups(*autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	panic("Not implemented")
}

// FakeKMS is a fake KMS client used for testing
type FakeKMS struct {
	aws *FakeAWSServices
}

// DescribeKey is not implemented but is required for interface conformance
func (kms *FakeKMS) DescribeKey(ctx context.Context, input *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	panic("Not implemented")
}

func instanceMatchesFilter(instance *ec2types.Instance, filter ec2types.Filter) bool {
	name := *filter.Name
	if name == "private-dns-name" {
		if instance.PrivateDnsName == nil {
			return false
		}
		return contains(filter.Values, aws.ToString(instance.PrivateDnsName))
	}

	if name == "instance-state-name" {
		return contains(filter.Values, string(instance.State.Name))
	}

	if name == "tag-key" {
		for _, instanceTag := range instance.Tags {
			if contains(filter.Values, aws.ToString(instanceTag.Key)) {
				return true
			}
		}
		return false
	}

	if strings.HasPrefix(name, "tag:") {
		tagName := name[4:]
		for _, instanceTag := range instance.Tags {
			if aws.ToString(instanceTag.Key) == tagName && contains(filter.Values, aws.ToString(instanceTag.Value)) {
				return true
			}
		}
		return false
	}

	panic("Unknown filter name: " + name)
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		// (deliberately panic if s == nil)
		if needle == s {
			return true
		}
	}
	return false
}

// DescribeNetworkInterfaces returns list of ENIs for testing
func (ec2i *FakeEC2Impl) DescribeNetworkInterfaces(ctx context.Context, input *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	fargateNodeNamePrefix := "fargate-"
	networkInterface := []ec2types.NetworkInterface{
		{
			PrivateIpAddress: aws.String("1.2.3.4"),
			AvailabilityZone: aws.String("us-west-2c"),
		},
	}
	for _, filter := range input.Filters {
		if strings.HasPrefix(filter.Values[0], fargateNodeNamePrefix) {
			// verify filter doesn't have fargate prefix
			panic(fmt.Sprintf("invalid endpoint specified for DescribeNetworkInterface call %s", filter.Values[0]))
		} else if strings.HasPrefix(filter.Values[0], "not-found") {
			// for negative testing
			return &ec2.DescribeNetworkInterfacesOutput{}, nil
		}

		if strings.Contains(filter.Values[0], "return.private.dns.name") {
			networkInterface[0].PrivateDnsName = aws.String("ip-1-2-3-4.compute.amazon.com")
		}

		if filter.Values[0] == "return.private.dns.name.ipv6" {
			networkInterface[0].Ipv6Addresses = []ec2types.NetworkInterfaceIpv6Address{
				{
					Ipv6Address: aws.String("2001:db8:3333:4444:5555:6666:7777:8888"),
				},
			}
		}
	}

	return &ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: networkInterface,
	}, nil
}
