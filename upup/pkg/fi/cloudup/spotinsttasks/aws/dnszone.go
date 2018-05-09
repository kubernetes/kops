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

package aws

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

// DNSZone is a zone object in a dns provider
//go:generate fitask -type=DNSZone
type DNSZone struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	DNSName *string
	ZoneID  *string

	Private    *bool
	PrivateVPC *VPC
}

var _ fi.CompareWithID = &DNSZone{}

func (e *DNSZone) CompareWithID() *string {
	return e.Name
}

func (e *DNSZone) Find(c *fi.Context) (*DNSZone, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	z, err := e.findExisting(cloud)
	if err != nil {
		return nil, err
	}

	if z == nil {
		return nil, nil
	}

	actual := &DNSZone{}
	actual.Name = e.Name
	if z.HostedZone.Name != nil {
		actual.DNSName = fi.String(strings.TrimSuffix(*z.HostedZone.Name, "."))
	}
	if z.HostedZone.Id != nil {
		actual.ZoneID = fi.String(strings.TrimPrefix(*z.HostedZone.Id, "/hostedzone/"))
	}
	actual.Private = z.HostedZone.Config.PrivateZone

	// If the zone is private, but we don't want it to be, that will be an error
	// e.PrivateVPC won't be set, so we can't find the "right" VPC (without cheating)
	if e.PrivateVPC != nil {
		for _, vpc := range z.VPCs {
			if cloud.Region() != aws.StringValue(vpc.VPCRegion) {
				continue
			}

			if aws.StringValue(e.PrivateVPC.ID) == aws.StringValue(vpc.VPCId) {
				actual.PrivateVPC = e.PrivateVPC
			}
		}
	}

	if e.ZoneID == nil {
		e.ZoneID = actual.ZoneID
	}
	if e.DNSName == nil {
		e.DNSName = actual.DNSName
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *DNSZone) findExisting(cloud awsup.AWSCloud) (*route53.GetHostedZoneOutput, error) {
	findID := ""
	if e.ZoneID != nil {
		request := &route53.GetHostedZoneInput{
			Id: e.ZoneID,
		}

		response, err := cloud.Route53().GetHostedZone(request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "NoSuchHostedZone" {
				return nil, nil
			} else {
				return nil, fmt.Errorf("error fetching DNS HostedZone %q: %v", findID, err)
			}
		} else {
			return response, nil
		}
	}

	findName := fi.StringValue(e.DNSName)
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
		if aws.StringValue(zone.Name) == findName && fi.BoolValue(zone.Config.PrivateZone) == fi.BoolValue(e.Private) {
			zones = append(zones, zone)
		}
	}

	if len(zones) == 0 {
		return nil, nil
	} else if len(zones) != 1 {
		return nil, fmt.Errorf("found multiple hosted zones matched name %q", findName)
	} else {
		request := &route53.GetHostedZoneInput{
			Id: zones[0].Id,
		}

		response, err := cloud.Route53().GetHostedZone(request)
		if err != nil {
			return nil, fmt.Errorf("error fetching DNS HostedZone by id %q: %v", *request.Id, err)
		}

		return response, nil
	}
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

func (_ *DNSZone) Render(t *spotinst.Target, a, e, changes *DNSZone) error {
	name := aws.StringValue(e.DNSName)
	if a == nil {
		request := &route53.CreateHostedZoneInput{}
		request.Name = e.DNSName
		nonce := rand.Int63()
		request.CallerReference = aws.String(strconv.FormatInt(nonce, 10))

		if e.PrivateVPC != nil {
			request.VPC = &route53.VPC{
				VPCId:     e.PrivateVPC.ID,
				VPCRegion: aws.String(t.Cloud.(spotinst.Cloud).Region()),
			}
		}

		glog.V(2).Infof("Creating Route53 HostedZone with Name %q", name)

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).Route53().CreateHostedZone(request)
		if err != nil {
			return fmt.Errorf("error creating DNS HostedZone %q: %v", name, err)
		}

		e.ZoneID = response.HostedZone.Id
	} else {
		if changes.PrivateVPC != nil {
			request := &route53.AssociateVPCWithHostedZoneInput{
				HostedZoneId: a.ZoneID,
				VPC: &route53.VPC{
					VPCId:     e.PrivateVPC.ID,
					VPCRegion: aws.String(t.Cloud.(spotinst.Cloud).Region()),
				},
			}

			changes.PrivateVPC = nil

			glog.V(2).Infof("Updating DNSZone %q", name)

			_, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).Route53().AssociateVPCWithHostedZone(request)
			if err != nil {
				return fmt.Errorf("error associating VPC with hosted zone %q: %v", name, err)
			}
		}

		empty := &DNSZone{}
		if !reflect.DeepEqual(empty, changes) {
			glog.Warningf("cannot apply changes to DNSZone %q: %v", name, changes)
		}
	}

	// We don't tag the zone - we expect it to be shared
	return nil
}
