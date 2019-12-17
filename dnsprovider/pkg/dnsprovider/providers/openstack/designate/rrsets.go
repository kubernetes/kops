/*
Copyright 2019 The Kubernetes Authors.

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

package designate

import (
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/pagination"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

var _ dnsprovider.ResourceRecordSets = ResourceRecordSets{}

type ResourceRecordSets struct {
	zone *Zone
}

func (rrsets ResourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	var list []dnsprovider.ResourceRecordSet
	err := recordsets.ListByZone(rrsets.zone.zones.iface.sc, rrsets.zone.impl.ID, nil).EachPage(func(page pagination.Page) (bool, error) {
		rrs, err := recordsets.ExtractRecordSets(page)
		if err != nil {
			return false, err
		}
		for _, rr := range rrs {
			list = append(list, &ResourceRecordSet{&rr, &rrsets})
		}
		return true, nil
	})

	return list, err
}

func (rrsets ResourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	// This list implementation is very similar to the one implemented in
	// the List() method above, but it restricts the retrieved list to
	// the records whose name match the given `name`.
	opts := &recordsets.ListOpts{
		Name: name,
	}

	var list []dnsprovider.ResourceRecordSet
	err := recordsets.ListByZone(rrsets.zone.zones.iface.sc, rrsets.zone.impl.ID, opts).EachPage(func(page pagination.Page) (bool, error) {
		rrs, err := recordsets.ExtractRecordSets(page)
		if err != nil {
			return false, err
		}
		for _, rr := range rrs {
			list = append(list, &ResourceRecordSet{&rr, &rrsets})
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (rrsets ResourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &ResourceRecordChangeset{
		zone:   rrsets.zone,
		rrsets: &rrsets,
	}
}

func (rrsets ResourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	rrs := &recordsets.RecordSet{
		Name: name,
		Type: string(rrstype),
		TTL:  int(ttl),
	}
	for _, rrdata := range rrdatas {
		rrs.Records = append(rrs.Records, string(rrdata))
	}

	return ResourceRecordSet{
		rrs,
		&rrsets,
	}
}

// Zone returns the parent zone
func (rrset ResourceRecordSets) Zone() dnsprovider.Zone {
	return rrset.zone
}
