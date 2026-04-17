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

package linodetasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"k8s.io/klog/v2"
	dnspkg "k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type DNSRecord struct {
	Name         *string
	ResourceName *string
	Lifecycle    fi.Lifecycle

	RecordType *string
	Target     fi.HasAddress
}

var _ fi.CloudupTask = &DNSRecord{}
var _ fi.HasName = &DNSRecord{}

func (d *DNSRecord) GetName() *string {
	return d.Name
}

func (d *DNSRecord) String() string {
	return fi.CloudupTaskAsString(d)
}

func (d *DNSRecord) Find(c *fi.CloudupContext) (*DNSRecord, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	dns, err := cloud.DNS()
	if err != nil {
		return nil, fmt.Errorf("error getting DNS provider: %w", err)
	}

	if dns == nil {
		klog.V(2).Infof("DNS provider not available for %s", fi.ValueOf(d.ResourceName))
		return nil, nil
	}

	zone, err := findDNSZone(dns, fi.ValueOf(d.ResourceName))
	if err != nil {
		return nil, err
	}
	if zone == nil {
		klog.V(2).Infof("DNS zone not found for %s, skipping DNS record", fi.ValueOf(d.ResourceName))
		return nil, nil
	}

	rrs, supported := zone.ResourceRecordSets()
	if !supported {
		return nil, fmt.Errorf("zone %q does not support resource record sets", zone.Name())
	}

	// Look for existing record
	recordName := dnspkg.EnsureDotSuffix(fi.ValueOf(d.ResourceName))
	records, err := rrs.Get(recordName)
	if err != nil {
		return nil, fmt.Errorf("error querying DNS records for %q: %w", recordName, err)
	}

	klog.V(4).Infof("Found %d DNS records for %s in zone %s", len(records), recordName, zone.Name())

	recordType := fi.ValueOf(d.RecordType)
	for _, record := range records {
		klog.V(4).Infof("Checking record type %s (want %s)", record.Type(), recordType)
		if string(record.Type()) != recordType {
			continue
		}

		// Found existing record with matching type
		klog.V(2).Infof("Found existing DNS record for %s type %s", recordName, recordType)
		actual := &DNSRecord{
			Name:         d.Name,
			ResourceName: d.ResourceName,
			Lifecycle:    d.Lifecycle,
			RecordType:   d.RecordType,
			Target:       d.Target,
		}
		return actual, nil
	}

	klog.V(2).Infof("No existing DNS record found for %s type %s", recordName, recordType)
	return nil, nil
}

func (d *DNSRecord) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(d, c)
}

func (*DNSRecord) CheckChanges(a, e, changes *DNSRecord) error {
	if e.ResourceName == nil {
		return fi.RequiredField("ResourceName")
	}
	if e.RecordType == nil {
		return fi.RequiredField("RecordType")
	}
	if e.Target == nil {
		return fi.RequiredField("Target")
	}
	return nil
}

func (*DNSRecord) RenderLinode(t *linode.APITarget, a, e, changes *DNSRecord) error {
	cloud := t.Cloud

	dnsProvider, err := cloud.DNS()
	if err != nil {
		return fmt.Errorf("error getting DNS provider: %w", err)
	}

	if dnsProvider == nil {
		klog.Infof("DNS provider not available, skipping DNS record creation for %s", fi.ValueOf(e.ResourceName))
		return nil
	}

	zone, err := findDNSZone(dnsProvider, fi.ValueOf(e.ResourceName))
	if err != nil {
		return err
	}
	if zone == nil {
		return fmt.Errorf("DNS zone not found for %s", fi.ValueOf(e.ResourceName))
	}

	rrs, supported := zone.ResourceRecordSets()
	if !supported {
		return fmt.Errorf("zone %q does not support resource record sets", zone.Name())
	}

	// Get the target's ID to query the load balancer
	targetLB, ok := e.Target.(*LoadBalancer)
	if !ok {
		return fmt.Errorf("target is not a LoadBalancer")
	}

	if targetLB.ID == nil {
		klog.V(4).Infof("Target LoadBalancer has no ID yet, skipping DNS record creation for %s", fi.ValueOf(e.ResourceName))
		return nil
	}

	// Query the actual load balancer to get its IP
	nodebalancer, err := cloud.Client().GetNodeBalancer(context.Background(), fi.ValueOf(targetLB.ID))
	if err != nil {
		if linodego.IsNotFound(err) {
			klog.V(4).Infof("Target LoadBalancer not yet created, skipping DNS record creation for %s", fi.ValueOf(e.ResourceName))
			return nil
		}
		return fmt.Errorf("error getting load balancer: %w", err)
	}

	if nodebalancer == nil || nodebalancer.IPv4 == nil || *nodebalancer.IPv4 == "" {
		klog.V(4).Infof("Target LoadBalancer has no IP yet, skipping DNS record creation for %s", fi.ValueOf(e.ResourceName))
		return nil
	}

	targetIP := *nodebalancer.IPv4
	recordName := dnspkg.EnsureDotSuffix(fi.ValueOf(e.ResourceName))
	recordType := rrstype.RrsType(fi.ValueOf(e.RecordType))

	// Check existing records before creating changeset
	existing, err := rrs.Get(recordName)
	if err != nil {
		return fmt.Errorf("error querying existing DNS records: %w", err)
	}

	// Check if a record already exists with the correct IP
	for _, record := range existing {
		if record.Type() == recordType {
			existingIPs := record.Rrdatas()
			if len(existingIPs) == 1 && existingIPs[0] == targetIP {
				klog.V(2).Infof("DNS record %s %s already points to %s, skipping update", recordName, recordType, targetIP)
				return nil
			}
			klog.V(2).Infof("DNS record %s %s exists but points to %v (want %s), updating", recordName, recordType, existingIPs, targetIP)
			break
		}
	}

	klog.V(2).Infof("Updating DNS record %s %s -> %s", recordName, recordType, targetIP)

	// Create changeset and update the record
	changeset := rrs.StartChangeset()

	// Remove old records with the same type
	for _, record := range existing {
		if record.Type() == recordType {
			klog.V(4).Infof("Removing existing DNS record %s %s -> %v", recordName, record.Type(), record.Rrdatas())
			changeset.Remove(record)
		}
	}

	// Add new record
	newRecord := rrs.New(recordName, []string{targetIP}, 300, recordType)
	changeset.Add(newRecord)

	if err := changeset.Apply(context.Background()); err != nil {
		return fmt.Errorf("error applying DNS changeset: %w", err)
	}

	klog.V(2).Infof("Updated DNS record %s %s -> %s", recordName, recordType, targetIP)
	return nil
}

// findDNSZone finds the DNS zone that contains the given hostname
func findDNSZone(dnsProvider dnsprovider.Interface, hostname string) (dnsprovider.Zone, error) {
	zones, supported := dnsProvider.Zones()
	if !supported {
		return nil, fmt.Errorf("DNS provider does not support zones")
	}

	allZones, err := zones.List()
	if err != nil {
		return nil, fmt.Errorf("error listing DNS zones: %w", err)
	}

	hostname = dnspkg.EnsureDotSuffix(hostname)

	var matches []dnsprovider.Zone
	for _, zone := range allZones {
		zoneName := dnspkg.EnsureDotSuffix(zone.Name())
		if strings.HasSuffix(hostname, zoneName) {
			matches = append(matches, zone)
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}

	// Return the most specific match (longest zone name)
	bestMatch := matches[0]
	for _, match := range matches {
		if len(match.Name()) > len(bestMatch.Name()) {
			bestMatch = match
		}
	}

	return bestMatch, nil
}
