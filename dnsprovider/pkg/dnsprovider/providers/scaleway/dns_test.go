package dns

import (
	"context"
	"os"
	"testing"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

const (
	validScalewayProfileName = "normal"
	validDNSZone             = "leila.sieben.fr"
)

func createValidTestClient(t *testing.T) *scw.Client {
	err := os.Setenv("SCW_DNS_ZONE", validDNSZone)
	if err != nil {
		t.Errorf("error setting DNS_ZONE in environment: %v", err)
	}
	config, _ := scw.LoadConfig()
	profile := config.Profiles[validScalewayProfileName]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	return client
}

func createInvalidTestClient(t *testing.T) *scw.Client {
	client, err := scw.NewClient(scw.WithoutAuth())
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	return client
}

func getDNSProviderZones(client *scw.Client) dnsprovider.Zones {
	dnsProvider := NewProvider(client)
	zs, _ := dnsProvider.Zones()
	return zs
}

func TestZonesListValid(t *testing.T) {
	client := createValidTestClient(t)
	z := &zones{client: client}

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

func TestZonesListShouldFail(t *testing.T) {
	client := createInvalidTestClient(t)
	z := &zones{client: client}

	zoneList, err := z.List()

	if err == nil {
		t.Errorf("expected non-nil err")
	}
	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}
}

func TestAddValid(t *testing.T) {
	client := createValidTestClient(t)
	zs := getDNSProviderZones(client)

	inZone := &zone{name: "kops-dns-test", client: client}
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

func TestAddShouldFail(t *testing.T) {
	client := createValidTestClient(t)
	err := os.Setenv("SCW_DNS_ZONE", "invalid.domain")
	zs := getDNSProviderZones(client)

	inZone := &zone{name: "kops-dns-test", client: client}
	outZone, err := zs.Add(inZone)

	if outZone != nil {
		t.Errorf("expected zone to be nil, got :%v", outZone)
	}
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestRemoveValid(t *testing.T) {
	client := createValidTestClient(t)
	zs := getDNSProviderZones(client)

	inZone := &zone{name: "kops-dns-test", client: client}
	err := zs.Remove(inZone)

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestRemoveShouldFail(t *testing.T) {
	client := createValidTestClient(t)
	err := os.Setenv("SCW_DNS_ZONE", "invalid.domain")
	zs := getDNSProviderZones(client)

	inZone := &zone{name: "kops-dns-test", client: client}
	err = zs.Remove(inZone)

	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestNewZone(t *testing.T) {
	client := createValidTestClient(t)
	zs := getDNSProviderZones(client)

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

func TestNewResourceRecordSet(t *testing.T) {
	client := createValidTestClient(t)
	zs := getDNSProviderZones(client)

	recordsIds, err := createRecord(client, &domain.UpdateDNSZoneRecordsRequest{
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

	records, err := rrset.Get("test." + validDNSZone)
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("unexpected records from resource record set: %d, expected 1 record", len(records))
	}
	if records[0].Name() != "test."+validDNSZone {
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
		err = deleteRecord(client, validDNSZone, id)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
}

func TestResourceRecordChangeset(t *testing.T) {
	ctx := context.Background()
	client := createValidTestClient(t)
	zs := getDNSProviderZones(client)

	recordsIds, err := createRecord(client, &domain.UpdateDNSZoneRecordsRequest{
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

	rrset, _ := zone.ResourceRecordSets()
	changeset := rrset.StartChangeset()

	if !changeset.IsEmpty() {
		t.Error("expected empty changeset")
	}

	record := rrset.New("to-add", []string{"127.0.0.1"}, 3600, rrstype.A)
	changeset.Add(record)

	record = rrset.New("to-remove", []string{"127.0.0.1"}, 3600, rrstype.A)
	changeset.Remove(record)

	record = rrset.New("to-upsert", []string{"127.0.0.1"}, 3601, rrstype.A)
	changeset.Upsert(record)

	err = changeset.Apply(ctx)
	if err != nil {
		t.Errorf("error applying changeset: %v", err)
	}

	records, err := rrset.Get("test." + validDNSZone)
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	records, err = rrset.Get("to-upsert." + validDNSZone)
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	if records[0].Ttl() != 3601 {
		t.Errorf("unexpected record TTL: %d, expected 3601", records[0].Ttl())
	}
	records, err = rrset.Get("to-remove." + validDNSZone)
	if records != nil {
		t.Errorf("record set 'to-remove' should have been deleted")
	}
	records, err = rrset.Get("to-add." + validDNSZone)
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}

	// Cleaning up created zones
	api := domain.NewAPI(client)
	addedRecords, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: validDNSZone,
		Name:    records[0].Name(),
	})
	for _, addedRecord := range addedRecords.Records {
		err = deleteRecord(client, validDNSZone, addedRecord.ID)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
	for _, id := range recordsIds {
		err = deleteRecord(client, validDNSZone, id)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}

}
