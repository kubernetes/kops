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

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/monitors"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/klog"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeLB   = "LoadBalancer"
	typeLBL  = "LBListener"
	typeLBP  = "LBPool"
	typeLBPM = "LBPoolMonitor"
)

func (os *clusterDiscoveryOS) DeleteSubnetLBs(subnet subnets.Subnet) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	preExistingSubnet := false
	if !strings.HasSuffix(subnet.Name, os.clusterName) {
		preExistingSubnet = true
	}

	opts := loadbalancers.ListOpts{
		VipSubnetID: subnet.ID,
	}
	lbs, err := os.osCloud.ListLBs(opts)
	if err != nil {
		return nil, err
	}

	filteredLBs := []loadbalancers.LoadBalancer{}
	if preExistingSubnet {
		// if we have preExistingSubnet, we cannot delete others than api LB
		for _, lb := range lbs {
			if lb.Name == fmt.Sprintf("api.%s", os.clusterName) {
				filteredLBs = append(filteredLBs, lb)
			}
		}
	} else {
		filteredLBs = lbs
	}

	for _, lb := range filteredLBs {
		resourceTracker := &resources.Resource{
			Name: lb.Name,
			ID:   lb.ID,
			Type: typeLB,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				opts := loadbalancers.DeleteOpts{
					Cascade: true,
				}
				return cloud.(openstack.OpenstackCloud).DeleteLB(r.ID, opts)
			},
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)

		if os.osCloud.UseOctavia() {
			klog.V(2).Info("skipping LB Children because using Octavia")
			continue
		}

		//Identify pools associated to this LB
		for _, pool := range lb.Pools {

			monitorList, err := os.cloud.(openstack.OpenstackCloud).ListMonitors(monitors.ListOpts{
				PoolID: pool.ID,
			})
			if err != nil {
				return nil, err
			}
			for _, monitor := range monitorList {
				resourceTracker := &resources.Resource{
					Name: monitor.Name,
					ID:   monitor.ID,
					Type: typeLBPM,
					Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
						return cloud.(openstack.OpenstackCloud).DeleteMonitor(r.ID)
					},
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			}

			resourceTracker := &resources.Resource{
				Name: pool.Name,
				ID:   pool.ID,
				Type: typeLBP,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeletePool(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}

		//Identify listeners associated to this LB
		for _, listener := range lb.Listeners {
			resourceTracker := &resources.Resource{
				Name: listener.Name,
				ID:   listener.ID,
				Type: typeLBL,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteListener(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}
