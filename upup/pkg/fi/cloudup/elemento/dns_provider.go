/*
Copyright 2026 The Kubernetes Authors.

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

package elemento

import (
	"context"
	"fmt"
	"strings"

	"github.com/Elemento-Modular-Cloud/ecloud-go/ecloud"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

type dnsClient interface {
	Create(ctx context.Context, zoneName string) (*ecloud.Dns, *ecloud.Response, error)
	AddDnsRecord(ctx context.Context, zoneName, recordName, recordValue string) (*ecloud.DnsRecord, *ecloud.Response, error)
	Get(ctx context.Context, zoneName string) (*ecloud.Dns, *ecloud.Response, error)
	ListDnsRecords(ctx context.Context, zoneName string) ([]*ecloud.DnsRecord, *ecloud.Response, error)
	GetDnsRecord(ctx context.Context, zoneName, recordName, recordType string) (*ecloud.DnsRecord, *ecloud.Response, error)
}

type dnsListClient interface {
	List(ctx context.Context) ([]*ecloud.Dns, *ecloud.Response, error)
}

// NewDNSProvider adapts the Elemento DNS SDK to kOps' generic dnsprovider
// interface. Deletion is intentionally unsupported until the SDK exposes a
// delete operation; Apply returns an explicit error for removals.
func NewDNSProvider(client ecloud.DnsClient, defaultZoneName string) (dnsprovider.Interface, error) {
	dnsClient, ok := any(&client).(dnsClient)
	if !ok {
		return nil, fmt.Errorf("Elemento DNS SDK does not implement the read methods required by the kOps DNS provider")
	}

	return &dnsProvider{
		client:          dnsClient,
		defaultZoneName: defaultZoneName,
	}, nil
}

type dnsProvider struct {
	client          dnsClient
	defaultZoneName string
}

var _ dnsprovider.Interface = &dnsProvider{}

func (p *dnsProvider) Zones() (dnsprovider.Zones, bool) {
	return &dnsZones{
		client:          p.client,
		defaultZoneName: p.defaultZoneName,
	}, true
}

type dnsZones struct {
	client          dnsClient
	defaultZoneName string
}

var _ dnsprovider.Zones = &dnsZones{}

func (z *dnsZones) List() ([]dnsprovider.Zone, error) {
	if listClient, ok := z.client.(dnsListClient); ok {
		dnsServices, _, err := listClient.List(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("listing Elemento DNS zones: %w", err)
		}
		return zonesFromElementoDNS(z.client, dnsServices), nil
	}

	if z.defaultZoneName == "" {
		return nil, fmt.Errorf("listing Elemento DNS zones requires DnsClient.List or a default zone name")
	}

	dnsService, _, err := z.client.Get(context.TODO(), z.defaultZoneName)
	if err != nil {
		return nil, fmt.Errorf("getting Elemento DNS zone %q: %w", z.defaultZoneName, err)
	}
	if dnsService == nil {
		return nil, nil
	}

	return zonesFromElementoDNS(z.client, []*ecloud.Dns{dnsService}), nil
}

func (z *dnsZones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	dnsService, _, err := z.client.Create(context.TODO(), newZone.Name())
	if err != nil && !IsDNSAlreadyExists(err) {
		return nil, fmt.Errorf("creating Elemento DNS zone %q: %w", newZone.Name(), err)
	}
	if dnsService == nil {
		dnsService, _, err = z.client.Get(context.TODO(), newZone.Name())
		if err != nil {
			return nil, fmt.Errorf("getting Elemento DNS zone %q after create: %w", newZone.Name(), err)
		}
	}

	return zoneFromElementoDNS(z.client, newZone.Name(), dnsService), nil
}

func (z *dnsZones) Remove(zone dnsprovider.Zone) error {
	return fmt.Errorf("deleting Elemento DNS zones is not supported yet")
}

func (z *dnsZones) New(name string) (dnsprovider.Zone, error) {
	return &dnsZone{
		client: z.client,
		name:   name,
		id:     name,
	}, nil
}

func zonesFromElementoDNS(client dnsClient, dnsServices []*ecloud.Dns) []dnsprovider.Zone {
	var zones []dnsprovider.Zone
	for _, dnsService := range dnsServices {
		if dnsService == nil {
			continue
		}
		zones = append(zones, zoneFromElementoDNS(client, dnsService.ZoneName, dnsService))
	}
	return zones
}

func zoneFromElementoDNS(client dnsClient, name string, dnsService *ecloud.Dns) dnsprovider.Zone {
	id := name
	if dnsService != nil {
		if dnsService.ZoneName != "" {
			name = dnsService.ZoneName
		}
		if dnsService.ID != "" {
			id = dnsService.ID
		}
	}

	return &dnsZone{
		client: client,
		name:   name,
		id:     id,
	}
}

type dnsZone struct {
	client dnsClient
	name   string
	id     string
}

var _ dnsprovider.Zone = &dnsZone{}

func (z *dnsZone) Name() string {
	return z.name
}

func (z *dnsZone) ID() string {
	return z.id
}

func (z *dnsZone) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return &dnsResourceRecordSets{
		client: z.client,
		zone:   z,
	}, true
}

type dnsResourceRecordSets struct {
	client dnsClient
	zone   *dnsZone
}

var _ dnsprovider.ResourceRecordSets = &dnsResourceRecordSets{}

func (r *dnsResourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	records, _, err := r.client.ListDnsRecords(context.TODO(), r.zone.Name())
	if err != nil {
		return nil, fmt.Errorf("listing Elemento DNS records in zone %q: %w", r.zone.Name(), err)
	}

	var rrsets []dnsprovider.ResourceRecordSet
	for _, record := range records {
		if record == nil {
			continue
		}
		rrsets = append(rrsets, rrsetFromElementoRecord(record))
	}

	return rrsets, nil
}

func (r *dnsResourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	recordSets, err := r.List()
	if err != nil {
		return nil, err
	}

	var matches []dnsprovider.ResourceRecordSet
	for _, recordSet := range recordSets {
		if recordSet.Name() == name {
			matches = append(matches, recordSet)
		}
	}
	return matches, nil
}

func (r *dnsResourceRecordSets) New(name string, rrdatas []string, ttl int64, recordType rrstype.RrsType) dnsprovider.ResourceRecordSet {
	if len(rrdatas) == 0 {
		return nil
	}
	return &dnsResourceRecordSet{
		name:       name,
		rrdatas:    rrdatas,
		ttl:        ttl,
		recordType: recordType,
	}
}

func (r *dnsResourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &dnsResourceRecordChangeset{
		client:   r.client,
		rrsets:   r,
		zoneName: r.zone.Name(),
	}
}

func (r *dnsResourceRecordSets) Zone() dnsprovider.Zone {
	return r.zone
}

func rrsetFromElementoRecord(record *ecloud.DnsRecord) dnsprovider.ResourceRecordSet {
	return &dnsResourceRecordSet{
		name:       record.Name,
		rrdatas:    []string{record.Value},
		ttl:        int64(record.TTL),
		recordType: rrstype.RrsType(record.Type),
	}
}

type dnsResourceRecordSet struct {
	name       string
	rrdatas    []string
	ttl        int64
	recordType rrstype.RrsType
}

var _ dnsprovider.ResourceRecordSet = &dnsResourceRecordSet{}

func (r *dnsResourceRecordSet) Name() string {
	return r.name
}

func (r *dnsResourceRecordSet) Rrdatas() []string {
	return r.rrdatas
}

func (r *dnsResourceRecordSet) Ttl() int64 {
	return r.ttl
}

func (r *dnsResourceRecordSet) Type() rrstype.RrsType {
	return r.recordType
}

type dnsResourceRecordChangeset struct {
	client   dnsClient
	rrsets   dnsprovider.ResourceRecordSets
	zoneName string

	additions []dnsprovider.ResourceRecordSet
	removals  []dnsprovider.ResourceRecordSet
	upserts   []dnsprovider.ResourceRecordSet
}

var _ dnsprovider.ResourceRecordChangeset = &dnsResourceRecordChangeset{}

func (c *dnsResourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.additions = append(c.additions, rrset)
	return c
}

func (c *dnsResourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.removals = append(c.removals, rrset)
	return c
}

func (c *dnsResourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.upserts = append(c.upserts, rrset)
	return c
}

func (c *dnsResourceRecordChangeset) Apply(ctx context.Context) error {
	if c.IsEmpty() {
		return nil
	}
	if len(c.removals) != 0 {
		return fmt.Errorf("deleting Elemento DNS records is not supported yet")
	}

	if _, _, err := c.client.Create(ctx, c.zoneName); err != nil && !IsDNSAlreadyExists(err) {
		return fmt.Errorf("creating Elemento DNS zone %q: %w", c.zoneName, err)
	}

	for _, rrset := range append(c.additions, c.upserts...) {
		if rrset == nil {
			continue
		}
		if rrset.Type() != rrstype.A {
			return fmt.Errorf("Elemento DNS currently supports only A records, got %q for %q", rrset.Type(), rrset.Name())
		}

		recordName := trimZoneSuffix(rrset.Name(), c.zoneName)
		for _, rrdata := range rrset.Rrdatas() {
			if _, _, err := c.client.AddDnsRecord(ctx, c.zoneName, recordName, rrdata); err != nil {
				return fmt.Errorf("ensuring Elemento DNS record %q in zone %q: %w", recordName, c.zoneName, err)
			}
		}
	}

	return nil
}

func (c *dnsResourceRecordChangeset) IsEmpty() bool {
	return len(c.additions) == 0 && len(c.upserts) == 0 && len(c.removals) == 0
}

func (c *dnsResourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return c.rrsets
}

// IsDNSAlreadyExists reports whether an Elemento DNS create/upsert operation
// found an existing resource that is already compatible with the request.
func IsDNSAlreadyExists(err error) bool {
	if ecloud.IsError(err, ecloud.ErrorCodeUniquenessError, ecloud.ErrorCodeConflict) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "already exists") ||
		strings.Contains(message, "already defined") ||
		strings.Contains(message, "uniqueness")
}

func trimZoneSuffix(name, zone string) string {
	zone = strings.TrimSuffix(zone, ".")
	return strings.TrimSuffix(name, "."+zone)
}
