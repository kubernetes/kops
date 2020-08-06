/*
Copyright 2020 The Kubernetes Authors.

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

package mockloadbalancer

import (
	"net/http/httptest"
	"sync"

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"k8s.io/kops/cloudmock/openstack"
)

// MockClient represents a mocked networks (nebula) client
type MockClient struct {
	openstack.MockOpenstackServer
	mutex sync.Mutex

	loadbalancers map[string]loadbalancers.LoadBalancer
	listeners     map[string]listeners.Listener
	pools         map[string]pools.Pool
}

// CreateClient will create a new mock networking client
func CreateClient() *MockClient {
	m := &MockClient{}
	m.Reset()
	m.SetupMux()
	m.mockListeners()
	m.mockLoadBalancers()
	m.mockPools()
	m.Server = httptest.NewServer(m.Mux)
	return m
}

// Reset will empty the state of the mock data
func (m *MockClient) Reset() {
	m.loadbalancers = make(map[string]loadbalancers.LoadBalancer)
	m.listeners = make(map[string]listeners.Listener)
	m.pools = make(map[string]pools.Pool)
}

// All returns a map of all resource IDs to their resources
func (m *MockClient) All() map[string]interface{} {
	all := make(map[string]interface{})
	for id, sg := range m.loadbalancers {
		all[id] = sg
	}
	for id, l := range m.listeners {
		all[id] = l
	}
	for id, p := range m.pools {
		all[id] = p
	}
	return all
}
