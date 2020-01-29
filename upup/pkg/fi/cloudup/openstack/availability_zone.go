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

	"github.com/gophercloud/gophercloud"
	az "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) ListAvailabilityZones(serviceClient *gophercloud.ServiceClient) (azList []az.AvailabilityZone, err error) {

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		azPage, err := az.List(serviceClient).AllPages()

		if err != nil {
			return false, fmt.Errorf("failed to list storage availability zones: %v", err)
		}
		azList, err = az.ExtractAvailabilityZones(azPage)
		if err != nil {
			return false, fmt.Errorf("failed to extract storage availability zones: %v", err)
		}
		return true, nil
	})
	if !done {

		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return azList, err
	}
	return azList, nil
}

func (c *openstackCloud) GetStorageAZFromCompute(computeAZ string) (*az.AvailabilityZone, error) {
	// TODO: This is less than desirable, but openstack differs here
	// Check to see if the availability zone exists.
	azList, err := c.ListAvailabilityZones(c.BlockStorageClient())
	if err != nil {
		return nil, fmt.Errorf("Volume.RenderOpenstack: %v", err)
	}
	for _, az := range azList {
		if az.ZoneName == computeAZ {
			return &az, nil
		}
	}
	// Determine if there is a meaningful storage AZ here
	if len(azList) == 1 {
		return &azList[0], nil
	}
	return nil, fmt.Errorf("no decernable storage availability zone could be mapped to compute availability zone %s", computeAZ)
}
