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

package awsup

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	dnsproviderroute53 "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
)

type MockAWSCloud struct {
	MockCloud
	region string
	tags   map[string]string

	zones []*ec2.AvailabilityZone
}

var _ fi.Cloud = (*MockAWSCloud)(nil)

func InstallMockAWSCloud(region string, zoneLetters string) *MockAWSCloud {
	i := BuildMockAWSCloud(region, zoneLetters)
	awsCloudInstances[region] = i
	allRegions = []*ec2.Region{
		{RegionName: aws.String(region)},
	}
	return i
}

func BuildMockAWSCloud(region string, zoneLetters string) *MockAWSCloud {
	i := &MockAWSCloud{region: region}
	for _, c := range zoneLetters {
		azName := fmt.Sprintf("%s%c", region, c)
		az := &ec2.AvailabilityZone{
			RegionName: aws.String(region),
			ZoneName:   aws.String(azName),
			State:      aws.String("available"),
		}
		i.zones = append(i.zones, az)
	}
	return i
}

type MockCloud struct {
	MockAutoscaling    autoscalingiface.AutoScalingAPI
	MockCloudFormation *cloudformation.CloudFormation
	MockEC2            ec2iface.EC2API
	MockRoute53        route53iface.Route53API
}

func (c *MockCloud) ProviderID() kops.CloudProviderID {
	return "mock"
}

func (c *MockCloud) DNS() (dnsprovider.Interface, error) {
	if c.MockRoute53 == nil {
		return nil, fmt.Errorf("MockRoute53 not set")
	}
	return dnsproviderroute53.New(c.MockRoute53), nil
}

func (c *MockAWSCloud) Region() string {
	return c.region
}

func (c *MockAWSCloud) DescribeAvailabilityZones() ([]*ec2.AvailabilityZone, error) {
	return c.zones, nil
}

func (c *MockAWSCloud) AddTags(name *string, tags map[string]string) {
	if name != nil {
		tags["Name"] = *name
	}
	for k, v := range c.tags {
		tags[k] = v
	}
}

func (c *MockAWSCloud) BuildFilters(name *string) []*ec2.Filter {
	return buildFilters(c.tags, name)
}

func (c *MockAWSCloud) AddAWSTags(id string, expected map[string]string) error {
	return addAWSTags(c, id, expected)
}

func (c *MockAWSCloud) BuildTags(name *string) map[string]string {
	return buildTags(c.tags, name)
}

func (c *MockAWSCloud) Tags() map[string]string {
	glog.Fatalf("MockAWSCloud Tags not implemented")
	return nil
}

func (c *MockAWSCloud) CreateTags(resourceId string, tags map[string]string) error {
	return createTags(c, resourceId, tags)
}

func (c *MockAWSCloud) GetTags(resourceID string) (map[string]string, error) {
	return getTags(c, resourceID)
}

func (c *MockAWSCloud) GetELBTags(loadBalancerName string) (map[string]string, error) {
	return nil, fmt.Errorf("MockAWSCloud GetELBTags not implemented")
}

func (c *MockAWSCloud) CreateELBTags(loadBalancerName string, tags map[string]string) error {
	return fmt.Errorf("MockAWSCloud CreateELBTags not implemented")
}

func (c *MockAWSCloud) DescribeInstance(instanceID string) (*ec2.Instance, error) {
	return nil, fmt.Errorf("MockAWSCloud DescribeInstance not implemented")
}

func (c *MockAWSCloud) DescribeVPC(vpcID string) (*ec2.Vpc, error) {
	return nil, fmt.Errorf("MockAWSCloud DescribeVPC not implemented")
}

func (c *MockAWSCloud) ResolveImage(name string) (*ec2.Image, error) {
	return resolveImage(c.MockEC2, name)
}

func (c *MockAWSCloud) WithTags(tags map[string]string) AWSCloud {
	m := &MockAWSCloud{}
	*m = *c
	m.tags = tags
	return m
}

func (c *MockAWSCloud) CloudFormation() *cloudformation.CloudFormation {
	if c.MockEC2 == nil {
		glog.Fatalf("MockAWSCloud MockCloudFormation not set")
	}
	return c.MockCloudFormation
}

func (c *MockAWSCloud) EC2() ec2iface.EC2API {
	if c.MockEC2 == nil {
		glog.Fatalf("MockAWSCloud MockEC2 not set")
	}
	return c.MockEC2
}

func (c *MockAWSCloud) IAM() *iam.IAM {
	glog.Fatalf("MockAWSCloud IAM not implemented")
	return nil
}

func (c *MockAWSCloud) ELB() *elb.ELB {
	glog.Fatalf("MockAWSCloud ELB not implemented")
	return nil
}

func (c *MockAWSCloud) Autoscaling() autoscalingiface.AutoScalingAPI {
	if c.MockAutoscaling == nil {
		glog.Fatalf("MockAWSCloud Autoscaling not implemented")
	}
	return c.MockAutoscaling
}

func (c *MockAWSCloud) Route53() route53iface.Route53API {
	if c.MockRoute53 == nil {
		glog.Fatalf("MockRoute53 not set")
	}
	return c.MockRoute53
}

func (c *MockAWSCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("MockAWSCloud FindVPCInfo not implemented")
}

// DefaultInstanceType determines an instance type for the specified cluster & instance group
func (c *MockAWSCloud) DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		return "m3.medium", nil
	case kops.InstanceGroupRoleNode:
		return "t2.medium", nil
	case kops.InstanceGroupRoleBastion:
		return "t2.micro", nil
	default:
		return "", fmt.Errorf("MockAWSCloud DefaultInstanceType does not handle %s", ig.Spec.Role)
	}
}
