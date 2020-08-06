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
	"net/http"
)

// Theres no type that represents a network with the "external" extension
// Copied from https://github.com/gophercloud/gophercloud/blob/bd999d0da882fe8c5b0077b7af2dcc019c1ab458/openstack/networking/v2/networks/results.go#L51
type externalNetwork struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	AdminStateUp bool     `json:"admin_state_up"`
	Status       string   `json:"status"`
	External     bool     `json:"router:external"`
	Tags         []string `json:"tags"`
}

func (m *MockClient) mockNetworks() {

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
	m.Mux.HandleFunc("/networks/", handler)
	m.Mux.HandleFunc("/networks", handler)
}
