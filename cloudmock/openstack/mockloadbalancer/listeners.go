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

package mockloadbalancer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
)

type listenerListResponse struct {
	Listeners []listeners.Listener `json:"listeners"`
}

type listenerGetResponse struct {
	Listener listeners.Listener `json:"listener"`
}

type listenerCreateRequest struct {
	Listener listeners.CreateOpts `json:"listener"`
}

func (m *MockClient) mockListeners() {
	re := regexp.MustCompile(`/lbaas/listeners/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		listenerID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			r.ParseForm()
			m.listListeners(w, r.Form)
		case http.MethodPost:
			m.createListener(w, r)
		case http.MethodDelete:
			m.deleteListener(w, listenerID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/lbaas/listeners/", handler)
	m.Mux.HandleFunc("/lbaas/listeners", handler)
}

func (m *MockClient) listListeners(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	listeners := make([]listeners.Listener, 0)
	for _, l := range m.listeners {
		name := vals.Get("name")
		id := vals.Get("id")
		if name != "" && l.Name != name {
			continue
		}
		if id != "" && l.ID != id {
			continue
		}
		listeners = append(listeners, l)
	}

	resp := listenerListResponse{
		Listeners: listeners,
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

func (m *MockClient) deleteListener(w http.ResponseWriter, listenerID string) {
	if _, ok := m.listeners[listenerID]; ok {
		delete(m.listeners, listenerID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createListener(w http.ResponseWriter, r *http.Request) {
	var create listenerCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create listener request")
	}

	w.WriteHeader(http.StatusAccepted)

	l := listeners.Listener{
		ID:            uuid.New().String(),
		Name:          create.Listener.Name,
		DefaultPoolID: create.Listener.DefaultPoolID,
		Loadbalancers: []listeners.LoadBalancerID{{ID: create.Listener.LoadbalancerID}},
		Protocol:      string(create.Listener.Protocol),
		ProtocolPort:  create.Listener.ProtocolPort,
		AllowedCIDRs:  create.Listener.AllowedCIDRs,
	}
	m.listeners[l.ID] = l

	resp := listenerGetResponse{
		Listener: l,
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
