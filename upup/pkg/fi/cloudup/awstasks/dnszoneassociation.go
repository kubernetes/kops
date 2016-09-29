package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

//go:generate fitask -type=DNSZoneAssociation
type DNSZoneAssociation struct {
	Name    *string
	DNSZone *DNSZone
	VPC     *VPC
}

var _ fi.CompareWithID = &AutoscalingGroup{}

func (e *DNSZoneAssociation) CompareWithID() *string {
	return e.Name
}

func (e *DNSZoneAssociation) Find(c *fi.Context) (*DNSZoneAssociation, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	dnsZoneID := e.DNSZone.ID
	vpcID := e.VPC.ID

	if dnsZoneID == nil || vpcID == nil {
		return nil, nil
	}

	request := &route53.GetHostedZoneInput{
		Id: dnsZoneID,
	}

	response, err := cloud.Route53().GetHostedZone(request)
	if err != nil {
		return nil, fmt.Errorf("error listing HostedZone %q: %v", dnsZoneID, err)
	}
	if response == nil {
		return nil, nil
	}

	for _, v := range response.VPCs {
		if aws.StringValue(v.VPCId) != *vpcID {
			continue
		}
		if aws.StringValue(v.VPCRegion) != cloud.Region() {
			// This seems incredibly unlikely
			glog.Warningf("Found matching VPC ID, but region mismatch: %q vs %q", aws.StringValue(v.VPCRegion), cloud.Region())
			continue
		}
		actual := &DNSZoneAssociation{
			DNSZone: e.DNSZone,
			VPC:     e.VPC,
		}
		glog.V(2).Infof("found matching DNSZoneAssociation")
		return actual, nil
	}

	return nil, nil
}

func (e *DNSZoneAssociation) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *DNSZoneAssociation) CheckChanges(a, e, changes *DNSZoneAssociation) error {
	basePath := field.NewPath("")
	if e.VPC == nil {
		return field.Required(basePath.Child("VPC"), "")
	}
	if e.DNSZone == nil {
		return field.Required(basePath.Child("DNSZone"), "")
	}
	return nil
}

func (_ *DNSZoneAssociation) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *DNSZoneAssociation) error {
	if a == nil {
		glog.V(2).Infof("Associating VPC %q with hosted zone %q", fi.StringValue(e.VPC.ID), fi.StringValue(e.DNSZone.ID))
		request := &route53.AssociateVPCWithHostedZoneInput{
			HostedZoneId: e.DNSZone.ID,
			VPC: &route53.VPC{
				VPCId: e.VPC.ID,
				VPCRegion: aws.String(t.Cloud.Region()),
			},
		}

		_, err := t.Cloud.Route53().AssociateVPCWithHostedZone(request)
		if err != nil {
			return fmt.Errorf("error associating VPC %q with hosted zone %q: %v", fi.StringValue(e.VPC.ID), fi.StringValue(e.DNSZone.ID), err)
		}
	}

	return nil // no tags
}

type terraformDNSZoneAssociation struct {
	VpcID     *terraform.Literal `json:"vpc_id,omitempty"`
	VpcRegion *string `json:"vpc_region,omitempty"`
	ZoneID    *terraform.Literal `json:"zone_id,omitempty"`
}

func (_ *DNSZoneAssociation) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DNSZoneAssociation) error {
	tf := &terraformDNSZoneAssociation{
		VpcID:     e.VPC.TerraformLink(),
		VpcRegion: fi.String(t.Cloud.(awsup.AWSCloud).Region()),
		ZoneID: e.DNSZone.TerraformLink(),
	}

	return t.RenderResource("aws_route53_zone_association", *e.Name, tf)
}
