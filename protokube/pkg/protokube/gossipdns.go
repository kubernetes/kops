package protokube

import (
	"k8s.io/kops/protokube/pkg/gossip/dns"
)

type GossipDnsProvider struct {
	DNSView *dns.DNSView
	Zone    dns.DNSZoneInfo
}

var _ DNSProvider = &GossipDnsProvider{}

func (p *GossipDnsProvider) Replace(fqdn string, values []string) error {
	record := &dns.DNSRecord{
		Name:    fqdn,
		RrsType: "A",
	}
	for _, value := range values {
		record.Rrdatas = append(record.Rrdatas, value)
	}
	return p.DNSView.ApplyChangeset(p.Zone, nil, []*dns.DNSRecord{record})
}

func (p *GossipDnsProvider) Run() {

}
