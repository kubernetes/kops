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
	"io"
	"os"
	"strings"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

var _ dnsprovider.Interface = Interface{}

const (
	ProviderName = "scaleway"
)

func init() {
	dnsprovider.RegisterDNSProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		client, err := newClient()
		if err != nil {
			return nil, err
		}

		return NewProvider(domain.NewAPI(client)), nil
	})
}

// TokenSource implements oauth2.TokenSource
type TokenSource struct {
	AccessToken string
}

// Token returns oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func newClient() (*scw.Client, error) {
	if accessKey := os.Getenv("SCW_ACCESS_KEY"); accessKey == "" {
		return nil, fmt.Errorf("SCW_ACCESS_KEY is required")
	}
	if secretKey := os.Getenv("SCW_SECRET_KEY"); secretKey == "" {
		return nil, fmt.Errorf("SCW_SECRET_KEY is required")
	}

	scwClient, err := scw.NewClient(
		scw.WithUserAgent("kubernetes-kops/"+kopsv.Version),
		scw.WithEnv(),
	)
	if err != nil {
		return nil, err
	}

	return scwClient, nil
}

// Interface implements dnsprovider.Interface
type Interface struct {
	domainAPI DomainAPI
}

// NewProvider returns an implementation of dnsprovider.Interface
func NewProvider(api DomainAPI) dnsprovider.Interface {
	return &Interface{domainAPI: api}
}

// Zones returns an implementation of dnsprovider.Zones
func (d Interface) Zones() (dnsprovider.Zones, bool) {
	return &zones{
		domainAPI: d.domainAPI,
	}, true
}

// zones is an implementation of dnsprovider.Zones
type zones struct {
	domainAPI DomainAPI
}

// List returns a list of all dns zones
func (z *zones) List() ([]dnsprovider.Zone, error) {
	dnsZones, err := z.domainAPI.ListDNSZones(&domain.ListDNSZonesRequest{}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS zones: %w", err)
	}

	zonesList := []dnsprovider.Zone(nil)
	for _, dnsZone := range dnsZones.DNSZones {
		if dnsZone.Domain == "privatedns" {
			continue
		}
		newZone := &zone{
			name:      dnsZone.Domain,
			domainAPI: z.domainAPI,
		}
		zonesList = append(zonesList, newZone)
	}

	return zonesList, nil
}

// Add adds a new DNS zone. The name of the new zone should be of the form "name.domain", otherwise we can't infer the
// domain name from anywhere else in this function
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	newZoneNameSplit := strings.SplitN(newZone.Name(), ".", 2)
	if len(newZoneNameSplit) < 2 {
		return nil, fmt.Errorf("new zone name should contain at least 1 '.', got %q", newZone.Name())
	}
	newZoneName := newZoneNameSplit[0]
	domainName := newZoneNameSplit[1]
	klog.V(8).Infof("Adding new DNS zone %s to domain %s", newZoneName, domainName)

	_, err := z.domainAPI.CreateDNSZone(&domain.CreateDNSZoneRequest{
		Subdomain: newZoneName,
		Domain:    domainName,
	})
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Added new DNS zone %s to domain %s", newZoneName, domainName)

	return &zone{
		name:      newZoneName,
		domainAPI: z.domainAPI,
	}, nil
}

// Remove deletes a zone
func (z *zones) Remove(zone dnsprovider.Zone) error {
	_, err := z.domainAPI.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
		DNSZone: zone.Name(),
	})
	if err != nil {
		return err
	}

	return nil
}

// New returns a new implementation of dnsprovider.Zone
func (z *zones) New(name string) (dnsprovider.Zone, error) {
	return &zone{
		name:      name,
		domainAPI: z.domainAPI,
	}, nil
}

// zone implements dnsprovider.Zone
type zone struct {
	name      string
	domainAPI DomainAPI
}

// Name returns the Name of a dns zone
func (z *zone) Name() string {
	return z.name
}

// ID returns the ID of a dns zone, here we use the name as an identifier
func (z *zone) ID() string {
	return z.name
}

// ResourceRecordSets returns an implementation of dnsprovider.ResourceRecordSets
func (z *zone) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return &resourceRecordSets{zone: z, domainAPI: z.domainAPI}, true
}

// resourceRecordSets implements dnsprovider.ResourceRecordSet
type resourceRecordSets struct {
	zone      *zone
	domainAPI DomainAPI
}

