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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type DNSName struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID           *string
	Zone         *DNSZone
	ResourceName *string
	ResourceType *string

	TargetLoadBalancer DNSTarget
}

type DNSTarget interface {
	fi.Task
	getDNSName() *string
	getHostedZoneId() *string
	CloudformationAttrDNSName() *cloudformation.Literal
	CloudformationAttrCanonicalHostedZoneNameID() *cloudformation.Literal
	TerraformLink(...string) *terraformWriter.Literal
}

func (e *DNSName) Find(c *fi.Context) (*DNSName, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.Zone == nil || e.Zone.ZoneID == nil {
		klog.V(4).Infof("Zone / ZoneID not found for %s, skipping Find", fi.ValueOf(e.ResourceName))
		return nil, nil
	}

	findName := fi.ValueOf(e.ResourceName)
	if findName == "" {
		return nil, nil
	}
	findName = strings.TrimSuffix(findName, ".")

	findType := fi.ValueOf(e.ResourceType)
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
	actual.Name = e.Name
	actual.Zone = e.Zone
	actual.ResourceName = e.ResourceName
	actual.ResourceType = e.ResourceType
	actual.Lifecycle = e.Lifecycle

	if found.AliasTarget != nil {
		dnsName := aws.StringValue(found.AliasTarget.DNSName)
		klog.Infof("AliasTarget for %q is %q", aws.StringValue(found.Name), dnsName)
		if dnsName != "" {
			if actual.TargetLoadBalancer, err = findDNSTarget(cloud, found.AliasTarget, dnsName, e.ResourceName); err != nil {
				return nil, err
			}
		}
	}

	return actual, nil
}

func findDNSTarget(cloud awsup.AWSCloud, aliasTarget *route53.AliasTarget, dnsName string, targetDNSName *string) (DNSTarget, error) {
	// TODO: I would like to search dnsName for presence of ".elb" or ".nlb" to simply searching, however both nlb and elb have .elb. in the name at present
	if ELB, err := findDNSTargetELB(cloud, aliasTarget, dnsName, targetDNSName); err != nil {
		return nil, err
	} else if ELB != nil {
		return ELB, nil
	}

	if NLB, err := findDNSTargetNLB(cloud, aliasTarget, dnsName, targetDNSName); err != nil {
		return nil, err
	} else if NLB != nil {
		return NLB, nil
	}

	return nil, nil
}

func findDNSTargetNLB(cloud awsup.AWSCloud, aliasTarget *route53.AliasTarget, dnsName string, targetDNSName *string) (DNSTarget, error) {
	lb, err := findNetworkLoadBalancerByAlias(cloud, aliasTarget)
	if err != nil {
		return nil, fmt.Errorf("error mapping DNSName %q to LoadBalancer: %v", dnsName, err)
	}
	if lb != nil {
		loadBalancerName := aws.StringValue(lb.LoadBalancerName) // TODO: can we keep these on object
		loadBalancerArn := aws.StringValue(lb.LoadBalancerArn)   // TODO: can we keep these on object
		tagMap, err := cloud.DescribeELBV2Tags([]string{loadBalancerArn})
		if err != nil {
			return nil, err
		}
		tags := tagMap[loadBalancerArn]
		nameTag, _ := awsup.FindELBV2Tag(tags, "Name")
		if nameTag == "" {
			return nil, fmt.Errorf("Found NLB %q linked to DNS name %q, but it did not have a Name tag", loadBalancerName, fi.ValueOf(targetDNSName))
		}
		nameTag = strings.Replace(nameTag, ".", "-", -1)
		return &NetworkLoadBalancer{Name: fi.PtrTo(nameTag)}, nil
	}
	return nil, nil
}

func findDNSTargetELB(cloud awsup.AWSCloud, aliasTarget *route53.AliasTarget, dnsName string, targetDNSName *string) (DNSTarget, error) {
	lb, err := findLoadBalancerByAlias(cloud, aliasTarget)
	if err != nil {
		return nil, fmt.Errorf("error mapping DNSName %q to LoadBalancer: %v", dnsName, err)
	}
	if lb != nil {
		loadBalancerName := aws.StringValue(lb.LoadBalancerName)
		tagMap, err := cloud.DescribeELBTags([]string{loadBalancerName})
		if err != nil {
			return nil, err
		}
		tags := tagMap[loadBalancerName]
		nameTag, _ := awsup.FindELBTag(tags, "Name")
		if nameTag == "" {
			return nil, fmt.Errorf("Found ELB %q linked to DNS name %q, but it did not have a Name tag", loadBalancerName, fi.ValueOf(targetDNSName))
		}
		return &ClassicLoadBalancer{Name: fi.PtrTo(nameTag)}, nil
	}
	return nil, nil
}

func (e *DNSName) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *DNSName) CheckChanges(a, e, changes *DNSName) error {
	if a == nil {
		if fi.ValueOf(e.Name) == "" {
			return fi.RequiredField("Name")
		}
		if fi.ValueOf(e.ResourceName) == "" {
			return fi.RequiredField("ResourceName")
		}
		if fi.ValueOf(e.ResourceType) == "" {
			return fi.RequiredField("ResourceType")
		}
		if e.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *DNSName) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *DNSName) error {
	rrs := &route53.ResourceRecordSet{
		Name: e.ResourceName,
		Type: e.ResourceType,
	}

	if e.TargetLoadBalancer != nil {
		rrs.AliasTarget = &route53.AliasTarget{
			DNSName:              e.TargetLoadBalancer.getDNSName(),
			EvaluateTargetHealth: aws.Bool(false),
			HostedZoneId:         e.TargetLoadBalancer.getHostedZoneId(),
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

	klog.V(2).Infof("Updating DNS record %q", *e.ResourceName)

	response, err := t.Cloud.Route53().ChangeResourceRecordSets(request)
	if err != nil {
		return fmt.Errorf("error creating ResourceRecordSets: %v", err)
	}

	klog.V(2).Infof("Change id is %q", aws.StringValue(response.ChangeInfo.Id))

	return nil
}

type terraformRoute53Record struct {
	Name    *string  `cty:"name"`
	Type    *string  `cty:"type"`
	TTL     *string  `cty:"ttl"`
	Records []string `cty:"records"`

	Alias  *terraformAlias          `cty:"alias"`
	ZoneID *terraformWriter.Literal `cty:"zone_id"`
}

type terraformAlias struct {
	Name                 *terraformWriter.Literal `cty:"name"`
	Type                 *terraformWriter.Literal `cty:"type"`
	ZoneID               *terraformWriter.Literal `cty:"zone_id"`
	EvaluateTargetHealth *bool                    `cty:"evaluate_target_health"`
}

func (_ *DNSName) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DNSName) error {
	tf := &terraformRoute53Record{
		Name:   e.ResourceName,
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

func (e *DNSName) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralSelfLink("aws_route53_record", *e.Name)
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
		Name:   e.ResourceName,
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
