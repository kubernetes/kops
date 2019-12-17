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

package dns

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/klog"
	"k8s.io/kops/protokube/pkg/gossip"
)

// We don't really support multiple zone ids, but we could
// Also, not supporting multiple zone ids makes implementing dnsprovider painful
const DefaultZoneName = "local"

type DNSTarget interface {
	Update(snapshot *DNSViewSnapshot) error
}

func RunDNSUpdates(target DNSTarget, src *DNSView) {
	var lastSnapshot *DNSViewSnapshot
	for {
		time.Sleep(5 * time.Second)

		// Consider replace with a watch?  But debouncing is also nice

		// Snapshot is very cheap if we are in-sync
		snapshot := src.Snapshot()
		if lastSnapshot != nil && lastSnapshot.version == snapshot.version {
			klog.V(4).Infof("DNSView unchanged: %v", lastSnapshot.version)
			continue
		}

		// TODO: We might want to keep old records alive for a bit

		klog.V(2).Infof("DNSView changed: %v", snapshot.version)

		err := target.Update(snapshot)
		if err != nil {
			klog.Warningf("error applying DNS changes to target: %v", err)
			continue
		}

		lastSnapshot = snapshot
	}

}

type DNSView struct {
	gossipState gossip.GossipState

	mutex        sync.Mutex
	lastSnapshot *DNSViewSnapshot
}

type DNSViewSnapshot struct {
	version uint64
	zoneMap map[string]*dnsViewSnapshotZone
}

type dnsViewSnapshotZone struct {
	Name    string
	Records map[string]DNSRecord
}

// RecordsForZone returns records matching the specified zone
func (s *DNSViewSnapshot) RecordsForZone(zoneInfo DNSZoneInfo) []DNSRecord {
	var records []DNSRecord

	zone := s.zoneMap[zoneInfo.Name]
	if zone != nil {
		for k := range zone.Records {
			records = append(records, zone.Records[k])
		}
	}

	return records
}

// RecordsForZoneAndName returns records matching the specified zone and name
func (s *DNSViewSnapshot) RecordsForZoneAndName(zoneInfo DNSZoneInfo, name string) []DNSRecord {
	var records []DNSRecord

	zone := s.zoneMap[zoneInfo.Name]
	if zone != nil {
		for k := range zone.Records {
			if zone.Records[k].Name != name {
				continue
			}

			records = append(records, zone.Records[k])
		}
	}

	return records
}

// ListZones returns all zones
func (s *DNSViewSnapshot) ListZones() []DNSZoneInfo {
	var zones []DNSZoneInfo
	for _, z := range s.zoneMap {
		zones = append(zones, DNSZoneInfo{
			Name: z.Name,
		})
	}
	return zones
}

// RemoveZone removes the specified zone, though this is currently not supported and returns an error.
func (v *DNSView) RemoveZone(info DNSZoneInfo) error {
	return fmt.Errorf("zone deletion is implicit")
}

// AddZone adds the specified zone; this creates a fake NS record just so that the zone has records
func (v *DNSView) AddZone(info DNSZoneInfo) (*DNSZoneInfo, error) {
	createRecords := []*DNSRecord{
		{
			RrsType: "NS",
			Name:    info.Name,
			Rrdatas: []string{"gossip"},
		},
	}

	err := v.ApplyChangeset(info, nil, createRecords)
	if err != nil {
		return nil, err
	}

	return &DNSZoneInfo{
		Name: info.Name,
	}, nil
}

// ApplyChangeset applies a DNS changeset to the records.
func (v *DNSView) ApplyChangeset(zone DNSZoneInfo, removeRecords []*DNSRecord, createRecords []*DNSRecord) error {
	var removeTags []string
	for _, record := range removeRecords {
		tagKey, err := buildTagKey(zone, record)
		if err != nil {
			return err
		}
		removeTags = append(removeTags, tagKey)
	}

	createTags := make(map[string]string)
	for _, record := range createRecords {
		tagKey, err := buildTagKey(zone, record)
		if err != nil {
			return err
		}
		if createTags[tagKey] != "" {
			return fmt.Errorf("duplicate record %q being created", tagKey)
		}
		createTags[tagKey] = strings.Join(record.Rrdatas, ",")
	}

	return v.gossipState.UpdateValues(removeTags, createTags)
}

func buildTagKey(zone DNSZoneInfo, record *DNSRecord) (string, error) {
	fqdn := strings.TrimSuffix(record.Name, ".")
	if fqdn != zone.Name && !strings.HasSuffix(fqdn, "."+zone.Name) {
		return "", fmt.Errorf("record %q not in zone %q", record.Name, zone.Name)
	}

	tokens := []string{
		"dns",
		zone.Name,
		record.RrsType,
		fqdn,
	}
	return strings.Join(tokens, "/"), nil
}

type DNSRecord struct {
	Name    string
	Rrdatas []string
	RrsType string
}

type DNSZoneInfo struct {
	Name string
}

func NewDNSView(gossipState gossip.GossipState) *DNSView {
	return &DNSView{
		gossipState: gossipState,
	}
}

// Snapshot returns a copy of the current desired DNS state-of-the-world
func (v *DNSView) Snapshot() *DNSViewSnapshot {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	gossipSnapshot := v.gossipState.Snapshot()
	// Snapshot must be cheap if nothing has changed
	if v.lastSnapshot != nil && gossipSnapshot.Version == v.lastSnapshot.version {
		return v.lastSnapshot
	}

	snapshot := &DNSViewSnapshot{
		version: gossipSnapshot.Version,
	}

	zoneMap := make(map[string]*dnsViewSnapshotZone)
	for k, v := range gossipSnapshot.Values {
		if strings.HasPrefix(k, "dns/") {
			tokens := strings.Split(k, "/")
			if len(tokens) != 4 {
				klog.Warningf("key had invalid format: %q", k)
				continue
			}

			zoneID := tokens[1]
			recordType := tokens[2]
			name := tokens[3]

			zone := zoneMap[zoneID]
			if zone == nil {
				zone = &dnsViewSnapshotZone{
					Name:    zoneID,
					Records: make(map[string]DNSRecord),
				}
				zoneMap[zoneID] = zone
			}

			key := recordType + "::" + name

			record, found := zone.Records[key]
			if !found {
				record.Name = name
				record.RrsType = recordType
			}

			addresses := strings.Split(v, ",")
			record.Rrdatas = append(record.Rrdatas, addresses...)
			zone.Records[key] = record
		} else {
			klog.Warningf("unknown tag %q=%q", k, v)
		}
	}

	snapshot.zoneMap = zoneMap
	v.lastSnapshot = snapshot

	return snapshot
}
