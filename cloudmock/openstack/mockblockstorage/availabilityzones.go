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

package mockblockstorage

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
)

type availabilityZoneListResponse struct {
	AvailabilityZones []availabilityzones.AvailabilityZone `json:"availabilityZoneInfo"`
}

func (m *MockClient) mockAvailabilityZones() {
	// There is no "create" API for zones so they are directly added here
	m.availabilityZones["us-test1-a"] = availabilityzones.AvailabilityZone{
		ZoneName: "us-east1-a",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			m.listAvailabilityZones(w)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/os-availability-zone", handler)
}

func (m *MockClient) listAvailabilityZones(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)

	availabilityzones := make([]availabilityzones.AvailabilityZone, 0)
	for _, k := range m.availabilityZones {
		availabilityzones = append(availabilityzones, k)
	}

	resp := availabilityZoneListResponse{
		AvailabilityZones: availabilityzones,
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
