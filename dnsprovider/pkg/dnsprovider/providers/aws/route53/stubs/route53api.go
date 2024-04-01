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

/* internal implements a stub for the AWS Route53 API, used primarily for unit testing purposes */
package stubs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// Compile time check for interface conformance
var _ Route53API = &Route53APIStub{}

/* Route53API is the subset of the AWS Route53 API that we actually use.  Add methods as required. Signatures must match exactly. */
type Route53API interface {
	ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
	ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error)
	CreateHostedZone(ctx context.Context, params *route53.CreateHostedZoneInput, optFns ...func(*route53.Options)) (*route53.CreateHostedZoneOutput, error)
	DeleteHostedZone(ctx context.Context, params *route53.DeleteHostedZoneInput, optFns ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error)
}

// Route53APIStub is a minimal implementation of Route53API, used primarily for unit testing.
// See https://docs.aws.amazon.com/sdk-for-go/api/service/route53/
// of all of its methods.
type Route53APIStub struct {
	zones      map[string]route53types.HostedZone
	recordSets map[string]map[string][]route53types.ResourceRecordSet
}

// NewRoute53APIStub returns an initialized Route53APIStub
func NewRoute53APIStub() *Route53APIStub {
	return &Route53APIStub{
		zones:      make(map[string]route53types.HostedZone),
		recordSets: make(map[string]map[string][]route53types.ResourceRecordSet),
	}
}

func (r *Route53APIStub) ListResourceRecordSets(ctx context.Context, input *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	output := &route53.ListResourceRecordSetsOutput{} // TODO: Support optional input args.
	if len(r.recordSets) <= 0 {
		output.ResourceRecordSets = []route53types.ResourceRecordSet{}
	} else if _, ok := r.recordSets[*input.HostedZoneId]; !ok {
		output.ResourceRecordSets = []route53types.ResourceRecordSet{}
	} else {
		for _, rrsets := range r.recordSets[*input.HostedZoneId] {
			output.ResourceRecordSets = append(output.ResourceRecordSets, rrsets...)
		}
	}
	return output, nil
}

func (r *Route53APIStub) ChangeResourceRecordSets(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	output := &route53.ChangeResourceRecordSetsOutput{}
	recordSets, ok := r.recordSets[*input.HostedZoneId]
	if !ok {
		recordSets = make(map[string][]route53types.ResourceRecordSet)
	}

	for _, change := range input.ChangeBatch.Changes {
		key := *change.ResourceRecordSet.Name + "::" + string(change.ResourceRecordSet.Type)
		switch change.Action {
		case route53types.ChangeActionCreate:
			if _, found := recordSets[key]; found {
				return nil, fmt.Errorf("attempt to create duplicate rrset %s", key) // TODO: Return AWS errors with codes etc
			}
			recordSets[key] = append(recordSets[key], *change.ResourceRecordSet)
		case route53types.ChangeActionDelete:
			if _, found := recordSets[key]; !found {
				return nil, fmt.Errorf("attempt to delete non-existent rrset %s", key) // TODO: Check other fields too
			}
			delete(recordSets, key)
		case route53types.ChangeActionUpsert:
			// TODO - not used yet
		}
	}
	r.recordSets[*input.HostedZoneId] = recordSets
	return output, nil // TODO: We should ideally return status etc, but we don't' use that yet.
}

func (r *Route53APIStub) ListHostedZones(ctx context.Context, input *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
	output := &route53.ListHostedZonesOutput{}
	for _, zone := range r.zones {
		output.HostedZones = append(output.HostedZones, zone)
	}
	return output, nil
}

func (r *Route53APIStub) CreateHostedZone(ctx context.Context, input *route53.CreateHostedZoneInput, optFns ...func(*route53.Options)) (*route53.CreateHostedZoneOutput, error) {
	name := aws.ToString(input.Name)
	id := "/hostedzone/" + name
	if _, ok := r.zones[id]; ok {
		return nil, fmt.Errorf("error creating hosted DNS zone: %s already exists", id)
	}
	r.zones[id] = route53types.HostedZone{
		Id:   aws.String(id),
		Name: aws.String(name),
	}
	z := r.zones[id]
	return &route53.CreateHostedZoneOutput{HostedZone: &z}, nil
}

func (r *Route53APIStub) DeleteHostedZone(ctx context.Context, input *route53.DeleteHostedZoneInput, optFns ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
	if _, ok := r.zones[*input.Id]; !ok {
		return nil, fmt.Errorf("error deleting hosted DNS zone: %s does not exist", *input.Id)
	}
	if len(r.recordSets[*input.Id]) > 0 {
		return nil, fmt.Errorf("error deleting hosted DNS zone: %s has resource records", *input.Id)
	}
	delete(r.zones, *input.Id)
	return &route53.DeleteHostedZoneOutput{}, nil
}
