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

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	v2pools "github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeLB  = "LoadBalancer"
	typeLBL = "LBListener"
	typeLBP = "LBPool"
)

func (os *clusterDiscoveryOS) ListLB() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	opts := loadbalancers.ListOpts{
		Name: "api." + os.clusterName,
	}
	lbs, err := os.osCloud.ListLBs(opts)
	if err != nil {
		return nil, err
	}

	for _, lb := range lbs {
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
	}
	return resourceTrackers, nil
}

func (os *clusterDiscoveryOS) ListLBPools() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	if os.osCloud.UseOctavia() {
		glog.V(2).Info("skipping ListLBPools because using Octavia")
		return nil, nil
	}

	pools, err := os.osCloud.ListPools(v2pools.ListOpts{})
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		if strings.Contains(pool.Name, os.clusterName) {
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
	}
	return resourceTrackers, nil
}

func (os *clusterDiscoveryOS) ListLBListener() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	if os.osCloud.UseOctavia() {
		glog.V(2).Info("skipping ListLBListener because using Octavia")
		return nil, nil
	}

	listeners, err := os.osCloud.ListListeners(listeners.ListOpts{})
	if err != nil {
		return nil, err
	}

	for _, listener := range listeners {
		if strings.Contains(listener.Name, os.clusterName) {
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
