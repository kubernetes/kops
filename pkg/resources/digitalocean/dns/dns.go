/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/context"

	"github.com/golang/glog"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

const ipPlaceholder = "203.0.113.123"

// DNS implements dnsprovider.Interface
type DNS struct {
	client *godo.Client
}

// NewProvider returns an implementation of dnsprovider.Interface
func NewProvider(client *godo.Client) dnsprovider.Interface {
	return &DNS{client: client}
}

// Zones returns an implementation of dnsprovider.Zones
func (d *DNS) Zones() (dnsprovider.Zones, bool) {
	return &zones{
		client: d.client,
	}, true
}

// zones is an implementation of dnsprovider.Zones
type zones struct {
	client *godo.Client
}

// List returns a list of all dns zones
func (z *zones) List() ([]dnsprovider.Zone, error) {
	domains, err := listDomains(z.client)
	if err != nil {
		return nil, err
	}

	var newZone *zone
	var zones []dnsprovider.Zone
	for _, domain := range domains {
		newZone = &zone{
			name:   domain.Name,
			client: z.client,
		}
		zones = append(zones, newZone)
	}

	return zones, nil
}

// Add adds a new DNS zone
func (z *zones) Add(newZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	domainCreateRequest := &godo.DomainCreateRequest{
		Name:      newZone.Name(),
		IPAddress: ipPlaceholder,
	}

	domain, err := createDomain(z.client, domainCreateRequest)
	if err != nil {
		return nil, err
	}

	return &zone{
		name:   domain.Name,
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
	client *godo.Client
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
	client *godo.Client
}

// List returns a list of dnsprovider.ResourceRecordSet
func (r *resourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	records, err := getRecords(r.client, r.zone.Name())
	if err != nil {
		return nil, err
	}

	var rrset *resourceRecordSet
	var rrsets []dnsprovider.ResourceRecordSet
	for _, record := range records {
		rrset = &resourceRecordSet{
			name:       record.Name,
			data:       record.Data,
			ttl:        record.TTL,
			recordType: rrstype.RrsType(record.Type),
		}

		rrsets = append(rrsets, rrset)
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
	if len(rrdatas) > 1 {
		return nil
	}

	return &resourceRecordSet{
		name:       name,
		data:       rrdatas[0],
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
	data       string
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
	return []string{r.data}
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
	client *godo.Client
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

// Upsert adds a new resource record set to the list of upesrts to apply
func (r *resourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	r.upserts = append(r.upserts, rrset)
	return r
}

// Apply adds new records stored in r.additions, updates records stored
// in r.upserts and deletes records stored in r.removals
func (r *resourceRecordChangeset) Apply() error {
	glog.V(2).Infof("applying changes in record change set")
	if r.IsEmpty() {
		glog.V(2).Infof("record change set is empty")
		return nil
	}

	if len(r.additions) > 0 {
		for _, record := range r.additions {
			recordCreateRequest := &godo.DomainRecordEditRequest{
				Name: record.Name(),
				Data: record.Rrdatas()[0],
				TTL:  int(record.Ttl()),
				Type: string(record.Type()),
			}
			err := createRecord(r.client, r.zone.Name(), recordCreateRequest)
			if err != nil {
				return fmt.Errorf("could not create record: %v", err)
			}
		}

		glog.V(2).Infof("record change set additions complete")
	}

	if len(r.removals) > 0 {
		records, err := getRecords(r.client, r.zone.Name())
		if err != nil {
			return err
		}

		for _, record := range r.removals {
			var desiredRecord godo.DomainRecord
			found := false
			for _, domainRecord := range records {
				if domainRecord.Name == record.Name() {
					desiredRecord = domainRecord
					found = true
				}
			}
			if !found {
				return fmt.Errorf("could not find desired record to remove")
			}

			err := deleteRecord(r.client, r.zone.Name(), desiredRecord.ID)
			if err != nil {
				return err
			}
		}

		glog.V(2).Infof("record change set removals complete")
	}

	if len(r.upserts) > 0 {
		records, err := getRecords(r.client, r.zone.Name())
		if err != nil {
			return err
		}

		for _, record := range r.upserts {
			var desiredRecord godo.DomainRecord
			found := false
			for _, domainRecord := range records {
				if domainRecord.Name == record.Name() {
					desiredRecord = domainRecord
					found = true
				}
			}

			if !found {
				return fmt.Errorf("could not find desired record to upsert")
			}

			domainEditRequest := &godo.DomainRecordEditRequest{
				Name: record.Name(),
				Data: record.Rrdatas()[0],
				TTL:  int(record.Ttl()),
				Type: string(record.Type()),
			}
			err := editRecord(r.client, r.zone.Name(), desiredRecord.ID, domainEditRequest)
			if err != nil {
				return err
			}
		}

		glog.V(2).Infof("record change set upserts complete")
	}

	glog.V(2).Infof("record change sets successfully applied")
	return nil
}

// IsEmpty returns true if a changeset is empty, false otherwise
func (r *resourceRecordChangeset) IsEmpty() bool {
	if len(r.additions) == 0 && len(r.removals) == 0 && len(r.upserts) == 0 {
		return true
	}

	return false
}

// ResourceRecordSet returns the associated resourceRecordSets of a changset
func (r *resourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return r.rrsets
}

// listDomains returns a list of godo.Domain
func listDomains(c *godo.Client) ([]godo.Domain, error) {
	// TODO (andrewsykim): pagination in ListOptions
	domains, resp, err := c.Domains.List(context.TODO(), &godo.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %v", err)
	}

	if err = handleResponse(resp); err != nil {
		return nil, err
	}

	return domains, err
}

// createDomain creates a domain provided godo.DomainCreateRequest
func createDomain(c *godo.Client, createRequest *godo.DomainCreateRequest) (*godo.Domain, error) {
	domain, resp, err := c.Domains.Create(context.TODO(), createRequest)
	if err != nil {
		return nil, err
	}

	if err = handleResponse(resp); err != nil {
		return nil, err
	}

	return domain, nil
}

// deleteDomain deletes a domain given its name
func deleteDomain(c *godo.Client, name string) error {
	resp, err := c.Domains.Delete(context.TODO(), name)
	if err != nil {
		return err
	}

	if err = handleResponse(resp); err != nil {
		return err
	}

	return nil
}

// getRecords returns a list of godo.DomainRecord given a zone name
func getRecords(c *godo.Client, zoneName string) ([]godo.DomainRecord, error) {
	records, resp, err := c.Domains.Records(context.TODO(), zoneName, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	if err = handleResponse(resp); err != nil {
		return nil, err
	}

	return records, nil
}

// createRecord creates a record given an assoicated zone and a godo.DomainRecordEditRequest
func createRecord(c *godo.Client, zoneName string, createRequest *godo.DomainRecordEditRequest) error {
	_, resp, err := c.Domains.CreateRecord(context.TODO(), zoneName, createRequest)
	if err != nil {
		return fmt.Errorf("error applying changeset: %v", err)
	}

	if err = handleResponse(resp); err != nil {
		return err
	}

	return nil
}

// editRecord edits a record given an associated ozone and a godo.DomainRecordEditRequest
func editRecord(c *godo.Client, zoneName string, recordID int, editRequest *godo.DomainRecordEditRequest) error {
	_, resp, err := c.Domains.EditRecord(context.TODO(), zoneName, recordID, editRequest)
	if err != nil {
		return fmt.Errorf("error applying changeset: %v", err)
	}

	if err = handleResponse(resp); err != nil {
		return err
	}

	return nil
}

// deleteRecord deletes a record given an associated zone and a record ID
func deleteRecord(c *godo.Client, zoneName string, recordID int) error {
	resp, err := c.Domains.DeleteRecord(context.TODO(), zoneName, recordID)
	if err != nil {
		return fmt.Errorf("error applying changeset: %v", err)
	}

	if err = handleResponse(resp); err != nil {
		return err
	}

	return nil
}

func handleResponse(resp *godo.Response) error {
	if resp.StatusCode != http.StatusOK {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}

		return fmt.Errorf("received non 200 status code: %d from api: %v",
			resp.StatusCode, string(respData))
	}

	return nil
}
