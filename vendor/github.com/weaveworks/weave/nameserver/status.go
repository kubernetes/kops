package nameserver

type Status struct {
	Domain   string
	Upstream []string
	Address  string
	TTL      uint32
	Entries  []EntryStatus
}

type EntryStatus struct {
	Hostname    string
	Origin      string
	ContainerID string
	Address     string
	Version     int
	Tombstone   int64
}

func NewStatus(ns *Nameserver, dnsServer *DNSServer) *Status {
	if dnsServer == nil {
		return nil
	}

	ns.RLock()
	defer ns.RUnlock()

	var entryStatusSlice []EntryStatus
	for _, entry := range ns.entries {
		entryStatusSlice = append(entryStatusSlice, EntryStatus{
			entry.Hostname,
			entry.Origin.String(),
			entry.ContainerID,
			entry.Addr.String(),
			entry.Version,
			entry.Tombstone})
	}

	upstreamConfig, _ := dnsServer.upstream.Config()
	return &Status{
		dnsServer.domain,
		upstreamConfig.Servers,
		dnsServer.address,
		dnsServer.ttl,
		entryStatusSlice}
}