// List returns a list of dnsprovider.ResourceRecordSet
func (r *resourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	records, err := listRecords(r.domainAPI, r.zone.Name())
	if err != nil {
		return nil, err
	}

	var rrsets []dnsprovider.ResourceRecordSet
	rrsetsWithoutDups := make(map[string]*resourceRecordSet)

	for _, record := range records {
		// The scaleway API returns the record without the zone
		// but the consumers of this interface expect the zone to be included
		recordName := dns.EnsureDotSuffix(record.Name) + r.Zone().Name()
		recordKey := recordName + "_" + record.Type.String()
		if rrset, ok := rrsetsWithoutDups[recordKey]; !ok {
			rrsetsWithoutDups[recordKey] = &resourceRecordSet{
				name:       recordName,
				data:       []string{record.Data},
				ttl:        int(record.TTL),
				recordType: rrstype.RrsType(record.Type),
			}
		} else {
			rrset.data = append(rrset.data, record.Data)
		}
	}

	for _, rrset := range rrsetsWithoutDups {
		rrsets = append(rrsets, rrset)
	}

	return rrsets, nil
}

// Get returns a list of dnsprovider.ResourceRecordSet that matches the name parameter. The name should contain the domain name.
func (r *resourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	rrsetList, err := r.List()
	if err != nil {
		return nil, err
	}

	var recordSets []dnsprovider.ResourceRecordSet
	for _, rrset := range rrsetList {
		if rrset.Name() == name {
			recordSets = append(recordSets, rrset)
		}
	}

	return recordSets, nil
}

// New returns an implementation of dnsprovider.ResourceRecordSet. The name should contain the domain name.
func (r *resourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	if len(rrdatas) == 0 {
		return nil
	}

	return &resourceRecordSet{
		name:       name,
		data:       rrdatas,
		ttl:        int(ttl),
		recordType: rrstype,
	}
}

// StartChangeset returns an implementation of dnsprovider.ResourceRecordChangeset
func (r *resourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &resourceRecordChangeset{
		domainAPI: r.domainAPI,
		zone:      r.zone,
		rrsets:    r,
		additions: []dnsprovider.ResourceRecordSet{},
		removals:  []dnsprovider.ResourceRecordSet{},
		upserts:   []dnsprovider.ResourceRecordSet{},
	}
}

// Zone returns the associated implementation of dnsprovider.Zone
func (r *resourceRecordSets) Zone() dnsprovider.Zone {
	return r.zone
}

// recordRecordSet implements dnsprovider.ResourceRecordSet which represents
// a single record associated with a zone
type resourceRecordSet struct {
	name       string
	data       []string
	ttl        int
	recordType rrstype.RrsType
}

// Name returns the name of a resource record set
func (r *resourceRecordSet) Name() string {
	return r.name
}

// Rrdatas returns a list of data associated with a resource record set
func (r *resourceRecordSet) Rrdatas() []string {
	return r.data
}

// Ttl returns the time-to-live of a record
func (r *resourceRecordSet) Ttl() int64 {
	return int64(r.ttl)
}

// Type returns the type of record a resource record set is
func (r *resourceRecordSet) Type() rrstype.RrsType {
	return r.recordType
}

// resourceRecordChangeset implements dnsprovider.ResourceRecordChangeset
type resourceRecordChangeset struct {
	domainAPI DomainAPI
	zone      *zone
	rrsets    dnsprovider.ResourceRecordSets

	additions []dnsprovider.ResourceRecordSet
	removals  []dnsprovider.ResourceRecordSet
	upserts   []dnsprovider.ResourceRecordSet
}

// Add adds a new resource record set to the list of additions to apply
func (r *resourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.additions = append(r.additions, rrset)
	return r
}

// Remove adds a new resource record set to the list of removals to apply
func (r *resourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.removals = append(r.removals, rrset)
	return r
}

// Upsert adds a new resource record set to the list of upserts to apply
func (r *resourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.upserts = append(r.upserts, rrset)
	return r
}

