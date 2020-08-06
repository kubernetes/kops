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

package mockcompute

import (
	"net/http/httptest"
	"sync"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"k8s.io/kops/cloudmock/openstack"
)

// MockClient represents a mocked networks (nebula) client
type MockClient struct {
	openstack.MockOpenstackServer
	mutex sync.Mutex

	serverGroups map[string]servergroups.ServerGroup
	servers      map[string]servers.Server
	keyPairs     map[string]keypairs.KeyPair
	images       map[string]images.Image
	flavors      map[string]flavors.Flavor
}

// CreateClient will create a new mock networking client
func CreateClient() *MockClient {
	m := &MockClient{}
	m.SetupMux()
	m.Reset()
	m.mockServerGroups()
	m.mockServers()
	m.mockKeyPairs()
	m.mockImages()
	m.mockFlavors()
	m.Server = httptest.NewServer(m.Mux)
	return m
}

// Reset will empty the state of the mock data
func (m *MockClient) Reset() {
	m.serverGroups = make(map[string]servergroups.ServerGroup)
	m.servers = make(map[string]servers.Server)
	m.keyPairs = make(map[string]keypairs.KeyPair)
	m.images = make(map[string]images.Image)
	m.flavors = make(map[string]flavors.Flavor)
}

// All returns a map of all resource IDs to their resources
func (m *MockClient) All() map[string]interface{} {
	all := make(map[string]interface{})
	for id, sg := range m.serverGroups {
		all[id] = sg
	}
	for id, kp := range m.keyPairs {
		all[id] = kp
	}
	return all
}
