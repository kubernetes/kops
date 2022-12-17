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
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type VPCAmazonIPv6CIDRBlock struct {
	Name      *string
	Lifecycle fi.Lifecycle

	VPC *VPC

	// Shared is set if this is a shared VPC
	Shared *bool
}

func (e *VPCAmazonIPv6CIDRBlock) Find(c *fi.Context) (*VPCAmazonIPv6CIDRBlock, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	// If the VPC doesn't (yet) exist, there is no association
	if e.VPC.ID == nil {
		return nil, nil
	}

	vpcIPv6CIDR, err := findVPCIPv6CIDR(cloud, e.VPC.ID)
	if err != nil {
		return nil, err
	}
	if vpcIPv6CIDR == nil {
		return nil, nil
	}

	actual := &VPCAmazonIPv6CIDRBlock{
		VPC: &VPC{ID: e.VPC.ID},
	}

	// Prevent spurious changes
	actual.Name = e.Name
	actual.Shared = e.Shared
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

func (_ *VPCAmazonIPv6CIDRBlock) RenderAWS(ctx *fi.Context, t *awsup.AWSAPITarget, a, e, changes *VPCAmazonIPv6CIDRBlock) error {
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

	// Response doesn't contain the new CIDR block
	_, err := t.Cloud.EC2().AssociateVpcCidrBlock(request)
	if err != nil {
		return fmt.Errorf("error associating Amazon IPv6 provided CIDR block to VPC: %v", err)
	}

	return nil // no tags
}

func (_ *VPCAmazonIPv6CIDRBlock) RenderTerraform(ctx *fi.Context, t *terraform.TerraformTarget, a, e, changes *VPCAmazonIPv6CIDRBlock) error {
	// At the moment, this can only be done via the aws_vpc resource
	return nil
}

func findVPCIPv6CIDR(cloud awsup.AWSCloud, vpcID *string) (*string, error) {
	vpc, err := cloud.DescribeVPC(aws.StringValue(vpcID))
	if err != nil {
		return nil, err
	}

	var byoIPv6CidrBlock *string

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if association == nil || association.Ipv6CidrBlockState == nil {
			continue
		}

		// Ipv6CidrBlock is available only when state is "associated"
		if aws.StringValue(association.Ipv6CidrBlockState.State) != ec2.VpcCidrBlockStateCodeAssociated {
			continue
		}

		if aws.StringValue(association.Ipv6Pool) == "Amazon" {
			return association.Ipv6CidrBlock, nil
		}

		if byoIPv6CidrBlock == nil {
			byoIPv6CidrBlock = association.Ipv6CidrBlock
		}
	}

	return byoIPv6CidrBlock, nil
}
