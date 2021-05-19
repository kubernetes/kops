/*
Copyright 2021 The Kubernetes Authors.

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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type VPCAmazonIPv6CIDRBlock struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	VPC       *VPC
	CIDRBlock *string

	// Shared is set if this is a shared VPC
	Shared *bool
}

func (e *VPCAmazonIPv6CIDRBlock) Find(c *fi.Context) (*VPCAmazonIPv6CIDRBlock, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	vpc, err := cloud.DescribeVPC(aws.StringValue(e.VPC.ID))
	if err != nil {
		return nil, err
	}

	var cidr *string
	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if association == nil || association.Ipv6CidrBlockState == nil {
			continue
		}

		state := aws.StringValue(association.Ipv6CidrBlockState.State)
		if state != ec2.VpcCidrBlockStateCodeAssociated && state != ec2.VpcCidrBlockStateCodeAssociating {
			continue
		}

		if aws.StringValue(association.Ipv6Pool) == "Amazon" {
			cidr = association.Ipv6CidrBlock
			break
		}
	}
	if cidr == nil {
		return nil, nil
	}

	actual := &VPCAmazonIPv6CIDRBlock{
		VPC:       &VPC{ID: vpc.VpcId},
		CIDRBlock: cidr,
	}

	// Expose the Amazon provided IPv6 CIDR block to other tasks
	e.CIDRBlock = cidr

	// Prevent spurious changes
	actual.Shared = e.Shared
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *VPCAmazonIPv6CIDRBlock) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *VPCAmazonIPv6CIDRBlock) CheckChanges(a, e, changes *VPCAmazonIPv6CIDRBlock) error {
	if e.VPC == nil {
		return fi.RequiredField("VPC")
	}

	if a != nil && changes != nil {
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *VPCAmazonIPv6CIDRBlock) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *VPCAmazonIPv6CIDRBlock) error {
	shared := aws.BoolValue(e.Shared)
	if shared && a == nil {
		// VPC not owned by kOps, no changes will be applied
		// Verify that the Amazon IPv6 provided CIDR block was found.
		return fmt.Errorf("IPv6 CIDR block provided by Amazon not found")
	}

	request := &ec2.AssociateVpcCidrBlockInput{
		VpcId:                       e.VPC.ID,
		AmazonProvidedIpv6CidrBlock: aws.Bool(true),
	}

	_, err := t.Cloud.EC2().AssociateVpcCidrBlock(request)
	if err != nil {
		return fmt.Errorf("error associating Amazon IPv6 provided CIDR block to VPC: %v", err)
	}

	return nil // no tags
}

func (_ *VPCAmazonIPv6CIDRBlock) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPCAmazonIPv6CIDRBlock) error {
	// At the moment, this can only be done via the aws_vpc resource
	return nil
}

type cloudformationVPCAmazonIPv6CIDRBlock struct {
	VPCID      *cloudformation.Literal `json:"VpcId"`
	AmazonIPv6 *bool                   `json:"AmazonProvidedIpv6CidrBlock"`
}

func (_ *VPCAmazonIPv6CIDRBlock) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *VPCAmazonIPv6CIDRBlock) error {
	shared := aws.BoolValue(e.Shared)
	if shared && a == nil {
		// VPC not owned by kOps, no changes will be applied
		// Verify that the Amazon IPv6 provided CIDR block was found.
		return fmt.Errorf("IPv6 CIDR block provided by Amazon not found")
	}

	cf := &cloudformationVPCAmazonIPv6CIDRBlock{
		VPCID:      e.VPC.CloudformationLink(),
		AmazonIPv6: aws.Bool(true),
	}

	return t.RenderResource("AWS::EC2::VPCCidrBlock", *e.Name, cf)
}
