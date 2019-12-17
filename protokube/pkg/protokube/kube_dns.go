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

package protokube

import (
	"time"

	"k8s.io/klog"
	"k8s.io/kops/dns-controller/pkg/dns"
)

const defaultTTL = time.Minute

type DNSProvider interface {
	Replace(fqdn string, values []string) error

	// RemoveRecordsImmediate deletes the specified DNS records, without batching etc
	RemoveRecordsImmediate(records []dns.Record) error

	Run()
}

// CreateInternalDNSNameRecord maps a FQDN to the internal IP address of the current machine
func (k *KubeBoot) CreateInternalDNSNameRecord(fqdn string) error {
	values := []string{k.InternalIP.String()}
	klog.Infof("Creating DNS record: %s => %s", fqdn, values)
	return k.DNS.Replace(fqdn, values)
}

// BuildInternalDNSName builds a DNS name for use inside the cluster, adding our internal DNS suffix to the key
func (k *KubeBoot) BuildInternalDNSName(key string) string {
	fqdn := key + k.InternalDNSSuffix
	return fqdn
}

type KopsDnsProvider struct {
	DNSScope      dns.Scope
	DNSController *dns.DNSController
}

var _ DNSProvider = &KopsDnsProvider{}

func (p *KopsDnsProvider) RemoveRecordsImmediate(records []dns.Record) error {
	return p.DNSController.RemoveRecordsImmediate(records)
}

func (p *KopsDnsProvider) Replace(fqdn string, values []string) error {
	ttl := defaultTTL
	if ttl != dns.DefaultTTL {
		klog.Infof("Ignoring ttl %v for %q", ttl, fqdn)
	}

	var records []dns.Record
	for _, value := range values {
		records = append(records, dns.Record{
			RecordType: dns.RecordTypeA,
			FQDN:       fqdn,
			Value:      value,
		})
	}
	p.DNSScope.Replace(fqdn, records)

	return nil
}

func (p *KopsDnsProvider) Run() {
	p.DNSController.Run()
}
