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

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
)

type groupListResponse struct {
	SecurityGroups []groups.SecGroup `json:"security_groups"`
}

type groupGetResponse struct {
	SecurityGroup groups.SecGroup `json:"security_group"`
}

type groupCreateRequest struct {
	SecurityGroup groups.CreateOpts `json:"security_group"`
}

func (m *MockClient) mockSecurityGroups() {
	re := regexp.MustCompile(`/security-groups/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		sgID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if sgID == "" {
				r.ParseForm()
				m.listSecurityGroups(w, r.Form)
			} else {
				m.getSecurityGroup(w, sgID)
			}
		case http.MethodPost:
			m.createSecurityGroup(w, r)
		case http.MethodDelete:
			m.deleteSecurityGroup(w, sgID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/security-groups/", handler)
	m.Mux.HandleFunc("/security-groups", handler)
}

func (m *MockClient) listSecurityGroups(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	sgs := make([]groups.SecGroup, 0)
	nameFilter := vals.Get("name")
	for _, s := range m.securityGroups {
		if nameFilter != "" && s.Name != nameFilter {
			continue
		}
		sgs = append(sgs, s)
	}

	resp := groupListResponse{
		SecurityGroups: sgs,
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

func (m *MockClient) getSecurityGroup(w http.ResponseWriter, groupID string) {
	if sg, ok := m.securityGroups[groupID]; ok {
		resp := groupGetResponse{
			SecurityGroup: sg,
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

func (m *MockClient) deleteSecurityGroup(w http.ResponseWriter, groupID string) {
	if _, ok := m.securityGroups[groupID]; ok {
		delete(m.securityGroups, groupID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createSecurityGroup(w http.ResponseWriter, r *http.Request) {
	var create groupCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create group request")
	}

	w.WriteHeader(http.StatusAccepted)

	s := groups.SecGroup{
		ID:          uuid.New().String(),
		Name:        create.SecurityGroup.Name,
		Description: create.SecurityGroup.Description,
	}
	m.securityGroups[s.ID] = s

	resp := groupGetResponse{
		SecurityGroup: s,
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
