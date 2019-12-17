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
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeInstance = "Instance"
)

func (os *clusterDiscoveryOS) ListInstances() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	instances, err := os.osCloud.ListInstances(servers.ListOpts{})
	if err != nil {
		return resourceTrackers, err
	}

	for _, instance := range instances {
		val, ok := instance.Metadata["k8s"]
		if ok && val == os.clusterName {
			// Clean up any bound floating IP's
			floatingIPs, err := os.listFloatingIPs(instance.ID)
			if err != nil {
				return resourceTrackers, err
			}
			resourceTrackers = append(resourceTrackers, floatingIPs...)

			resourceTracker := &resources.Resource{
				Name: instance.Name,
				ID:   instance.ID,
				Type: typeInstance,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteInstanceWithID(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}
