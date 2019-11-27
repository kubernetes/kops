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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/attributestags"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/pagination"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) AppendTag(resource string, id string, tag string) error {
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		err := attributestags.Add(c.neutronClient, resource, id, tag).ExtractErr()
		if err != nil {
			return false, fmt.Errorf("error appending tag %s: %v", tag, err)
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

func (c *openstackCloud) DeleteTag(resource string, id string, tag string) error {
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		err := attributestags.Delete(c.neutronClient, resource, id, tag).ExtractErr()
		if err != nil {
			return false, fmt.Errorf("error deleting tag %s: %v", tag, err)
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

func (c *openstackCloud) FindNetworkBySubnetID(subnetID string) (*networks.Network, error) {
	var rslt *networks.Network
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		subnet, err := c.GetSubnet(subnetID)
		if err != nil {
			return false, fmt.Errorf("error retrieving subnet with id %s: %v", subnetID, err)
		}

		netID := subnet.NetworkID
		net, err := c.GetNetwork(netID)
		if err != nil {
			return false, fmt.Errorf("error retrieving network with id %s: %v", netID, err)
		}
		rslt = net
		return true, nil
	})
	if err != nil {
		return nil, err
	} else if done {
		return rslt, nil
	} else {
		return nil, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetNetwork(id string) (*networks.Network, error) {
	var network *networks.Network
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		r, err := networks.Get(c.neutronClient, id).Extract()
		if err != nil {
			return false, fmt.Errorf("error retrieving network with id %s: %v", id, err)
		}
		network = r
		return true, nil
	})
	if err != nil {
		return network, err
	} else if done {
		return network, nil
	} else {
		return network, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListNetworks(opt networks.ListOptsBuilder) ([]networks.Network, error) {
	var ns []networks.Network

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := networks.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing networks: %v", err)
		}

		r, err := networks.ExtractNetworks(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting networks from pages: %v", err)
		}
		ns = r
		return true, nil
	})
	if err != nil {
		return ns, err
	} else if done {
		return ns, nil
	} else {
		return ns, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetExternalNetwork() (net *networks.Network, err error) {
	type NetworkWithExternalExt struct {
		networks.Network
		external.NetworkExternalExt
	}

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {

		err = networks.List(c.NetworkingClient(), networks.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
			var externalNetwork []NetworkWithExternalExt
			err := networks.ExtractNetworksInto(page, &externalNetwork)
			if err != nil {
				return false, err
			}
			for _, externalNet := range externalNetwork {
				if externalNet.External && externalNet.Name == fi.StringValue(c.extNetworkName) {
					net = &externalNet.Network
					return true, nil
				}
			}
			return true, nil
		})
		if err != nil {
			return false, nil
		}
		return net != nil, nil
	})

	if err != nil {
		return net, err
	} else if done {
		return net, nil
	} else {
		return net, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateNetwork(opt networks.CreateOptsBuilder) (*networks.Network, error) {
	var n *networks.Network

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		r, err := networks.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating network: %v", err)
		}
		n = r
		return true, nil
	})
	if err != nil {
		return n, err
	} else if done {
		return n, nil
	} else {
		return n, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) DeleteNetwork(networkID string) error {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err := networks.Delete(c.neutronClient, networkID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting network: %v", err)
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
