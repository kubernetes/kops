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
	typeRouterIF   = "Router-IF"
	typeRouter     = "Router"
	typeSubnet     = "Subnet"
	typeNetwork    = "Network"
	typeNetworkTag = "NetworkTag"
	typeSubnetTag  = "SubnetTag"
)

func (os *clusterDiscoveryOS) ListNetwork() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	routerName := strings.Replace(os.clusterName, ".", "-", -1)

	projectNetworks, err := os.osCloud.ListNetworks(networks.ListOpts{})
	if err != nil {
		return resourceTrackers, err
	}

	filteredNetwork := []networks.Network{}
	for _, net := range projectNetworks {
		if net.Name == os.clusterName || fi.ArrayContains(net.Tags, os.clusterName) {
			filteredNetwork = append(filteredNetwork, net)
		}
	}

	for _, network := range filteredNetwork {

		preExistingNet := true
		if os.clusterName == network.Name {
			preExistingNet = false
		}

		optRouter := osrouter.ListOpts{
			Name: routerName,
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
		networkSubnets, err := os.osCloud.ListSubnets(optSubnet)
		if err != nil {
			return resourceTrackers, err
		}
		filteredSubnets := []subnets.Subnet{}
		if preExistingNet {
			// if we have preExistingNet, the subnet must have cluster tag
			for _, sub := range networkSubnets {
				if fi.ArrayContains(sub.Tags, os.clusterName) {
					filteredSubnets = append(filteredSubnets, sub)
				}
			}
		} else {
			filteredSubnets = networkSubnets
		}

		for _, subnet := range filteredSubnets {
			// router interfaces
			preExistingSubnet := false
			if !strings.HasSuffix(subnet.Name, os.clusterName) {
				preExistingSubnet = true
			}

			if !preExistingSubnet {
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
			}
			//associated load balancers
			lbTrackers, err := os.DeleteSubnetLBs(subnet)
			if err != nil {
				return resourceTrackers, err
			}
			resourceTrackers = append(resourceTrackers, lbTrackers...)

			if !preExistingSubnet {
				resourceTracker := &resources.Resource{
					Name: subnet.Name,
					ID:   subnet.ID,
					Type: typeSubnet,
					Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
						return cloud.(openstack.OpenstackCloud).DeleteSubnet(r.ID)
					},
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			} else {
				resourceTracker := &resources.Resource{
					Name: os.clusterName,
					ID:   subnet.ID,
					Type: typeSubnetTag,
					Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
						return cloud.(openstack.OpenstackCloud).DeleteTag(openstack.ResourceTypeSubnet, r.ID, r.Name)
					},
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			}
		}

		// Ports
		portTrackers, err := os.ListPorts(network)
		if err != nil {
			return resourceTrackers, err
		}
		resourceTrackers = append(resourceTrackers, portTrackers...)

		if !preExistingNet {
			resourceTracker := &resources.Resource{
				Name: network.Name,
				ID:   network.ID,
				Type: typeNetwork,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteNetwork(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		} else {
			resourceTracker := &resources.Resource{
				Name: os.clusterName,
				ID:   network.ID,
				Type: typeNetworkTag,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteTag(openstack.ResourceTypeNetwork, r.ID, r.Name)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}
