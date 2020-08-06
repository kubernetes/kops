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

package mocknetworking

import (
	"net/http/httptest"
	"sync"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/kops/cloudmock/openstack"
)

// MockClient represents a mocked networks (nebula) client
type MockClient struct {
	openstack.MockOpenstackServer
	mutex sync.Mutex

	networks           map[string]externalNetwork
	ports              map[string]ports.Port
	routers            map[string]routers.Router
	routerInterfaces   map[string][]routers.InterfaceInfo
	securityGroups     map[string]groups.SecGroup
	securityGroupRules map[string]rules.SecGroupRule
	subnets            map[string]subnets.Subnet
}

// CreateClient will create a new mock networking client
func CreateClient() *MockClient {
	m := &MockClient{}
	m.Reset()
	m.SetupMux()
	m.mockNetworks()
	m.mockPorts()
	m.mockRouters()
	m.mockSecurityGroups()
	m.mockSecurityGroupRules()
	m.mockSubnets()
	m.Server = httptest.NewServer(m.Mux)
	return m
}

// Reset will empty the state of the mock data
func (m *MockClient) Reset() {
	m.networks = make(map[string]externalNetwork)
	m.ports = make(map[string]ports.Port)
	m.routers = make(map[string]routers.Router)
	m.routerInterfaces = make(map[string][]routers.InterfaceInfo)
	m.securityGroups = make(map[string]groups.SecGroup)
	m.securityGroupRules = make(map[string]rules.SecGroupRule)
	m.subnets = make(map[string]subnets.Subnet)
}

// All returns a map of all resource IDs to their resources
func (m *MockClient) All() map[string]interface{} {
	all := make(map[string]interface{})
	for id, n := range m.networks {
		all[id] = n
	}
	for id, p := range m.ports {
		all[id] = p
	}
	for id, r := range m.routers {
		all[id] = r
	}
	for id, ri := range m.routerInterfaces {
		all[id] = ri
	}
	for id, sg := range m.securityGroups {
		all[id] = sg
	}
	for id, sgr := range m.securityGroupRules {
		all[id] = sgr
	}
	for id, s := range m.subnets {
		all[id] = s
	}
	return all
}
