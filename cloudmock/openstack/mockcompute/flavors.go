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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"k8s.io/kops/upup/pkg/fi"
)

type flavorListResponse struct {
	Flavors []flavors.Flavor `json:"flavors"`
}

type flavorGetResponse struct {
	Flavor flavors.Flavor `json:"flavor"`
}

type flavorCreateRequest struct {
	Flavor flavors.CreateOpts `json:"flavor"`
}

func (m *MockClient) mockFlavors() {
	re := regexp.MustCompile(`/flavors/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		flavorID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if flavorID == "detail" {
				m.listFlavors(w)
			} else {
				m.getFlavor(w, flavorID)
			}
		case http.MethodPost:
			m.createFlavor(w, r)
		case http.MethodDelete:
			m.deleteFlavor(w, flavorID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/flavors/", handler)
	m.Mux.HandleFunc("/flavors", handler)
}

func (m *MockClient) listFlavors(w http.ResponseWriter) {

	w.WriteHeader(http.StatusOK)

	flavors := make([]flavors.Flavor, 0)
	for _, flavor := range m.flavors {
		flavors = append(flavors, flavor)
	}

	resp := flavorListResponse{
		Flavors: flavors,
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

func (m *MockClient) getFlavor(w http.ResponseWriter, flavorID string) {
	if flavor, ok := m.flavors[flavorID]; ok {
		resp := flavorGetResponse{
			Flavor: flavor,
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

func (m *MockClient) deleteFlavor(w http.ResponseWriter, flavorID string) {
	if _, ok := m.flavors[flavorID]; ok {
		delete(m.flavors, flavorID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createFlavor(w http.ResponseWriter, r *http.Request) {
	var create flavorCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create flavor request")
	}
	w.WriteHeader(http.StatusCreated)

	flavor := flavors.Flavor{
		ID:    uuid.New().String(),
		Name:  create.Flavor.Name,
		RAM:   create.Flavor.RAM,
		VCPUs: create.Flavor.VCPUs,
		Disk:  fi.IntValue(create.Flavor.Disk),
	}
	m.flavors[flavor.ID] = flavor

	resp := flavorGetResponse{
		Flavor: flavor,
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
