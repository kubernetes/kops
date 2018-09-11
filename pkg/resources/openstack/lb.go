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

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeLB = "lb"
)

func listLBs(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	opts := loadbalancers.ListOpts{
		Name: clusterName,
	}
	lbs, err := cloud.ListLBs(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list lbs: %s", err)
	}

	var rts []*resources.Resource
	for _, t := range lbs {
		rt := &resources.Resource{
			Name: t.Name,
			ID:   t.ID,
			Type: typeLB,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				opts := loadbalancers.DeleteOpts{
					Cascade: true,
				}
				return loadbalancers.Delete(cloud.(openstack.OpenstackCloud).LoadBalancerClient(), t.ID, opts).ExtractErr()
			},
			Obj: lbs,
		}
		rts = append(rts, rt)
	}

	return rts, nil
}
