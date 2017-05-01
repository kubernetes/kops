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

package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
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

	// SharedID is set to the ID if this is shared, matching by ID
	SharedID *string
	// SharedNetworkKey is set to the ID if this is shared, matching by network key
	SharedNetworkKey *string

	Tags map[string]string
}

var _ fi.CompareWithID = &VPC{}

func (e *VPC) CompareWithID() *string {
	return e.ID
}

func buildEc2FiltersForSharedNetworkKey(name *string, sharedNetworkKey string) []*ec2.Filter {
	filters := []*ec2.Filter{
		awsup.NewEC2Filter("tag:"+awsup.TagNameSharedNetworkKey, sharedNetworkKey),
	}
	if fi.StringValue(name) != "" {
		filters = append(filters, awsup.NewEC2Filter("tag:Name", *name))
	}
	return filters
}

func intersectTags(tags []*ec2.Tag, desired map[string]string) map[string]string {
	if tags == nil {
		return nil
	}
	actual := make(map[string]string)
	for _, t := range tags {
		k := aws.StringValue(t.Key)
		v := aws.StringValue(t.Value)

		if _, found := desired[k]; found {
			actual[k] = v
		}
	}
	return actual
}

func (e *VPC) Find(c *fi.Context) (*VPC, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &ec2.DescribeVpcsInput{}

	if fi.StringValue(e.ID) != "" {
		request.VpcIds = []*string{e.ID}
	} else if fi.StringValue(e.SharedNetworkKey) != "" {
		request.Filters = buildEc2FiltersForSharedNetworkKey(e.Name, *e.SharedNetworkKey)
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
	actual.SharedID = e.SharedID
	actual.SharedNetworkKey = e.SharedNetworkKey
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

func (e *VPC) isShared() bool {
	return fi.StringValue(e.SharedNetworkKey) != "" || e.isSharedByID()
}

func (e *VPC) isSharedByID() bool {
	return fi.StringValue(e.SharedID) != ""
}

func (_ *VPC) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *VPC) error {
	if e.isSharedByID() {
		// If we're sharing by ID, verify the VPC was found and matches our required settings
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

	tags := e.Tags
	if e.isSharedByID() {
		// Don't tag shared resources
		tags = nil
	}
	return t.AddAWSTags(*e.ID, tags)
}

type terraformVPC struct {
	CIDR               *string           `json:"cidr_block,omitempty"`
	EnableDNSHostnames *bool             `json:"enable_dns_hostnames,omitempty"`
	EnableDNSSupport   *bool             `json:"enable_dns_support,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
}

func (_ *VPC) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPC) error {
	if err := t.AddOutputVariable("vpc_id", e.TerraformLink()); err != nil {
		return err
	}

	if e.isShared() {
		// Not terraform owned / managed
		return nil
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
	if e.isShared() {
		if e.ID == nil {
			glog.Fatalf("ID must be set, if VPC is shared: %s", e)
		}

		glog.V(4).Infof("reusing existing VPC with id %q", *e.ID)
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
	if e.isShared() {
		// Not cloudformation owned / managed
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
	if e.isShared() {
		if e.ID == nil {
			glog.Fatalf("ID must be set, if VPC is shared: %s", e)
		}

		glog.V(4).Infof("reusing existing VPC with id %q", *e.ID)
		return cloudformation.LiteralString(*e.ID)
	}

	return cloudformation.Ref("AWS::EC2::VPC", *e.Name)
}
