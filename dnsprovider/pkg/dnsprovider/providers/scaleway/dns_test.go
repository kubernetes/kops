package dns

import (
	"context"
	"errors"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"testing"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type fakeDomainService struct {
	listFunc                 func(ctx context.Context, listOpt *scw.RequestOption) ([]domain.DomainSummary, error)
	getFunc                  func(ctx context.Context, name string) (*domain.DomainSummary, error)
	createFunc               func(ctx context.Context, domainCreateRequest *domain.CreateDNSZoneRequest) (*domain.DomainSummary, error)
	deleteFunc               func(ctx context.Context, name string) error
	recordsFunc              func(ctx context.Context, domain string, listOpts *scw.RequestOption) ([]domain.Record, error)
	recordsByNameFunc        func(ctx context.Context, domain string, name string, listOpts *scw.RequestOption) ([]domain.Record, error)
	recordsByTypeFunc        func(ctx context.Context, domain string, name string, listOpts *scw.RequestOption) ([]domain.Record, error)
	recordsByTypeAndNameFunc func(ctx context.Context, domain string, ofType string, name string, listOpts *scw.RequestOption) ([]domain.Record, error)

	recordFunc       func(ctx context.Context, domain string, id int) (*domain.Record, error)
	deleteRecordFunc func(ctx context.Context, domain string, id int) error
	editRecordFunc   func(ctx context.Context, domain string, id int, editRequest *domain.CreateDNSZoneRequest) (*domain.Record, error)
	createRecordFunc func(ctx context.Context, domain string, createRequest *domain.CreateDNSZoneRequest) (*domain.Record, error)
}

func (f *fakeDomainService) List(ctx context.Context, listOpt *scw.RequestOption) ([]domain.DomainSummary, error) {
	return f.listFunc(ctx, listOpt)
}

func (f *fakeDomainService) Get(ctx context.Context, name string) (*domain.DomainSummary, error) {
	return f.getFunc(ctx, name)
}

func (f *fakeDomainService) Create(ctx context.Context, domainCreateRequest *domain.CreateDNSZoneRequest) (*domain.DomainSummary, error) {
	return f.createFunc(ctx, domainCreateRequest)
}

func (f *fakeDomainService) Delete(ctx context.Context, name string) error {
	return f.deleteFunc(ctx, name)
}

func (f *fakeDomainService) Records(ctx context.Context, domain string, listOpts *scw.RequestOption) ([]domain.Record, error) {
	return f.recordsFunc(ctx, domain, listOpts)
}

func (f *fakeDomainService) Record(ctx context.Context, domain string, id int) (*domain.Record, error) {
	return f.recordFunc(ctx, domain, id)
}

func (f *fakeDomainService) DeleteRecord(ctx context.Context, domain string, id int) error {
	return f.deleteRecordFunc(ctx, domain, id)
}

func (f *fakeDomainService) EditRecord(ctx context.Context, domain string,
	id int, editRequest *domain.CreateDNSZoneRequest) (*domain.Record, error) {
	return f.editRecordFunc(ctx, domain, id, editRequest)
}

func (f *fakeDomainService) CreateRecord(ctx context.Context, domain string,
	createRequest *domain.CreateDNSZoneRequest) (*domain.Record, error) {
	return f.createRecordFunc(ctx, domain, createRequest)
}

func (f *fakeDomainService) RecordsByName(ctx context.Context, domain string, name string, listOpts *scw.RequestOption) ([]domain.Record, error) {
	return f.recordsByNameFunc(ctx, domain, name, listOpts)
}

func (f *fakeDomainService) RecordsByType(ctx context.Context, domain string, oftype string, listOpts *scw.RequestOption) ([]domain.Record, error) {
	return f.recordsByTypeFunc(ctx, domain, oftype, listOpts)
}

func (f *fakeDomainService) RecordsByTypeAndName(ctx context.Context, domain string, ofType string, name string, listOpts *scw.RequestOption) ([]domain.Record, error) {
	return f.recordsByTypeAndNameFunc(ctx, domain, ofType, name, listOpts)
}

func TestZonesList(t *testing.T) {
	client, err := scw.NewClient(nil)
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	fake := &fakeDomainService{}

	// happy path
	fake.listFunc = func(ctx context.Context, listOpts *scw.RequestOption) ([]domain.DomainSummary, error) {
		domains := []domain.DomainSummary{
			{
				Domain: "example.com",
			},
		}

		return domains, nil
	}

	z := &zones{client}
	zoneList, err := z.List()
	if err != nil {
		t.Errorf("error listing zones: %v", err)
	}

	if len(zoneList) != 1 {
		t.Errorf("expected only 1 zone, got %d", len(zoneList))
	}

	zone := zoneList[0]
	if zone.Name() != "example.com" {
		t.Errorf("expected example.com as zone name, got: %s", zone.Name())
	}

	// bad response path
	fake.listFunc = func(ctx context.Context, listOpts *scw.RequestOption) ([]domain.DomainSummary, error) {
		domains := []domain.DomainSummary{
			{
				Domain: "example.com",
			},
		}

		return domains, errors.New("internal error!")
	}

	z = &zones{client}
	zoneList, err = z.List()
	if err == nil {
		t.Errorf("expected non-nil err")
	}

	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}

	// scw client returned error path
	fake.listFunc = func(ctx context.Context, listOpts *scw.RequestOption) ([]domain.DomainSummary, error) {
		return nil, errors.New("error!")
	}

	z = &zones{client}
	zoneList, err = z.List()
	if err == nil {
		t.Errorf("expected non-nil err")
	}

	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}
}

