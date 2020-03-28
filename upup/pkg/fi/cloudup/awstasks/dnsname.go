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

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=DNSName
type DNSName struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID           *string
	Zone         *DNSZone
	ResourceType *string

	TargetLoadBalancer *LoadBalancer
}

func (e *DNSName) Find(c *fi.Context) (*DNSName, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.Zone == nil || e.Zone.ZoneID == nil {
		klog.V(4).Infof("Zone / ZoneID not found for %s, skipping Find", fi.StringValue(e.Name))
		return nil, nil
	}

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
		HostedZoneId: e.Zone.ZoneID,
		// TODO: Start at correct name?
	}

	var found *route53.ResourceRecordSet

	err := cloud.Route53().ListResourceRecordSetsPages(request, func(p *route53.ListResourceRecordSetsOutput, lastPage bool) (shouldContinue bool) {
		for _, rr := range p.ResourceRecordSets {
			resourceType := aws.StringValue(rr.Type)
			name := aws.StringValue(rr.Name)

			klog.V(4).Infof("Found DNS resource %q %q", resourceType, name)

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
	actual.Lifecycle = e.Lifecycle

	if found.AliasTarget != nil {
		dnsName := aws.StringValue(found.AliasTarget.DNSName)
		klog.Infof("AliasTarget for %q is %q", aws.StringValue(found.Name), dnsName)
		if dnsName != "" {
			// TODO: check "looks like" an ELB?
			lb, err := findLoadBalancerByAlias(cloud, found.AliasTarget)
			if err != nil {
				return nil, fmt.Errorf("error mapping DNSName %q to LoadBalancer: %v", dnsName, err)
			}
			if lb == nil {
				klog.Warningf("Unable to find load balancer with DNS name: %q", dnsName)
			} else {
				loadBalancerName := aws.StringValue(lb.LoadBalancerName)
				tagMap, err := describeLoadBalancerTags(cloud, []string{loadBalancerName})
				if err != nil {
					return nil, err
				}
				tags := tagMap[loadBalancerName]
				nameTag, _ := awsup.FindELBTag(tags, "Name")
				if nameTag == "" {
					return nil, fmt.Errorf("Found ELB %q linked to DNS name %q, but it did not have a Name tag", loadBalancerName, fi.StringValue(e.Name))
				}
				actual.TargetLoadBalancer = &LoadBalancer{Name: fi.String(nameTag)}
			}
		}
	}

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
	request.HostedZoneId = e.Zone.ZoneID
	request.ChangeBatch = changeBatch

	klog.V(2).Infof("Updating DNS record %q", *e.Name)

	response, err := t.Cloud.Route53().ChangeResourceRecordSets(request)
	if err != nil {
		return fmt.Errorf("error creating ResourceRecordSets: %v", err)
	}

	klog.V(2).Infof("Change id is %q", aws.StringValue(response.ChangeInfo.Id))

	return nil
}

type terraformRoute53Record struct {
	Name    *string  `json:"name" cty:"name"`
	Type    *string  `json:"type" cty:"type"`
	TTL     *string  `json:"ttl,omitempty" cty:"ttl"`
	Records []string `json:"records,omitempty" cty:"records"`

	Alias  *terraformAlias    `json:"alias,omitempty" cty:"alias"`
	ZoneID *terraform.Literal `json:"zone_id" cty:"zone_id"`
}

type terraformAlias struct {
	Name                 *terraform.Literal `json:"name" cty:"name"`
	ZoneID               *terraform.Literal `json:"zone_id" cty:"zone_id"`
	EvaluateTargetHealth *bool              `json:"evaluate_target_health" cty:"evaluate_target_health"`
}

func (_ *DNSName) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DNSName) error {
	tf := &terraformRoute53Record{
		Name:   e.Name,
		ZoneID: e.Zone.TerraformLink(),
		Type:   e.ResourceType,
	}

	if e.TargetLoadBalancer != nil {
		tf.Alias = &terraformAlias{
			Name:                 e.TargetLoadBalancer.TerraformLink("dns_name"),
			EvaluateTargetHealth: aws.Bool(false),
			ZoneID:               e.TargetLoadBalancer.TerraformLink("zone_id"),
		}
	}

	return t.RenderResource("aws_route53_record", *e.Name, tf)
}

func (e *DNSName) TerraformLink() *terraform.Literal {
	return terraform.LiteralSelfLink("aws_route53_record", *e.Name)
}

type cloudformationRoute53Record struct {
	Name            *string  `json:"Name"`
	Type            *string  `json:"Type"`
	TTL             *string  `json:"TTL,omitempty"`
	ResourceRecords []string `json:"ResourceRecords,omitempty"`

	AliasTarget *cloudformationAlias    `json:"AliasTarget,omitempty"`
	ZoneID      *cloudformation.Literal `json:"HostedZoneId"`
}

type cloudformationAlias struct {
	DNSName              *cloudformation.Literal `json:"DNSName,omitempty"`
	ZoneID               *cloudformation.Literal `json:"HostedZoneId,omitempty"`
	EvaluateTargetHealth *bool                   `json:"EvaluateTargetHealth,omitempty"`
}

func (_ *DNSName) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *DNSName) error {
	cf := &cloudformationRoute53Record{
		Name:   e.Name,
		ZoneID: e.Zone.CloudformationLink(),
		Type:   e.ResourceType,
	}

	if e.TargetLoadBalancer != nil {
		cf.AliasTarget = &cloudformationAlias{
			DNSName:              e.TargetLoadBalancer.CloudformationAttrDNSName(),
			EvaluateTargetHealth: aws.Bool(false),
			ZoneID:               e.TargetLoadBalancer.CloudformationAttrCanonicalHostedZoneNameID(),
		}
	}

	return t.RenderResource("AWS::Route53::RecordSet", *e.Name, cf)
}

func (e *DNSName) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::Route53::RecordSet", *e.Name)
}
