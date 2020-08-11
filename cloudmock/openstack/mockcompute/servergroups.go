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
	"regexp"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
)

type serverGroupListResponse struct {
	ServerGroups []servergroups.ServerGroup `json:"server_groups"`
}

type serverGroupGetResponse struct {
	ServerGroup servergroups.ServerGroup `json:"server_group"`
}

type serverGroupCreateRequest struct {
	ServerGroup servergroups.CreateOpts `json:"server_group"`
}

func (m *MockClient) mockServerGroups() {
	re := regexp.MustCompile(`/os-server-groups/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		sgID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if sgID == "" {
				m.listServerGroups(w)
			} else {
				m.getServerGroup(w, sgID)
			}
		case http.MethodPost:
			m.createServerGroup(w, r)
		case http.MethodDelete:
			m.deleteServerGroup(w, sgID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/os-server-groups/", handler)
	m.Mux.HandleFunc("/os-server-groups", handler)
}

func (m *MockClient) listServerGroups(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)

	servergroups := make([]servergroups.ServerGroup, 0)
	for _, s := range m.serverGroups {
		servergroups = append(servergroups, s)
	}

	resp := serverGroupListResponse{
		ServerGroups: servergroups,
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

func (m *MockClient) getServerGroup(w http.ResponseWriter, serverGroupID string) {
	if serverGroup, ok := m.serverGroups[serverGroupID]; ok {
		resp := serverGroupGetResponse{
			ServerGroup: serverGroup,
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

func (m *MockClient) deleteServerGroup(w http.ResponseWriter, serverGroupID string) {
	if _, ok := m.serverGroups[serverGroupID]; ok {
		delete(m.serverGroups, serverGroupID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createServerGroup(w http.ResponseWriter, r *http.Request) {
	var create serverGroupCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create serverGroup request")
	}

	w.WriteHeader(http.StatusOK)

	serverGroup := servergroups.ServerGroup{
		ID:       uuid.New().String(),
		Name:     create.ServerGroup.Name,
		Policies: create.ServerGroup.Policies,
	}
	m.serverGroups[serverGroup.ID] = serverGroup

	resp := serverGroupGetResponse{
		ServerGroup: serverGroup,
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