// Apply adds new records stored in r.additions, updates records stored in r.upserts and deletes records stored in r.removals
func (r *resourceRecordChangeset) Apply(ctx context.Context) error {
	// Empty changesets should be a relatively quick no-op
	if r.IsEmpty() {
		klog.V(4).Info("record change set is empty")
		return nil
	}

	changeBatch := []*domain.RecordChange(nil)
	klog.V(8).Infof("applying changes in record change set : [ %d additions | %d upserts | %d removals ]",
		len(r.additions), len(r.upserts), len(r.removals))

	records, err := listRecords(r.domainAPI, r.zone.Name())
	if err != nil {
		return err
	}

	// Scaleway's Domain API doesn't allow more than one edit with the same record name in one request, which happens
	// when there are several control-planes, so we have to check for duplicates in the upsert category and if there are,
	// treat them as additions instead
	if len(r.upserts) > 0 {
		for _, rrset := range r.upserts {
			for i, rrdata := range rrset.Rrdatas() {
				if i == 0 {
					changeBatch = putRecordToUpdateInChangeBatch(changeBatch, rrset, r.zone.Name(), records, rrdata)
				} else {
					rrsetFromIndex1 := r.rrsets.New(rrset.Name(), rrset.Rrdatas()[1:], rrset.Ttl(), rrset.Type())
					changeBatch = putRecordToAddInChangeBatch(changeBatch, rrsetFromIndex1, r.zone.Name())
					break
				}
			}
		}
	}
	if len(r.additions) > 0 {
		for _, rrset := range r.additions {
			changeBatch = putRecordToAddInChangeBatch(changeBatch, rrset, r.zone.Name())
		}
	}
	if len(r.removals) > 0 {
		for _, rrset := range r.removals {
			changeBatch = putRecordToDeleteInChangeBatch(changeBatch, rrset, r.zone.Name(), records)
		}
	}

	_, err = r.domainAPI.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone: r.zone.Name(),
		Changes: changeBatch,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to apply resource record set: %w", err)
	}

	klog.V(2).Info("record change sets successfully applied")
	return nil
}

// IsEmpty returns true if a changeset is empty, false otherwise
func (r *resourceRecordChangeset) IsEmpty() bool {
	if len(r.additions) == 0 && len(r.removals) == 0 && len(r.upserts) == 0 {
		return true
	}

	return false
}

// ResourceRecordSets returns the associated resourceRecordSets of a changeset
func (r *resourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return r.rrsets
}

// listRecords returns a list of scaleway records given a zone name (the name of the record doesn't end with the zone name)
func listRecords(api DomainAPI, zoneName string) ([]*domain.Record, error) {
	records, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: zoneName,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	return records.Records, err
}

func putRecordToAddInChangeBatch(changeBatch []*domain.RecordChange, rrset dnsprovider.ResourceRecordSet, zoneName string) []*domain.RecordChange {
	recordsToAdd := []*domain.Record(nil)
	recordName := strings.TrimSuffix(rrset.Name(), ".")
	recordName = strings.TrimSuffix(recordName, "."+zoneName)
	for _, rrdata := range rrset.Rrdatas() {
		recordsToAdd = append(recordsToAdd, &domain.Record{
			Name: recordName,
			Data: rrdata,
			TTL:  uint32(rrset.Ttl()),
			Type: domain.RecordType(rrset.Type()),
		})
	}
	klog.V(8).Infof("adding new DNS record %q to zone %q", recordName, zoneName)
	return append(changeBatch, &domain.RecordChange{
		Add: &domain.RecordChangeAdd{
			Records: recordsToAdd,
		},
	})
}

func putRecordToUpdateInChangeBatch(changeBatch []*domain.RecordChange, rrset dnsprovider.ResourceRecordSet, zoneName string, records []*domain.Record, rrdata string) []*domain.RecordChange {
	for _, record := range records {
		recordNameWithZone := fmt.Sprintf("%s.%s.", record.Name, zoneName)
		if recordNameWithZone == dns.EnsureDotSuffix(rrset.Name()) && rrset.Type() == rrstype.RrsType(record.Type) {
			klog.V(8).Infof("changing DNS record %q of zone %q", record.Name, zoneName)
			return append(changeBatch, &domain.RecordChange{
				Set: &domain.RecordChangeSet{
					ID: &record.ID,
					Records: []*domain.Record{
						{
							Name: record.Name,
							Data: rrdata,
							TTL:  uint32(rrset.Ttl()),
							Type: domain.RecordType(rrset.Type()),
						},
					},
				},
			})
		}
	}
	return changeBatch
}

func putRecordToDeleteInChangeBatch(changeBatch []*domain.RecordChange, rrset dnsprovider.ResourceRecordSet, zoneName string, records []*domain.Record) []*domain.RecordChange {
	for _, record := range records {
		recordNameWithZone := fmt.Sprintf("%s.%s.", record.Name, zoneName)
		if recordNameWithZone == dns.EnsureDotSuffix(rrset.Name()) && record.Data == rrset.Rrdatas()[0] &&
			rrset.Type() == rrstype.RrsType(record.Type) {
			klog.V(8).Infof("removing DNS record %q of zone %q", record.Name, zoneName)
			return append(changeBatch, &domain.RecordChange{
				Delete: &domain.RecordChangeDelete{
					ID: &record.ID,
				},
			})
		}
	}
	return changeBatch
}
