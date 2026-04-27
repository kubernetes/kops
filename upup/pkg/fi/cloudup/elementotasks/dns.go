/*
Copyright 2026 The Kubernetes Authors.

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

package elementotasks

import (
	"time"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
)

type Dns struct {
	ID           string
	ZoneName     *string
	Created      time.Time
	AtomOsTarget string
	Status       string
	Records      []DnsRecord
}

type DnsRecord struct {
	Name  string
	Type  string
	Value string
	TTL   int
}

func (d *Dns) Find(c *fi.CloudupContext) (*Dns, error) {
	cloud := c.T.Cloud.(elemento.ElementoCloud)
	client := cloud.DnsClient()

}

func (d *Dns) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(d, c)
}

func (*Dns) CheckChanges(a, e, changes *Dns) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.DNSZone != nil {
			return fi.CannotChangeField("DNSZone")
		}
	}

	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	if e.DNSZone == nil {
		return fi.RequiredField("DNSZone")
	}
	if e.Type == nil {
		return fi.RequiredField("Type")
	}
	if e.Data == nil {
		return fi.RequiredField("Data")
	}
	if e.TTL == nil {
		return fi.RequiredField("TTL")
	}

	return nil
}

func (*Dns) RenderElemento(t *elemento.ElementoAPITarget, a, e, changes *Dns) error {
	cloud := t.Cloud

	if a == nil {
		//return createManagedDNSRecord(context.TODO(), cloud, e)
	}

	if changes.Data != nil || changes.TTL != nil || changes.Type != nil {
		//return updateManagedDNSRecord(context.TODO(), cloud, e)
	}

	return nil
}
