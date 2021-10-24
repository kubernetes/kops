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
	"fmt"
	"sync"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type networkClient struct {
	// networks are networks keyed by project and network name.
	networks map[string]map[string]*compute.Network
	sync.Mutex
}

var _ gce.NetworkClient = &networkClient{}

func newNetworkClient() *networkClient {
	return &networkClient{
		networks: map[string]map[string]*compute.Network{},
	}
}

func (c *networkClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, nws := range c.networks {
		for n, nw := range nws {
			m[n] = nw
		}
	}
	return m
}

func (c *networkClient) Insert(project string, network *compute.Network) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	networks, ok := c.networks[project]
	if !ok {
		networks = map[string]*compute.Network{}
		c.networks[project] = networks
	}
	network.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, network.Name)
	networks[network.Name] = network
	return doneOperation(), nil
}

func (c *networkClient) Delete(project, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	networks, ok := c.networks[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := networks[name]; !ok {
		return nil, notFoundError()
	}
	delete(networks, name)
	return doneOperation(), nil
}

func (c *networkClient) Get(project, name string) (*compute.Network, error) {
	c.Lock()
	defer c.Unlock()
	networks, ok := c.networks[project]
	if !ok {
		return nil, notFoundError()
	}
	network, ok := networks[name]
	if !ok {
		return nil, notFoundError()
	}
	return network, nil
}

func (c *networkClient) List(project string) (*compute.NetworkList, error) {
	c.Lock()
	defer c.Unlock()
	networks, ok := c.networks[project]
	if !ok {
		return nil, notFoundError()
	}
	networkList := &compute.NetworkList{}
	for _, network := range networks {
		networkList.Items = append(networkList.Items, network)
	}
	return networkList, nil
}
