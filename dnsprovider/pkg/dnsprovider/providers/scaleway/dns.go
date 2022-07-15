package dns

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"k8s.io/klog/v2"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"os"

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

		return NewProvider(client, ""), nil //TODO: remplir le nom de domaine
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
	accessToken := os.Getenv("SCW_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("SCW_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)

	scwClient, err := scw.NewClient(scw.WithHTTPClient(oauthClient))
	if err != nil {
		return nil, err
	}
	return scwClient, nil
}

// Interface implements dnsprovider.Interface
type Interface struct {
	client       *scw.Client
	parentDomain string
}

// NewProvider returns an implementation of dnsprovider.Interface
func NewProvider(client *scw.Client, parentDomain string) dnsprovider.Interface {
	return &Interface{client: client, parentDomain: parentDomain}
}

// Zones returns an implementation of dnsprovider.Zones
func (d Interface) Zones() (dnsprovider.Zones, bool) {
	return &zones{
		client:       d.client,
		parentDomain: d.parentDomain,
	}, true
}

// zones is an implementation of dnsprovider.Zones
type zones struct {
	client       *scw.Client
	parentDomain string
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
			name:         domainSummary.Domain,
			parentDomain: z.parentDomain,
			client:       z.client,
		}
		zones = append(zones, newZone)
	}

	return zones, nil
}

// Add adds a new DNS zone
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	domainCreateRequest := &domain.CreateDNSZoneRequest{
		Subdomain: newZone.Name(),
		Domain:    z.parentDomain,
	}

	d, err := createDomain(z.client, domainCreateRequest)
	if err != nil {
		return nil, err
	}

	return &zone{
		name:         d.Subdomain,
		parentDomain: d.Domain,
		client:       z.client,
	}, nil
}

// Remove deletes a zone
func (z *zones) Remove(zone dnsprovider.Zone) error {
	return deleteDomain(z.client, zone.Name()+"."+z.parentDomain)
}

// New returns a new implementation of dnsprovider.Zone
func (z *zones) New(name string) (dnsprovider.Zone, error) {
	return &zone{
		name:         name,
		parentDomain: z.parentDomain,
		client:       z.client,
	}, nil
}

// zone implements dnsprovider.Zone
type zone struct {
	name         string
	client       *scw.Client
	parentDomain string
	//id           string
}

// Name returns the Name of a dns zone
func (z *zone) Name() string {
	return z.name
}

// ID returns the ID of a dns zone
func (z *zone) ID() string {
	//return z.id
	// TODO: shall we use the name as ID ? or handle the zone as a record to be able to get the ID ?
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
		klog.V(4).Info(ctx, "record change set is empty")
		return nil
	}

	klog.V(2).Info(ctx, "applying changes in record change set")

	if len(r.additions) > 0 {
		for _, rrset := range r.additions {
			err := r.applyResourceRecordSet(rrset)
			if err != nil {
				return fmt.Errorf("failed to apply resource record set: %s, err: %s", rrset.Name(), err)
			}
		}

		klog.V(2).Info(ctx, "record change set additions complete")
	}

	if len(r.upserts) > 0 {
		for _, rrset := range r.upserts {
			err := r.applyResourceRecordSet(rrset)
			if err != nil {
				return fmt.Errorf("failed to apply resource record set: %s, err: %s", rrset.Name(), err)
			}
		}

		klog.V(2).Info(ctx, "record change set upserts complete")
	}

	if len(r.removals) > 0 {
		records, err := getRecords(r.client, r.zone.Name())
		if err != nil {
			return err
		}

		for _, record := range r.removals {
			for _, domainRecord := range records {
				if domainRecord.Name == record.Name() {
					err := deleteRecord(r.client, r.zone.Name(), domainRecord.ID)
					if err != nil {
						return fmt.Errorf("failed to delete record: %v", err)
					}
				}
			}
		}

		klog.V(2).Info(ctx, "record change set removals complete")
	}

	klog.V(2).Info(ctx, "record change sets successfully applied")
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

// applyResourceRecordSet will create records of a domain as required by resourceRecordChangeset
// and delete any previously created records matching the same name.
// This is required for scaleway since it's API does not handle record sets, but
// only individual records
func (r *resourceRecordChangeset) applyResourceRecordSet(rrset dnsprovider.ResourceRecordSet) error {
	deleteRecords, err := getRecordsByName(r.client, r.zone.Name(), rrset.Name())
	if err != nil {
		return fmt.Errorf("failed to get record IDs to delete")
	}

	addRecords := []*domain.Record(nil)

	for range rrset.Rrdatas() {
		for _, rrdata := range rrset.Rrdatas() {
			addRecords = append(addRecords, &domain.Record{
				Name: rrset.Name(),
				Data: rrdata,
				TTL:  uint32(rrset.Ttl()),
				Type: domain.RecordType(rrset.Type()),
			})
		}
	}

	recordCreateRequest := &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: r.zone.parentDomain,
		Changes: []*domain.RecordChange{
			{
				Add: &domain.RecordChangeAdd{
					Records: addRecords,
				},
			},
		},
	}

	_, err = createRecord(r.client, recordCreateRequest)
	if err != nil {
		return fmt.Errorf("could not create record: %v", err)
	}

	for _, record := range deleteRecords {
		err = deleteRecord(r.client, r.zone.Name(), record.ID)
		if err != nil {
			return fmt.Errorf("error cleaning up old records: %v", err)
		}
	}

	return nil
}

// listDomains returns a list of scaleway Domain objects
//func listDomains(c *scw.Client) ([]*domain.DNSZone, error) {
func listDomains(c *scw.Client) ([]*domain.DomainSummary, error) {
	registrarApi := domain.NewRegistrarAPI(c)

	domains, err := registrarApi.ListDomains(&domain.RegistrarAPIListDomainsRequest{
		Page:           nil,
		PageSize:       nil,
		OrderBy:        "",
		Registrar:      nil,
		Status:         "",
		ProjectID:      nil,
		OrganizationID: nil,
		IsExternal:     nil,
	})
	//api := domain.NewAPI(c)
	//
	//domains, err := api.ListDNSZones(&domain.ListDNSZonesRequest{
	//	OrganizationID: nil,
	//	ProjectID:      nil,
	//	OrderBy:        "",
	//	Page:           nil,
	//	PageSize:       nil,
	//	Domain:         "",
	//	DNSZone:        "",
	//})

	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %v", err)
	}

	return domains.Domains, err
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

// getRecords returns a list of scaleway given a zone name
func getRecords(c *scw.Client, zoneName string) ([]*domain.Record, error) {
	api := domain.NewAPI(c)

	records, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: zoneName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %v", err)
	}

	return records.Records, err
}

// getRecordsByName returns a list of domain Records based on the provided zone and name
func getRecordsByName(client *scw.Client, zoneName, recordName string) ([]*domain.Record, error) {
	api := domain.NewAPI(client)

	records, err := api.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: zoneName,
		Name:    recordName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %v", err)
	}

	return records.Records, err
}

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
