package protokube

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"strings"
	"time"
)

type Route53DNSProvider struct {
	client *route53.Route53

	zoneName string
	zone     *route53.HostedZone
}

func NewRoute53DNSProvider(zoneName string) (*Route53DNSProvider, error) {
	if zoneName == "" {
		return nil, fmt.Errorf("zone name is required")
	}

	p := &Route53DNSProvider{
		zoneName: zoneName,
	}

	s := session.New()
	s.Handlers.Send.PushFront(func(r *request.Request) {
		// Log requests
		glog.V(4).Infof("AWS API Request: %s/%s", r.ClientInfo.ServiceName, r.Operation.Name)
	})

	config := aws.NewConfig()

	p.client = route53.New(s, config)

	return p, nil
}

func (p *Route53DNSProvider) getZone() (*route53.HostedZone, error) {
	if p.zone != nil {
		return p.zone, nil
	}

	findZone := p.zoneName
	if !strings.HasSuffix(findZone, ".") {
		findZone += "."
	}
	request := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(findZone),
	}

	response, err := p.client.ListHostedZonesByName(request)
	if err != nil {
		return nil, fmt.Errorf("error querying for DNS HostedZones %q: %v", findZone, err)
	}

	var zones []*route53.HostedZone
	for _, zone := range response.HostedZones {
		if aws.StringValue(zone.Name) == findZone {
			zones = append(zones, zone)
		}
	}
	if len(zones) == 0 {
		return nil, nil
	}
	if len(zones) != 1 {
		return nil, fmt.Errorf("found multiple hosted zones matched name %q", findZone)
	}

	p.zone = zones[0]

	return p.zone, nil
}

func (p *Route53DNSProvider) Set(fqdn string, recordType string, value string, ttl time.Duration) error {
	zone, err := p.getZone()
	if err != nil {
		return err
	}

	rrs := &route53.ResourceRecordSet{
		Name: aws.String(fqdn),
		Type: aws.String(recordType),
		TTL:  aws.Int64(int64(ttl.Seconds())),
		ResourceRecords: []*route53.ResourceRecord{
			{Value: aws.String(value)},
		},
	}

	change := &route53.Change{
		Action:            aws.String("UPSERT"),
		ResourceRecordSet: rrs,
	}

	changeBatch := &route53.ChangeBatch{}
	changeBatch.Changes = []*route53.Change{change}

	request := &route53.ChangeResourceRecordSetsInput{}
	request.HostedZoneId = zone.Id
	request.ChangeBatch = changeBatch

	glog.V(2).Infof("Updating DNS record %q", fqdn)
	glog.V(4).Infof("route53 request: %s", DebugString(request))

	response, err := p.client.ChangeResourceRecordSets(request)
	if err != nil {
		return fmt.Errorf("error creating ResourceRecordSets: %v", err)
	}

	glog.V(2).Infof("Change id is %q", aws.StringValue(response.ChangeInfo.Id))

	return nil
}
