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

	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) GetL3FloatingIP(id string) (fip *l3floatingip.FloatingIP, err error) {
	return getL3FloatingIP(c, id)
}

func getL3FloatingIP(c OpenstackCloud, id string) (fip *l3floatingip.FloatingIP, err error) {
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {

		fip, err = l3floatingip.Get(c.NetworkingClient(), id).Extract()
		if err != nil {
			return false, fmt.Errorf("GetFloatingIP: fetching floating IP (%s) failed: %v", id, err)
		}
		return true, nil
	})
	if !done {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return fip, err
	}
	return fip, nil
}

func (c *openstackCloud) CreateL3FloatingIP(opts l3floatingip.CreateOpts) (fip *l3floatingip.FloatingIP, err error) {
	return createL3FloatingIP(c, opts)
}

func createL3FloatingIP(c OpenstackCloud, opts l3floatingip.CreateOpts) (fip *l3floatingip.FloatingIP, err error) {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {

		fip, err = l3floatingip.Create(c.NetworkingClient(), opts).Extract()
		if err != nil {
			return false, fmt.Errorf("CreateL3FloatingIP: create L3 floating IP failed: %v", err)
		}
		return true, nil
	})
	if !done {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return fip, err
	}
	return fip, nil
}

func (c *openstackCloud) ListFloatingIPs() (fips []floatingips.FloatingIP, err error) {
	return listFloatingIPs(c)
}

func listFloatingIPs(c OpenstackCloud) (fips []floatingips.FloatingIP, err error) {

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		pages, err := floatingips.List(c.ComputeClient()).AllPages()
		if err != nil {
			return false, fmt.Errorf("failed to list floating ip: %v", err)
		}
		fips, err = floatingips.ExtractFloatingIPs(pages)
		if err != nil {
			return false, fmt.Errorf("failed to extract floating ip: %v", err)
		}
		return true, nil
	})
	if !done {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return fips, err
	}
	return fips, nil
}

func (c *openstackCloud) ListL3FloatingIPs(opts l3floatingip.ListOpts) (fips []l3floatingip.FloatingIP, err error) {
	return listL3FloatingIPs(c, opts)
}

func listL3FloatingIPs(c OpenstackCloud, opts l3floatingip.ListOpts) (fips []l3floatingip.FloatingIP, err error) {

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		page, err := l3floatingip.List(c.NetworkingClient(), opts).AllPages()
		if err != nil {
			return false, fmt.Errorf("failed to list L3 floating ip: %v", err)
		}
		fips, err = l3floatingip.ExtractFloatingIPs(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract L3 floating ip: %v", err)
		}
		return true, nil
	})
	if !done {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return fips, err
	}
	return fips, nil
}

func (c *openstackCloud) DeleteFloatingIP(id string) (err error) {
	return deleteFloatingIP(c, id)
}

func deleteFloatingIP(c OpenstackCloud, id string) (err error) {

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err = l3floatingip.Delete(c.ComputeClient(), id).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("failed to delete floating ip %s: %v", id, err)
		}
		return true, nil
	})
	if !done && err == nil {
		err = wait.ErrWaitTimeout
	}
	return err
}

func (c *openstackCloud) DeleteL3FloatingIP(id string) (err error) {
	return deleteL3FloatingIP(c, id)
}

func deleteL3FloatingIP(c OpenstackCloud, id string) (err error) {

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err = l3floatingip.Delete(c.NetworkingClient(), id).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("failed to delete L3 floating ip %s: %v", id, err)
		}
		return true, nil
	})
	if !done && err == nil {
		err = wait.ErrWaitTimeout
	}
	return err
}
