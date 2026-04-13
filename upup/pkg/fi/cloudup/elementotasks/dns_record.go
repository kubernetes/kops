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
	"context"
	"fmt"
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
)

// +kops:fitask
type DNSRecord struct {
	Name      *string
	Lifecycle fi.Lifecycle

	// DNSZone is the parent zone, for example "example.com".
	DNSZone *string
	// Type is the record type, typically "A" for bootstrap.
	Type *string
	// Data is the current record target. During the first implementation this is
	// usually a placeholder value, until the final API VIP or LB address is known.
	Data *string
	// TTL is the DNS TTL in seconds.
	TTL *int64
	// Comment is not used by kOps itself, but helps document why this bootstrap
	// record exists while the Elemento DNS API integration is being completed.
	Comment *string
}

type managedDNSRecord struct {
	ID   string
	Name string
	Zone string
	Type string
	Data string
	TTL  int64
}

func (r *DNSRecord) Find(c *fi.CloudupContext) (*DNSRecord, error) {
	cloud := c.T.Cloud.(elemento.ElementoCloud)

	actual, err := findManagedDNSRecord(context.TODO(), cloud, r)
	if err != nil {
		return nil, err
	}
	if actual == nil {
		return nil, nil
	}

	matches := &DNSRecord{
		Name:      r.Name,
		Lifecycle: r.Lifecycle,
		DNSZone:   fi.PtrTo(actual.Zone),
		Type:      fi.PtrTo(actual.Type),
		Data:      fi.PtrTo(actual.Data),
		TTL:       fi.PtrTo(actual.TTL),
		Comment:   r.Comment,
	}
	return matches, nil
}

func (r *DNSRecord) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(r, c)
}

func (*DNSRecord) CheckChanges(a, e, changes *DNSRecord) error {
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

func (*DNSRecord) RenderElemento(t *elemento.ElementoAPITarget, a, e, changes *DNSRecord) error {
	cloud := t.Cloud

	if a == nil {
		return createManagedDNSRecord(context.TODO(), cloud, e)
	}

	if changes.Data != nil || changes.TTL != nil || changes.Type != nil {
		return updateManagedDNSRecord(context.TODO(), cloud, e)
	}

	return nil
}

func fullyQualifiedRecordName(recordName string, zone string) string {
	recordName = strings.TrimSuffix(recordName, ".")
	zone = strings.TrimSuffix(zone, ".")
	if zone == "" {
		return recordName
	}
	if strings.HasSuffix(recordName, "."+zone) || recordName == zone {
		return recordName
	}
	return recordName + "." + zone
}

func findManagedDNSRecord(ctx context.Context, cloud elemento.ElementoCloud, desired *DNSRecord) (*managedDNSRecord, error) {
	_ = ctx
	_ = cloud

	// TODO(elemento-dns): Replace this mock with a real lookup against the Elemento DNS API.
	//
	// Expected implementation shape:
	// 1. Resolve the Elemento zone matching fi.ValueOf(desired.DNSZone)
	// 2. List records in that zone
	// 3. Find the record with:
	//    - fullyQualifiedRecordName(fi.ValueOf(desired.Name), fi.ValueOf(desired.DNSZone))
	//    - type fi.ValueOf(desired.Type)
	// 4. Return its current target and TTL so kOps can decide whether it changed
	return nil, nil
}

func createManagedDNSRecord(ctx context.Context, cloud elemento.ElementoCloud, desired *DNSRecord) error {
	_ = ctx
	_ = cloud

	fqdn := fullyQualifiedRecordName(fi.ValueOf(desired.Name), fi.ValueOf(desired.DNSZone))

	// TODO(elemento-dns): Replace this error with the real Elemento DNS create call.
	//
	// Expected create payload:
	// - zone: fi.ValueOf(desired.DNSZone)
	// - name: fqdn
	// - type: fi.ValueOf(desired.Type)
	// - value: fi.ValueOf(desired.Data)
	// - ttl: fi.ValueOf(desired.TTL)
	//
	// Suggested first target records:
	// - api.internal.<cluster>
	// - kops-controller.internal.<cluster>
	// - api.<cluster> if the public API endpoint should resolve through your zone
	return fmt.Errorf("Elemento DNS create is not implemented yet for %q; implement createManagedDNSRecord in elementotasks/dns_record.go", fqdn)
}

func updateManagedDNSRecord(ctx context.Context, cloud elemento.ElementoCloud, desired *DNSRecord) error {
	_ = ctx
	_ = cloud

	fqdn := fullyQualifiedRecordName(fi.ValueOf(desired.Name), fi.ValueOf(desired.DNSZone))

	// TODO(elemento-dns): Replace this error with the real Elemento DNS update call.
	//
	// This should update at least:
	// - target/value when the bootstrap placeholder is replaced by the final VIP/LB IP
	// - ttl if you choose to lower it for bootstrap and raise it later
	return fmt.Errorf("Elemento DNS update is not implemented yet for %q; implement updateManagedDNSRecord in elementotasks/dns_record.go", fqdn)
}
