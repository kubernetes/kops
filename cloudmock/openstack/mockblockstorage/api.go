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

package mockblockstorage

import (
	"net/http/httptest"
	"sync"

	cinderv3 "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
	"k8s.io/kops/cloudmock/openstack"
)

// MockClient represents a mocked blockstorage (cinderv3) client
type MockClient struct {
	openstack.MockOpenstackServer
	mutex sync.Mutex

	volumes           map[string]cinderv3.Volume
	availabilityZones map[string]availabilityzones.AvailabilityZone
}

// CreateClient will create a new mock blockstorage client
func CreateClient() *MockClient {
	m := &MockClient{}
	m.Reset()
	m.SetupMux()
	m.mockVolumes()
	m.mockAvailabilityZones()
	m.Server = httptest.NewServer(m.Mux)
	return m
}

// Reset will empty the state of the mock data
func (m *MockClient) Reset() {
	m.volumes = make(map[string]cinderv3.Volume)
	m.availabilityZones = make(map[string]availabilityzones.AvailabilityZone)
}
