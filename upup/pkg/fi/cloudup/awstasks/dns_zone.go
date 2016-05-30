package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"strings"
)

//go:generate fitask -type=DNSZone
type DNSZone struct {
	Name *string
	ID   *string
}

func (e *DNSZone) Find(c *fi.Context) (*DNSZone, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

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

	z := zones[0]

	actual := &DNSZone{}
	actual.Name = e.Name
	actual.ID = z.Id

	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
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

	return nil
}
