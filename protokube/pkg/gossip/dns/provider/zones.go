package provider

import (
	"fmt"
	"k8s.io/kops/protokube/pkg/gossip/dns"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
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
