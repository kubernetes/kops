package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"strings"
)

//go:generate fitask -type=DNSZone
type DNSZone struct {
	Name *string
	ID   *string
}

var _ fi.CompareWithID = &DNSZone{}

func (e *DNSZone) CompareWithID() *string {
	return e.Name
}

func (e *DNSZone) Find(c *fi.Context) (*DNSZone, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	z, err := e.findExisting(cloud)
	if err != nil {
		return nil, err
	}

	if z == nil {
		return nil, nil
	}

	actual := &DNSZone{}
	actual.Name = e.Name
	actual.ID = z.Id

	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (e *DNSZone) findExisting(cloud *awsup.AWSCloud) (*route53.HostedZone, error) {
	findName := fi.StringValue(e.Name)
	if findName == "" {
		return nil, nil
	}
	if !strings.HasSuffix(findName, ".") {
		findName += "."
	}
	request := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(findName),
	}

	response, err := cloud.Route53.ListHostedZonesByName(request)
	if err != nil {
		return nil, fmt.Errorf("error listing DNS HostedZones: %v", err)
	}

	var zones []*route53.HostedZone
	for _, zone := range response.HostedZones {
		if aws.StringValue(zone.Name) == findName {
			zones = append(zones, zone)
		}
	}
	if len(zones) == 0 {
		return nil, nil
	}
	if len(zones) != 1 {
		return nil, fmt.Errorf("found multiple hosted zones matched name %q", findName)
	}

	return zones[0], nil
}

func (e *DNSZone) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *DNSZone) CheckChanges(a, e, changes *DNSZone) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (_ *DNSZone) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *DNSZone) error {
	if a == nil {
		request := &route53.CreateHostedZoneInput{}
		request.Name = e.Name

		glog.V(2).Infof("Creating Route53 HostedZone with Name %q", e.Name)

		response, err := t.Cloud.Route53.CreateHostedZone(request)
		if err != nil {
			return fmt.Errorf("error creating DNS HostedZone: %v", err)
		}

		e.ID = response.HostedZone.Id
	}

	// We don't tag the zone - we expect it to be shared
	return nil
}

type terraformRoute53Zone struct {
	Name *string           `json:"name"`
	Tags map[string]string `json:"tags,omitempty"`
}

func (_ *DNSZone) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DNSZone) error {
	cloud := t.Cloud.(*awsup.AWSCloud)

	// As a special case, we check for an existing zone
	// It is really painful to have TF create a new one...
	// (you have to reconfigure the DNS NS records)
	glog.Infof("Check for existing route53 zone to re-use with name %q", *e.Name)
	z, err := e.findExisting(cloud)
	if err != nil {
		return err
	}

	if z != nil {
		glog.Infof("Existing zone %q found; will configure TF to reuse", aws.StringValue(z.Name))

		e.ID = z.Id
		return nil
	}

	tf := &terraformRoute53Zone{
		Name: e.Name,
		//Tags:               cloud.BuildTags(e.Name, nil),
	}

	return t.RenderResource("aws_route53_zone", *e.Name, tf)
}

func (e *DNSZone) TerraformLink() *terraform.Literal {
	if e.ID != nil {
		glog.V(4).Infof("reusing existing route53 zone with id %q", *e.ID)
		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralSelfLink("aws_route53_zone", *e.Name)
}
