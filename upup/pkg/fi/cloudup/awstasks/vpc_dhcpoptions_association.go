package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

type VPCDHCPOptionsAssociation struct {
	VPC         *VPC
	DHCPOptions *DHCPOptions
}

func (e *VPCDHCPOptionsAssociation) String() string {
	return fi.TaskAsString(e)
}

func (e *VPCDHCPOptionsAssociation) Find(c *fi.Context) (*VPCDHCPOptionsAssociation, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

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
		glog.V(2).Infof("calling EC2 AssociateDhcpOptions")
		request := &ec2.AssociateDhcpOptionsInput{
			VpcId:         e.VPC.ID,
			DhcpOptionsId: e.DHCPOptions.ID,
		}

		_, err := t.Cloud.EC2.AssociateDhcpOptions(request)
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

	return t.RenderResource("aws_vpc_dhcp_options_association", *e.VPC.Name, tf)
}
