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
	apiTimeout   = 30 * time.Second
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

func (i *Interface) Zones() (dnsprovider.Zones, bool) {
	return &zones{
		client: i.client,
	}, true
}

// zones implements dnsprovider.Zones
type zones struct {
	client *linodego.Client
}

func (z *zones) List() ([]dnsprovider.Zone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	domains, err := z.client.ListDomains(ctx, &linodego.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) domains: %w", err)
	}
	klog.V(2).Infof("Found %d Linode (Akamai) domains", len(domains))

	var zoneList []dnsprovider.Zone
	for _, domain := range domains {
		zoneList = append(zoneList, &zone{
			name:     domain.Domain,
			id:       domain.Domain,
			domainID: domain.ID,
			client:   z.client,
		})
	}

	return zoneList, nil
}

func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	domain, err := z.client.CreateDomain(ctx, linodego.DomainCreateOptions{
		Domain: newZone.Name(),
		Type:   linodego.DomainTypeMaster,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating domain: %w", err)
	}

	return &zone{
		name:     domain.Domain,
		id:       domain.Domain,
		domainID: domain.ID,
		client:   z.client,
	}, nil
}

func (z *zones) Remove(zn dnsprovider.Zone) error {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	linodeZone, ok := zn.(*zone)
	if !ok {
		return fmt.Errorf("unexpected zone type %T", zn)
	}
	domainID, err := linodeZone.getDomainID(ctx)
	if err != nil {
		return err
	}
	return z.client.DeleteDomain(ctx, domainID)
}

func (z *zones) New(name string) (dnsprovider.Zone, error) {
	return &zone{
		name:   name,
		id:     name,
		client: z.client,
	}, nil
}

// zone implements dnsprovider.Zone
type zone struct {
	name     string
	id       string // domain name; satisfies dnsprovider.Zone.ID()
	domainID int    // Linode integer domain ID, cached after first lookup
	client   *linodego.Client
}

// getDomainID returns the integer Linode domain ID, fetching from the API if not already cached.
func (z *zone) getDomainID(ctx context.Context) (int, error) {
	if z.domainID != 0 {
		return z.domainID, nil
	}
	domains, err := z.client.ListDomains(ctx, &linodego.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("error listing Linode (Akamai) domains: %w", err)
	}
	for _, domain := range domains {
		if domain.Domain == z.name {
			z.domainID = domain.ID
			return z.domainID, nil
		}
	}
	return 0, fmt.Errorf("domain %s not found", z.name)
}

func (z *zone) Name() string {
	return z.name
}

func (z *zone) ID() string {
	return z.id
}

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

func (r *resourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	domainID, err := r.zone.getDomainID(ctx)
	if err != nil {
		return nil, err
	}

	records, err := r.client.ListDomainRecords(ctx, domainID, &linodego.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) domain records for %s: %w", r.zone.Name(), err)
	}

	// Group records by name+type to coalesce multi-value sets
	rrsetMap := make(map[string]*resourceRecordSet)
	for _, record := range records {
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

func (r *resourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &resourceRecordChangeset{
		client: r.client,
		zone:   r.zone,
		rrsets: r,
	}
}

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

func (r *resourceRecordSet) Name() string {
	return r.name
}

func (r *resourceRecordSet) Rrdatas() []string {
	return r.data
}

func (r *resourceRecordSet) Ttl() int64 {
	return r.ttl
}

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

func (r *resourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return r.rrsets
}

func (r *resourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.additions = append(r.additions, rrset)
	return r
}

func (r *resourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.removals = append(r.removals, rrset)
	return r
}

func (r *resourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.upserts = append(r.upserts, rrset)
	return r
}

func (r *resourceRecordChangeset) Apply(ctx context.Context) error {
	if r.IsEmpty() {
		klog.V(4).Info("record change set is empty")
		return nil
	}

	klog.V(2).Info("applying changes in record change set")

	domainID, err := r.zone.getDomainID(ctx)
	if err != nil {
		return err
	}

	// Fetch existing records once for all deletions (removals + upsert pre-delete).
	var existingRecords []linodego.DomainRecord
	if len(r.removals) > 0 || len(r.upserts) > 0 {
		existingRecords, err = r.client.ListDomainRecords(ctx, domainID, &linodego.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing domain records: %w", err)
		}
	}

	for _, rrset := range r.removals {
		if err := r.deleteRecord(ctx, domainID, rrset, existingRecords); err != nil {
			return fmt.Errorf("error removing record %s: %w", rrset.Name(), err)
		}
	}
	if len(r.removals) > 0 {
		klog.V(2).Info("record change set removals complete")
	}

	for _, rrset := range r.additions {
		if err := r.createRecord(ctx, domainID, rrset); err != nil {
			return fmt.Errorf("error adding record %s: %w", rrset.Name(), err)
		}
	}
	if len(r.additions) > 0 {
		klog.V(2).Info("record change set additions complete")
	}

	for _, rrset := range r.upserts {
		if err := r.deleteRecord(ctx, domainID, rrset, existingRecords); err != nil {
			klog.V(2).Infof("error deleting existing record %s before upsert (may not exist): %v", rrset.Name(), err)
		}
		if err := r.createRecord(ctx, domainID, rrset); err != nil {
			return fmt.Errorf("error upserting record %s: %w", rrset.Name(), err)
		}
	}
	if len(r.upserts) > 0 {
		klog.V(2).Info("record change set upserts complete")
	}

	return nil
}

func (r *resourceRecordChangeset) IsEmpty() bool {
	return len(r.additions) == 0 && len(r.removals) == 0 && len(r.upserts) == 0
}

func (r *resourceRecordChangeset) createRecord(ctx context.Context, domainID int, rrset dnsprovider.ResourceRecordSet) error {
	recordName := strings.TrimSuffix(rrset.Name(), ".")
	recordName = strings.TrimSuffix(recordName, "."+r.zone.Name())

	for _, data := range rrset.Rrdatas() {
		_, err := r.client.CreateDomainRecord(ctx, domainID, linodego.DomainRecordCreateOptions{
			Type:   linodego.DomainRecordType(rrset.Type()),
			Name:   recordName,
			Target: data,
			TTLSec: int(rrset.Ttl()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteRecord deletes all records matching the rrset's name and type from the pre-fetched existingRecords list.
func (r *resourceRecordChangeset) deleteRecord(ctx context.Context, domainID int, rrset dnsprovider.ResourceRecordSet, existingRecords []linodego.DomainRecord) error {
	recordName := strings.TrimSuffix(rrset.Name(), ".")
	recordName = strings.TrimSuffix(recordName, "."+r.zone.Name())

	for _, record := range existingRecords {
		if record.Name == recordName && string(record.Type) == string(rrset.Type()) {
			if err := r.client.DeleteDomainRecord(ctx, domainID, record.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
