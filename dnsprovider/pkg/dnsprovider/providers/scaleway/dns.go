package dns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"

	"github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
		return nil, errors.New("SCW_ACCESS_KEY is required")
	}
	if secretKey := os.Getenv("SCW_SECRET_KEY"); secretKey == "" {
		return nil, errors.New("SCW_SECRET_KEY is required")
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
	domains, err := listDomains(z.client)
	if err != nil {
		return nil, err
	}

	var newZone *zone
	var zones []dnsprovider.Zone
	for _, domainSummary := range domains {
		newZone = &zone{
			name:   domainSummary.Domain,
			client: z.client,
		}
		zones = append(zones, newZone)
	}

	return zones, nil
}

// Add adds a new DNS zone
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	domainCreateRequest := &domain.CreateDNSZoneRequest{
		Subdomain: newZone.Name(),
		Domain:    os.Getenv("SCW_DNS_ZONE"),
	}

	klog.V(8).Infof("Adding new DNS zone %s to domain %s", newZone.Name(), os.Getenv("SCW_DNS_ZONE"))
	d, err := createDomain(z.client, domainCreateRequest)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Added new DNS zone %s to domain %s", d.Subdomain, d.Domain)

	return &zone{
		name:   d.Subdomain,
		client: z.client,
	}, nil
}

// Remove deletes a zone
func (z *zones) Remove(zone dnsprovider.Zone) error {
	return deleteDomain(z.client, zone.Name()+"."+os.Getenv("SCW_DNS_ZONE"))
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
	//id           string
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
	records, err := getRecords(r.client, r.zone.Name())
	if err != nil {
		return nil, err
	}

	var rrsets []dnsprovider.ResourceRecordSet
	rrsetsWithoutDups := make(map[string]*resourceRecordSet)

	for _, record := range records {
		// The scaleway API returns the record without the zone
		// but the consumers of this interface expect the zone to be included
		recordName := dns.EnsureDotSuffix(record.Name) + r.Zone().Name()
		if set, ok := rrsetsWithoutDups[recordName]; !ok {
			rrsetsWithoutDups[recordName] = &resourceRecordSet{
				name:       recordName,
				data:       []string{record.Data},
				ttl:        int(record.TTL),
				recordType: rrstype.RrsType(record.Type),
			}
		} else {
			set.data = append(set.data, record.Data)
		}
	}

	for _, set := range rrsetsWithoutDups {
		rrsets = append(rrsets, set)
	}

	return rrsets, nil
}

// Get returns a list of dnsprovider.ResourceRecordSet that matches the name parameter
func (r *resourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	records, err := r.List()
	if err != nil {
		return nil, err
	}

	var recordSets []dnsprovider.ResourceRecordSet
	for _, record := range records {
		if record.Name() == name {
			recordSets = append(recordSets, record)
		}
	}

	return recordSets, nil
}

// New returns an implementation of dnsprovider.ResourceRecordSet
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

