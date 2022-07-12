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

package dns

import (
	"context"
	"fmt"
	"testing"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"

	//"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

const (
	validScalewayProfileName = "default"
	validDNSZone             = "leila.sieben.fr"
)

func createValidTestClientReal(t *testing.T) DomainAPI {
	config, err := scw.LoadConfig()
	if err != nil {
		t.Fatalf("could not load config")
	}
	profile, ok := config.Profiles[validScalewayProfileName]
	if !ok || profile == nil {
		t.Fatalf("could not find profile %q", validScalewayProfileName)
	}
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating domainAPI: %v", err)
	}
	return domain.NewAPI(client)
}

func createInvalidTestClientReal(t *testing.T) DomainAPI {
	client, err := scw.NewClient(scw.WithoutAuth())
	if err != nil {
		t.Errorf("error creating domainAPI: %v", err)
	}
	return domain.NewAPI(client)
}

func getDNSProviderZonesReal(api DomainAPI) dnsprovider.Zones {
	dnsProvider := NewProvider(api)
	zs, _ := dnsProvider.Zones()
	return zs
}

func TestZonesListValidReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	z := &zones{domainAPI: domainAPI}

	zoneList, err := z.List()

	if err != nil {
		t.Errorf("error listing zones: %v", err)
	}
	if len(zoneList) < 1 {
		t.Errorf("expected at least 1 zone, got 0")
	}
	zone := zoneList[0]
	if zone.Name() != validDNSZone {
		t.Errorf("expected %s as zone name, got: %s", validDNSZone, zone.Name())
	}
}

func TestZonesListShouldFailReal(t *testing.T) {
	domainAPI := createInvalidTestClientReal(t)
	z := &zones{domainAPI: domainAPI}

	zoneList, err := z.List()

	if err == nil {
		t.Errorf("expected non-nil err")
	}
	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}
}

func TestAddValidReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	inZone := &zone{
		name:      fmt.Sprintf("%s.%s", "kops-dns-test", validDNSZone),
		domainAPI: domainAPI,
	}
	outZone, err := zs.Add(inZone)

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if outZone == nil {
		t.Errorf("zone is nil, exiting test early")
	}
	if outZone.Name() != "kops-dns-test" {
		t.Errorf("unexpected zone name: %s", outZone.Name())
	}
}

func TestAddShouldFailReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	inZone := &zone{
		name:      fmt.Sprintf("%s.%s", "kops-dns-test", validDNSZone),
		domainAPI: domainAPI,
	}
	outZone, err := zs.Add(inZone)

	if outZone != nil {
		t.Errorf("expected zone to be nil, got :%v", outZone)
	}
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestRemoveValidReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	inZone := &zone{
		name:      fmt.Sprintf("%s.%s", "kops-dns-test", validDNSZone),
		domainAPI: domainAPI,
	}
	err := zs.Remove(inZone)

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestRemoveShouldFailReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	inZone := &zone{
		name:      fmt.Sprintf("%s.%s", "kops-dns-test", validDNSZone),
		domainAPI: domainAPI,
	}
	err := zs.Remove(inZone)

	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestNewZoneReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	zone, err := zs.New("kops-dns-test")

	if err != nil {
		t.Errorf("error creating zone: %v", err)
		return
	}
	if zone == nil {
		t.Errorf("zone is nil, exiting test early")
	}
	if zone.Name() != "kops-dns-test" {
		t.Errorf("unexpected zone name: %v", zone.Name())
	}
}

// createRecord creates a record given an associated zone and an UpdateDNSZoneRecordsRequest
func createRecord(api DomainAPI, recordsCreateRequest *domain.UpdateDNSZoneRecordsRequest) ([]string, error) {
	resp, err := api.UpdateDNSZoneRecords(recordsCreateRequest)
	if err != nil {
		return nil, fmt.Errorf("error creating record: %v", err)
	}

	recordsIds := []string(nil)
	for _, record := range resp.Records {
		recordsIds = append(recordsIds, record.ID)
	}

	return recordsIds, nil
}

// deleteRecord deletes a record given an associated zone and a record ID
func deleteRecord(api DomainAPI, zoneName string, recordID string) error {
	recordDeleteRequest := &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: zoneName,
		Changes: []*domain.RecordChange{
			{
				Delete: &domain.RecordChangeDelete{
					ID: &recordID,
				},
			},
		},
	}

	_, err := api.UpdateDNSZoneRecords(recordDeleteRequest)
	if err != nil {
		return fmt.Errorf("error deleting record: %v", err)
	}

	return nil
}

