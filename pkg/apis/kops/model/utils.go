/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
)

// FindSubnet returns the subnet with the specified name, or returns nil
func FindSubnet(c *kops.Cluster, subnetName string) *kops.ClusterSubnetSpec {
	for i := range c.Spec.Subnets {
		if c.Spec.Subnets[i].Name == subnetName {
			return &c.Spec.Subnets[i]
		}
	}
	return nil
}

// FindZonesForInstanceGroup computes the zones for an instance group, which are the zones directly declared in the InstanceGroup, or the subnet zones
func FindZonesForInstanceGroup(c *kops.Cluster, ig *kops.InstanceGroup) ([]string, error) {
	zones := sets.NewString(ig.Spec.Zones...)
	for _, subnetName := range ig.Spec.Subnets {
		subnet := FindSubnet(c, subnetName)
		if subnet == nil {
			return nil, fmt.Errorf("cannot find subnet %q (declared in instance group %q, not found in cluster)", subnetName, ig.ObjectMeta.Name)
		}

		if subnet.Zone != "" {
			zones.Insert(subnet.Zone)
		}
	}
	return zones.List(), nil
}
