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

package designate

import (
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"

	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
)

var _ dnsprovider.ResourceRecordSet = ResourceRecordSet{}

type ResourceRecordSet struct {
	impl   *recordsets.RecordSet
	rrsets *ResourceRecordSets
}

func (rrset ResourceRecordSet) Name() string {
	return rrset.impl.Name
}

func (rrset ResourceRecordSet) Rrdatas() []string {
	return rrset.impl.Records
}

func (rrset ResourceRecordSet) Ttl() int64 {
	return int64(rrset.impl.TTL)
}

func (rrset ResourceRecordSet) Type() rrstype.RrsType {
	return rrstype.RrsType(rrset.impl.Type)
}
