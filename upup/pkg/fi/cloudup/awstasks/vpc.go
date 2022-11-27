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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type VPC struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID   *string
	CIDR *string

	// AmazonIPv6 is used only for Terraform rendering.
	// Direct rendering is handled via the VPCAmazonIPv6CIDRBlock task
	AmazonIPv6 *bool
	IPv6CIDR   *string

	EnableDNSHostnames *bool
	EnableDNSSupport   *bool

	// Shared is set if this is a shared VPC
	Shared *bool

	Tags map[string]string

	// AssociateExtraCIDRBlocks contains a list of cidr blocks that should be
	// associated with the VPC; any other CIDR blocks should be disassociated.
	// The associations themselves are created through the VPCCIDRBlock awstask.
	AssociateExtraCIDRBlocks []string
}

var (
	_ fi.CompareWithID     = &VPC{}
	_ fi.ProducesDeletions = &VPC{}
)

func (e *VPC) CompareWithID() *string {
	return e.ID
}

func (e *VPC) Find(c *fi.Context) (*VPC, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &ec2.DescribeVpcsInput{}

	if fi.ValueOf(e.ID) != "" {
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
		ID:         vpc.VpcId,
		CIDR:       vpc.CidrBlock,
		AmazonIPv6: aws.Bool(false),
		Name:       findNameTag(vpc.Tags),
		Tags:       intersectTags(vpc.Tags, e.Tags),
	}

	klog.V(4).Infof("found matching VPC %v", actual)

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if association == nil || association.Ipv6CidrBlockState == nil {
			continue
		}

		state := aws.StringValue(association.Ipv6CidrBlockState.State)
		if state != ec2.VpcCidrBlockStateCodeAssociated && state != ec2.VpcCidrBlockStateCodeAssociating {
			continue
		}

		pool := aws.StringValue(association.Ipv6Pool)
		if pool == "Amazon" {
			actual.AmazonIPv6 = aws.Bool(true)
			actual.IPv6CIDR = association.Ipv6CidrBlock
			e.IPv6CIDR = association.Ipv6CidrBlock
			break
		} else if actual.IPv6CIDR == nil {
			actual.IPv6CIDR = association.Ipv6CidrBlock
			e.IPv6CIDR = association.Ipv6CidrBlock
		}
	}

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
	actual.Name = e.Name // Name is part of Tags
	actual.AssociateExtraCIDRBlocks = e.AssociateExtraCIDRBlocks

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
			return fi.FieldIsImmutable(e.CIDR, a.CIDR, field.NewPath("CIDR"))
		}
	}
	return nil
}

func (e *VPC) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *VPC) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *VPC) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Verify the VPC was found and matches our required settings
		if a == nil {
			return fmt.Errorf("VPC with id %q not found", fi.ValueOf(e.ID))
		}

		if changes != nil && changes.EnableDNSSupport != nil {
			if featureflag.VPCSkipEnableDNSSupport.Enabled() {
				klog.Warningf("VPC did not have EnableDNSSupport=true, but ignoring because of VPCSkipEnableDNSSupport feature-flag")
			} else {
				// TODO: We could easily just allow kops to fix this...
				return fmt.Errorf("VPC with id %q was set to be shared, but did not have EnableDNSSupport=true.", fi.ValueOf(e.ID))
			}
		}
	}

	if a == nil {
		klog.V(2).Infof("Creating VPC with CIDR: %q", *e.CIDR)

		request := &ec2.CreateVpcInput{
			CidrBlock:         e.CIDR,
			TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeVpc, e.Tags),
		}

		response, err := t.Cloud.EC2().CreateVpc(request)
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

		_, err := t.Cloud.EC2().ModifyVpcAttribute(request)
		if err != nil {
			return fmt.Errorf("error modifying VPC attribute: %v", err)
		}
	}

	if changes.EnableDNSHostnames != nil {
		request := &ec2.ModifyVpcAttributeInput{
			VpcId:              e.ID,
			EnableDnsHostnames: &ec2.AttributeBooleanValue{Value: changes.EnableDNSHostnames},
		}

		_, err := t.Cloud.EC2().ModifyVpcAttribute(request)
		if err != nil {
			return fmt.Errorf("error modifying VPC attribute: %v", err)
		}
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

func (e *VPC) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	if fi.IsNilOrEmpty(e.ID) || fi.ValueOf(e.Shared) {
		return nil, nil
	}

	var removals []fi.Deletion
	request := &ec2.DescribeVpcsInput{
		VpcIds: []*string{e.ID},
	}
	cloud := c.Cloud.(awsup.AWSCloud)
	response, err := cloud.EC2().DescribeVpcs(request)
	if err != nil {
		return nil, err
	}
	if response == nil || len(response.Vpcs) == 0 {
		return nil, nil
	}

	if len(response.Vpcs) != 1 {
		return nil, fmt.Errorf("found multiple VPCs matching tags")
	}
	vpc := response.Vpcs[0]
	for _, association := range vpc.CidrBlockAssociationSet {
		// We'll only delete CIDR associations that are not the primary association
		// and that have a state of "associated"
		if fi.ValueOf(association.CidrBlock) == fi.ValueOf(vpc.CidrBlock) ||
			association.CidrBlockState != nil && fi.ValueOf(association.CidrBlockState.State) != ec2.VpcCidrBlockStateCodeAssociated {
			continue
		}
		match := false
		for _, cidr := range e.AssociateExtraCIDRBlocks {
			if fi.ValueOf(association.CidrBlock) == cidr {
				match = true
				break
			}
		}
		if !match {
			removals = append(removals, &deleteVPCCIDRBlock{
				vpcID:         vpc.VpcId,
				cidrBlock:     association.CidrBlock,
				associationID: association.AssociationId,
			})
		}
	}
	return removals, nil
}

