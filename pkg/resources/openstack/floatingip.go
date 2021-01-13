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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeFloatingIP = "FloatingIP"
)

func DeleteFloatingIP(cloud fi.Cloud, r *resources.Resource) error {
	return cloud.(openstack.OpenstackCloud).DeleteFloatingIP(r.ID)
}

func DeleteL3FloatingIP(cloud fi.Cloud, r *resources.Resource) error {
	return cloud.(openstack.OpenstackCloud).DeleteL3FloatingIP(r.ID)
}

func (os *clusterDiscoveryOS) listL3FloatingIPs(routerID string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	floatingIPs, err := os.osCloud.ListL3FloatingIPs(l3floatingip.ListOpts{})
	if err != nil {
		return resourceTrackers, err
	}
	for _, floatingIP := range floatingIPs {
		if floatingIP.RouterID == routerID {
			resourceTracker := &resources.Resource{
				Name:    floatingIP.FloatingIP,
				ID:      floatingIP.ID,
				Type:    typeFloatingIP,
				Deleter: DeleteL3FloatingIP,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}

func (os *clusterDiscoveryOS) listFloatingIPs(instance servers.Server) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	// we can find real instance name from instance name in old format
	// however, in new format the real name can be found in metadata
	name := instance.Name
	if val, ok := instance.Metadata[openstack.TagKopsName]; ok {
		name = val
	}
	floatingIPs, err := os.osCloud.ListL3FloatingIPs(l3floatingip.ListOpts{
		Description: "fip-" + name,
	})
	if err != nil {
		return resourceTrackers, err
	}
	for _, floatingIP := range floatingIPs {
		resourceTracker := &resources.Resource{
			Name:    floatingIP.Description,
			ID:      floatingIP.ID,
			Type:    typeFloatingIP,
			Deleter: DeleteFloatingIP,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}
