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
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

var _ dnsprovider.ResourceRecordChangeset = &ResourceRecordChangeset{}

type ResourceRecordChangeset struct {
	zone   *Zone
	rrsets *ResourceRecordSets

	additions []dnsprovider.ResourceRecordSet
	removals  []dnsprovider.ResourceRecordSet
	upserts   []dnsprovider.ResourceRecordSet
}

func (c *ResourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.additions = append(c.additions, rrset)
	return c
}

func (c *ResourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.removals = append(c.removals, rrset)
	return c
}

func (c *ResourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.upserts = append(c.upserts, rrset)
	return c
}

func (c *ResourceRecordChangeset) Apply(ctx context.Context) error {
	// Empty changesets should be a relatively quick no-op
	if c.IsEmpty() {
		return nil
	}

	zoneID := c.zone.impl.ID

	for _, removal := range c.removals {
		rrID, err := c.nameToID(removal.Name())
		if err != nil {
			return err
		}
		err = recordsets.Delete(c.zone.zones.iface.sc, zoneID, rrID).ExtractErr()
		if err != nil {
			return err
		}
	}

	for _, addition := range c.additions {
		opts := recordsets.CreateOpts{
			Name:    addition.Name(),
			TTL:     int(addition.Ttl()),
			Type:    string(addition.Type()),
			Records: addition.Rrdatas(),
		}
		_, err := recordsets.Create(c.zone.zones.iface.sc, zoneID, opts).Extract()
		if err != nil {
			return err
		}
	}

	for _, upsert := range c.upserts {
		rrID, err := c.nameToID(upsert.Name())
		if err != nil {
			return err
		}
		ttl := int(upsert.Ttl())
		uopts := recordsets.UpdateOpts{
			TTL:     &ttl,
			Records: upsert.Rrdatas(),
		}
		_, err = recordsets.Update(c.zone.zones.iface.sc, zoneID, rrID, uopts).Extract()
		if err != nil {
			copts := recordsets.CreateOpts{
				Name:    upsert.Name(),
				TTL:     int(upsert.Ttl()),
				Type:    string(upsert.Type()),
				Records: upsert.Rrdatas(),
			}
			_, err := recordsets.Create(c.zone.zones.iface.sc, zoneID, copts).Extract()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *ResourceRecordChangeset) IsEmpty() bool {
	return len(c.removals) == 0 && len(c.additions) == 0 && len(c.upserts) == 0
}

// ResourceRecordSets returns the parent ResourceRecordSets
func (c *ResourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return c.rrsets
}

func (c *ResourceRecordChangeset) nameToID(name string) (string, error) {
	opts := recordsets.ListOpts{
		Name: name,
	}
	allPages, err := recordsets.ListByZone(c.zone.zones.iface.sc, c.zone.impl.ID, opts).AllPages()
	if err != nil {
		return "", err
	}
	rrs, err := recordsets.ExtractRecordSets(allPages)
	if err != nil {
		return "", err
	}
	switch len(rrs) {
	case 0:
		return "", fmt.Errorf("couldn't find recordset with name: %s, expected 1", name)
	case 1:
		return rrs[0].ID, nil
	default:
		return "", fmt.Errorf("found multiple recordsets with name: %s, expected 1", name)
	}
}
