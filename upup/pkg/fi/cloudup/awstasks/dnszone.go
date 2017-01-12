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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"math/rand"
	"strconv"
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
	cloud := c.Cloud.(awsup.AWSCloud)

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

func (e *DNSZone) findExisting(cloud awsup.AWSCloud) (*route53.HostedZone, error) {
	findID := ""
	if e.ID != nil {
		findID = *e.ID
	} else if e.Name != nil && !strings.Contains(*e.Name, ".") {
		// Looks like a hosted zone ID
		findID = *e.Name
	}
	if findID != "" {
		request := &route53.GetHostedZoneInput{
			Id: aws.String(findID),
		}

		response, err := cloud.Route53().GetHostedZone(request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "NoSuchHostedZone" {
				if e.ID != nil {
					return nil, nil
				}
				// Otherwise continue ... maybe the name was not an id after all...
			} else {
				return nil, fmt.Errorf("error fetching DNS HostedZone %q: %v", findID, err)
			}
		} else {
			return response.HostedZone, nil
		}
	}

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

	response, err := cloud.Route53().ListHostedZonesByName(request)
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
		nonce := rand.Int63()
		request.CallerReference = aws.String(strconv.FormatInt(nonce, 10))

		glog.V(2).Infof("Creating Route53 HostedZone with Name %q", e.Name)

		response, err := t.Cloud.Route53().CreateHostedZone(request)
		if err != nil {
			return fmt.Errorf("error creating DNS HostedZone: %v", err)
		}

		e.ID = response.HostedZone.Id
	}

	// We don't tag the zone - we expect it to be shared
	return nil
}

type terraformRoute53Zone struct {
	Name      *string              `json:"name"`
	Tags      map[string]string    `json:"tags,omitempty"`
	Lifecycle *terraform.Lifecycle `json:"lifecycle,omitempty"`
}

func (_ *DNSZone) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DNSZone) error {
	cloud := t.Cloud.(awsup.AWSCloud)

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
	}

	if z == nil {
		// Because we expect most users to create their zones externally,
		// we now block hostedzone creation in terraform.
		// This lets us perform deeper DNS validation, but also solves the problem
		// that otherwise we don't know if TF created the hosted zone
		// (in which case we should output it) or whether it already existed (in which case we should not)
		// The root problem here is that TF doesn't have a strong notion of an unmanaged resource
		return fmt.Errorf("Creation of Route53 hosted zones is not supported for terraform")
		//tf := &terraformRoute53Zone{
		//	Name: e.Name,
		//	//Tags:               cloud.BuildTags(e.Name, nil),
		//}
		//
		//tf.Lifecycle = &terraform.Lifecycle{
		//	PreventDestroy: fi.Bool(true),
		//}
		//
		//return t.RenderResource("aws_route53_zone", *e.Name, tf)
	} else {
		return nil
	}
}

func (e *DNSZone) TerraformLink() *terraform.Literal {
	if e.ID != nil {
		glog.V(4).Infof("reusing existing route53 zone with id %q", *e.ID)
		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralSelfLink("aws_route53_zone", *e.Name)
}