// Apply adds new records stored in r.additions, updates records stored
// in r.upserts and deletes records stored in r.removals
func (r *resourceRecordChangeset) Apply(ctx context.Context) error {
	// Empty changesets should be a relatively quick no-op
	if r.IsEmpty() {
		klog.V(4).Info("record change set is empty")
		return nil
	}

	klog.V(2).Info("applying changes in record change set")
	updateRecordsRequest := []*domain.RecordChange(nil)
	dnsZone := os.Getenv("SCW_DNS_ZONE")
	api := domain.NewAPI(r.client)

	records, err := getRecords(r.client, r.zone.Name())
	if err != nil {
		return err
	}

	if len(r.additions) > 0 {
		recordsToAdd := []*domain.Record(nil)
		for _, rrset := range r.additions {
			for _, rrdata := range rrset.Rrdatas() {
				recordsToAdd = append(recordsToAdd, &domain.Record{
					Name: rrset.Name(),
					Data: rrdata,
					TTL:  uint32(rrset.Ttl()),
					Type: domain.RecordType(rrset.Type()),
				})
			}
			klog.V(8).Infof("adding new DNS record %s to zone %s", rrset.Name(), r.zone.name)
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
					if record.Name == rrset.Name() {
						klog.V(8).Infof("changing DNS record %s of zone %s", rrset.Name(), r.zone.name)
						updateRecordsRequest = append(updateRecordsRequest, &domain.RecordChange{
							Set: &domain.RecordChangeSet{
								ID: &record.ID,
								Records: []*domain.Record{
									{
										Name: rrset.Name(),
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
				if record.Name == rrset.Name() && record.Data == rrset.Rrdatas()[0] {
					klog.V(8).Infof("removing DNS record %s of zone %s", rrset.Name(), r.zone.name)
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
		DNSZone: dnsZone,
		Changes: updateRecordsRequest,
	})
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

// listDomains returns a list of scaleway Domain objects
func listDomains(c *scw.Client) ([]*domain.DNSZone, error) {
	api := domain.NewAPI(c)

	domains, err := api.ListDNSZones(&domain.ListDNSZonesRequest{
		//Domain:         "",
		//DNSZone:        "",
	}, scw.WithAllPages())

	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %v", err)
	}

	return domains.DNSZones, err
}

// createDomain creates a domain provided scw.DomainCreateRequest
func createDomain(c *scw.Client, createRequest *domain.CreateDNSZoneRequest) (*domain.DNSZone, error) {
	api := domain.NewAPI(c)

	dnsZone, err := api.CreateDNSZone(createRequest)

	if err != nil {
		return nil, err
	}

	return dnsZone, nil
}

// deleteDomain deletes a domain given its name
func deleteDomain(c *scw.Client, name string) error {
	api := domain.NewAPI(c)

	_, err := api.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
		DNSZone: name,
	})
	if err != nil {
		return err
	}

	return nil
}

// getRecords returns a list of scaleway records given a zone name (the name of the record doesn't end with the zone name)
func getRecords(c *scw.Client, zoneName string) ([]*domain.Record, error) {
	api := domain.NewAPI(c)

	records, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: zoneName,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %v", err)
	}

	return records.Records, err
}

//// getRecordsByName returns a list of domain Records based on the provided zone and name
//func getRecordsByName(client *scw.Client, zoneName, recordName string) ([]*domain.Record, error) {
//	api := domain.NewAPI(client)
//
//	records, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
//		DNSZone: zoneName,
//		Name:    recordName,
//	}, scw.WithAllPages())
//	if err != nil {
//		return nil, fmt.Errorf("failed to list records: %v", err)
//	}
//
//	return records.Records, err
//}

// createRecord creates a record given an associated zone and an UpdateDNSZoneRecordsRequest
func createRecord(c *scw.Client, recordsCreateRequest *domain.UpdateDNSZoneRecordsRequest) ([]string, error) {
	api := domain.NewAPI(c)

	resp, err := api.UpdateDNSZoneRecords(recordsCreateRequest)
	if err != nil {
		return nil, fmt.Errorf("error creating record: %v", err)
	}

	recordsIds := []string(nil)
	for _, record := range resp.Records {
		recordsIds = append(recordsIds, record.ID)
	}

	return recordsIds, nil
}

// deleteRecord deletes a record given an associated zone and a record ID
func deleteRecord(c *scw.Client, zoneName string, recordID string) error {
	api := domain.NewAPI(c)

	recordDeleteRequest := &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: zoneName,
		Changes: []*domain.RecordChange{
			{
				Delete: &domain.RecordChangeDelete{
					ID: &recordID,
				},
			},
		},
	}

	_, err := api.UpdateDNSZoneRecords(recordDeleteRequest)
	if err != nil {
		return fmt.Errorf("error deleting record: %v", err)
	}

	return nil
}
