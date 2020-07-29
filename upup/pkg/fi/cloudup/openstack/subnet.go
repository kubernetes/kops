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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) ListSubnets(opt subnets.ListOptsBuilder) ([]subnets.Subnet, error) {
	return listSubnets(c, opt)
}

func listSubnets(c OpenstackCloud, opt subnets.ListOptsBuilder) ([]subnets.Subnet, error) {
	var s []subnets.Subnet

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := subnets.List(c.NetworkingClient(), opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing subnets: %v", err)
		}

		r, err := subnets.ExtractSubnets(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting subnets from pages: %v", err)
		}
		s = r
		return true, nil
	})
	if err != nil {
		return s, err
	} else if done {
		return s, nil
	} else {
		return s, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetSubnet(subnetID string) (*subnets.Subnet, error) {
	return getSubnet(c, subnetID)
}

func getSubnet(c OpenstackCloud, subnetID string) (*subnets.Subnet, error) {
	var subnet *subnets.Subnet
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		sub, err := subnets.Get(c.NetworkingClient(), subnetID).Extract()
		if err != nil {
			return false, fmt.Errorf("error retrieving subnet: %v", err)
		}
		subnet = sub
		return true, nil
	})
	if err != nil {
		return nil, err
	} else if done {
		return subnet, nil
	} else {
		return nil, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateSubnet(opt subnets.CreateOptsBuilder) (*subnets.Subnet, error) {
	return createSubnet(c, opt)
}

func createSubnet(c OpenstackCloud, opt subnets.CreateOptsBuilder) (*subnets.Subnet, error) {
	var s *subnets.Subnet

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := subnets.Create(c.NetworkingClient(), opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating subnet: %v", err)
		}
		s = v
		return true, nil
	})
	if err != nil {
		return s, err
	} else if done {
		return s, nil
	} else {
		return s, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) DeleteSubnet(subnetID string) error {
	return deleteSubnet(c, subnetID)
}

func deleteSubnet(c OpenstackCloud, subnetID string) error {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err := subnets.Delete(c.NetworkingClient(), subnetID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting subnet: %v", err)
		}
		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetExternalSubnet() (subnet *subnets.Subnet, err error) {
	return getExternalSubnet(c, c.extSubnetName)
}

func getExternalSubnet(c OpenstackCloud, subnetName *string) (subnet *subnets.Subnet, err error) {
	if subnetName == nil {
		return nil, nil
	}

	subnets, err := c.ListSubnets(subnets.ListOpts{
		Name: fi.StringValue(subnetName),
	})
	if err != nil {
		return nil, err
	}

	if len(subnets) == 1 {
		return &subnets[0], nil
	}
	return nil, fmt.Errorf("did not find floatingsubnet for external router")
}

func (c *openstackCloud) GetLBFloatingSubnet() (subnet *subnets.Subnet, err error) {
	return getLBFloatingSubnet(c, c.floatingSubnet)
}

func getLBFloatingSubnet(c OpenstackCloud, floatingSubnet *string) (subnet *subnets.Subnet, err error) {
	if floatingSubnet == nil {
		return nil, nil
	}

	subnets, err := c.ListSubnets(subnets.ListOpts{
		Name: fi.StringValue(floatingSubnet),
	})
	if err != nil {
		return nil, err
	}

	if len(subnets) == 1 {
		return &subnets[0], nil
	}
	return nil, fmt.Errorf("did not find floatingsubnet for LB")
}
