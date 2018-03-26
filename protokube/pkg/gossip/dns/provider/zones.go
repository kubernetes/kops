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
	"k8s.io/kops/protokube/pkg/gossip/dns"
)

type zones struct {
	dnsView *dns.DNSView
}

var _ dnsprovider.Zones = &zones{}

// List returns the managed Zones, or an error if the list operation failed.
func (z *zones) List() ([]dnsprovider.Zone, error) {
	snapshot := z.dnsView.Snapshot()

	var zones []dnsprovider.Zone
	zoneInfos := snapshot.ListZones()
	for i := range zoneInfos {
		zones = append(zones, &zone{dnsView: z.dnsView, zoneInfo: zoneInfos[i]})
	}
	return zones, nil
}

// Add creates and returns a new managed zone, or an error if the operation failed
func (z *zones) Add(addZone dnsprovider.Zone) (dnsprovider.Zone, error) {
	zoneToAdd, ok := addZone.(*zone)
	if !ok {
		return nil, fmt.Errorf("unexpected zone type: %T", addZone)
	}

	zoneInfo, err := z.dnsView.AddZone(zoneToAdd.zoneInfo)
	if err != nil {
		return nil, err
	}
	return &zone{dnsView: z.dnsView, zoneInfo: *zoneInfo}, nil
}

// Remove deletes a managed zone, or returns an error if the operation failed.
func (z *zones) Remove(removeZone dnsprovider.Zone) error {
	zone, ok := removeZone.(*zone)
	if !ok {
		return fmt.Errorf("unexpected zone type: %T", removeZone)
	}

	return z.dnsView.RemoveZone(zone.zoneInfo)
}

// New allocates a new Zone, which can then be passed to Add()
// Arguments are as per the Zone interface below.
func (z *zones) New(name string) (dnsprovider.Zone, error) {
	a := &zone{
		dnsView: z.dnsView,
		zoneInfo: dns.DNSZoneInfo{
			Name: name,
		},
	}

	return a, nil
}
