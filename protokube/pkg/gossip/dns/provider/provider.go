package provider

import (
	"k8s.io/kops/protokube/pkg/gossip/dns"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type Provider struct {
	zones   zones
	dnsView *dns.DNSView
}

var _ dnsprovider.Interface = &Provider{}

// Zones returns the provider's Zones interface, or false if not supported.
func (p *Provider) Zones() (dnsprovider.Zones, bool) {
	return &p.zones, true
}

func New(dnsView *dns.DNSView) (dnsprovider.Interface, error) {
	p := &Provider{
		dnsView: dnsView,
	}

	p.zones.dnsView = dnsView

	return p, nil

}
