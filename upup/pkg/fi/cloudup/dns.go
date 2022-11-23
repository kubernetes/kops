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

package cloudup

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/viper"

	"k8s.io/klog/v2"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
	"k8s.io/kops/pkg/apis/kops"
	apimodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	// PlaceholderIP is from TEST-NET-3
	// https://en.wikipedia.org/wiki/Reserved_IP_addresses
	PlaceholderIP   = "203.0.113.123"
	PlaceholderIPv6 = "fd00:dead:add::"
	PlaceholderTTL  = 10
	// DigitalOcean's DNS servers require a certain minimum TTL (it's 30), keeping 60 here.
	PlaceholderTTLDigitialOcean = 60
)

type recordKey struct {
	hostname string
	rrsType  rrstype.RrsType
}

func findZone(cluster *kops.Cluster, cloud fi.Cloud) (dnsprovider.Zone, error) {
	dns, err := cloud.DNS()
	if err != nil {
		return nil, fmt.Errorf("error building DNS provider: %v", err)
	}
	if dns == nil {
		return nil, nil
	}

	zonesProvider, ok := dns.Zones()
	if !ok {
		return nil, fmt.Errorf("error getting DNS zones provider")
	}

	zones, err := zonesProvider.List()
	if err != nil {
		return nil, fmt.Errorf("error listing DNS zones: %v", err)
	}

	var matches []dnsprovider.Zone
	findName := strings.TrimSuffix(cluster.Spec.DNSZone, ".")
	for _, zone := range zones {
		id := zone.ID()
		name := strings.TrimSuffix(zone.Name(), ".")
		if id == cluster.Spec.DNSZone || name == findName {
			matches = append(matches, zone)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("cannot find DNS Zone %q.  Please pre-create the zone and set up NS records so that it resolves", cluster.Spec.DNSZone)
	}

	if len(matches) > 1 {
		klog.Infof("Found multiple DNS Zones matching %q, please set the cluster's spec.dnsZone to the desired Zone ID:", cluster.Spec.DNSZone)
		for _, zone := range zones {
			id := zone.ID()
			klog.Infof("\t%s", id)
		}
		return nil, fmt.Errorf("found multiple DNS Zones matching %q", cluster.Spec.DNSZone)
	}

	zone := matches[0]
	return zone, nil
}

func validateDNS(cluster *kops.Cluster, cloud fi.Cloud) error {
	if cluster.IsGossip() || cluster.UsesPrivateDNS() || cluster.UsesNoneDNS() {
		klog.V(2).Infof("Skipping DNS validation for non-public DNS")
		return nil
	}

	zone, err := findZone(cluster, cloud)
	if err != nil {
		return err
	}
	if zone == nil {
		return nil
	}
	dnsName := strings.TrimSuffix(zone.Name(), ".")

	klog.V(2).Infof("Doing DNS lookup to verify NS records for %q", dnsName)
	ns, err := net.LookupNS(dnsName)
	if err != nil {
		return fmt.Errorf("error doing DNS lookup for NS records for %q: %v", dnsName, err)
	}

	if len(ns) == 0 {
		if viper.GetString("DNS_IGNORE_NS_CHECK") == "" {
			return fmt.Errorf("NS records not found for %q - please make sure they are correctly configured", dnsName)
		}
		klog.Warningf("Ignoring failed NS record check because DNS_IGNORE_NS_CHECK is set")
	} else {
		var hosts []string
		for _, n := range ns {
			hosts = append(hosts, n.Host)
		}
		klog.V(2).Infof("Found NS records for %q: %v", dnsName, hosts)
	}

	return nil
}

func precreateDNS(ctx context.Context, cluster *kops.Cluster, cloud fi.Cloud) error {
	// TODO: Move to update

	// We precreate some DNS names (where they don't exist), with a dummy IP address
	// This avoids hitting negative TTL on DNS lookups, which tend to be very long
	// If we get the names wrong here, it doesn't really matter (extra DNS name, slower boot)

	// Nothing to do for Gossip clusters and clusters without DNS
	if cluster.IsGossip() || cluster.UsesNoneDNS() {
		return nil
	}

	recordKeys := buildPrecreateDNSHostnames(cluster)
	if len(recordKeys) == 0 {
		klog.V(2).Infof("No DNS records to pre-create")
		return nil
	}

	klog.V(2).Infof("Checking DNS records")
	zone, err := findZone(cluster, cloud)
	if err != nil {
		return err
	}
	if zone == nil {
		return nil
	}

	rrs, ok := zone.ResourceRecordSets()
	if !ok {
		return fmt.Errorf("error getting DNS resource records for %q", zone.Name())
	}

	recordsMap := make(map[string]dnsprovider.ResourceRecordSet)
	// TODO: We should change the filter to be a suffix match instead
	// records, err := rrs.List("", "")
	records, err := rrs.List()
	if err != nil {
		return fmt.Errorf("error listing DNS resource records for %q: %v", zone.Name(), err)
	}

	for _, record := range records {
		name := dns.EnsureDotSuffix(record.Name())
		key := string(record.Type()) + "::" + name
		recordsMap[key] = record
	}

	changeset := rrs.StartChangeset()
	// TODO: Add ChangeSet.IsEmpty() method
	var created []recordKey

	for _, recordKey := range recordKeys {
		recordKey.hostname = dns.EnsureDotSuffix(recordKey.hostname)
		foundAddress := false
		{
			dnsRecord := recordsMap[string(recordKey.rrsType)+"::"+recordKey.hostname]
			if dnsRecord != nil {
				rrdatas := dnsRecord.Rrdatas()
				if len(rrdatas) > 0 {
					klog.V(4).Infof("Found DNS record %s => %s; won't create", recordKey, rrdatas)
				} else {
					// This is probably an alias target; leave it alone...
					klog.V(4).Infof("Found DNS record %s, but no records", recordKey)
				}
				foundAddress = true
			}
		}

		foundTXT := false
		{
			dnsRecord := recordsMap["TXT::"+recordKey.hostname]
			if dnsRecord != nil {
				foundTXT = true
			}
		}
		if foundAddress && foundTXT {
			continue
		}

		ip := PlaceholderIP
		if recordKey.rrsType != rrstype.A {
			ip = PlaceholderIPv6
		}
		klog.V(2).Infof("Pre-creating DNS record %s => %s", recordKey, ip)

		if !foundAddress {
			if cloud.ProviderID() == kops.CloudProviderDO {
				changeset.Add(rrs.New(recordKey.hostname, []string{ip}, PlaceholderTTLDigitialOcean, recordKey.rrsType))
			} else {
				changeset.Add(rrs.New(recordKey.hostname, []string{ip}, PlaceholderTTL, recordKey.rrsType))
			}
		}
		if !foundTXT {
			if cluster.Spec.ExternalDNS != nil && cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderExternalDNS {
				changeset.Add(rrs.New(recordKey.hostname, []string{fmt.Sprintf("\"heritage=external-dns,external-dns/owner=kops-%s\"", cluster.ObjectMeta.Name)}, PlaceholderTTL, rrstype.TXT))
			}
		}
		created = append(created, recordKey)
	}

	if len(created) != 0 {
		klog.Infof("Pre-creating DNS records")

		err := changeset.Apply(ctx)
		if err != nil {
			return fmt.Errorf("error pre-creating DNS records: %v", err)
		}
		klog.V(2).Infof("Pre-created DNS names: %v", created)
	}

	return nil
}

// buildPrecreateDNSHostnames returns the hostnames we should precreate
func buildPrecreateDNSHostnames(cluster *kops.Cluster) []recordKey {
	var recordKeys []recordKey
	internalType := rrstype.A
	if cluster.Spec.IsIPv6Only() {
		internalType = rrstype.AAAA
	}

	hasAPILoadbalancer := cluster.Spec.API.LoadBalancer != nil
	useLBForInternalAPI := hasAPILoadbalancer && cluster.Spec.API.LoadBalancer.UseForInternalAPI

	if cluster.Spec.API.PublicName != "" && !hasAPILoadbalancer {
		recordKeys = append(recordKeys, recordKey{
			hostname: cluster.Spec.API.PublicName,
			rrsType:  rrstype.A,
		})
		if internalType != rrstype.A {
			recordKeys = append(recordKeys, recordKey{
				hostname: cluster.Spec.API.PublicName,
				rrsType:  internalType,
			})
		}
	}

	if !useLBForInternalAPI {
		recordKeys = append(recordKeys, recordKey{
			hostname: cluster.APIInternalName(),
			rrsType:  internalType,
		})
	}

	if apimodel.UseKopsControllerForNodeBootstrap(cluster) {
		name := "kops-controller.internal." + cluster.ObjectMeta.Name
		recordKeys = append(recordKeys, recordKey{
			hostname: name,
			rrsType:  internalType,
		})
	}

	return recordKeys
}
