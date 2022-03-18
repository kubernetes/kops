/*
Copyright 2021 The Kubernetes Authors.

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

package gce

import (
	"context"
	"fmt"

	dns "google.golang.org/api/dns/v1"
)

type DNSClient interface {
	ManagedZones() ManagedZoneClient
	ResourceRecordSets() ResourceRecordSetClient
	Changes() ChangeClient
}

type dnsClientImpl struct {
	srv *dns.Service
}

var _ DNSClient = &dnsClientImpl{}

func newDNSClientImpl(ctx context.Context) (*dnsClientImpl, error) {
	srv, err := dns.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building DNS API client: %v", err)
	}
	return &dnsClientImpl{
		srv: srv,
	}, nil
}

func (c *dnsClientImpl) ManagedZones() ManagedZoneClient {
	return &managedZoneClientImpl{
		srv: c.srv.ManagedZones,
	}
}

func (c *dnsClientImpl) ResourceRecordSets() ResourceRecordSetClient {
	return &resourceRecordSetClientImpl{
		srv: c.srv.ResourceRecordSets,
	}
}

func (c *dnsClientImpl) Changes() ChangeClient {
	return &changeClientImpl{
		srv: c.srv.Changes,
	}
}

type ManagedZoneClient interface {
	List(project string) ([]*dns.ManagedZone, error)
	Insert(project string, zone *dns.ManagedZone) error
	Delete(project string, zoneName string) error
}

type managedZoneClientImpl struct {
	srv *dns.ManagedZonesService
}

var _ ManagedZoneClient = &managedZoneClientImpl{}

func (c *managedZoneClientImpl) Insert(project string, zone *dns.ManagedZone) error {
	_, err := c.srv.Create(project, zone).Do()
	return err
}

func (c *managedZoneClientImpl) Delete(project string, zoneName string) error {
	err := c.srv.Delete(project, zoneName).Do()
	return err
}

func (c *managedZoneClientImpl) List(project string) ([]*dns.ManagedZone, error) {
	r, err := c.srv.List(project).Do()
	if err != nil {
		return nil, err
	}
	return r.ManagedZones, nil
}

type ResourceRecordSetClient interface {
	List(project, zone string) ([]*dns.ResourceRecordSet, error)
}

type resourceRecordSetClientImpl struct {
	srv *dns.ResourceRecordSetsService
}

var _ ResourceRecordSetClient = &resourceRecordSetClientImpl{}

func (c *resourceRecordSetClientImpl) List(project, zone string) ([]*dns.ResourceRecordSet, error) {
	r, err := c.srv.List(project, zone).Do()
	if err != nil {
		return nil, err
	}
	return r.Rrsets, nil
}

type ChangeClient interface {
	Create(project, zone string, ch *dns.Change) (*dns.Change, error)
}

type changeClientImpl struct {
	srv *dns.ChangesService
}

var _ ChangeClient = &changeClientImpl{}

func (c *changeClientImpl) Create(project, zone string, ch *dns.Change) (*dns.Change, error) {
	return c.srv.Create(project, zone, ch).Do()
}
