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

	"github.com/Elemento-Modular-Cloud/ecloud-go/ecloud"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
)

// +kops:fitask
type DNSRecord struct {
	Name      *string
	Data      *string
	DNSZone   *string
	Type      *string
	TTL       *int64
	Lifecycle fi.Lifecycle
	Comment   *string
}

var _ fi.CloudupTask = &DNSRecord{}

func (d *DNSRecord) Find(c *fi.CloudupContext) (*DNSRecord, error) {
	// The Elemento SDK currently exposes create-only DNS methods. We therefore
	// reconcile DNS by issuing create calls and tolerating "already exists" errors.
	return nil, nil
}

func (d *DNSRecord) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(d, c)
}

func (_ *DNSRecord) CheckChanges(actual, expected, changes *DNSRecord) error {
	if expected.Name == nil {
		return fi.RequiredField("Name")
	}
	if expected.DNSZone == nil {
		return fi.RequiredField("DNSZone")
	}
	if expected.Type == nil {
		return fi.RequiredField("Type")
	}
	if fi.ValueOf(expected.Type) != "A" {
		return fmt.Errorf("Elemento DNS currently supports only A records, got %q", fi.ValueOf(expected.Type))
	}
	if expected.Data == nil {
		return fi.RequiredField("Data")
	}

	return nil
}

func (_ *DNSRecord) RenderElemento(t *elemento.ElementoAPITarget, actual, expected, changes *DNSRecord) error {
	client := t.Cloud.DnsClient()
	zoneName := fi.ValueOf(expected.DNSZone)
	recordName := fi.ValueOf(expected.Name)
	recordValue := fi.ValueOf(expected.Data)

	if err := ensureElementoDNSZone(context.TODO(), client, zoneName); err != nil {
		return err
	}
	if err := ensureElementoDNSRecord(context.TODO(), client, zoneName, recordName, recordValue); err != nil {
		return err
	}

	return nil
}

func ensureElementoDNSZone(ctx context.Context, client ecloud.DnsClient, zoneName string) error {
	_, _, err := client.Create(ctx, zoneName)
	if err != nil {
		if elemento.IsDNSAlreadyExists(err) {
			klog.V(2).Infof("Elemento DNS zone %q already exists", zoneName)
			return nil
		}
		return fmt.Errorf("creating Elemento DNS zone %q: %w", zoneName, err)
	}

	klog.V(2).Infof("Created Elemento DNS zone %q", zoneName)
	return nil
}

func ensureElementoDNSRecord(ctx context.Context, client ecloud.DnsClient, zoneName, recordName, recordValue string) error {
	record, _, err := client.AddDnsRecord(ctx, zoneName, recordName, recordValue)
	if err != nil {
		if elemento.IsDNSAlreadyExists(err) {
			klog.V(2).Infof("Elemento DNS record %q in zone %q already exists", recordName, zoneName)
			return nil
		}
		return fmt.Errorf("creating Elemento DNS record %q in zone %q: %w", recordName, zoneName, err)
	}

	klog.V(2).Infof("Created Elemento DNS record %q in zone %q as %q", recordName, zoneName, record.Name)
	return nil
}
