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

	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeDNSRecord = "dNSRecord"
)

func (os *clusterDiscoveryOS) ListDNSRecordsets() ([]*resources.Resource, error) {
	zopts := zones.ListOpts{
		Name: os.clusterName,
	}

	// if dnsclient does not exist (designate disabled)
	if os.osCloud.DNSClient() == nil {
		return nil, nil
	}

	zs, err := os.osCloud.ListDNSZones(zopts)
	if err != nil {
		return nil, fmt.Errorf("failed to list dns zones: %s", err)
	}

	switch len(zs) {
	case 0:
	case 1:
	default:
	}

	z := zs[0]

	rrs, err := os.osCloud.ListDNSRecordsets(z.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract recordsets pages for zone %s: %v", z.Name, err)
	}

	var resourceTrackers []*resources.Resource
	for _, rr := range rrs {
		if rr.Type != "A" {
			continue
		}

		resourceTracker := &resources.Resource{
			Name: rr.Name,
			ID:   rr.ID,
			Type: typeDNSRecord,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				// TODO: not tested and this should have retry similar to what we have in another resources
				return recordsets.Delete(cloud.(openstack.OpenstackCloud).DNSClient(), z.ID, rr.ID).ExtractErr()
			},
			Obj: rr,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
