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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
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

type networkListResponse struct {
	Networks []externalNetwork `json:"networks"`
}

type networkGetResponse struct {
	Network externalNetwork `json:"network"`
}

type networkCreateRequest struct {
	Network externalNetwork `json:"network"`
}

func (m *MockClient) mockNetworks() {
	re := regexp.MustCompile(`/networks/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		networkID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if networkID == "" {
				r.ParseForm()
				m.listNetworks(w, r.Form)
			} else {
				m.getNetwork(w, networkID)
			}
		case http.MethodPut:
			m.tagNetwork(w, r)
		case http.MethodPost:
			m.createNetwork(w, r)
		case http.MethodDelete:
			m.deleteNetwork(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/networks/", handler)
	m.Mux.HandleFunc("/networks", handler)
}

func (m *MockClient) listNetworks(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	networks := make([]externalNetwork, 0)
	nameFilter := vals.Get("name")
	for _, n := range m.networks {
		if nameFilter != "" && nameFilter != n.Name {
			continue
		}
		networks = append(networks, n)
	}

	resp := networkListResponse{
		Networks: networks,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}

func (m *MockClient) getNetwork(w http.ResponseWriter, networkID string) {
	if network, ok := m.networks[networkID]; ok {
		resp := networkGetResponse{
			Network: network,
		}
		respB, err := json.Marshal(resp)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %+v", resp))
		}
		_, err = w.Write(respB)
		if err != nil {
			panic("failed to write body")
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 && len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	networkID := parts[1]
	if _, ok := m.networks[networkID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if len(parts) == 4 && parts[3] == "tags" {
		// /networks/<networkid>/tags/<tag>
		tagToDelete := parts[3]
		network := m.networks[networkID]
		tags := make([]string, 0)
		for _, tag := range network.Tags {
			if tag != tagToDelete {
				tags = append(tags, tag)
			}
		}
		network.Tags = tags
	} else {
		// /networks/<networkid>
		delete(m.networks, networkID)
	}
	w.WriteHeader(http.StatusOK)
}

func (m *MockClient) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create networkCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create network request")
	}

	w.WriteHeader(http.StatusAccepted)

	n := externalNetwork{
		ID:           uuid.New().String(),
		Name:         create.Network.Name,
		AdminStateUp: create.Network.AdminStateUp,
		External:     create.Network.External,
		Tags:         make([]string, 0),
	}
	m.networks[n.ID] = n

	resp := networkGetResponse{
		Network: n,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}

func (m *MockClient) tagNetwork(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	networkID := parts[1]
	tag := parts[3]

	if _, ok := m.networks[networkID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	network := m.networks[networkID]
	network.Tags = append(network.Tags, tag)
	m.networks[networkID] = network

	w.WriteHeader(http.StatusCreated)
}
