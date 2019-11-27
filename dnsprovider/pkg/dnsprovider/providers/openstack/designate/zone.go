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

	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
)

var _ dnsprovider.Zone = &Zone{}

type Zone struct {
	impl  zones.Zone
	zones *Zones
}

func (z *Zone) Name() string {
	return z.impl.Name
}

func (z *Zone) ID() string {
	return z.impl.ID
}

func (z *Zone) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return &ResourceRecordSets{z}, true
}
