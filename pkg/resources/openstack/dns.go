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

package openstack

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/dns"

	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	typeDNSRecord = "DNSRecord"
)

func (os *clusterDiscoveryOS) ListDNSRecordsets() ([]*resources.Resource, error) {
	// if dnsclient does not exist (designate disabled) or using gossip DNS
	if os.osCloud.DNSClient() == nil || dns.IsGossipClusterName(os.clusterName) {
		return nil, nil
	}

	zs, err := os.osCloud.ListDNSZones(zones.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list dns zones: %s", err)
	}

	var clusterZone zones.Zone
	for _, zone := range zs {
		if strings.HasSuffix(os.clusterName, strings.TrimSuffix(zone.Name, ".")) {
			clusterZone = zone
			break
		}
	}

	if clusterZone.ID == "" {
		return nil, fmt.Errorf("failed to find cluster dns zone")
	}

	rrs, err := os.osCloud.ListDNSRecordsets(clusterZone.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract recordsets pages for zone %s: %v", clusterZone.Name, err)
	}

	var resourceTrackers []*resources.Resource
	for _, rr := range rrs {
		if rr.Type != "A" || !strings.HasSuffix(strings.TrimSuffix(rr.Name, "."), os.clusterName) {
			continue
		}

		resourceTracker := &resources.Resource{
			Name: rr.Name,
			ID:   rr.ID,
			Type: typeDNSRecord,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return os.osCloud.DeleteDNSRecordset(clusterZone.ID, r.ID)
			},
			Obj: rr,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
