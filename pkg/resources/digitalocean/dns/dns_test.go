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

package dns

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/digitalocean/godo"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

type fakeDomainService struct {
	listFunc         func(ctx context.Context, listOpt *godo.ListOptions) ([]godo.Domain, *godo.Response, error)
	getFunc          func(ctx context.Context, name string) (*godo.Domain, *godo.Response, error)
	createFunc       func(ctx context.Context, domainCreateRequest *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error)
	deleteFunc       func(ctx context.Context, name string) (*godo.Response, error)
	recordsFunc      func(ctx context.Context, domain string, listOpts *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error)
	recordFunc       func(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error)
	deleteRecordFunc func(ctx context.Context, domain string, id int) (*godo.Response, error)
	editRecordFunc   func(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error)
	createRecordFunc func(ctx context.Context, domain string, createRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error)
}

func (f *fakeDomainService) List(ctx context.Context, listOpt *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return f.listFunc(ctx, listOpt)
}

func (f *fakeDomainService) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return f.getFunc(ctx, name)
}

func (f *fakeDomainService) Create(ctx context.Context, domainCreateRequest *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return f.createFunc(ctx, domainCreateRequest)
}

func (f *fakeDomainService) Delete(ctx context.Context, name string) (*godo.Response, error) {
	return f.deleteFunc(ctx, name)
}

func (f *fakeDomainService) Records(ctx context.Context, domain string, listOpts *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return f.recordsFunc(ctx, domain, listOpts)
}

func (f *fakeDomainService) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return f.recordFunc(ctx, domain, id)
}

func (f *fakeDomainService) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return f.deleteRecordFunc(ctx, domain, id)
}

func (f *fakeDomainService) EditRecord(ctx context.Context, domain string,
	id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return f.editRecordFunc(ctx, domain, id, editRequest)
}

func (f *fakeDomainService) CreateRecord(ctx context.Context, domain string,
	createRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return f.createRecordFunc(ctx, domain, createRequest)
}

func TestZonesList(t *testing.T) {
	client := godo.NewClient(nil)
	fake := &fakeDomainService{}

	// happy path
	fake.listFunc = func(ctx context.Context, listOpts *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
		domains := []godo.Domain{
			{
				Name: "example.com",
			},
		}

		return domains, nil, nil
	}
	client.Domains = fake

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
	fake.listFunc = func(ctx context.Context, listOpts *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
		domains := []godo.Domain{
			{
				Name: "example.com",
			},
		}

		return domains, nil, errors.New("internal error!")
	}
	client.Domains = fake

	z = &zones{client}
	zoneList, err = z.List()
	if err == nil {
		t.Errorf("expected non-nil err")
	}

	if zoneList != nil {
		t.Errorf("expected nil zone, got %v", zoneList)
	}

	// godo client returned error path
	fake.listFunc = func(ctx context.Context, listOpts *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
		return nil, nil, errors.New("error!")
	}
	client.Domains = fake

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
	client := godo.NewClient(nil)
	fake := &fakeDomainService{}

	// happy path
	fake.createFunc = func(ctx context.Context, domainCreateRequest *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
		domain := &godo.Domain{Name: domainCreateRequest.Name}
		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK

		return domain, resp, nil
	}
	client.Domains = fake

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
	fake.createFunc = func(ctx context.Context, domainCreateRequest *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
		domain := &godo.Domain{Name: domainCreateRequest.Name}

		return domain, nil, errors.New("bad response!")
	}
	client.Domains = fake

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

	// godo returns error
	fake.createFunc = func(ctx context.Context, domainCreateRequest *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
		domain := &godo.Domain{Name: domainCreateRequest.Name}

		return domain, nil, errors.New("error!")
	}
	client.Domains = fake

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
	client := godo.NewClient(nil)
	fake := &fakeDomainService{}

	// happy path
	fake.deleteFunc = func(ctx context.Context, name string) (*godo.Response, error) {
		return nil, nil
	}
	client.Domains = fake

	dnsProvider := NewProvider(client)
	zs, _ := dnsProvider.Zones()
	inZone := &zone{name: "test", client: client}

	err := zs.Remove(inZone)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	// bad status code
	fake.deleteFunc = func(ctx context.Context, name string) (*godo.Response, error) {
		return nil, errors.New("bad response!")
	}
	client.Domains = fake

	dnsProvider = NewProvider(client)
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "test", client: client}

	err = zs.Remove(inZone)
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}

	// godo returns error
	fake.deleteFunc = func(ctx context.Context, name string) (*godo.Response, error) {
		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK
		resp.Body = ioutil.NopCloser(bytes.NewBufferString("error!"))
		return resp, errors.New("error!")
	}
	client.Domains = fake

	dnsProvider = NewProvider(client)
	zs, _ = dnsProvider.Zones()
	inZone = &zone{name: "test", client: client}

	err = zs.Remove(inZone)
	if err == nil {
		t.Errorf("expected non-nil err: %v", err)
	}
}

func TestNewZone(t *testing.T) {
	client := godo.NewClient(nil)
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
	fake.recordsFunc = func(ctx context.Context, domain string, listOpts *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
		domainRecords := []godo.DomainRecord{
			{
				Name: "test",
				Data: "127.0.0.1",
				TTL:  3600,
				Type: "A",
			},
		}

		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK

		return domainRecords, resp, nil
	}
	client := godo.NewClient(nil)
	client.Domains = fake

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
	fake := &fakeDomainService{}
	fake.recordsFunc = func(ctx context.Context, domain string, listOpts *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
		domainRecords := []godo.DomainRecord{
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

		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK

		return domainRecords, resp, nil
	}

	fake.createRecordFunc = func(ctx context.Context, domain string, createRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK
		return &godo.DomainRecord{}, resp, nil
	}

	fake.deleteRecordFunc = func(ctx context.Context, domain string, id int) (*godo.Response, error) {
		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK
		return resp, nil
	}

	fake.editRecordFunc = func(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
		resp := &godo.Response{
			Response: &http.Response{},
		}
		resp.StatusCode = http.StatusOK
		return &godo.DomainRecord{}, resp, nil
	}

	client := godo.NewClient(nil)
	client.Domains = fake

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

	err = changeset.Apply()
	if err != nil {
		t.Errorf("error applying changeset: %v", err)
	}
}
