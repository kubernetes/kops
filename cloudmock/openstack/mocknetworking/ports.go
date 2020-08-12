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
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type portListResponse struct {
	Ports []ports.Port `json:"ports"`
}

type portGetResponse struct {
	Port ports.Port `json:"port"`
}

type portCreateRequest struct {
	Port ports.CreateOpts `json:"port"`
}

func (m *MockClient) mockPorts() {
	re := regexp.MustCompile(`/ports/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		portID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if portID == "" {
				r.ParseForm()
				m.listPorts(w, r.Form)
			} else {
				m.getPort(w, portID)
			}
		case http.MethodPut:
			m.tagPort(w, r)
		case http.MethodPost:
			m.createPort(w, r)
		case http.MethodDelete:
			m.deletePort(w, portID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/ports/", handler)
	m.Mux.HandleFunc("/ports", handler)
}

func (m *MockClient) listPorts(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	ports := make([]ports.Port, 0)
	nameFilter := vals.Get("name")
	idFilter := vals.Get("id")
	networkFilter := vals.Get("network_id")
	deviceFilter := vals.Get("device_id")
	for _, p := range m.ports {
		if nameFilter != "" && nameFilter != p.Name {
			continue
		}
		if deviceFilter != "" && deviceFilter != p.DeviceID {
			continue
		}
		if networkFilter != "" && networkFilter != p.NetworkID {
			continue
		}
		if idFilter != "" && idFilter != p.ID {
			continue
		}
		ports = append(ports, p)
	}

	resp := portListResponse{
		Ports: ports,
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

func (m *MockClient) getPort(w http.ResponseWriter, portID string) {
	if port, ok := m.ports[portID]; ok {
		resp := portGetResponse{
			Port: port,
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

func (m *MockClient) deletePort(w http.ResponseWriter, portID string) {
	if _, ok := m.ports[portID]; ok {
		delete(m.ports, portID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createPort(w http.ResponseWriter, r *http.Request) {
	var create portCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create port request")
	}

	fixedIPs := make([]ports.IP, 0)
	// The request type uses a []interface{map[string]interface{}]} to represent []ports.IP
	if rawIPs, ok := create.Port.FixedIPs.([]interface{}); ok {
		for _, rawFixedIP := range rawIPs {
			if rawIP, ok := rawFixedIP.(map[string]interface{}); ok {
				if subnetID, ok := rawIP["subnet_id"]; ok {
					if subnet, ok := subnetID.(string); ok {
						fixedIPs = append(fixedIPs, ports.IP{SubnetID: subnet})
					}
				}
			}
		}
	}
	w.WriteHeader(http.StatusAccepted)

	p := ports.Port{
		ID:             uuid.New().String(),
		Name:           create.Port.Name,
		NetworkID:      create.Port.NetworkID,
		SecurityGroups: *create.Port.SecurityGroups,
		DeviceID:       create.Port.DeviceID,
		FixedIPs:       fixedIPs,
	}
	m.ports[p.ID] = p

	resp := portGetResponse{
		Port: p,
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

func (m *MockClient) tagPort(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /ports/<portid>/tags/<tag>
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	portID := parts[1]
	tag := parts[3]

	if _, ok := m.ports[portID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	port := m.ports[portID]
	port.Tags = append(port.Tags, tag)
	m.ports[portID] = port

	w.WriteHeader(http.StatusCreated)
}