type terraformVPCData struct {
	ID *string `cty:"id"`
}

type terraformVPC struct {
	CIDR               *string           `cty:"cidr_block"`
	EnableDNSHostnames *bool             `cty:"enable_dns_hostnames"`
	EnableDNSSupport   *bool             `cty:"enable_dns_support"`
	AmazonIPv6         *bool             `cty:"assign_generated_ipv6_cidr_block"`
	Tags               map[string]string `cty:"tags"`
}

func (_ *VPC) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPC) error {
	if err := t.AddOutputVariable("vpc_id", e.TerraformLink()); err != nil {
		return err
	}

	cidrPrefixLengthCaptureList := terraformWriter.LiteralFunctionExpression("regex",
		terraformWriter.LiteralFromStringValue(".*/(\\\\d+)"),
		terraformWriter.LiteralTokens("local", "vpc_ipv6_cidr_block"),
	)
	cidrPrefixLengthString := terraformWriter.LiteralIndexExpression(
		cidrPrefixLengthCaptureList,
		terraformWriter.LiteralFromIntValue(0),
	)
	if err := t.AddOutputVariable("vpc_ipv6_cidr_length", terraformWriter.LiteralNullConditionalExpression(
		terraformWriter.LiteralTokens("local", "vpc_ipv6_cidr_block"),
		terraformWriter.LiteralFunctionExpression("tonumber", cidrPrefixLengthString),
	)); err != nil {
		return err
	}

	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed
		// We won't apply changes, but our validation (kops update) will still warn

		if err := t.AddOutputVariable("vpc_cidr_block", terraformWriter.LiteralData("aws_vpc", *e.Name, "cidr_block")); err != nil {
			return err
		}

		if err := t.AddOutputVariable("vpc_ipv6_cidr_block", terraformWriter.LiteralData("aws_vpc", *e.Name, "ipv6_cidr_block")); err != nil {
			return err
		}

		tf := terraformVPCData{
			ID: e.ID,
		}

		return t.RenderDataSource("aws_vpc", *e.Name, tf)
	}

	if err := t.AddOutputVariable("vpc_cidr_block", terraformWriter.LiteralProperty("aws_vpc", *e.Name, "cidr_block")); err != nil {
		return err
	}

	if err := t.AddOutputVariable("vpc_ipv6_cidr_block", terraformWriter.LiteralProperty("aws_vpc", *e.Name, "ipv6_cidr_block")); err != nil {
		return err
	}

	tf := &terraformVPC{
		CIDR:               e.CIDR,
		Tags:               e.Tags,
		EnableDNSHostnames: e.EnableDNSHostnames,
		EnableDNSSupport:   e.EnableDNSSupport,
		AmazonIPv6:         e.AmazonIPv6,
	}

	return t.RenderResource("aws_vpc", *e.Name, tf)
}

func (e *VPC) TerraformLink() *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if VPC is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing VPC with id %q", *e.ID)
		return terraformWriter.LiteralFromStringValue(*e.ID)
	}

	return terraformWriter.LiteralProperty("aws_vpc", *e.Name, "id")
}

type deleteVPCCIDRBlock struct {
	vpcID         *string
	cidrBlock     *string
	associationID *string
}

var _ fi.Deletion = &deleteVPCCIDRBlock{}

func (d *deleteVPCCIDRBlock) Delete(t fi.Target) error {
	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}
	request := &ec2.DisassociateVpcCidrBlockInput{
		AssociationId: d.associationID,
	}
	_, err := awsTarget.Cloud.EC2().DisassociateVpcCidrBlock(request)
	return err
}

func (d *deleteVPCCIDRBlock) TaskName() string {
	return "VPCCIDRBlock"
}

func (d *deleteVPCCIDRBlock) Item() string {
	return fmt.Sprintf("%v: cidr=%v", *d.vpcID, *d.cidrBlock)
}
