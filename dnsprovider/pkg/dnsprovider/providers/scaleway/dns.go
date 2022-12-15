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

		return NewProvider(client), nil
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
	client *scw.Client
}

// NewProvider returns an implementation of dnsprovider.Interface
func NewProvider(client *scw.Client) dnsprovider.Interface {
	return &Interface{client: client}
}

// Zones returns an implementation of dnsprovider.Zones
func (d Interface) Zones() (dnsprovider.Zones, bool) {
	return &zones{
		client: d.client,
	}, true
}

// zones is an implementation of dnsprovider.Zones
type zones struct {
	client *scw.Client
}

// List returns a list of all dns zones
func (z *zones) List() ([]dnsprovider.Zone, error) {
	api := domain.NewAPI(z.client)

	dnsZones, err := api.ListDNSZones(&domain.ListDNSZonesRequest{}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS zones: %v", err)
	}

	zonesList := []dnsprovider.Zone(nil)
	for _, dnsZone := range dnsZones.DNSZones {
		newZone := &zone{
			name:   dnsZone.Domain,
			client: z.client,
		}
		zonesList = append(zonesList, newZone)
	}

	return zonesList, nil
}

// Add adds a new DNS zone. The name of the new zone should be of the form "name.domain", otherwise we can't infer the
// domain name from anywhere else in this function
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	api := domain.NewAPI(z.client)

	newZoneNameSplit := strings.SplitN(newZone.Name(), ".", 2)
	if len(newZoneNameSplit) < 2 {
		return nil, fmt.Errorf("new zone name should contain at least 1 '.', got %q", newZone.Name())
	}
	newZoneName := newZoneNameSplit[0]
	domainName := newZoneNameSplit[1]
	klog.V(8).Infof("Adding new DNS zone %s to domain %s", newZoneName, domainName)

	_, err := api.CreateDNSZone(&domain.CreateDNSZoneRequest{
		Subdomain: newZoneName,
		Domain:    domainName,
	})
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Added new DNS zone %s to domain %s", newZoneName, domainName)

	return &zone{
		name:   newZoneName,
		client: z.client,
	}, nil
}

// Remove deletes a zone
func (z *zones) Remove(zone dnsprovider.Zone) error {
	api := domain.NewAPI(z.client)

	_, err := api.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
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
		name:   name,
		client: z.client,
	}, nil
}

// zone implements dnsprovider.Zone
type zone struct {
	name   string
	client *scw.Client
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
	return &resourceRecordSets{zone: z, client: z.client}, true
}

// resourceRecordSets implements dnsprovider.ResourceRecordSet
type resourceRecordSets struct {
	zone   *zone
	client *scw.Client
}

// List returns a list of dnsprovider.ResourceRecordSet
func (r *resourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	records, err := listRecords(r.client, r.zone.Name())
	if err != nil {
		return nil, err
	}

	var rrsets []dnsprovider.ResourceRecordSet
	rrsetsWithoutDups := make(map[string]*resourceRecordSet)

	for _, record := range records {
		// The scaleway API returns the record without the zone
		// but the consumers of this interface expect the zone to be included
		recordName := dns.EnsureDotSuffix(record.Name) + r.Zone().Name()
		if rrset, ok := rrsetsWithoutDups[recordName]; !ok {
			rrsetsWithoutDups[recordName] = &resourceRecordSet{
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
		client:    r.client,
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
	client *scw.Client
	zone   *zone
	rrsets dnsprovider.ResourceRecordSets

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

	api := domain.NewAPI(r.client)
	updateRecordsRequest := []*domain.RecordChange(nil)
	klog.V(8).Infof("applying changes in record change set : [ %d additions | %d upserts | %d removals ]",
		len(r.additions), len(r.upserts), len(r.removals))

	records, err := listRecords(r.client, r.zone.Name())
	if err != nil {
		return err
	}

	if len(r.additions) > 0 {
		recordsToAdd := []*domain.Record(nil)
		for _, rrset := range r.additions {
			recordName := strings.TrimSuffix(rrset.Name(), ".")
			recordName = strings.TrimSuffix(recordName, "."+r.zone.Name())
			for _, rrdata := range rrset.Rrdatas() {
				recordsToAdd = append(recordsToAdd, &domain.Record{
					Name: recordName,
					Data: rrdata,
					TTL:  uint32(rrset.Ttl()),
					Type: domain.RecordType(rrset.Type()),
				})
			}
			klog.V(8).Infof("adding new DNS record %q to zone %q", recordName, r.zone.name)
			updateRecordsRequest = append(updateRecordsRequest, &domain.RecordChange{
				Add: &domain.RecordChangeAdd{
					Records: recordsToAdd,
				},
			})
		}
	}

	if len(r.upserts) > 0 {
		for _, rrset := range r.upserts {
			for _, rrdata := range rrset.Rrdatas() {
				for _, record := range records {
					recordNameWithZone := fmt.Sprintf("%s.%s.", record.Name, r.zone.Name())
					if recordNameWithZone == dns.EnsureDotSuffix(rrset.Name()) {
						klog.V(8).Infof("changing DNS record %q of zone %q", record.Name, r.zone.Name())
						updateRecordsRequest = append(updateRecordsRequest, &domain.RecordChange{
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
			}
		}
	}

	if len(r.removals) > 0 {
		for _, rrset := range r.removals {
			for _, record := range records {
				recordNameWithZone := fmt.Sprintf("%s.%s.", record.Name, r.zone.Name())
				if recordNameWithZone == dns.EnsureDotSuffix(rrset.Name()) && record.Data == rrset.Rrdatas()[0] {
					klog.V(8).Infof("removing DNS record %q of zone %q", record.Name, r.zone.name)
					updateRecordsRequest = append(updateRecordsRequest, &domain.RecordChange{
						Delete: &domain.RecordChangeDelete{
							ID: &record.ID,
						},
					})
				}

			}
		}
	}

	_, err = api.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone: r.zone.Name(),
		Changes: updateRecordsRequest,
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
func listRecords(c *scw.Client, zoneName string) ([]*domain.Record, error) {
	api := domain.NewAPI(c)

	records, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: zoneName,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %v", err)
	}

	return records.Records, err
}
