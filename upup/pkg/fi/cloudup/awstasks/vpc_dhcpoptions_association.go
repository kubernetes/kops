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

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=VPCDHCPOptionsAssociation
type VPCDHCPOptionsAssociation struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	VPC         *VPC
	DHCPOptions *DHCPOptions
}

func (e *VPCDHCPOptionsAssociation) Find(c *fi.Context) (*VPCDHCPOptionsAssociation, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	vpcID := e.VPC.ID
	dhcpOptionsID := e.DHCPOptions.ID

	if vpcID == nil || dhcpOptionsID == nil {
		return nil, nil
	}

	vpc, err := cloud.DescribeVPC(*vpcID)
	if err != nil {
		return nil, err
	}

	actual := &VPCDHCPOptionsAssociation{}
	actual.VPC = &VPC{ID: vpc.VpcId}
	actual.DHCPOptions = &DHCPOptions{ID: vpc.DhcpOptionsId}

	// Prevent spurious changes
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *VPCDHCPOptionsAssociation) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *VPCDHCPOptionsAssociation) CheckChanges(a, e, changes *VPCDHCPOptionsAssociation) error {
	if e.VPC == nil {
		return fi.RequiredField("VPC")
	}
	if e.DHCPOptions == nil {
		return fi.RequiredField("DHCPOptions")
	}

	if a != nil && changes != nil {
		if changes.VPC != nil {
			// Should be impossible anyway because VPC is our primary key...
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *VPCDHCPOptionsAssociation) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *VPCDHCPOptionsAssociation) error {
	if changes.DHCPOptions != nil {
		klog.V(2).Infof("calling EC2 AssociateDhcpOptions")
		request := &ec2.AssociateDhcpOptionsInput{
			VpcId:         e.VPC.ID,
			DhcpOptionsId: e.DHCPOptions.ID,
		}

		_, err := t.Cloud.EC2().AssociateDhcpOptions(request)
		if err != nil {
			return fmt.Errorf("error creating VPCDHCPOptionsAssociation: %v", err)
		}
	}

	return nil // no tags
}

type terraformVPCDHCPOptionsAssociation struct {
	VPCID         *terraform.Literal `json:"vpc_id"`
	DHCPOptionsID *terraform.Literal `json:"dhcp_options_id"`
}

func (_ *VPCDHCPOptionsAssociation) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPCDHCPOptionsAssociation) error {
	tf := &terraformVPCDHCPOptionsAssociation{
		VPCID:         e.VPC.TerraformLink(),
		DHCPOptionsID: e.DHCPOptions.TerraformLink(),
	}

	return t.RenderResource("aws_vpc_dhcp_options_association", *e.Name, tf)
}

type cloudformationVPCDHCPOptionsAssociation struct {
	VpcId         *cloudformation.Literal `json:"VpcId"`
	DhcpOptionsId *cloudformation.Literal `json:"DhcpOptionsId"`
}

func (_ *VPCDHCPOptionsAssociation) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *VPCDHCPOptionsAssociation) error {
	tf := &cloudformationVPCDHCPOptionsAssociation{
		VpcId:         e.VPC.CloudformationLink(),
		DhcpOptionsId: e.DHCPOptions.CloudformationLink(),
	}

	return t.RenderResource("AWS::EC2::VPCDHCPOptionsAssociation", *e.Name, tf)
}
