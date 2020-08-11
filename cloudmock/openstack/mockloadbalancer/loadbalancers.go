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
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
)

type loadbalancerListResponse struct {
	LoadBalancers []loadbalancers.LoadBalancer `json:"loadbalancers"`
}

type loadbalancerGetResponse struct {
	LoadBalancer loadbalancers.LoadBalancer `json:"loadbalancer"`
}

type loadbalancerCreateRequest struct {
	LoadBalancer loadbalancers.CreateOpts `json:"loadbalancer"`
}

func (m *MockClient) mockLoadBalancers() {
	re := regexp.MustCompile(`/lbaas/loadbalancers/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		loadbalancerID := re.ReplaceAllString(r.URL.Path, "")
		// TODO: handle /members subresource
		switch r.Method {
		case http.MethodGet:
			if loadbalancerID == "" {
				r.ParseForm()
				m.listLoadBalancers(w, r.Form)
			} else {
				m.getLoadBalancer(w, loadbalancerID)
			}
		case http.MethodPost:
			m.createLoadBalancer(w, r)
		case http.MethodDelete:
			m.deleteLoadBalancer(w, loadbalancerID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/lbaas/loadbalancers/", handler)
	m.Mux.HandleFunc("/lbaas/loadbalancers", handler)
}

func (m *MockClient) listLoadBalancers(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	loadbalancers := make([]loadbalancers.LoadBalancer, 0)
	for _, l := range m.loadbalancers {
		name := vals.Get("name")
		id := vals.Get("id")
		if name != "" && name != l.Name {
			continue
		}
		if id != "" && id != l.ID {
			continue
		}
		loadbalancers = append(loadbalancers, populateLB(l, m.pools, m.listeners))
	}

	resp := loadbalancerListResponse{
		LoadBalancers: loadbalancers,
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

func (m *MockClient) getLoadBalancer(w http.ResponseWriter, loadbalancerID string) {
	if loadbalancer, ok := m.loadbalancers[loadbalancerID]; ok {
		resp := loadbalancerGetResponse{
			LoadBalancer: populateLB(loadbalancer, m.pools, m.listeners),
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

func (m *MockClient) deleteLoadBalancer(w http.ResponseWriter, loadbalancerID string) {
	if _, ok := m.loadbalancers[loadbalancerID]; ok {
		delete(m.loadbalancers, loadbalancerID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var create loadbalancerCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create loadbalancer request")
	}

	w.WriteHeader(http.StatusAccepted)
	l := loadbalancers.LoadBalancer{
		ID:                 uuid.New().String(),
		Name:               create.LoadBalancer.Name,
		VipSubnetID:        create.LoadBalancer.VipSubnetID,
		ProvisioningStatus: "ACTIVE",
		// TODO: create a Port and set VipPortID
	}
	m.loadbalancers[l.ID] = l

	resp := loadbalancerGetResponse{
		LoadBalancer: l,
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

func populateLB(lb loadbalancers.LoadBalancer, lbPools map[string]pools.Pool, lbListeners map[string]listeners.Listener) loadbalancers.LoadBalancer {
	lb.Pools = make([]pools.Pool, 0)
	for _, p := range lbPools {
		match := false
		for _, poolLB := range p.Loadbalancers {
			if lb.ID == poolLB.ID {
				match = true
				break
			}
		}
		if match {
			lb.Pools = append(lb.Pools, p)
		}
	}
	for _, l := range lbListeners {
		match := false
		for _, listenerLB := range l.Loadbalancers {
			if lb.ID == listenerLB.ID {
				match = true
				break
			}
		}
		if match {
			lb.Listeners = append(lb.Listeners, l)
		}
	}

	return lb
}
