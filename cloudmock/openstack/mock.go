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

package openstack

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gophercloud/gophercloud"
)

type MockOpenstackServer struct {
	Mux *http.ServeMux

	Server *httptest.Server
}

// SetupMux prepares the Mux and Server.
func (m *MockOpenstackServer) SetupMux() {
	m.Mux = http.NewServeMux()
	m.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		panic(fmt.Sprintf("Unhandled mock request: %+v\n", r))
	})
}

// TeardownHTTP releases HTTP-related resources.
func (m *MockOpenstackServer) TeardownHTTP() {
	m.Server.Close()
}

func (m *MockOpenstackServer) ServiceClient() *gophercloud.ServiceClient {
	return &gophercloud.ServiceClient{
		ProviderClient: &gophercloud.ProviderClient{},
		Endpoint:       m.Server.URL + "/",
	}
}
