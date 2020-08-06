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
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
)

type poolListResponse struct {
	Pools []pools.Pool `json:"pools"`
}

type poolGetResponse struct {
	Pool pools.Pool `json:"pool"`
}

type poolCreateRequest struct {
	Pool pools.CreateOpts `json:"pool"`
}

func (m *MockClient) mockPools() {
	re := regexp.MustCompile(`/lbaas/pools/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		poolID := re.ReplaceAllString(r.URL.Path, "")
		// TODO: handle /members subresource
		switch r.Method {
		case http.MethodGet:
			if poolID == "" {
				r.ParseForm()
				m.listPools(w, r.Form)
			} else {
				m.getPool(w, poolID)
			}
		case http.MethodPost:
			m.createPool(w, r)
		case http.MethodDelete:
			m.deletePool(w, poolID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/lbaas/pools/", handler)
	m.Mux.HandleFunc("/lbaas/pools", handler)
}

func (m *MockClient) listPools(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	pools := make([]pools.Pool, 0)
	id := vals.Get("id")
	name := vals.Get("name")
	for _, p := range m.pools {
		if id != "" && id != p.ID {
			continue
		}
		if name != "" && name != p.Name {
			continue
		}
		pools = append(pools, p)
	}

	resp := poolListResponse{
		Pools: pools,
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

func (m *MockClient) getPool(w http.ResponseWriter, poolID string) {
	if pool, ok := m.pools[poolID]; ok {
		resp := poolGetResponse{
			Pool: pool,
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

func (m *MockClient) deletePool(w http.ResponseWriter, poolID string) {
	if _, ok := m.pools[poolID]; ok {
		delete(m.pools, poolID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createPool(w http.ResponseWriter, r *http.Request) {
	var create poolCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create pool request")
	}

	w.WriteHeader(http.StatusAccepted)

	p := pools.Pool{
		ID:            uuid.New().String(),
		Name:          create.Pool.Name,
		LBMethod:      string(create.Pool.LBMethod),
		Protocol:      string(create.Pool.Protocol),
		Loadbalancers: []pools.LoadBalancerID{{ID: create.Pool.LoadbalancerID}},
	}
	m.pools[p.ID] = p

	resp := poolGetResponse{
		Pool: p,
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
