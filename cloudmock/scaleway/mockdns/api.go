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

package mockdns

import (
	"fmt"

	"github.com/google/uuid"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type FakeDomainAPI struct {
	DNSZones []*domain.DNSZone
	Records  map[string]*domain.Record
}

func (f *FakeDomainAPI) ListDNSZones(req *domain.ListDNSZonesRequest, opts ...scw.RequestOption) (*domain.ListDNSZonesResponse, error) {
	if f.DNSZones == nil {
		return nil, fmt.Errorf("error response")
	}
	return &domain.ListDNSZonesResponse{
		TotalCount: uint32(len(f.DNSZones)),
		DNSZones:   f.DNSZones,
	}, nil
}

func (f *FakeDomainAPI) CreateDNSZone(req *domain.CreateDNSZoneRequest, opts ...scw.RequestOption) (*domain.DNSZone, error) {
	if f.DNSZones == nil {
		return nil, fmt.Errorf("error response")
	}
	newZone := &domain.DNSZone{
		Domain:    req.Domain,
		Subdomain: req.Subdomain,
	}
	f.DNSZones = append(f.DNSZones, newZone)
	return newZone, nil
}

func (f *FakeDomainAPI) DeleteDNSZone(req *domain.DeleteDNSZoneRequest, opts ...scw.RequestOption) (*domain.DeleteDNSZoneResponse, error) {
	if f.DNSZones == nil {
		return nil, fmt.Errorf("error response")
	}
	var newZoneList []*domain.DNSZone
	for _, zone := range f.DNSZones {
		if req.DNSZone == fmt.Sprintf("%s.%s", zone.Subdomain, zone.Domain) {
			continue
		}
		newZoneList = append(newZoneList, zone)
	}
	f.DNSZones = newZoneList
	return &domain.DeleteDNSZoneResponse{}, nil
}

func (f *FakeDomainAPI) ListDNSZoneRecords(req *domain.ListDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.ListDNSZoneRecordsResponse, error) {
	var recordsList []*domain.Record
	for _, record := range f.Records {
		recordsList = append(recordsList, record)
	}
	records := &domain.ListDNSZoneRecordsResponse{
		TotalCount: uint32(len(f.Records)),
		Records:    recordsList,
	}
	return records, nil
}

func (f *FakeDomainAPI) UpdateDNSZoneRecords(req *domain.UpdateDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.UpdateDNSZoneRecordsResponse, error) {
	for _, change := range req.Changes {

		if change.Add != nil {
			for _, toAdd := range change.Add.Records {
				toAdd.ID = uuid.New().String()
				f.Records[toAdd.Name] = toAdd
			}

		} else if change.Set != nil {
			if len(change.Set.Records) != 1 {
				fmt.Printf("only 1 record change will be applied from %d changes requested", len(change.Set.Records))
			}
			for _, toUpsert := range change.Set.Records {
				if _, ok := f.Records[toUpsert.Name]; !ok {
					return nil, fmt.Errorf("could not find record %s to upsert", toUpsert.Name)
				}
				toUpsert.ID = *change.Set.ID
				f.Records[toUpsert.Name] = toUpsert
			}
		} else if change.Delete != nil {
			found := false
			for name, record := range f.Records {
				if record.ID == *change.Delete.ID {
					delete(f.Records, name)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("could not find record %s to delete", *change.Delete.ID)
			}

		} else {
			return nil, fmt.Errorf("mock DNS not implemented for this method")
		}
	}

	var recordsList []*domain.Record
	for _, record := range f.Records {
		recordsList = append(recordsList, record)
	}
	return &domain.UpdateDNSZoneRecordsResponse{Records: recordsList}, nil
}
