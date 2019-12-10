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
	"sort"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/kops/pkg/apis/kops"
)

const (
	openstackExternalIPType = "OS-EXT-IPS:type"
	openstackAddressFixed   = "fixed"
	openstackAddress        = "addr"
)

type flavorList []flavors.Flavor

func (s flavorList) Len() int {
	return len(s)
}

func (s flavorList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s flavorList) Less(i, j int) bool {
	if s[i].VCPUs < s[j].VCPUs {
		return true
	}
	if s[i].VCPUs > s[j].VCPUs {
		return false
	}
	return s[i].RAM < s[j].RAM
}

func (c *openstackCloud) DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	flavorPage, err := flavors.ListDetail(c.ComputeClient(), flavors.ListOpts{
		MinRAM: 1024,
	}).AllPages()
	if err != nil {
		return "", fmt.Errorf("Could not list flavors: %v", err)
	}
	var fList flavorList
	fList, err = flavors.ExtractFlavors(flavorPage)
	if err != nil {
		return "", fmt.Errorf("Could not extract flavors: %v", err)
	}
	sort.Sort(&fList)

	var candidates flavorList
	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		// Requirements based on awsCloudImplementation.DefaultInstanceType
		for _, flavor := range fList {
			if flavor.RAM >= 4096 && flavor.VCPUs >= 1 {
				candidates = append(candidates, flavor)
			}
		}

	case kops.InstanceGroupRoleNode:
		for _, flavor := range fList {
			if flavor.RAM >= 4096 && flavor.VCPUs >= 2 {
				candidates = append(candidates, flavor)
			}
		}

	case kops.InstanceGroupRoleBastion:
		for _, flavor := range fList {
			if flavor.RAM >= 1024 {
				candidates = append(candidates, flavor)
			}
		}

	default:
		return "", fmt.Errorf("unhandled role %q", ig.Spec.Role)
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("No suitable flavor for role %q", ig.Spec.Role)
	}
	return candidates[0].Name, nil
}

func GetServerFixedIP(server *servers.Server, interfaceName string) (poolAddress string, err error) {
	if localAddr, ok := server.Addresses[interfaceName]; ok {

		if localAddresses, ok := localAddr.([]interface{}); ok {
			for _, addr := range localAddresses {
				addrMap := addr.(map[string]interface{})
				if addrType, ok := addrMap[openstackExternalIPType]; ok && addrType == openstackAddressFixed {
					if fixedIP, ok := addrMap[openstackAddress]; ok {
						if fixedIPStr, ok := fixedIP.(string); ok {
							poolAddress = fixedIPStr
						} else {
							err = fmt.Errorf("Fixed IP was not a string: %v", fixedIP)
						}
					} else {
						err = fmt.Errorf("Type fixed did not contain addr: %v", addr)
					}
				}
			}
		}
	} else {
		err = fmt.Errorf("server `%s` interface name `%s` not found", server.ID, interfaceName)
	}
	return poolAddress, err
}