func TestNewResourceRecordSetReal(t *testing.T) {
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	recordsIds, err := createRecord(domainAPI, &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: validDNSZone,
		Changes: []*domain.RecordChange{
			{
				Add: &domain.RecordChangeAdd{
					Records: []*domain.Record{
						{
							Name: "test",
							Data: "127.0.0.1",
							TTL:  3600,
							Type: "A",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Errorf("error creating record: %v", err)
	}

	zone, err := zs.New(validDNSZone)
	if err != nil {
		t.Errorf("error creating zone: %v", err)

	}
	if zone == nil {
		t.Errorf("zone is nil, exiting test early")
	}
	if zone.Name() != validDNSZone {
		t.Errorf("unexpected zone name: %v", zone.Name())
	}

	rrset, _ := zone.ResourceRecordSets()
	rrsets, err := rrset.List()

	if err != nil {
		t.Errorf("error listing resource record sets: %v", err)
	}
	if len(rrsets) < 1 {
		t.Errorf("unexpected number of records: %d", len(rrsets))
	}

	records, err := rrset.Get(fmt.Sprintf("%s.%s", "test", validDNSZone))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("unexpected records from resource record set: %d, expected 1 record", len(records))
	}
	if records[0].Name() != fmt.Sprintf("%s.%s", "test", validDNSZone) {
		t.Errorf("unexpected record name: %s, expected 'test'", records[0].Name())
	}
	if len(records[0].Rrdatas()) != 1 {
		t.Errorf("unexpected number of resource record data: %d", len(records[0].Rrdatas()))
	}
	if records[0].Rrdatas()[0] != "127.0.0.1" {
		t.Errorf("unexpected resource record data: %s", records[0].Rrdatas()[0])
	}
	if records[0].Ttl() != 3600 {
		t.Errorf("unexpected record TTL: %d, expected 3600", records[0].Ttl())
	}
	if records[0].Type() != rrstype.A {
		t.Errorf("unexpected resource record type: %s, expected %s", records[0].Type(), rrstype.A)
	}

	// Cleaning up created zones
	for _, id := range recordsIds {
		err = deleteRecord(domainAPI, validDNSZone, id)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
}

func TestResourceRecordChangesetReal(t *testing.T) {
	ctx := context.Background()
	domainAPI := createValidTestClientReal(t)
	zs := getDNSProviderZonesReal(domainAPI)

	recordsIds, err := createRecord(domainAPI, &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: validDNSZone,
		Changes: []*domain.RecordChange{
			{
				Add: &domain.RecordChangeAdd{
					Records: []*domain.Record{
						{
							Name: "test",
							Data: "127.0.0.1",
							TTL:  3600,
							Type: "A",
						},
						{
							Name: "to-remove",
							Data: "127.0.0.1",
							TTL:  3600,
							Type: "A",
						},
						{
							Name: "to-upsert",
							Data: "127.0.0.1",
							TTL:  3600,
							Type: "A",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Errorf("error creating record: %v", err)
	}

	zone, err := zs.New(validDNSZone)
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}
	if zone == nil {
		t.Errorf("zone is nil, exiting test early")
	}
	if zone.Name() != validDNSZone {
		t.Errorf("unexpected zone name: %v", zone.Name())
	}

	rrsets, _ := zone.ResourceRecordSets()
	changeset := rrsets.StartChangeset()

	if !changeset.IsEmpty() {
		t.Error("expected empty changeset")
	}

	rrset := rrsets.New(fmt.Sprintf("%s.%s", "to-add", validDNSZone), []string{"127.0.0.1"}, 3600, rrstype.A)
	changeset.Add(rrset)

	rrset = rrsets.New(fmt.Sprintf("%s.%s", "to-remove", validDNSZone), []string{"127.0.0.1"}, 3600, rrstype.A)
	changeset.Remove(rrset)

	rrset = rrsets.New(fmt.Sprintf("%s.%s", "to-upsert", validDNSZone), []string{"127.0.0.1"}, 3601, rrstype.A)
	changeset.Upsert(rrset)

	err = changeset.Apply(ctx)
	if err != nil {
		t.Errorf("error applying changeset: %v", err)
	}

	_, err = rrsets.Get(fmt.Sprintf("%s.%s", "test", validDNSZone))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	_, err = rrsets.Get(fmt.Sprintf("%s.%s", "to-add", validDNSZone))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	recordsRemove, _ := rrsets.Get(fmt.Sprintf("%s.%s", "to-remove", validDNSZone))
	if recordsRemove != nil {
		t.Errorf("record set 'to-remove' should have been deleted")
	}
	recordsUpsert, err := rrsets.Get(fmt.Sprintf("%s.%s", "to-upsert", validDNSZone))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	if recordsUpsert[0].Ttl() != 3601 {
		t.Errorf("unexpected record TTL: %d, expected 3601", recordsUpsert[0].Ttl())
	}

	// Cleaning up created zones
	for _, id := range recordsIds {
		err = deleteRecord(domainAPI, validDNSZone, id)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
	addedRecords, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: validDNSZone,
		Name:    "to-add",
	})
	if err != nil {
		t.Fatalf("error getting added record for deletion: %v", err)
	}
	for _, addedRecord := range addedRecords.Records {
		err = deleteRecord(domainAPI, validDNSZone, addedRecord.ID)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
}
