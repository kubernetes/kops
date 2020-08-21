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

package mockimage

import (
	"net/http/httptest"
	"sync"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"k8s.io/kops/cloudmock/openstack"
)

// MockClient represents a mocked networks (nebula) client
type MockClient struct {
	openstack.MockOpenstackServer
	mutex sync.Mutex

	images map[string]images.Image
}

// CreateClient will create a new mock networking client
func CreateClient() *MockClient {
	m := &MockClient{}
	m.SetupMux()
	m.Reset()
	m.mockImages()
	m.Server = httptest.NewServer(m.Mux)
	return m
}

// Reset will empty the state of the mock data
func (m *MockClient) Reset() {
	m.images = make(map[string]images.Image)
}

// All returns a map of all resource IDs to their resources
func (m *MockClient) All() map[string]interface{} {
	all := make(map[string]interface{})
	for id, i := range m.images {
		all[id] = i
	}
	return all
}
