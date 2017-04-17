package dns

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/protokube/pkg/gossip"
	"strings"
	"sync"
	"time"
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

		// TODO: Replace with a watch - but debouncing is also nice

		// Snapshot is very cheap if we are in-sync
		snapshot := src.Snapshot()
		if lastSnapshot != nil && lastSnapshot.version == snapshot.version {
			glog.Infof("DNSView unchanged: %v", lastSnapshot.version)
			continue
		}

		// TODO: We might want to keep old records alive for a bit

		glog.Infof("DNSView changed: %v", snapshot.version)

		err := target.Update(snapshot)
		if err != nil {
			glog.Warningf("error applying DNS changes to target: %v", err)
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

func (s *DNSViewSnapshot) ListZones() []DNSZoneInfo {
	var zones []DNSZoneInfo
	for _, z := range s.zoneMap {
		zones = append(zones, DNSZoneInfo{
			Name: z.Name,
		})
	}
	return zones
}

func (s *DNSView) RemoveZone(info DNSZoneInfo) error {
	// TODO: Not sure if we should support this
	return fmt.Errorf("zone deletion is implicit")
}

func (s *DNSView) AddZone(info DNSZoneInfo) (*DNSZoneInfo, error) {
	//if info.ID != "" {
	//	return nil, fmt.Errorf("zone already created")
	//}

	// TODO: Not sure if we should support this
	return nil, fmt.Errorf("zone creation is implicit")
}

func (s *DNSView) ApplyChangeset(zone DNSZoneInfo, removeRecords []*DNSRecord, createRecords []*DNSRecord) error {
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

	return s.gossipState.UpdateValues(removeTags, createTags)
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
	//zoneID string

	Name    string
	Rrdatas []string
	//Ttl     int
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
func (g *DNSView) Snapshot() *DNSViewSnapshot {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	gossipSnapshot := g.gossipState.Snapshot()
	// Snapshot must be cheap if nothing has changed
	if g.lastSnapshot != nil && gossipSnapshot.Version == g.lastSnapshot.version {
		return g.lastSnapshot
	}

	snapshot := &DNSViewSnapshot{
		version: gossipSnapshot.Version,
	}

	zoneMap := make(map[string]*dnsViewSnapshotZone)
	for k, v := range gossipSnapshot.Values {
		if strings.HasPrefix(k, "dns/") {
			tokens := strings.Split(k, "/")
			if len(tokens) != 4 {
				glog.Warningf("key had invalid format: %q", k)
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
			glog.Warningf("unknown tag %q=%q", k, v)
		}
	}

	snapshot.zoneMap = zoneMap
	g.lastSnapshot = snapshot

	return snapshot
}
