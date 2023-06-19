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
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/cloudmock/scaleway/mockdns"
)

type DomainAPI interface {
	ListDNSZones(req *domain.ListDNSZonesRequest, opts ...scw.RequestOption) (*domain.ListDNSZonesResponse, error)
	CreateDNSZone(req *domain.CreateDNSZoneRequest, opts ...scw.RequestOption) (*domain.DNSZone, error)
	DeleteDNSZone(req *domain.DeleteDNSZoneRequest, opts ...scw.RequestOption) (*domain.DeleteDNSZoneResponse, error)
	ListDNSZoneRecords(req *domain.ListDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.ListDNSZoneRecordsResponse, error)
	UpdateDNSZoneRecords(req *domain.UpdateDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.UpdateDNSZoneRecordsResponse, error)
}

var _ DomainAPI = &domain.API{}
var _ DomainAPI = &mockdns.FakeDomainAPI{}
