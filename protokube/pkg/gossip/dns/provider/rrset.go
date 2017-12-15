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
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"k8s.io/kops/protokube/pkg/gossip/dns"
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
