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

package dnstasks

import (
	"fmt"

	"strings"

	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/upup/pkg/fi"
)

// DNSZone is a zone object in a dns provider
//go:generate fitask -type=DNSZone
type DNSZone struct {
	Name *string
	ID   *string
}

var _ fi.CompareWithID = &DNSZone{}

func (e *DNSZone) CompareWithID() *string {
	return e.Name
}

func (e *DNSZone) Find(c *fi.Context) (*DNSZone, error) {
	dns := c.DNS

	z, err := e.findExisting(dns)
	if err != nil {
		return nil, err
	}

	if z == nil {
		return nil, nil
	}

	actual := &DNSZone{}
	actual.Name = e.Name
	actual.ID = fi.String(z.Name())

	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (e *DNSZone) findExisting(dns dnsprovider.Interface) (dnsprovider.Zone, error) {
	findName := fi.StringValue(e.Name)
	if findName == "" {
		return nil, nil
	}
	if !strings.HasSuffix(findName, ".") {
		findName += "."
	}
	zonesProvider, ok := dns.Zones()
	if !ok {
		return nil, fmt.Errorf("DNS provider does not support zones")
	}
	// TODO: Support filtering!
	zones, err := zonesProvider.List()
	if err != nil {
		return nil, fmt.Errorf("error listing DNS zones: %v", err)
	}

	for _, zone := range zones {
		if zone.Name() == findName {
			zones = append(zones, zone)
		}
	}
	if len(zones) == 0 {
		return nil, nil
	}
	if len(zones) != 1 {
		return nil, fmt.Errorf("found multiple hosted zones matching name %q", findName)
	}

	return zones[0], nil
}

func (e *DNSZone) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *DNSZone) CheckChanges(a, e, changes *DNSZone) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (_ *DNSZone) Render(c *fi.Context, a, e, changes *DNSZone) error {
	dns := c.DNS
	zonesProvider, ok := dns.Zones()
	if !ok {
		return fmt.Errorf("DNS provider does not support zones")
	}

	if a == nil {
		name := fi.StringValue(e.Name)

		klog.V(2).Infof("Creating DNS Zone with Name %q", name)
		zone, err := zonesProvider.New(name)
		if err != nil {
			return fmt.Errorf("error creating DNS Zone %q: %v", name, err)
		}

		e.ID = fi.String(zone.Name())
	}

	return nil
}
