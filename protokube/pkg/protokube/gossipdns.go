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
