/*
Copyright 2017 The Kubernetes Authors.

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

package provider

import (
	"fmt"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"k8s.io/kops/protokube/pkg/gossip/dns"
)

type zone struct {
	dnsView  *dns.DNSView
	zoneInfo dns.DNSZoneInfo
}

var _ dnsprovider.Zone = &zone{}

// Name returns the name of the zone, e.g. "example.com"
func (z *zone) Name() string {
	return z.zoneInfo.Name
}

// ID returns the unique provider identifier for the zone
func (z *zone) ID() string {
	// gossip does not allow multiple zones with the same name
	return "gossip:" + z.zoneInfo.Name
}

// ResourceRecordsets returns the provider's ResourceRecordSets interface, or false if not supported.
func (z *zone) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return &resourceRecordSets{
		zone: z,
	}, true
}

func keyForDNSRecord(rrs *dns.DNSRecord) string {
	return string(rrs.RrsType) + "::" + rrs.Name
}

func (z *zone) applyChangeset(c *resourceRecordChangeset) error {
	snapshot := z.dnsView.Snapshot()

	existingRecords := make(map[string]*dns.DNSRecord)
	{
		records := snapshot.RecordsForZone(z.zoneInfo)
		for i := range records {
			k := keyForDNSRecord(&records[i])
			existingRecords[k] = &records[i]
		}
	}

	remove := make(map[string]*resourceRecordSet)
	for _, r := range c.remove {
		rrs, ok := r.(*resourceRecordSet)
		if !ok {
			return fmt.Errorf("unexpected type for ResourceRecordSet: %T", r)
		}

		k := keyForDNSRecord(&rrs.data)
		existing := existingRecords[k]
		if existing == nil {
			return fmt.Errorf("resource record set not found: %v", rrs)
		}
		if remove[k] != nil {
			return fmt.Errorf("resource record deleted twice: %v", rrs)
		}
		remove[k] = rrs
	}

	add := make(map[string]*resourceRecordSet)
	for _, r := range c.add {
		rrs, ok := r.(*resourceRecordSet)
		if !ok {
			return fmt.Errorf("unexpected type for ResourceRecordSet: %T", r)
		}
		k := keyForDNSRecord(&rrs.data)
		existing := existingRecords[k]
		if existing != nil && remove[k] == nil {
			return fmt.Errorf("resource record set already exists: %v", rrs)
		}
		if add[k] != nil {
			return fmt.Errorf("resource record added twice: %v", rrs)
		}
		add[k] = rrs
	}

	for _, r := range c.upsert {
		rrs, ok := r.(*resourceRecordSet)
		if !ok {
			return fmt.Errorf("unexpected type for ResourceRecordSet: %T", r)
		}

		k := keyForDNSRecord(&rrs.data)
		// TODO: Check existing?  Probably not...
		add[k] = rrs
	}

	var addRecords []*dns.DNSRecord
	var removeRecords []*dns.DNSRecord

	newRecords := make(map[string]*dns.DNSRecord)
	for k, v := range existingRecords {
		newRecords[k] = v
	}

	for _, rrs := range remove {
		k := keyForDNSRecord(&rrs.data)
		existing := newRecords[k]
		if existing == nil {
			return fmt.Errorf("resource record set not found: %v", rrs)
		}
		delete(newRecords, k)
		removeRecords = append(removeRecords, &rrs.data)
	}
	for _, rrs := range add {
		k := keyForDNSRecord(&rrs.data)
		existing := newRecords[k]
		if existing != nil {
			return fmt.Errorf("resource record already exists: %v", rrs)
		}
		newRecords[k] = &rrs.data
		addRecords = append(addRecords, &rrs.data)
	}

	return z.dnsView.ApplyChangeset(z.zoneInfo, removeRecords, addRecords)
}

type resourceRecordSets struct {
	zone *zone
}

var _ dnsprovider.ResourceRecordSets = &resourceRecordSets{}

// List returns the ResourceRecordSets of the Zone, or an error if the list operation failed.
func (r *resourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	snapshot := r.zone.dnsView.Snapshot()

	records := snapshot.RecordsForZone(r.zone.zoneInfo)
	if records == nil {
		return nil, nil
	}

	var rrs []dnsprovider.ResourceRecordSet
	for _, rr := range records {
		rrs = append(rrs, &resourceRecordSet{data: rr})
	}
	return rrs, nil
}

// Get returns the ResourceRecordSet with the name in the Zone. if the named resource record set does not exist, but no error occurred, the returned set, and error, are both nil.
func (r *resourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	snapshot := r.zone.dnsView.Snapshot()

	records := snapshot.RecordsForZoneAndName(r.zone.zoneInfo, name)
	if records == nil {
		return nil, nil
	}

	var rrs []dnsprovider.ResourceRecordSet
	for _, rr := range records {
		rrs = append(rrs, &resourceRecordSet{data: rr})
	}
	return rrs, nil
}

// New allocates a new ResourceRecordSet, which can then be passed to ResourceRecordChangeset Add() or Remove()
// Arguments are as per the ResourceRecordSet interface below.
func (r *resourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	data := dns.DNSRecord{
		Name:    name,
		Rrdatas: rrdatas,
		//Ttl:     int(ttl),
		RrsType: string(rrstype),
	}
	return &resourceRecordSet{data: data}
}

// StartChangeset begins a new batch operation of changes against the Zone
func (r *resourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	changeset := &resourceRecordChangeset{
		zone:               r.zone,
		resourceRecordSets: r,
	}
	return changeset
}

// Zone returns the parent zone
func (r *resourceRecordSets) Zone() dnsprovider.Zone {
	return r.zone
}
