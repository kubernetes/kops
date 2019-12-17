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

package dns

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

// dnsCache is a wrapper around the DNS provider, adding some caching
type dnsCache struct {
	// zonesProviders is a slice of configured DNS providers
	zonesProviders []dnsprovider.Zones

	// mutex protects the following mutable state
	mutex sync.Mutex

	cachedZones          []dnsprovider.Zone
	cachedZonesTimestamp int64
}

func newDNSCache(providers []dnsprovider.Interface) (*dnsCache, error) {
	var zonesProviders []dnsprovider.Zones
	for _, provider := range providers {
		zonesProvider, ok := provider.Zones()
		if !ok {
			return nil, fmt.Errorf("DNS provider does not support zones")
		}
		zonesProviders = append(zonesProviders, zonesProvider)
	}

	return &dnsCache{
		zonesProviders: zonesProviders,
	}, nil
}

// nanoTime is a stand-in until we get a monotonic clock
func nanoTime() int64 {
	return time.Now().UnixNano()
}

// ListZones returns the zones, using a cached copy if validity has not yet expired.
// This is not a cheap call with a large number of hosted zones, hence the caching.
func (d *dnsCache) ListZones(validity time.Duration) ([]dnsprovider.Zone, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := nanoTime()

	if d.cachedZones != nil {
		if (d.cachedZonesTimestamp + validity.Nanoseconds()) > now {
			return d.cachedZones, nil
		}
		klog.V(2).Infof("querying all DNS zones (cache expired)")
	} else {
		klog.V(2).Infof("querying all DNS zones (no cached results)")
	}

	var allZones []dnsprovider.Zone
	for _, zonesProvider := range d.zonesProviders {
		zones, err := zonesProvider.List()
		if err != nil {
			return nil, fmt.Errorf("error querying for DNS zones: %v", err)
		}

		allZones = append(allZones, zones...)
	}
	d.cachedZones = allZones
	d.cachedZonesTimestamp = now

	return allZones, nil
}
