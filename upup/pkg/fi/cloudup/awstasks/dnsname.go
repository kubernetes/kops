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

//go:generate fitask -type=DNSName
type DNSName struct {
	Name         *string
	ID           *string
	Zone         *DNSZone
	ResourceType *string

	TargetLoadBalancer *LoadBalancer
}

func (e *DNSName) Find(c *fi.Context) (*DNSName, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	findName := fi.StringValue(e.Name)
	if findName == "" {
		return nil, nil
	}
	findName = strings.TrimSuffix(findName, ".")

	findType := fi.StringValue(e.ResourceType)
	if findType == "" {
		return nil, nil
	}

	request := &route53.ListResourceRecordSetsInput{
		HostedZoneId: e.Zone.ID,
		// TODO: Start at correct name?
	}

	var found *route53.ResourceRecordSet

	err := cloud.Route53.ListResourceRecordSetsPages(request, func(p *route53.ListResourceRecordSetsOutput, lastPage bool) (shouldContinue bool) {
		for _, rr := range p.ResourceRecordSets {
			resourceType := aws.StringValue(rr.Type)
			name := aws.StringValue(rr.Name)

			glog.V(4).Infof("Found DNS resource %q %q", resourceType, name)

			if findType != resourceType {
				continue
			}

			name = strings.TrimSuffix(name, ".")

			if name == findName {
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

	actual := &DNSName{}
	actual.Zone = e.Zone
	actual.Name = e.Name
	actual.ResourceType = e.ResourceType

	return actual, nil
}

func (e *DNSName) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *DNSName) CheckChanges(a, e, changes *DNSName) error {
	if a == nil {
		if fi.StringValue(e.Name) == "" {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *DNSName) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *DNSName) error {
	rrs := &route53.ResourceRecordSet{
		Name: e.Name,
		Type: e.ResourceType,
	}

	if e.TargetLoadBalancer != nil {
		rrs.AliasTarget = &route53.AliasTarget{
			DNSName:              e.TargetLoadBalancer.DNSName,
			EvaluateTargetHealth: aws.Bool(false),
			HostedZoneId:         e.TargetLoadBalancer.HostedZoneId,
		}
	}

	change := &route53.Change{
		Action:            aws.String("UPSERT"),
		ResourceRecordSet: rrs,
	}

	changeBatch := &route53.ChangeBatch{}
	changeBatch.Changes = []*route53.Change{change}

	request := &route53.ChangeResourceRecordSetsInput{}
	request.HostedZoneId = e.Zone.ID
	request.ChangeBatch = changeBatch

	glog.V(2).Infof("Updating DNS record %q", *e.Name)

	response, err := t.Cloud.Route53.ChangeResourceRecordSets(request)
	if err != nil {
		return fmt.Errorf("error creating ResourceRecordSets: %v", err)
	}

	glog.V(2).Infof("Change id is %q", aws.StringValue(response.ChangeInfo.Id))

	return nil
}

type terraformRoute53Record struct {
	Name    *string  `json:"name"`
	Type    *string  `json:"type"`
	TTL     *string  `json:"ttl"`
	Records []string `json:"records"`

	Alias  *terraformAlias    `json:"alias"`
	ZoneID *terraform.Literal `json:"zone_id"`
}

type terraformAlias struct {
	Name                 *string `json:"name"`
	HostedZoneId         *string `json:"zone_id"`
	EvaluateTargetHealth *bool   `json:"evaluate_target_health"`
}

func (_ *DNSName) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DNSName) error {
	tf := &terraformRoute53Record{
		Name:   e.Name,
		ZoneID: e.Zone.TerraformLink(),
		Type:   e.ResourceType,
	}

	if e.TargetLoadBalancer != nil {
		tf.Alias = &terraformAlias{
			Name:                 e.TargetLoadBalancer.DNSName,
			EvaluateTargetHealth: aws.Bool(false),
			HostedZoneId:         e.TargetLoadBalancer.HostedZoneId,
		}
	}

	return t.RenderResource("aws_route53_record", *e.Name, tf)
}

func (e *DNSName) TerraformLink() *terraform.Literal {
	return terraform.LiteralSelfLink("aws_route53_record", *e.Name)
}
