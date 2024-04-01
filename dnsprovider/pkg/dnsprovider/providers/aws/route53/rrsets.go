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

package route53

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordSets = ResourceRecordSets{}

type ResourceRecordSets struct {
	zone *Zone
}

func (rrsets ResourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: rrsets.zone.impl.Id,
	}

	var list []dnsprovider.ResourceRecordSet
	paginator := route53.NewListResourceRecordSetsPaginator(rrsets.zone.zones.interface_.service, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("error listing resource record sets: %w", err)
		}
		for _, rrset := range page.ResourceRecordSets {
			list = append(list, &ResourceRecordSet{&rrset, &rrsets})
		}
	}
	return list, nil
}

func (rrsets ResourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	// This list implementation is very similar to the one implemented in
	// the List() method above, but it restricts the retrieved list to
	// the records whose name match the given `name`.
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    rrsets.zone.impl.Id,
		StartRecordName: aws.String(name),
	}

	var list []dnsprovider.ResourceRecordSet
	paginator := route53.NewListResourceRecordSetsPaginator(rrsets.zone.zones.interface_.service, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("error listing resource record sets: %w", err)
		}
		for _, rrset := range page.ResourceRecordSets {
			if aws.ToString(rrset.Name) != name {
				continue
			}
			list = append(list, &ResourceRecordSet{&rrset, &rrsets})
		}
	}
	return list, nil
}

func (r ResourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &ResourceRecordChangeset{
		zone:   r.zone,
		rrsets: &r,
	}
}

func (r ResourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	rrstypeStr := string(rrstype)
	rrs := &route53types.ResourceRecordSet{
		Name: &name,
		Type: route53types.RRType(rrstypeStr),
		TTL:  &ttl,
	}
	for _, rrdata := range rrdatas {
		rrs.ResourceRecords = append(rrs.ResourceRecords, route53types.ResourceRecord{
			Value: aws.String(rrdata),
		})
	}

	return ResourceRecordSet{
		rrs,
		&r,
	}
}

// Zone returns the parent zone
func (rrset ResourceRecordSets) Zone() dnsprovider.Zone {
	return rrset.zone
}
