package dns

import (
	"context"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"testing"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestZonesList(t *testing.T) {
	// happy path
	config, _ := scw.LoadConfig()
	profile := config.Profiles["devterraform"]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	z := &zones{client: client}
	zoneList, err := z.List()
	if err != nil {
		t.Errorf("error listing zones: %v", err)
	}
	if len(zoneList) != 1 {
		t.Errorf("expected only 1 zone, got %d", len(zoneList))
	}
	zone := zoneList[0]
	if zone.Name() != "scaleway-terraform.com" {
		t.Errorf("expected example.com as zone name, got: %s", zone.Name())
	}

	// bad response path
	client, err = scw.NewClient(scw.WithoutAuth())
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	z = &zones{client: client}
	zoneList, err = z.List()
	if err == nil {
		t.Errorf("expected non-nil err")
	}
	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}
}

func TestAdd(t *testing.T) {
	config, _ := scw.LoadConfig()
	profile := config.Profiles["devterraform"]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	// happy path
	dnsProvider := NewProvider(client, "scaleway-terraform.com")
	zs, _ := dnsProvider.Zones()
	inZone := &zone{name: "kops-dns-test", client: client}

	outZone, err := zs.Add(inZone)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if outZone.Name() != "kops-dns-test" {
		t.Errorf("unexpected zone name: %s", outZone.Name())
	}

	// bad status code
	dnsProvider = NewProvider(client, "api.k8s.fr-par.scw.cloud")
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "kops-dns-test", client: client}

	outZone, err = zs.Add(inZone)
	if outZone != nil {
		t.Errorf("expected zone to be nil, got :%v", outZone)
	}
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestRemove(t *testing.T) {
	config, _ := scw.LoadConfig()
	profile := config.Profiles["devterraform"]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	// happy path
	dnsProvider := NewProvider(client, "scaleway-terraform.com")
	zs, _ := dnsProvider.Zones()
	inZone := &zone{name: "kops-dns-test", client: client}

	err = zs.Remove(inZone)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	// bad status code
	dnsProvider = NewProvider(client, "api.k8s.fr-par.scw.cloud")
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "kops-dns-test", client: client}

	err = zs.Remove(inZone)
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestNewZone(t *testing.T) {
	config, _ := scw.LoadConfig()
	profile := config.Profiles["devterraform"]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	dnsProvider := NewProvider(client, "scaleway-terraform.com")
	zs, _ := dnsProvider.Zones()

	zone, err := zs.New("kops-dns-test")
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}
	if zone.Name() != "kops-dns-test" {
		t.Errorf("unexpected zone name: %v", zone.Name())
	}
}

func TestNewResourceRecordSet(t *testing.T) {
	config, _ := scw.LoadConfig()
	profile := config.Profiles["devterraform"]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	recordsIds, err := createRecord(client, &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: "scaleway-terraform.com",
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

	dnsProvider := NewProvider(client, "scaleway-terraform.com")
	zs, _ := dnsProvider.Zones()

	zone, err := zs.New("scaleway-terraform.com")
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}
	if zone.Name() != "scaleway-terraform.com" {
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

	records, err := rrset.Get("test.scaleway-terraform.com")
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("unexpected records from resource record set: %d, expected 1 record", len(records))
	}
	if records[0].Name() != "test.scaleway-terraform.com" {
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

	for _, id := range recordsIds {
		err = deleteRecord(client, "scaleway-terraform.com", id)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
}

func TestResourceRecordChangeset(t *testing.T) {
	ctx := context.Background()

	config, _ := scw.LoadConfig()
	profile := config.Profiles["devterraform"]
	client, err := scw.NewClient(scw.WithProfile(profile))
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	recordsIds, err := createRecord(client, &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: "scaleway-terraform.com",
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

	dnsProvider := NewProvider(client, "scaleway-terraform.com")
	zs, _ := dnsProvider.Zones()

	zone, err := zs.New("scaleway-terraform.com")
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}
	if zone.Name() != "scaleway-terraform.com" {
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

	records, err := rrset.Get("test.scaleway-terraform.com")
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	records, err = rrset.Get("to-add.scaleway-terraform.com")
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	records, err = rrset.Get("to-upsert.scaleway-terraform.com")
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}
	if records[0].Ttl() != 3601 {
		t.Errorf("unexpected record TTL: %d, expected 3601", records[0].Ttl())
	}
	records, err = rrset.Get("to-remove.scaleway-terraform.com")
	if records != nil {
		t.Errorf("record set 'to-remove' should have been deleted")
	}

	for _, id := range recordsIds {
		err = deleteRecord(client, "scaleway-terraform.com", id)
		if err != nil {
			t.Errorf("error deleting record: %v", err)
		}
	}
}
