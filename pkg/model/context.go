/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/apis/kops"
	"strings"
	"fmt"
)

type KopsModelContext struct {
	Cluster        *kops.Cluster
	InstanceGroups []*kops.InstanceGroup

	SSHPublicKeys  [][]byte
}


// Will attempt to calculate a meaningful name for an ELB given a prefix
// Will never return a string longer than 32 chars
func (m *KopsModelContext) GetELBName32(prefix string) (string, error) {
	var returnString string
	c := m.Cluster.ObjectMeta.Name
	s := strings.Split(c, ".")

	// TODO: strings.Split cannot return empty
	if len(s) > 0 {
		returnString = fmt.Sprintf("%s-%s", prefix, s[0])
	} else {
		returnString = fmt.Sprintf("%s-%s", prefix, c)
	}
	if len(returnString) > 32 {
		returnString = returnString[:32]
	}
	return returnString, nil
}

func (m*KopsModelContext) ClusterName() string {
	return m.Cluster.ObjectMeta.Name
}

// GatherSubnets maps the subnet names in an InstanceGroup to the ClusterSubnetSpec objects (which are stored on the Cluster)
func (m*KopsModelContext) GatherSubnets(ig *kops.InstanceGroup) ([]*kops.ClusterSubnetSpec, error) {
	var subnets []*kops.ClusterSubnetSpec
	for _, subnetName := range ig.Spec.Subnets {
		var matches []*kops.ClusterSubnetSpec
		for i := range m.Cluster.Spec.Subnets {
			clusterSubnet := &m.Cluster.Spec.Subnets[i]
			if clusterSubnet.SubnetName == subnetName {
				matches = append(matches, clusterSubnet)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("subnet not found: %q", subnetName)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("found multiple subnets with name: %q", subnetName)
		}
		subnets = append(subnets, matches[0])
	}
	return subnets, nil
}


// FindInstanceGroup returns the instance group with the matching Name (or nil if not found)
func (m*KopsModelContext) FindInstanceGroup(name string) (*kops.InstanceGroup) {
	for _, ig := range m.InstanceGroups {
		if ig.ObjectMeta.Name == name {
			return ig
		}
	}
	return nil
}


// FindSubnet returns the subnet with the matching Name (or nil if not found)
func (m*KopsModelContext) FindSubnet(name string) (*kops.ClusterSubnetSpec) {
	for  i := range m.Cluster.Spec.Subnets {
		s := &m.Cluster.Spec.Subnets[i]
		if s.SubnetName == name {
			return s
		}
	}
	return nil
}