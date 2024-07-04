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

package awstasks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type VPCCIDRBlock struct {
	Name      *string
	Lifecycle fi.Lifecycle

	VPC       *VPC
	CIDRBlock *string

	// Shared is set if this is a shared VPC
	Shared *bool
}

func (e *VPCCIDRBlock) Find(c *fi.CloudupContext) (*VPCCIDRBlock, error) {
	cloud := awsup.GetCloud(c)

	vpcID := aws.ToString(e.VPC.ID)

	// If the VPC doesn't (yet) exist, there is no association
	if vpcID == "" {
		return nil, nil
	}

	vpc, err := cloud.DescribeVPC(vpcID)
	if err != nil {
		return nil, err
	}

	found := false
	if e.CIDRBlock != nil {
		for _, cba := range vpc.CidrBlockAssociationSet {
			if cba.CidrBlockState == nil {
				continue
			}

			state := cba.CidrBlockState.State
			if state != ec2types.VpcCidrBlockStateCodeAssociated && state != ec2types.VpcCidrBlockStateCodeAssociating {
				continue
			}

			if aws.ToString(cba.CidrBlock) == aws.ToString(e.CIDRBlock) {
				found = true
				break
			}
		}
	}
	if !found {
		return nil, nil
	}

	actual := &VPCCIDRBlock{
		VPC:       &VPC{ID: vpc.VpcId},
		CIDRBlock: e.CIDRBlock,
	}

	// Prevent spurious changes
	actual.Shared = e.Shared
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *VPCCIDRBlock) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (s *VPCCIDRBlock) CheckChanges(a, e, changes *VPCCIDRBlock) error {
	if e.VPC == nil {
		return fi.RequiredField("VPC")
	}

	if e.CIDRBlock == nil {
		return fi.RequiredField("CIDRBlock")
	}

	if a != nil && changes != nil {
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}

		if changes.CIDRBlock != nil {
			return fi.CannotChangeField("CIDRBlock")
		}
	}

	return nil
}

func (_ *VPCCIDRBlock) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *VPCCIDRBlock) error {
	ctx := context.TODO()
	shared := aws.ToBool(e.Shared)
	if shared && a == nil {
		// VPC not owned by kOps, no changes will be applied
		// Verify that the CIDR block was found.
		return fmt.Errorf("CIDR block %q not found", aws.ToString(e.CIDRBlock))
	}

	if changes.CIDRBlock != nil {
		request := &ec2.AssociateVpcCidrBlockInput{
			VpcId:     e.VPC.ID,
			CidrBlock: e.CIDRBlock,
		}

		_, err := t.Cloud.EC2().AssociateVpcCidrBlock(ctx, request)
		if err != nil {
			return fmt.Errorf("error associating AdditionalCIDR to VPC: %v", err)
		}
	}

	return nil // no tags
}

type terraformVPCCIDRBlock struct {
	VPCID     *terraformWriter.Literal `cty:"vpc_id"`
	CIDRBlock *string                  `cty:"cidr_block"`
}

func (_ *VPCCIDRBlock) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPCCIDRBlock) error {
	shared := aws.ToBool(e.Shared)
	if shared && a == nil {
		// VPC not owned by kOps, no changes will be applied
		// Verify that the CIDR block was found.
		return fmt.Errorf("CIDR block %q not found", aws.ToString(e.CIDRBlock))
	}

	// When this has been enabled please fix test TestAdditionalCIDR in integration_test.go to run runTestAWS.
	tf := &terraformVPCCIDRBlock{
		VPCID:     e.VPC.TerraformLink(),
		CIDRBlock: e.CIDRBlock,
	}

	// Terraform 0.12 doesn't support resource names that start with digits. See #7052
	// and https://www.terraform.io/upgrade-guides/0-12.html#pre-upgrade-checklist
	name := fmt.Sprintf("cidr-%v", *e.Name)
	return t.RenderResource("aws_vpc_ipv4_cidr_block_association", name, tf)
}
