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
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
)

type floatingIPListResponse struct {
	FloatingIPs []floatingips.FloatingIP `json:"floatingips"`
}

type floatingIPType struct {
	ID         string `json:"id"`
	FloatingIP string `json:"floatingip"`
	TenantID   string `json:"tenant_id"`
	ProjectID  string `json:"project_id"`
}

type floatingIPCreateRequest struct {
	FloatingIP floatingIPType `json:"floatingip"`
}

type floatingIPGetResponse struct {
	FloatingIP floatingips.FloatingIP `json:"floatingip"`
}

func (m *MockClient) mockFloatingIPs() {
	re := regexp.MustCompile(`/floatingips/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		floatingIPID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if floatingIPID == "" {
				r.ParseForm()
				m.listFloatingIPs(w, r.Form)
			}
		case http.MethodPost:
			m.createFloatingIp(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/floatingips/", handler)
	m.Mux.HandleFunc("/floatingips", handler)
}

func (m *MockClient) listFloatingIPs(w http.ResponseWriter, vals url.Values) {
	w.WriteHeader(http.StatusOK)

	floatingips := make([]floatingips.FloatingIP, 0)
	for _, p := range m.floatingips {
		floatingips = append(floatingips, p)
	}
	resp := floatingIPListResponse{
		FloatingIPs: floatingips,
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

func (m *MockClient) createFloatingIp(w http.ResponseWriter, r *http.Request) {
	var create floatingIPCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create floating ip request")
	}

	w.WriteHeader(http.StatusAccepted)

	f := floatingips.FloatingIP{
		ID:         uuid.New().String(),
		FloatingIP: create.FloatingIP.FloatingIP,
		TenantID:   create.FloatingIP.TenantID,
		//UpdatedAt:      time.Now(),
		//CreatedAt:      time.Now(),
		ProjectID:      create.FloatingIP.ProjectID,
		Status:         "ACTIVE",
		RouterID:       "router",
		Tags:           nil,
		RevisionNumber: 0,
	}
	m.floatingips[f.ID] = f

	resp := floatingIPGetResponse{
		FloatingIP: f,
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
