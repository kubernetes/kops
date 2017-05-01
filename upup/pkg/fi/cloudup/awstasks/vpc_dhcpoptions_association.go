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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=VPCDHCPOptionsAssociation
type VPCDHCPOptionsAssociation struct {
	Name *string

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
	if changes.DHCPOptions == nil {
		return nil
	}
	glog.V(2).Infof("calling EC2 AssociateDhcpOptions")
	request := &ec2.AssociateDhcpOptionsInput{
		VpcId:         e.VPC.ID,
		DhcpOptionsId: e.DHCPOptions.ID,
	}

	_, err := t.Cloud.EC2().AssociateDhcpOptions(request)
	if err != nil {
		return fmt.Errorf("error creating VPCDHCPOptionsAssociation: %v", err)
	}

	// This part is a little annoying. If you're running in a region
	// with where there is no default-looking DHCP option set, when
	// you create any VPC, AWS will create a default-looking DHCP
	// option set for you. If you then re-associate (as below) or
	// delete the VPC, the option set will hang around. However, if
	// you have a default-looking DHCP option set (for example, an
	// unmodified default VPC) and create a VPC, AWS will associate
	// the VPC with the DHCP option set of the default VPC. There's no
	// signal as to whether the option set returned is the default or
	// was created along with the VPC.
	//
	// Solution: When we reassociate the DHCP option set, try a
	// courtesy delete on it. If that gets a DependencyViolation,
	// it's still in use and we move on.
	if *a.DHCPOptions.ID == "default" {
		return nil
	}
	resp, err := t.Cloud.EC2().DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: []*string{a.DHCPOptions.ID}})
	if err != nil {
		glog.V(2).Infof("ignoring error describing old DHCP option set %q: %v", *a.DHCPOptions.ID, err)
		return nil
	}
	if len(resp.DhcpOptions) != 1 {
		glog.V(2).Infof("old DHCP option set %q not found", *a.DHCPOptions.ID)
		return nil
	}
	opt := resp.DhcpOptions[0]
	if len(opt.Tags) != 0 {
		glog.V(2).Infof("old DHCP option set %q was tagged, not deleting", *a.DHCPOptions.ID)
		return nil
	}
	for _, conf := range opt.DhcpConfigurations {
		if *conf.Key == "domain-name" {
			var domain string
			if t.Cloud.Region() == "us-east-1" {
				domain = "ec2.internal"
			} else {
				domain = t.Cloud.Region() + ".compute.internal"
			}
			if len(conf.Values) != 1 || *conf.Values[0].Value != domain {
				glog.V(2).Infof("old DHCP option set %q has mismatched domain name, not deleting", *a.DHCPOptions.ID)
				return nil
			}
		} else if *conf.Key == "domain-name-servers" {
			if len(conf.Values) != 1 || *conf.Values[0].Value != "AmazonProvidedDNS" {
				glog.V(2).Infof("old DHCP option set %q has mismatched domain name servers, not deleting", *a.DHCPOptions.ID)
				return nil
			}
		} else {
			glog.V(2).Infof("old DHCP option set %q has unknown config key %q, not deleting", *a.DHCPOptions.ID, conf.Key)
			return nil
		}
	}
	glog.V(2).Infof("attempting to delete replaced default DHCP option set %q", *a.DHCPOptions.ID)
	if _, err := t.Cloud.EC2().DeleteDhcpOptions(&ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: a.DHCPOptions.ID,
	}); err != nil {
		if awsup.AWSErrorCode(err) != "DependencyViolation" {
			return fmt.Errorf("deleting disassociated DHCP option set %q failed: %v",
				*a.DHCPOptions.ID, err)
		}
		glog.V(2).Infof("ignoring DependencyViolation deleting disassociated DHCP option set %q", *a.DHCPOptions.ID)
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
