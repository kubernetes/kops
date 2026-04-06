/*
Copyright 2024 The Kubernetes Authors.

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

package linode

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/linode/linodego"
	"k8s.io/klog/v2"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

var _ dnsprovider.Interface = (*Interface)(nil)

const (
	ProviderName = "linode"
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

// newClient creates a new Linode (Akamai) API client using the LINODE_TOKEN environment variable
func newClient() (*linodego.Client, error) {
	apiToken := os.Getenv("LINODE_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("LINODE_TOKEN environment variable is required")
	}

	client := linodego.NewClient(nil)
	client.SetUserAgent("kops/dns-controller")
	client.SetToken(apiToken)
	return &client, nil
}

// Interface implements dnsprovider.Interface for Linode (Akamai) DNS
type Interface struct {
	client *linodego.Client
}

// NewProvider returns a Linode (Akamai) DNS provider
func NewProvider(client *linodego.Client) dnsprovider.Interface {
	return &Interface{
		client: client,
	}
}

// Zones returns the zones managed by this provider
func (i *Interface) Zones() (dnsprovider.Zones, bool) {
	return &zones{
		client: i.client,
	}, true
}

// zones implements dnsprovider.Zones
type zones struct {
	client *linodego.Client
}

// List returns all DNS zones
func (z *zones) List() ([]dnsprovider.Zone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &linodego.ListOptions{}
	klog.V(2).Infof("Listing Linode (Akamai) domains...")
	domains, err := z.client.ListDomains(ctx, opts)
	if err != nil {
		klog.Errorf("Failed to list Linode (Akamai) domains: %v", err)
		return nil, fmt.Errorf("failed to list Linode (Akamai) domains: %w", err)
	}
	klog.V(2).Infof("Found %d Linode (Akamai) domains", len(domains))

	var zoneList []dnsprovider.Zone
	for _, domain := range domains {
		klog.V(2).Infof("Adding zone: %s (ID: %d)", domain.Domain, domain.ID)
		zone := &zone{
			name:   domain.Domain,
			id:     domain.Domain,
			client: z.client,
		}
		zoneList = append(zoneList, zone)
	}

	return zoneList, nil
}

// Add creates a new DNS zone
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	opts := linodego.DomainCreateOptions{
		Domain: newZone.Name(),
		Type:   linodego.DomainTypeMaster,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	domain, err := z.client.CreateDomain(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	return &zone{
		name:   domain.Domain,
		id:     domain.Domain,
		client: z.client,
	}, nil
}

// Remove deletes a DNS zone
func (z *zones) Remove(zone dnsprovider.Zone) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	domains, err := z.client.ListDomains(ctx, &linodego.ListOptions{})
	if err != nil {
		return err
	}

	for _, domain := range domains {
		if domain.Domain == zone.Name() {
			return z.client.DeleteDomain(ctx, domain.ID)
		}
	}

	return fmt.Errorf("zone %s not found", zone.Name())
}

// New creates a new zone object
func (z *zones) New(name string) (dnsprovider.Zone, error) {
	return &zone{
		name:   name,
		id:     name,
		client: z.client,
	}, nil
}

// zone implements dnsprovider.Zone
type zone struct {
	name   string
	id     string
	client *linodego.Client
}

// Name returns the zone name
func (z *zone) Name() string {
	return z.name
}

// ID returns the zone ID (which is the domain name in Linode / Akamai)
func (z *zone) ID() string {
	return z.id
}

// ResourceRecordSets returns the resource record sets for this zone
func (z *zone) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return &resourceRecordSets{
		zone:   z,
		client: z.client,
	}, true
}

// resourceRecordSets implements dnsprovider.ResourceRecordSets
type resourceRecordSets struct {
	zone   *zone
	client *linodego.Client
}

// List returns all resource record sets in the zone
func (r *resourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	domains, err := r.client.ListDomains(ctx, &linodego.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Linode (Akamai) domains: %w", err)
	}

	var domainID int
	for _, domain := range domains {
		if domain.Domain == r.zone.Name() {
			domainID = domain.ID
			break
		}
	}

	if domainID == 0 {
		return nil, fmt.Errorf("domain %s not found", r.zone.Name())
	}

	records, err := r.client.ListDomainRecords(ctx, domainID, &linodego.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Linode (Akamai) domain records for domain %s: %w", r.zone.Name(), err)
	}

	// Group records by name and type to handle multiple data values
	rrsetMap := make(map[string]*resourceRecordSet)

	for _, record := range records {
		// Construct full name with zone suffix and ensure trailing dot
		fullName := dns.EnsureDotSuffix(record.Name) + dns.EnsureDotSuffix(r.zone.Name())
		key := fullName + "|" + string(record.Type)

		if rrset, exists := rrsetMap[key]; exists {
			rrset.data = append(rrset.data, record.Target)
		} else {
			rrsetMap[key] = &resourceRecordSet{
				name:       fullName,
				data:       []string{record.Target},
				ttl:        int64(record.TTLSec),
				recordType: rrstype.RrsType(record.Type),
			}
		}
	}

	var rrsets []dnsprovider.ResourceRecordSet
	for _, rrset := range rrsetMap {
		rrsets = append(rrsets, rrset)
	}

	return rrsets, nil
}

// Get returns resource record sets matching the name
func (r *resourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	records, err := r.List()
	if err != nil {
		return nil, err
	}

	var matches []dnsprovider.ResourceRecordSet
	for _, record := range records {
		if record.Name() == name {
			matches = append(matches, record)
		}
	}

	return matches, nil
}

// New creates a new resource record set
func (r *resourceRecordSets) New(name string, rrdatas []string, ttl int64, rrtype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	if len(rrdatas) == 0 {
		return nil
	}

	return &resourceRecordSet{
		name:       name,
		data:       rrdatas,
		ttl:        ttl,
		recordType: rrtype,
	}
}

// StartChangeset returns a new changeset
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

// Zone returns the associated zone
func (r *resourceRecordSets) Zone() dnsprovider.Zone {
	return r.zone
}

// resourceRecordSet implements dnsprovider.ResourceRecordSet
type resourceRecordSet struct {
	name       string
	data       []string
	ttl        int64
	recordType rrstype.RrsType
}

// Name returns the record name
func (r *resourceRecordSet) Name() string {
	return r.name
}

// Rrdatas returns the record data
func (r *resourceRecordSet) Rrdatas() []string {
	return r.data
}

// Ttl returns the time-to-live
func (r *resourceRecordSet) Ttl() int64 {
	return r.ttl
}

// Type returns the record type
func (r *resourceRecordSet) Type() rrstype.RrsType {
	return r.recordType
}

// resourceRecordChangeset implements dnsprovider.ResourceRecordChangeset
type resourceRecordChangeset struct {
	client *linodego.Client
	zone   *zone
	rrsets dnsprovider.ResourceRecordSets

	additions []dnsprovider.ResourceRecordSet
	removals  []dnsprovider.ResourceRecordSet
	upserts   []dnsprovider.ResourceRecordSet
}

// ResourceRecordSets returns the associated resource record sets
func (r *resourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return r.rrsets
}

// Add adds a new resource record set
func (r *resourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.additions = append(r.additions, rrset)
	return r
}

// Remove removes a resource record set
func (r *resourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.removals = append(r.removals, rrset)
	return r
}

// Upsert upserts a resource record set
func (r *resourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.upserts = append(r.upserts, rrset)
	return r
}

// Apply applies all pending changes
func (r *resourceRecordChangeset) Apply(ctx context.Context) error {
	if r.IsEmpty() {
		klog.V(4).Info("record change set is empty")
		return nil
	}

	klog.V(2).Info("applying changes in record change set")

	// Get domain ID
	domains, err := r.client.ListDomains(ctx, &linodego.ListOptions{})
	if err != nil {
		return err
	}

	var domainID int
	for _, domain := range domains {
		if domain.Domain == r.zone.Name() {
			domainID = domain.ID
			break
		}
	}

	if domainID == 0 {
		return fmt.Errorf("domain %s not found", r.zone.Name())
	}

	// Apply removals
	if len(r.removals) > 0 {
		for _, rrset := range r.removals {
			if err := r.deleteRecord(ctx, domainID, rrset); err != nil {
				return fmt.Errorf("failed to remove record %s: %w", rrset.Name(), err)
			}
		}
		klog.V(2).Info("record change set removals complete")
	}

	// Apply additions
	if len(r.additions) > 0 {
		for _, rrset := range r.additions {
			if err := r.createRecord(ctx, domainID, rrset); err != nil {
				return fmt.Errorf("failed to add record %s: %w", rrset.Name(), err)
			}
		}
		klog.V(2).Info("record change set additions complete")
	}

	// Apply upserts
	if len(r.upserts) > 0 {
		for _, rrset := range r.upserts {
			// Delete existing records first
			if err := r.deleteRecord(ctx, domainID, rrset); err != nil {
				klog.V(2).Infof("error deleting existing record %s (may not exist): %v", rrset.Name(), err)
			}
			// Then create new ones
			if err := r.createRecord(ctx, domainID, rrset); err != nil {
				return fmt.Errorf("failed to upsert record %s: %w", rrset.Name(), err)
			}
		}
		klog.V(2).Info("record change set upserts complete")
	}

	return nil
}

// IsEmpty returns true if no changes are pending
func (r *resourceRecordChangeset) IsEmpty() bool {
	return len(r.additions) == 0 && len(r.removals) == 0 && len(r.upserts) == 0
}

// createRecord creates a new DNS record
func (r *resourceRecordChangeset) createRecord(ctx context.Context, domainID int, rrset dnsprovider.ResourceRecordSet) error {
	// Extract record name without zone suffix
	recordName := strings.TrimSuffix(rrset.Name(), ".")
	recordName = strings.TrimSuffix(recordName, "."+r.zone.Name())

	// Create one record for each data entry
	for _, data := range rrset.Rrdatas() {
		opts := linodego.DomainRecordCreateOptions{
			Type:   linodego.DomainRecordType(rrset.Type()),
			Name:   recordName,
			Target: data,
			TTLSec: int(rrset.Ttl()),
		}

		_, err := r.client.CreateDomainRecord(ctx, domainID, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

// deleteRecord deletes DNS records matching the resource record set
func (r *resourceRecordChangeset) deleteRecord(ctx context.Context, domainID int, rrset dnsprovider.ResourceRecordSet) error {
	// Extract record name without zone suffix
	recordName := strings.TrimSuffix(rrset.Name(), ".")
	recordName = strings.TrimSuffix(recordName, "."+r.zone.Name())

	records, err := r.client.ListDomainRecords(ctx, domainID, &linodego.ListOptions{})
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Name == recordName && string(record.Type) == string(rrset.Type()) {
			if err := r.client.DeleteDomainRecord(ctx, domainID, record.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
