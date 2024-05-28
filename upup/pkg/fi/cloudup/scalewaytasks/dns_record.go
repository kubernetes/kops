/*
Copyright 2023 The Kubernetes Authors.

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

package scalewaytasks

import (
	"fmt"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

const PlaceholderIP = "203.0.113.123"

// +kops:fitask
type DNSRecord struct {
	ID          *string
	Name        *string
	Data        *string
	DNSZone     *string
	Type        *string
	TTL         *uint32
	IsInternal  *bool
	ClusterName *string
	Lifecycle   fi.Lifecycle
}

var _ fi.CloudupTask = &DNSRecord{}
var _ fi.CompareWithID = &DNSRecord{}

func (d *DNSRecord) CompareWithID() *string {
	return d.ID
}

var _ fi.CloudupHasDependencies = &Instance{}

func (d *DNSRecord) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Instance); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*PrivateNetwork); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (d *DNSRecord) Find(context *fi.CloudupContext) (*DNSRecord, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	records, err := cloud.DomainService().ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		ID:      d.ID,
		DNSZone: fi.ValueOf(d.DNSZone),
		Name:    fi.ValueOf(d.Name),
		Type:    domain.RecordType(fi.ValueOf(d.Type)),
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing DNS records named %q in zone %q: %w", fi.ValueOf(d.Name), fi.ValueOf(d.DNSZone), err)
	}

	if records.TotalCount == 0 {
		return nil, nil
	}
	if records.TotalCount > 1 && d.ID != nil {
		return nil, fmt.Errorf("expected exactly 1 DNS record with ID %s, got %d", *d.ID, records.TotalCount)
	}
	recordFound := records.Records[0]

	return &DNSRecord{
		ID:          fi.PtrTo(recordFound.ID),
		Name:        fi.PtrTo(recordFound.Name),
		Data:        fi.PtrTo(recordFound.Data),
		TTL:         fi.PtrTo(recordFound.TTL),
		DNSZone:     d.DNSZone,
		Type:        fi.PtrTo(recordFound.Type.String()),
		IsInternal:  d.IsInternal,
		ClusterName: d.ClusterName,
		Lifecycle:   d.Lifecycle,
	}, nil
}

func (d *DNSRecord) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(d, context)
}

func (_ *DNSRecord) CheckChanges(actual, expected, changes *DNSRecord) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.DNSZone != nil {
			return fi.CannotChangeField("DNSZone")
		}
		if changes.Type != nil {
			return fi.CannotChangeField("Type")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.DNSZone == nil {
			return fi.RequiredField("DNSZone")
		}
		if expected.Type == nil {
			return fi.RequiredField("Type")
		}
		if expected.Data == nil {
			return fi.RequiredField("Data")
		}
		if expected.TTL == nil {
			return fi.RequiredField("TTL")
		}
	}
	return nil
}

func (d *DNSRecord) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *DNSRecord) error {
	cloud := t.Cloud.(scaleway.ScwCloud)

	if *expected.Data == PlaceholderIP {
		controlPlanesIPs, err := scaleway.GetControlPlanesIPs(cloud, *expected.ClusterName, *expected.IsInternal)
		if err != nil || len(controlPlanesIPs) == 0 {
			return fmt.Errorf("error getting control plane IPs: %v", err)
		}
		expected.Data = &controlPlanesIPs[0]
	}

	if actual != nil {
		recordUpdated, err := cloud.DomainService().UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
			DNSZone: fi.ValueOf(actual.DNSZone),
			Changes: []*domain.RecordChange{
				{
					Set: &domain.RecordChangeSet{
						ID: actual.ID,
						Records: []*domain.Record{
							{
								Data: fi.ValueOf(expected.Data),
								TTL:  fi.ValueOf(expected.TTL),
							},
						},
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("updating DNS record %q (%s): %w", fi.ValueOf(actual.Name), fi.ValueOf(actual.ID), err)
		}
		expected.ID = &recordUpdated.Records[0].ID
		return nil
	}

	recordCreated, err := cloud.DomainService().UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone: fi.ValueOf(expected.DNSZone),
		Changes: []*domain.RecordChange{
			{
				Add: &domain.RecordChangeAdd{
					Records: []*domain.Record{
						{
							Data: fi.ValueOf(expected.Data),
							Name: fi.ValueOf(expected.Name),
							TTL:  fi.ValueOf(expected.TTL),
							Type: domain.RecordType(fi.ValueOf(expected.Type)),
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("creating DNS record %q in zone %q: %w", fi.ValueOf(expected.Name), fi.ValueOf(expected.DNSZone), err)
	}

	expected.ID = &recordCreated.Records[0].ID

	return nil
}

type terraformDNSRecord struct {
	Name      *string              `cty:"name"`
	Data      *string              `cty:"data"`
	DNSZone   *string              `cty:"dns_zone"`
	Type      *string              `cty:"type"`
	TTL       *int32               `cty:"ttl"`
	Lifecycle *terraform.Lifecycle `cty:"lifecycle"`
}

func (_ *DNSRecord) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *DNSRecord) error {
	tf := terraformDNSRecord{
		Name:    expected.Name,
		Data:    expected.Data,
		DNSZone: expected.DNSZone,
		Type:    expected.Type,
		TTL:     fi.PtrTo(int32(fi.ValueOf(expected.TTL))),
		Lifecycle: &terraform.Lifecycle{
			IgnoreChanges: []*terraformWriter.Literal{{String: "data"}},
		},
	}
	return t.RenderResource("scaleway_domain_record", fi.ValueOf(expected.Name), tf)
}
