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
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typePort = "Port"
)

func (os *clusterDiscoveryOS) ListPorts(network networks.Network) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	projectPorts, err := os.osCloud.ListPorts(ports.ListOpts{
		TenantID:  network.ProjectID,
		NetworkID: network.ID,
	})
	if err != nil {
		return nil, err
	}

	preExistingNet := true
	if os.clusterName == network.Name {
		preExistingNet = false
	}

	filteredPorts := []ports.Port{}
	if preExistingNet {
		// if we have preExistingNet, the port must have cluster tag
		for _, singlePort := range projectPorts {
			if fi.ArrayContains(singlePort.Tags, os.clusterName) {
				filteredPorts = append(filteredPorts, singlePort)
			}
		}
	} else {
		filteredPorts = projectPorts
	}

	for _, port := range filteredPorts {
		resourceTracker := &resources.Resource{
			Name: port.Name,
			ID:   port.ID,
			Type: typePort,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return cloud.(openstack.OpenstackCloud).DeletePort(r.ID)
			},
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}
