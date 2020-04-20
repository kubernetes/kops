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
	"k8s.io/klog"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
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
	cloud := c.Cloud.(awsup.AWSCloud)

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

	klog.V(4).Infof("found matching VPC %v", actual)

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
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the VPC was found and matches our required settings
		if a == nil {
			return fmt.Errorf("VPC with id %q not found", fi.StringValue(e.ID))
		}

		if changes != nil && changes.EnableDNSSupport != nil {
			if featureflag.VPCSkipEnableDNSSupport.Enabled() {
				klog.Warningf("VPC did not have EnableDNSSupport=true, but ignoring because of VPCSkipEnableDNSSupport feature-flag")
			} else {
				// TODO: We could easily just allow kops to fix this...
				return fmt.Errorf("VPC with id %q was set to be shared, but did not have EnableDNSSupport=true.", fi.StringValue(e.ID))
			}
		}
	}

	if a == nil {
		klog.V(2).Infof("Creating VPC with CIDR: %q", *e.CIDR)

		request := &ec2.CreateVpcInput{
			CidrBlock: e.CIDR,
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

type terraformVPC struct {
	CIDR               *string           `json:"cidr_block,omitempty" cty:"cidr_block"`
	EnableDNSHostnames *bool             `json:"enable_dns_hostnames,omitempty" cty:"enable_dns_hostnames"`
	EnableDNSSupport   *bool             `json:"enable_dns_support,omitempty" cty:"enable_dns_support"`
	Tags               map[string]string `json:"tags,omitempty" cty:"tags"`
}

func (_ *VPC) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPC) error {
	if err := t.AddOutputVariable("vpc_id", e.TerraformLink()); err != nil {
		return err
	}

	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed
		// We won't apply changes, but our validation (kops update) will still warn
		return nil
	}

	if err := t.AddOutputVariable("vpc_cidr_block", terraform.LiteralProperty("aws_vpc", *e.Name, "cidr_block")); err != nil {
		// TODO: Should we try to output vpc_cidr_block for shared vpcs?
		return err
	}

	tf := &terraformVPC{
		CIDR:               e.CIDR,
		Tags:               e.Tags,
		EnableDNSHostnames: e.EnableDNSHostnames,
		EnableDNSSupport:   e.EnableDNSSupport,
	}

	return t.RenderResource("aws_vpc", *e.Name, tf)
}

func (e *VPC) TerraformLink() *terraform.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if VPC is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing VPC with id %q", *e.ID)
		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralProperty("aws_vpc", *e.Name, "id")
}

type cloudformationVPC struct {
	CidrBlock          *string             `json:"CidrBlock,omitempty"`
	EnableDnsHostnames *bool               `json:"EnableDnsHostnames,omitempty"`
	EnableDnsSupport   *bool               `json:"EnableDnsSupport,omitempty"`
	Tags               []cloudformationTag `json:"Tags,omitempty"`
}

func (_ *VPC) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *VPC) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not cloudformation owned / managed
		// We won't apply changes, but our validation (kops update) will still warn
		return nil
	}

	tf := &cloudformationVPC{
		CidrBlock:          e.CIDR,
		EnableDnsHostnames: e.EnableDNSHostnames,
		EnableDnsSupport:   e.EnableDNSSupport,
		Tags:               buildCloudformationTags(e.Tags),
	}

	return t.RenderResource("AWS::EC2::VPC", *e.Name, tf)
}

func (e *VPC) CloudformationLink() *cloudformation.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if VPC is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing VPC with id %q", *e.ID)
		return cloudformation.LiteralString(*e.ID)
	}

	return cloudformation.Ref("AWS::EC2::VPC", *e.Name)
}
