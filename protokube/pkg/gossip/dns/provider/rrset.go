package provider

import (
	"k8s.io/kops/protokube/pkg/gossip/dns"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

const defaultTTL = 60

type resourceRecordSet struct {
	data dns.DNSRecord
}

var _ dnsprovider.ResourceRecordSet = &resourceRecordSet{}

// Name returns the name of the ResourceRecordSet, e.g. "www.example.com".
func (r *resourceRecordSet) Name() string {
	return r.data.Name
}

// Rrdatas returns the Resource Record Datas of the record set.
func (r *resourceRecordSet) Rrdatas() []string {
	return r.data.Rrdatas
}

// Ttl returns the time-to-live of the record set, in seconds.
func (r *resourceRecordSet) Ttl() int64 {
	return defaultTTL
}

// Type returns the type of the record set (A, CNAME, SRV, etc)
func (r *resourceRecordSet) Type() rrstype.RrsType {
	// TODO: Check if it is one of the well-known types?
	return rrstype.RrsType(r.data.RrsType)
}
