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

package mockcompute

import (
	"context"
	"fmt"
	"sync"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type addressClient struct {
	// addrs are addresses keyed by project, region, and address name.
	addrs map[string]map[string]map[string]*compute.Address
	sync.Mutex
}

var _ gce.AddressClient = &addressClient{}

func newAddressClient() *addressClient {
	return &addressClient{
		addrs: map[string]map[string]map[string]*compute.Address{},
	}
}

func (c *addressClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, regions := range c.addrs {
		for _, addrs := range regions {
			for n, a := range addrs {
				m[n] = a
			}
		}
	}
	return m
}

func (c *addressClient) Insert(project, region string, addr *compute.Address) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.addrs[project]
	if !ok {
		regions = map[string]map[string]*compute.Address{}
		c.addrs[project] = regions
	}
	addrs, ok := regions[region]
	if !ok {
		addrs = map[string]*compute.Address{}
		regions[region] = addrs
	}
	addr.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/addrs/%s", project, region, addr.Name)
	addrs[addr.Name] = addr
	return doneOperation(), nil
}

func (c *addressClient) Delete(project, region, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.addrs[project]
	if !ok {
		return nil, notFoundError()
	}
	addrs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := addrs[name]; !ok {
		return nil, notFoundError()
	}
	delete(addrs, name)
	return doneOperation(), nil
}

func (c *addressClient) Get(project, region, name string) (*compute.Address, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.addrs[project]
	if !ok {
		return nil, notFoundError()
	}
	addrs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	addr, ok := addrs[name]
	if !ok {
		return nil, notFoundError()
	}
	return addr, nil
}

func (c *addressClient) List(ctx context.Context, project, region string) ([]*compute.Address, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.addrs[project]
	if !ok {
		return nil, nil
	}
	addrs, ok := regions[region]
	if !ok {
		return nil, nil
	}
	var l []*compute.Address
	for _, a := range addrs {
		l = append(l, a)
	}
	return l, nil
}

func (c *addressClient) ListWithFilter(project, region, filter string) ([]*compute.Address, error) {
	return c.List(context.Background(), project, region)
}
