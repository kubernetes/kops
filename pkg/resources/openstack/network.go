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
	"strings"

	osrouter "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeRouterIF = "Router-IF"
	typeRouter   = "Router"
	typeSubnet   = "Subnet"
	typeNetwork  = "Network"
)

func (os *clusterDiscoveryOS) ListNetwork() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	opt := networks.ListOpts{
		Name: os.clusterName,
	}
	networks, err := os.osCloud.ListNetworks(opt)
	if err != nil {
		return resourceTrackers, err
	}

	for _, network := range networks {
		optRouter := osrouter.ListOpts{
			Name: strings.Replace(os.clusterName, ".", "-", -1),
		}
		routers, err := os.osCloud.ListRouters(optRouter)
		if err != nil {
			return resourceTrackers, err
		}
		for _, router := range routers {

			// Get the floating IP's associated to this router
			floatingIPs, err := os.listL3FloatingIPs(router.ID)
			if err != nil {
				return resourceTrackers, err
			}
			resourceTrackers = append(resourceTrackers, floatingIPs...)

			resourceTracker := &resources.Resource{
				Name: router.Name,
				ID:   router.ID,
				Type: typeRouter,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteRouter(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}

		optSubnet := subnets.ListOpts{
			NetworkID: network.ID,
		}
		subnets, err := os.osCloud.ListSubnets(optSubnet)
		if err != nil {
			return resourceTrackers, err
		}
		for _, subnet := range subnets {
			// router interfaces
			for _, router := range routers {
				resourceTracker := &resources.Resource{
					Name: router.ID,
					ID:   subnet.ID,
					Type: typeRouterIF,
					Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
						opts := osrouter.RemoveInterfaceOpts{
							SubnetID: r.ID,
						}
						return cloud.(openstack.OpenstackCloud).DeleteRouterInterface(r.Name, opts)
					},
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			}
			//associated load balancers
			lbTrackers, err := os.DeleteSubnetLBs(subnet)
			if err != nil {
				return resourceTrackers, err
			}
			resourceTrackers = append(resourceTrackers, lbTrackers...)

			resourceTracker := &resources.Resource{
				Name: subnet.Name,
				ID:   subnet.ID,
				Type: typeSubnet,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteSubnet(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)

		}

		// Ports
		portTrackers, err := os.ListPorts(network)
		if err != nil {
			return resourceTrackers, err
		}
		resourceTrackers = append(resourceTrackers, portTrackers...)

		resourceTracker := &resources.Resource{
			Name: network.Name,
			ID:   network.ID,
			Type: typeNetwork,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return cloud.(openstack.OpenstackCloud).DeleteNetwork(r.ID)
			},
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}
