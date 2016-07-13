package protokube

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"reflect"
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

	if !strings.Contains(p.zoneName, ".") {
		// Looks like a zone ID
		zoneID := p.zoneName
		glog.Infof("Querying for hosted zone by id: %q", zoneID)

		request := &route53.GetHostedZoneInput{
			Id: aws.String(zoneID),
		}

		response, err := p.client.GetHostedZone(request)
		if err != nil {
			if AWSErrorCode(err) == "NoSuchHostedZone" {
				glog.Infof("Zone not found with id %q; will reattempt by name", zoneID)
			} else {
				return nil, fmt.Errorf("error querying for DNS HostedZones %q: %v", zoneID, err)
			}
		} else {
			p.zone = response.HostedZone
			return p.zone, nil
		}
	}

	glog.Infof("Querying for hosted zone by name: %q", p.zoneName)

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

func (p *Route53DNSProvider) findResourceRecord(hostedZoneID string, name string, resourceType string) (*route53.ResourceRecordSet, error) {
	name = strings.TrimSuffix(name, ".")

	request := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		// TODO: Start at correct name?
	}

	var found *route53.ResourceRecordSet

	err := p.client.ListResourceRecordSetsPages(request, func(p *route53.ListResourceRecordSetsOutput, lastPage bool) (shouldContinue bool) {
		for _, rr := range p.ResourceRecordSets {
			if aws.StringValue(rr.Type) != resourceType {
				continue
			}

			rrName := aws.StringValue(rr.Name)
			rrName = strings.TrimSuffix(rrName, ".")

			if name == rrName {
				found = rr
				break
			}
		}

		// TODO: Also exit if we are on the 'next' name?

		return found == nil
	})

	if err != nil {
		return nil, fmt.Errorf("error listing DNS ResourceRecords: %v", err)
	}

	if found == nil {
		return nil, nil
	}

	return found, nil
}

func (p *Route53DNSProvider) Set(fqdn string, recordType string, value string, ttl time.Duration) error {
	zone, err := p.getZone()
	if err != nil {
		return err
	}

	// More correct, and makes the simple comparisons later on work correctly
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	existing, err := p.findResourceRecord(aws.StringValue(zone.Id), fqdn, recordType)
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

	if existing != nil {
		if reflect.DeepEqual(rrs, existing) {
			glog.V(2).Infof("DNS %q %s record already set to %q", fqdn, recordType, value)
			return nil
		} else {
			glog.Infof("ResourceRecordSet change:")
			glog.Infof("Existing: %v", DebugString(existing))
			glog.Infof("Desired:  %v", DebugString(rrs))
		}
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

// AWSErrorCode returns the aws error code, if it is an awserr.Error, otherwise ""
func AWSErrorCode(err error) string {
	if awsError, ok := err.(awserr.Error); ok {
		return awsError.Code()
	}
	return ""
}
