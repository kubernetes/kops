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
	"strings"
	"testing"

	"github.com/google/uuid"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"k8s.io/kops/cloudmock/scaleway/mockdns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

func setUpFakeZones() *mockdns.FakeDomainAPI {
	return &mockdns.FakeDomainAPI{
		DNSZones: []*domain.DNSZone{
			{
				Domain:    "example.com",
				Subdomain: "zone",
			},
			{
				Domain:    "example.com",
				Subdomain: "kops",
			},
			{
				Domain:    "domain.fr",
				Subdomain: "zone",
			},
		},
	}
}

func setUpFakeZonesNil() *mockdns.FakeDomainAPI {
	return &mockdns.FakeDomainAPI{}
}

func getDNSProviderZones(api DomainAPI) dnsprovider.Zones {
	dnsProvider := NewProvider(api)
	zs, _ := dnsProvider.Zones()
	return zs
}

func TestZonesListValid(t *testing.T) {
	domainAPI := setUpFakeZones()
	z := &zones{domainAPI: domainAPI}

	zoneList, err := z.List()

	if err != nil {
		t.Errorf("error listing zones: %v", err)
	}
	if len(zoneList) != 3 {
		t.Errorf("expected at least 1 zone, got 0")
	}
	for i, zone := range zoneList {
		if zone.Name() != domainAPI.DNSZones[i].Domain {
			t.Errorf("expected %s as zone name, got: %s", domainAPI.DNSZones[i].Domain, zone.Name())
		}
	}
}

func TestZonesListShouldFail(t *testing.T) {
	domainAPI := setUpFakeZonesNil()
	z := &zones{domainAPI: domainAPI}

	zoneList, err := z.List()

	if err == nil {
		t.Errorf("expected non-nil err")
	}
	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}
}

func TestAddValid(t *testing.T) {
	domainAPI := setUpFakeZones()
	zs := getDNSProviderZones(domainAPI)

	inZone := &zone{
		name:      "dns.example.com",
		domainAPI: domainAPI,
	}
	outZone, err := zs.Add(inZone)

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if outZone == nil {
		t.Errorf("zone is nil, exiting test early")
	}
	if outZone.Name() != "dns" {
		t.Errorf("unexpected zone name: %s", outZone.Name())
	}
}

func TestAddShouldFail(t *testing.T) {
	domainAPI := setUpFakeZonesNil()
	zs := getDNSProviderZones(domainAPI)

	inZone := &zone{
		name:      "dns.example.com",
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

func TestRemoveValid(t *testing.T) {
	domainAPI := setUpFakeZones()
	zs := getDNSProviderZones(domainAPI)

	inZone := &zone{
		name:      "kops.example.com",
		domainAPI: domainAPI,
	}
	err := zs.Remove(inZone)

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestRemoveShouldFail(t *testing.T) {
	domainAPI := setUpFakeZonesNil()
	zs := getDNSProviderZones(domainAPI)

	inZone := &zone{
		name:      "kops.example.com",
		domainAPI: domainAPI,
	}
	err := zs.Remove(inZone)

	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestNewZone(t *testing.T) {
	domainAPI := setUpFakeZones()
	zs := getDNSProviderZones(domainAPI)

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

func setUpFakeRecords() *mockdns.FakeDomainAPI {
	return &mockdns.FakeDomainAPI{
		Records: map[string]*domain.Record{
			"test": {
				Data: "1.2.3.4",
				Name: "test",
				TTL:  3600,
				Type: "A",
				ID:   uuid.New().String(),
			},
			"to-remove": {
				Data: "5.6.7.8",
				Name: "to-remove",
				TTL:  3600,
				Type: "A",
				ID:   uuid.New().String(),
			},
			"to-upsert": {
				Data: "127.0.0.1",
				Name: "to-upsert",
				TTL:  3600,
				Type: "A",
				ID:   uuid.New().String(),
			},
		},
	}
}

func TestNewResourceRecordSet(t *testing.T) {
	domainAPI := setUpFakeRecords()
	zoneName := "kops.example.com"
	zone := zone{
		domainAPI: domainAPI,
		name:      zoneName,
	}

	rrset, _ := zone.ResourceRecordSets()
	rrsets, err := rrset.List()

	if err != nil {
		t.Errorf("error listing resource record sets: %v", err)
	}
	if len(rrsets) != 3 {
		t.Errorf("unexpected number of records: %d", len(rrsets))
	}

	for _, record := range rrsets {
		recordNameShort := strings.TrimSuffix(record.Name(), "."+zoneName)
		expectedRecord, ok := domainAPI.Records[recordNameShort]
		if !ok {
			t.Errorf("could not find record %s in mock records list", record.Name())
		}

		expectedName := fmt.Sprintf("%s.%s", expectedRecord.Name, zoneName)
		if record.Name() != expectedName {
			t.Errorf("expected %q as record name, got %q", expectedName, record.Name())
		}
		if record.Ttl() != int64(expectedRecord.TTL) {
			t.Errorf("expected %d as record TTL, got %d", expectedRecord.TTL, record.Ttl())
		}
		if record.Type() != rrstype.RrsType(expectedRecord.Type) {
			t.Errorf("expected %q as record type, got %q", expectedRecord.Type, record.Type())
		}
		if len(record.Rrdatas()) < 1 {
			t.Errorf("expected at least 1 rrdata for record %s, got 0", record.Name())
		} else if record.Rrdatas()[0] != expectedRecord.Data {
			t.Errorf("expected %q as record data, got %q", expectedRecord.Data, record.Rrdatas())
		}
	}
}

func TestResourceRecordChangeset(t *testing.T) {
	ctx := context.Background()
	domainAPI := setUpFakeRecords()
	zoneName := "kops.example.com"
	zone := zone{
		domainAPI: domainAPI,
		name:      zoneName,
	}

	rrsets, _ := zone.ResourceRecordSets()
	changeset := rrsets.StartChangeset()

	if !changeset.IsEmpty() {
		t.Error("expected empty changeset")
	}

	rrset := rrsets.New(fmt.Sprintf("%s.%s", "to-add", zoneName), []string{"127.0.0.1"}, 3600, rrstype.A)
	changeset.Add(rrset)

	rrset = rrsets.New(fmt.Sprintf("%s.%s", "to-remove", zoneName), []string{"5.6.7.8"}, 3600, rrstype.A)
	changeset.Remove(rrset)

	rrset = rrsets.New(fmt.Sprintf("%s.%s", "to-upsert", zoneName), []string{"127.0.0.1"}, 3601, rrstype.A)
	changeset.Upsert(rrset)

	err := changeset.Apply(ctx)
	if err != nil {
		t.Errorf("error applying changeset: %v", err)
	}

	_, err = rrsets.Get(fmt.Sprintf("%s.%s", "test", zoneName))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	_, err = rrsets.Get(fmt.Sprintf("%s.%s", "to-add", zoneName))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	recordsRemove, _ := rrsets.Get(fmt.Sprintf("%s.%s", "to-remove", zoneName))
	if recordsRemove != nil {
		t.Errorf("record set 'to-remove' should have been deleted")
	}
	recordsUpsert, err := rrsets.Get(fmt.Sprintf("%s.%s", "to-upsert", zoneName))
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	if recordsUpsert[0].Ttl() != 3601 {
		t.Errorf("unexpected record TTL: %d, expected 3601", recordsUpsert[0].Ttl())
	}
}
