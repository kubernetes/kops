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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type listFn func(openstack.OpenstackCloud, string) ([]*resources.Resource, error)

// ListResources lists the OpenStack resources kops manages
func ListResources(cloud openstack.OpenstackCloud, clusterName string) (map[string]*resources.Resource, error) {
	rts := make(map[string]*resources.Resource)

	keypairs, err := listKeypairs(cloud, clusterName)
	if err != nil {
		return rts, err
	}
	keypairIDs := make([]string, len(keypairs))
	for i, t := range keypairs {
		id := t.Type + ":" + t.ID
		rts[id] = t
		keypairIDs[i] = id
	}

	serverGroups, err := listServerGroups(cloud, clusterName)
	if err != nil {
		return rts, err
	}
	serverGroupIDs := make([]string, len(serverGroups))
	for _, t := range serverGroups {
		id := t.Type + ":" + t.ID
		for _, m := range t.Obj.(servergroups.ServerGroup).Members {
			t.Blocked = append(t.Blocks, typeServer+":"+m)
		}
		serverGroupIDs = append(serverGroupIDs, id)
		rts[id] = t
	}

	instances, err := listInstances(cloud, clusterName)
	if err != nil {
		return rts, err
	}
	for _, t := range instances {
		rts[t.Type+":"+t.ID] = t
	}

	lbs, err := listLBs(cloud, clusterName)
	if err != nil {
		return rts, err
	}
	for _, t := range lbs {
		listeners := t.Obj.(loadbalancers.LoadBalancer).Listeners
		for _, l := range listeners {
			for _, p := range l.Pools {
				t.Blocks = append(t.Blocks, typeSubnet+":"+p.SubnetID)
			}
		}
		rts[t.Type+":"+t.ID] = t
	}

	return rts, nil
}
