/*
Copyright 2018 The Kubernetes Authors.

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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeNetwork = "network"
	typeSubnet  = "subnet"
	typePort    = "port"
	typeRouter  = "router"
)

func listNetworks(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	rts := make([]*resources.Resource, 0)
	opts := networks.ListOpts{
		Name: clusterName,
	}
	ns, err := cloud.ListNetworks(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %s", err)
	}
	for _, n := range ns {
		rt := &resources.Resource{
			Name: n.Name,
			ID:   n.ID,
			Type: typeNetwork,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return networks.Delete(cloud.(openstack.OpenstackCloud).NetworkingClient(), n.ID).ExtractErr()
			},
			Obj: n,
		}
		rts = append(rts, rt)
	}
	return rts, nil
}

func listSubnets(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	rts := make([]*resources.Resource, 0)
	opts := subnets.ListOpts{
		Name: clusterName,
	}
	subs, err := cloud.ListSubnets(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list subnets: %s", err)
	}
	for _, s := range subs {
		rt := &resources.Resource{
			Name: s.Name,
			ID:   s.ID,
			Type: typeSubnet,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return subnets.Delete(cloud.(openstack.OpenstackCloud).NetworkingClient(), s.ID).ExtractErr()
			},
			Obj: s,
		}
		rts = append(rts, rt)
	}
	return rts, nil
}

func listPorts(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	rts := make([]*resources.Resource, 0)
	opts := ports.ListOpts{
		Name: clusterName,
	}
	ss, err := cloud.ListPorts(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %s", err)
	}
	for _, s := range ss {
		rt := &resources.Resource{
			Name: s.Name,
			ID:   s.ID,
			Type: typeSubnet,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return ports.Delete(cloud.(openstack.OpenstackCloud).NetworkingClient(), s.ID).ExtractErr()
			},
			Obj: s,
		}
		rts = append(rts, rt)
	}
	return rts, nil
}
