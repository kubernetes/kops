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

// DNS implements dnsprovider.Interface
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
			name:   domainSummary.Domain, // TODO: check if .Domain == .Name
			client: z.client,
		}
		zones = append(zones, newZone)
	}

	return zones, nil
}

// Add adds a new DNS zone
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	domainCreateRequest := &domain.CreateDNSZoneRequest{
		Domain: newZone.Name(),
	}

	d, err := createDomain(z.client, domainCreateRequest)
	if err != nil {
		return nil, err
	}

	return &zone{
		name:   d.Domain,
		client: z.client,
	}, nil
}

// Remove deletes a zone
func (z *zones) Remove(zone dnsprovider.Zone) error {
	return deleteDomain(z.client, zone.Name())
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

// ID returns the name of a dns zone, in DO the ID is the name
func (z *zone) ID() string {
	return z.name
}

// ResourceRecordSet returns an implementation of dnsprovider.ResourceRecordSets
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
		// digitalocean API returns the record without the zone
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
// in DO this is almost always the IP of a record
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
					err := deleteDomain(r.client, r.zone.Name())
					//err := deleteRecord(r.client, r.zone.Name(), domainRecord.ID)
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

// ResourceRecordSet returns the associated resourceRecordSets of a changeset
func (r *resourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return r.rrsets
}

// applyResourceRecordSet will create records of a domain as required by resourceRecordChangeset
// and delete any previously created records matching the same name.
// This is required for digitalocean since it's API does not handle record sets, but
// only individual records
func (r *resourceRecordChangeset) applyResourceRecordSet(rrset dnsprovider.ResourceRecordSet) error {
	deleteRecords, err := getRecordsByName(r.client, r.zone.Name(), rrset.Name())
	if err != nil {
		return fmt.Errorf("failed to get record IDs to delete")
	}

	for range rrset.Rrdatas() {
		//for _, rrdata := range rrset.Rrdatas() {
		recordCreateRequest := &domain.CreateDNSZoneRequest{
			Subdomain: rrset.Name(),
			Domain:    r.zone.Name(),
			//Name: rrset.Name(),
			//Data: rrdata,
			//TTL:  int(rrset.Ttl()),
			//Type: string(rrset.Type()),
		}
		err := createRecord(r.client, recordCreateRequest)
		if err != nil {
			return fmt.Errorf("could not create record: %v", err)
		}
	}

	for range deleteRecords {
		err := deleteDomain(r.client, r.zone.Name())
		//err := deleteRecord(r.client, r.zone.Name(), record.ID)
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

	//domains, _, err := c.Domains.List(context.TODO(), &scw.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %v", err)
	}

	return domains.Domains, err
}

// createDomain creates a domain provided scw.DomainCreateRequest
func createDomain(c *scw.Client, createRequest *domain.CreateDNSZoneRequest) (*domain.DNSZone, error) {
	api := domain.NewAPI(c)

	dnsZone, err := api.CreateDNSZone(createRequest)
	//domain, _, err := c.Domains.Create(context.TODO(), createRequest)
	if err != nil {
		return nil, err
	}

	return dnsZone, nil
}

// deleteDomain deletes a domain given its name
func deleteDomain(c *scw.Client, name string) error {
	//_, err := c.Domains.Delete(context.TODO(), name)
	api := domain.NewAPI(c)

	_, err := api.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
		DNSZone: name,
	})
	if err != nil {
		return err
	}

	return nil
}

// getRecords returns a list of scaleway DomainRecord objects given a zone name
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

// getRecordsByName returns a list of godo.DomainRecord based on the provided zone and name
func getRecordsByName(client *scw.Client, zoneName, recordName string) ([]*domain.Record, error) {
	//records, err := getRecords(client, zoneName)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// digitalocean record.Name returns record without the zone suffix
	//// so normalize record by removing it
	//normalizedRecordName := strings.TrimSuffix(recordName, ".")
	//normalizedRecordName = strings.TrimSuffix(normalizedRecordName, "."+zoneName)
	//
	//var recordsByName []godo.DomainRecord
	//for _, record := range records {
	//	if record.Name == normalizedRecordName {
	//		recordsByName = append(recordsByName, record)
	//	}
	//}
	//
	//return recordsByName, nil
	//TODO: check if this is equivalent to the above
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

// createRecord creates a record given an associated zone and a godo.DomainRecordEditRequest
func createRecord(c *scw.Client, createRequest *domain.CreateDNSZoneRequest) error {
	api := domain.NewAPI(c)

	_, err := api.CreateDNSZone(createRequest)
	//_, _, err := c.Domains.CreateRecord(context.TODO(), zoneName, createRequest)
	if err != nil {
		return fmt.Errorf("error creating record: %v", err)
	}

	return nil
}

// deleteRecord deletes a record given an associated zone and a record ID
//func deleteRecord(c *scw.Client, zoneName string, recordID int) error {
//	api := domain.NewAPI(c)
//
//	//_, err := c.Domains.DeleteRecord(context.TODO(), zoneName, recordID)
//	_, err := api.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
//		DNSZone: zoneName,
//	})
//	if err != nil {
//		return fmt.Errorf("error deleting record: %v", err)
//	}
//
//	return nil
//}
