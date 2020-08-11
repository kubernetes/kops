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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type serverGetResponse struct {
	Server servers.Server `json:"server"`
}

type serverListResponse struct {
	Servers []servers.Server `json:"servers"`
}

type serverCreateRequest struct {
	Server servers.CreateOpts `json:"server"`
}

func (m *MockClient) mockServers() {
	re := regexp.MustCompile(`/servers/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		serverID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if serverID == "detail" {
				r.ParseForm()
				m.listServers(w, r.Form)
			}
		case http.MethodPost:
			m.createServer(w, r)
		case http.MethodDelete:
			m.deleteServer(w, serverID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/servers/", handler)
	m.Mux.HandleFunc("/servers", handler)
}

func (m *MockClient) listServers(w http.ResponseWriter, vals url.Values) {
	serverName := strings.Trim(vals.Get("name"), "^$")
	matched := make([]servers.Server, 0)
	for _, server := range m.servers {
		if server.Name == serverName {
			matched = append(matched, server)
		}
	}
	resp := serverListResponse{
		Servers: matched,
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

func (m *MockClient) deleteServer(w http.ResponseWriter, serverID string) {
	if _, ok := m.servers[serverID]; ok {
		delete(m.servers, serverID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createServer(w http.ResponseWriter, r *http.Request) {
	var create serverCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create server request")
	}

	w.WriteHeader(http.StatusCreated)

	server := servers.Server{
		ID:       uuid.New().String(),
		Name:     create.Server.Name,
		Metadata: create.Server.Metadata,
	}
	securityGroups := make([]map[string]interface{}, len(create.Server.SecurityGroups))
	for i, groupName := range create.Server.SecurityGroups {
		securityGroups[i] = map[string]interface{}{"name": groupName}
	}
	server.SecurityGroups = securityGroups

	// Assign an IP address
	private := make([]map[string]string, 1)
	private[0] = make(map[string]string)
	private[0]["OS-EXT-IPS:type"] = "fixed"
	private[0]["addr"] = "192.168.1.1"
	server.Addresses = map[string]interface{}{
		"private": private,
	}

	m.servers[server.ID] = server

	resp := serverGetResponse{
		Server: server,
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
