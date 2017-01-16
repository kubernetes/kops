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
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"sync"
	"time"
)

// dnsCache is a wrapper around the DNS provider, adding some caching
type dnsCache struct {
	// zones is the DNS provider
	zonesProvider dnsprovider.Zones

	// mutex protects the following mutable state
	mutex sync.Mutex

	cachedZones          []dnsprovider.Zone
	cachedZonesTimestamp int64
}

func NewDNSCache(provider dnsprovider.Interface) (*dnsCache, error) {
	zonesProvider, ok := provider.Zones()
	if !ok {
		return nil, fmt.Errorf("DNS provider does not support zones")
	}

	return &dnsCache{
		zonesProvider: zonesProvider,
	}, nil
}

// nanoTime is a stand-in until we get a monotonic clock
func nanoTime() int64 {
	return time.Now().UnixNano()
}

// ListZones returns the cached list of zones, as long as it is no older than validity
// This is not a cheap call with a large number of hosted zones, hence the caching
func (d *dnsCache) ListZones(validity time.Duration) ([]dnsprovider.Zone, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := nanoTime()

	if d.cachedZones != nil {
		if (d.cachedZonesTimestamp + validity.Nanoseconds()) > now {
			return d.cachedZones, nil
		} else {
			glog.V(2).Infof("Listing all DNS zones (cache expired)")
		}
	} else {
		glog.V(2).Infof("Listing all DNS zones (no cached results)")
	}

	zones, err := d.zonesProvider.List()
	if err != nil {
		return nil, fmt.Errorf("error querying for DNS zones: %v", err)
	}

	d.cachedZones = zones
	d.cachedZonesTimestamp = now

	return zones, nil
}