func TestAdd(t *testing.T) {
	client, err := scw.NewClient(nil)
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	fake := &fakeDomainService{}

	// happy path
	fake.createFunc = func(ctx context.Context, domainCreateRequest *domain.CreateDNSZoneRequest) (*domain.DomainSummary, error) {
		d := &domain.DomainSummary{Domain: domainCreateRequest.Domain}

		return d, nil
	}

	dnsProvider := NewProvider(client)
	zs, _ := dnsProvider.Zones()
	inZone := &zone{name: "test", client: client}

	outZone, err := zs.Add(inZone)

	if outZone.Name() != "test" {
		t.Errorf("unexpected zone name: %s", outZone.Name())
	}

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	// bad status code
	fake.createFunc = func(ctx context.Context, domainCreateRequest *domain.CreateDNSZoneRequest) (*domain.DomainSummary, error) {
		d := &domain.DomainSummary{Domain: domainCreateRequest.Domain}

		return d, errors.New("bad response!")
	}

	dnsProvider = NewProvider(client)
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "test", client: client}

	outZone, err = zs.Add(inZone)

	if outZone != nil {
		t.Errorf("expected zone to be nil, got :%v", outZone)
	}

	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}

	// scw returns error
	fake.createFunc = func(ctx context.Context, domainCreateRequest *domain.CreateDNSZoneRequest) (*domain.DomainSummary, error) {
		d := &domain.DomainSummary{Domain: domainCreateRequest.Domain}

		return d, errors.New("error!")
	}

	dnsProvider = NewProvider(client)
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "test", client: client}

	outZone, err = zs.Add(inZone)

	if outZone != nil {
		t.Errorf("expected zone to be nil, got :%v", outZone)
	}

	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestRemove(t *testing.T) {
	client, err := scw.NewClient(nil)
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	fake := &fakeDomainService{}

	// happy path
	fake.deleteFunc = func(ctx context.Context, name string) error {
		return nil
	}

	dnsProvider := NewProvider(client)
	zs, _ := dnsProvider.Zones()
	inZone := &zone{name: "test", client: client}

	err = zs.Remove(inZone)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	// bad status code
	fake.deleteFunc = func(ctx context.Context, name string) error {
		return errors.New("bad response!")
	}

	dnsProvider = NewProvider(client)
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "test", client: client}

	err = zs.Remove(inZone)
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}

	// scw returns error
	fake.deleteFunc = func(ctx context.Context, name string) error {
		return errors.New("error!")
	}

	dnsProvider = NewProvider(client)
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "test", client: client}

	err = zs.Remove(inZone)
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestNewZone(t *testing.T) {
	client, err := scw.NewClient(nil)
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}
	dnsprovider := NewProvider(client)
	zs, _ := dnsprovider.Zones()

	zone, err := zs.New("test")
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}

	if zone.Name() != "test" {
		t.Errorf("unexpected zone name: %v", zone.Name())
	}
}

func TestNewResourceRecordSet(t *testing.T) {
	fake := &fakeDomainService{}
	fake.recordsFunc = func(ctx context.Context, domainName string, listOpts *scw.RequestOption) ([]domain.Record, error) {
		domainRecords := []domain.Record{
			{
				Name: "test",
				Data: "127.0.0.1",
				TTL:  3600,
				Type: "A",
			},
		}
		return domainRecords, nil
	}
	client, err := scw.NewClient(nil)
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	dnsprovider := NewProvider(client)
	zs, _ := dnsprovider.Zones()

	zone, err := zs.New("example.com")
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}

	if zone.Name() != "example.com" {
		t.Errorf("unexpected zone name: %v", zone.Name())
	}

	rrset, _ := zone.ResourceRecordSets()
	rrsets, err := rrset.List()
	if err != nil {
		t.Errorf("error listing resource record sets: %v", err)
	}

	if len(rrsets) != 1 {
		t.Errorf("unexpected number of records: %d", len(rrsets))
	}

	records, err := rrset.Get("test.example.com")
	if err != nil {
		t.Errorf("unexpected error getting resource record set: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("unexpected records from resource record set: %d, expected 1 record", len(records))
	}

	if records[0].Name() != "test.example.com" {
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
}

func TestResourceRecordChangeset(t *testing.T) {
	ctx := context.Background()

	fake := &fakeDomainService{}
	fake.recordsFunc = func(ctx context.Context, domainName string, listOpts *scw.RequestOption) ([]domain.Record, error) {
		domainRecords := []domain.Record{
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
		}

		return domainRecords, nil
	}

	fake.createRecordFunc = func(ctx context.Context, domainName string, createRequest *domain.CreateDNSZoneRequest) (*domain.Record, error) {
		return &domain.Record{}, nil
	}

	fake.deleteRecordFunc = func(ctx context.Context, domainName string, id int) error {
		return nil
	}

	fake.editRecordFunc = func(ctx context.Context, domainName string, id int, editRequest *domain.CreateDNSZoneRequest) (*domain.Record, error) {
		return &domain.Record{}, nil
	}

	fake.editRecordFunc = func(ctx context.Context, domainName string, id int, editRequest *domain.CreateDNSZoneRequest) (*domain.Record, error) {
		return &domain.Record{}, nil
	}

	client, err := scw.NewClient(nil)
	if err != nil {
		t.Errorf("error creating client: %v", err)
	}

	dnsprovider := NewProvider(client)
	zs, _ := dnsprovider.Zones()

	zone, err := zs.New("example.com")
	if err != nil {
		t.Errorf("error creating zone: %v", err)
	}

	if zone.Name() != "example.com" {
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

	record = rrset.New("to-upsert", []string{"127.0.0.1"}, 3600, rrstype.A)
	changeset.Upsert(record)

	err = changeset.Apply(ctx)
	if err != nil {
		t.Errorf("error applying changeset: %v", err)
	}
}
