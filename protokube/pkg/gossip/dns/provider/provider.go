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
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/protokube/pkg/gossip/dns"
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
