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
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

var _ dnsprovider.Zones = Zones{}

type Zones struct {
	iface *Interface
}

func (kzs Zones) List() ([]dnsprovider.Zone, error) {
	var zoneList []dnsprovider.Zone

	allPages, err := zones.List(kzs.iface.sc, nil).AllPages()
	if err != nil {
		return zoneList, err
	}
	zs, err := zones.ExtractZones(allPages)
	if err != nil {
		return zoneList, err
	}
	for _, z := range zs {
		kz := &Zone{
			impl:  z,
			zones: &kzs,
		}
		zoneList = append(zoneList, kz)
	}
	return zoneList, nil
}

func (kzs Zones) Add(zone dnsprovider.Zone) (dnsprovider.Zone, error) {
	opts := &zones.CreateOpts{Name: zone.Name()}
	z, err := zones.Create(kzs.iface.sc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return &Zone{
		impl:  *z,
		zones: &kzs,
	}, nil
}

func (kzs Zones) Remove(zone dnsprovider.Zone) error {
	_, err := zones.Delete(kzs.iface.sc, zone.ID()).Extract()
	return err
}

func (kzs Zones) New(name string) (dnsprovider.Zone, error) {
	zone := zones.Zone{Name: name}
	return &Zone{zone, &kzs}, nil
}
