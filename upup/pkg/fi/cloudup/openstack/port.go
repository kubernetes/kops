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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) CreatePort(opt ports.CreateOptsBuilder) (*ports.Port, error) {
	var p *ports.Port

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := ports.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating port: %v", err)
		}
		p = v
		return true, nil
	})
	if err != nil {
		return p, err
	} else if done {
		return p, nil
	} else {
		return p, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetPort(id string) (*ports.Port, error) {
	var p *ports.Port

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		port, err := ports.Get(c.neutronClient, id).Extract()
		if err != nil {
			return false, err
		}
		p = port
		return true, nil
	})
	if err != nil {
		return p, err
	} else if done {
		return p, nil
	} else {
		return p, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListPorts(opt ports.ListOptsBuilder) ([]ports.Port, error) {
	var p []ports.Port

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := ports.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing ports: %v", err)
		}

		r, err := ports.ExtractPorts(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting ports from pages: %v", err)
		}
		p = r
		return true, nil
	})
	if err != nil {
		return p, err
	} else if done {
		return p, nil
	} else {
		return p, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) DeletePort(portID string) error {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err := ports.Delete(c.neutronClient, portID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting port: %v", err)
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
