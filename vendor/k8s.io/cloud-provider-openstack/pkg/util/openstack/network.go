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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

// GetFloatingIPs returns all the filtered floating IPs
func GetFloatingIPs(client *gophercloud.ServiceClient, opts floatingips.ListOpts) ([]floatingips.FloatingIP, error) {
	var floatingIPList []floatingips.FloatingIP

	allPages, err := floatingips.List(client, opts).AllPages()
	if err != nil {
		return floatingIPList, err
	}
	floatingIPList, err = floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return floatingIPList, err
	}

	return floatingIPList, nil
}

// GetFloatingIPByPortID get the floating IP of the given port.
func GetFloatingIPByPortID(client *gophercloud.ServiceClient, portID string) (*floatingips.FloatingIP, error) {
	opt := floatingips.ListOpts{
		PortID: portID,
	}
	ips, err := GetFloatingIPs(client, opt)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, nil
	}

	return &ips[0], nil
}
