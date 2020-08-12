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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
)

type routerListResponse struct {
	Routers []routers.Router `json:"routers"`
}

type routerGetResponse struct {
	Router routers.Router `json:"router"`
}

type routerCreateRequest struct {
	Router routers.CreateOpts `json:"router"`
}

func (m *MockClient) mockRouters() {
	re := regexp.MustCompile(`/routers/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		routerID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if routerID == "" {
				r.ParseForm()
				m.listRouters(w, r.Form)
			} else {
				m.getRouter(w, routerID)
			}
		case http.MethodPut:
			m.routerInterface(w, r)
		case http.MethodPost:
			m.createRouter(w, r)
		case http.MethodDelete:
			m.deleteRouter(w, routerID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/routers/", handler)
	m.Mux.HandleFunc("/routers", handler)
}

func (m *MockClient) listRouters(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	routers := make([]routers.Router, 0)
	nameFilter := vals.Get("name")
	idFilter := vals.Get("id")
	for _, r := range m.routers {
		if nameFilter != "" && r.Name != nameFilter {
			continue
		}
		if idFilter != "" && r.ID != idFilter {
			continue
		}
		routers = append(routers, r)
	}

	resp := routerListResponse{
		Routers: routers,
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

func (m *MockClient) getRouter(w http.ResponseWriter, routerID string) {
	if router, ok := m.routers[routerID]; ok {
		resp := routerGetResponse{
			Router: router,
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

func (m *MockClient) deleteRouter(w http.ResponseWriter, routerID string) {
	if _, ok := m.routers[routerID]; ok {
		delete(m.routers, routerID)
		delete(m.routerInterfaces, routerID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createRouter(w http.ResponseWriter, r *http.Request) {
	var create routerCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create router request")
	}

	w.WriteHeader(http.StatusAccepted)

	p := routers.Router{
		ID:           uuid.New().String(),
		Name:         create.Router.Name,
		AdminStateUp: *create.Router.AdminStateUp,
		GatewayInfo:  *create.Router.GatewayInfo,
	}
	m.routers[p.ID] = p

	resp := routerGetResponse{
		Router: p,
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

func (m *MockClient) routerInterface(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /routers/<routerID>/add_router_interface
	// /routers/<routerID>/remove_router_interface
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	routerID := parts[1]
	if _, ok := m.routers[routerID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var createInterface routers.AddInterfaceOpts
	err := json.NewDecoder(r.Body).Decode(&createInterface)
	if err != nil {
		panic("error decoding create router interface request")
	}
	if parts[2] == "add_router_interface" {
		subnet := m.subnets[createInterface.SubnetID]
		interfaces := m.routerInterfaces[routerID]
		interfaces = append(interfaces, routers.InterfaceInfo{
			SubnetID: subnet.ID,
		})
		m.routerInterfaces[routerID] = interfaces
		// If PortID is not sent, this creates a new port.

		port := ports.Port{
			ID:        uuid.New().String(),
			NetworkID: subnet.NetworkID,
			DeviceID:  routerID,
			FixedIPs: []ports.IP{
				{
					SubnetID: subnet.ID,
				},
			},
		}
		m.ports[port.ID] = port
	} else if parts[2] == "remove_router_interface" {
		interfaces := make([]routers.InterfaceInfo, 0)
		for _, i := range m.routerInterfaces[routerID] {
			if i.SubnetID != createInterface.SubnetID {
				interfaces = append(interfaces, i)
			}
		}
		m.routerInterfaces[routerID] = interfaces
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	respB, err := json.Marshal(createInterface)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", createInterface))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}
