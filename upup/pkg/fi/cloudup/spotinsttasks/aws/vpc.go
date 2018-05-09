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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=VPC
type VPC struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID                 *string
	CIDR               *string
	EnableDNSHostnames *bool
	EnableDNSSupport   *bool

	// Shared is set if this is a shared VPC
	Shared *bool

	Tags map[string]string
}

var _ fi.CompareWithID = &VPC{}

func (e *VPC) CompareWithID() *string {
	return e.ID
}

func (e *VPC) Find(c *fi.Context) (*VPC, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	request := &ec2.DescribeVpcsInput{}

	if fi.StringValue(e.ID) != "" {
		request.VpcIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2().DescribeVpcs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing VPCs: %v", err)
	}
	if response == nil || len(response.Vpcs) == 0 {
		return nil, nil
	}

	if len(response.Vpcs) != 1 {
		return nil, fmt.Errorf("found multiple VPCs matching tags")
	}
	vpc := response.Vpcs[0]
	actual := &VPC{
		ID:   vpc.VpcId,
		CIDR: vpc.CidrBlock,
		Name: findNameTag(vpc.Tags),
		Tags: intersectTags(vpc.Tags, e.Tags),
	}

	glog.V(4).Infof("found matching VPC %v", actual)

	if actual.ID != nil {
		request := &ec2.DescribeVpcAttributeInput{VpcId: actual.ID, Attribute: aws.String(ec2.VpcAttributeNameEnableDnsSupport)}
		response, err := cloud.EC2().DescribeVpcAttribute(request)
		if err != nil {
			return nil, fmt.Errorf("error querying for dns support: %v", err)
		}
		actual.EnableDNSSupport = response.EnableDnsSupport.Value
	}

	if actual.ID != nil {
		request := &ec2.DescribeVpcAttributeInput{VpcId: actual.ID, Attribute: aws.String(ec2.VpcAttributeNameEnableDnsHostnames)}
		response, err := cloud.EC2().DescribeVpcAttribute(request)
		if err != nil {
			return nil, fmt.Errorf("error querying for dns support: %v", err)
		}
		actual.EnableDNSHostnames = response.EnableDnsHostnames.Value
	}

	// Prevent spurious comparison failures
	actual.Shared = e.Shared
	if e.ID == nil {
		e.ID = actual.ID
	}
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (s *VPC) CheckChanges(a, e, changes *VPC) error {
	if a == nil {
		if e.CIDR == nil {
			// TODO: Auto-assign CIDR?
			return fi.RequiredField("CIDR")
		}
	}
	if a != nil {
		if changes.CIDR != nil {
			// TODO: Do we want to destroy & recreate the VPC?
			return fi.CannotChangeField("CIDR")
		}
	}
	return nil
}

func (e *VPC) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *VPC) Render(t *spotinst.Target, a, e, changes *VPC) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the VPC was found and matches our required settings
		if a == nil {
			return fmt.Errorf("VPC with id %q not found", fi.StringValue(e.ID))
		}

		if changes != nil && changes.EnableDNSSupport != nil {
			if featureflag.VPCSkipEnableDNSSupport.Enabled() {
				glog.Warningf("VPC did not have EnableDNSSupport=true, but ignoring because of VPCSkipEnableDNSSupport feature-flag")
			} else {
				return fmt.Errorf("VPC with id %q was set to be shared, but did not have EnableDNSSupport=true.", fi.StringValue(e.ID))
			}
		}

		return nil
	}

	if a == nil {
		glog.V(2).Infof("Creating VPC with CIDR: %q", *e.CIDR)

		request := &ec2.CreateVpcInput{
			CidrBlock: e.CIDR,
		}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateVpc(request)
		if err != nil {
			return fmt.Errorf("error creating VPC: %v", err)
		}

		e.ID = response.Vpc.VpcId
	}

	if changes.EnableDNSSupport != nil {
		request := &ec2.ModifyVpcAttributeInput{
			VpcId:            e.ID,
			EnableDnsSupport: &ec2.AttributeBooleanValue{Value: changes.EnableDNSSupport},
		}

		_, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().ModifyVpcAttribute(request)
		if err != nil {
			return fmt.Errorf("error modifying VPC attribute: %v", err)
		}
	}

	if changes.EnableDNSHostnames != nil {
		request := &ec2.ModifyVpcAttributeInput{
			VpcId:              e.ID,
			EnableDnsHostnames: &ec2.AttributeBooleanValue{Value: changes.EnableDNSHostnames},
		}

		_, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().ModifyVpcAttribute(request)
		if err != nil {
			return fmt.Errorf("error modifying VPC attribute: %v", err)
		}
	}

	tags := e.Tags
	if shared {
		// Don't tag shared resources
		tags = nil
	}
	return t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.ID, tags)
}
