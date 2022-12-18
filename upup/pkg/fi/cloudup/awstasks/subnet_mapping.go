/*
Copyright 2021 The Kubernetes Authors.

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

package awstasks

import (
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
)

type SubnetMapping struct {
	Subnet *Subnet

	// PrivateIPv4Address only valid for NLBs
	PrivateIPv4Address *string
	// AllocationID only valid for NLBs
	AllocationID *string
}

// OrderSubnetsById implements sort.Interface for []Subnet, based on ID
type OrderSubnetMappingsByID []*SubnetMapping

func (a OrderSubnetMappingsByID) Len() int      { return len(a) }
func (a OrderSubnetMappingsByID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderSubnetMappingsByID) Less(i, j int) bool {
	v1 := fi.ValueOf(a[i].Subnet.ID)
	v2 := fi.ValueOf(a[j].Subnet.ID)
	if v1 == v2 {
		if a[i].PrivateIPv4Address != nil && a[j].PrivateIPv4Address != nil {
			return fi.ValueOf(a[i].PrivateIPv4Address) < fi.ValueOf(a[j].PrivateIPv4Address)
		}
		if a[i].AllocationID != nil && a[j].AllocationID != nil {
			return fi.ValueOf(a[i].AllocationID) < fi.ValueOf(a[j].AllocationID)
		}
	}
	return v1 < v2
}

// OrderSubnetMappingsByName implements sort.Interface for []Subnet, based on Name
type OrderSubnetMappingsByName []*SubnetMapping

func (a OrderSubnetMappingsByName) Len() int      { return len(a) }
func (a OrderSubnetMappingsByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderSubnetMappingsByName) Less(i, j int) bool {
	v1 := fi.ValueOf(a[i].Subnet.Name)
	v2 := fi.ValueOf(a[j].Subnet.Name)
	return v1 < v2
}

func subnetMappingSlicesEqualIgnoreOrder(l, r []*SubnetMapping) bool {
	lBySubnet := make(map[string]*SubnetMapping)
	for _, s := range l {
		lBySubnet[*s.Subnet.ID] = s
	}
	rBySubnet := make(map[string]*SubnetMapping)
	for _, s := range r {
		if s.Subnet == nil || s.Subnet.ID == nil {
			klog.V(4).Infof("Subnet ID not set; returning not-equal: %v", s)
			return false
		}
		rBySubnet[*s.Subnet.ID] = s
	}
	if len(lBySubnet) != len(rBySubnet) {
		return false
	}

	for n, s := range lBySubnet {
		s2, ok := rBySubnet[n]
		if !ok {
			return false
		}
		if fi.ValueOf(s.PrivateIPv4Address) != fi.ValueOf(s2.PrivateIPv4Address) {
			return false
		}
		if fi.ValueOf(s.AllocationID) != fi.ValueOf(s2.AllocationID) {
			return false
		}
	}
	return true
}

func (e *SubnetMapping) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Subnet); ok {
			deps = append(deps, task)
		}
	}
	return deps
}
