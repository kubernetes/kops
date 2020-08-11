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

package mockdns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	zones "github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
)

type zoneListResponse struct {
	Zones []zones.Zone `json:"zones"`
}

type zoneGetResponse struct {
	Zone zones.Zone `json:"zone"`
}

type recordSetsListResponse struct {
	RecordSets []recordsets.RecordSet `json:"recordsets"`
}

func (m *MockClient) mockZones() {

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		path := r.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")

		zoneID := ""
		zoneName := ""
		if len(parts) > 1 {
			zoneID = parts[1]
		}
		r.ParseForm()
		zoneName = r.Form.Get("name")
		switch r.Method {
		case http.MethodGet:
			if zoneID == "" && zoneName == "" {
				// /zones
				m.listZones(w)
			} else if len(parts) == 3 && parts[2] == "recordsets" {
				// /zones/<zoneid>/recordsets
				m.listRecordSets(w, zoneID)
			} else if len(parts) == 4 && parts[2] == "recordsets" {
				// /zones/<zoneid>/recordsets/<recordsetid>
				m.getRecordSet(w, zoneID, parts[3])
			} else {
				// /zones?name=<zonename>
				m.getZone(w, zoneName)
			}
		case http.MethodPost:
			m.createZone(w, r)
		case http.MethodDelete:
			m.deleteZone(w, zoneID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/zones/", handler)
	m.Mux.HandleFunc("/zones", handler)
}

func (m *MockClient) listZones(w http.ResponseWriter) {

	w.WriteHeader(http.StatusOK)

	zones := make([]zones.Zone, 0)
	for _, z := range m.zones {
		zones = append(zones, z)
	}

	resp := zoneListResponse{
		Zones: zones,
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

func (m *MockClient) getZone(w http.ResponseWriter, zoneName string) {
	for _, zone := range m.zones {
		if zone.Name != zoneName {
			continue
		}
		resp := zoneListResponse{
			Zones: []zones.Zone{zone},
		}
		respB, err := json.Marshal(resp)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %+v", resp))
		}
		_, err = w.Write(respB)
		if err != nil {
			panic("failed to write body")
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (m *MockClient) deleteZone(w http.ResponseWriter, zoneID string) {
	if _, ok := m.zones[zoneID]; ok {
		delete(m.zones, zoneID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createZone(w http.ResponseWriter, r *http.Request) {
	var create zones.CreateOpts
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create zone request")
	}

	w.WriteHeader(http.StatusAccepted)

	z := zones.Zone{
		ID:   uuid.New().String(),
		Name: create.Name,
	}
	m.zones[z.ID] = z

	resp := zoneGetResponse{
		Zone: z,
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

func (m *MockClient) listRecordSets(w http.ResponseWriter, zoneID string) {
	w.WriteHeader(http.StatusOK)

	records := make([]recordsets.RecordSet, 0)
	for _, r := range m.recordSets {
		if r.ZoneID != zoneID {
			continue
		}
		records = append(records, r)
	}

	resp := recordSetsListResponse{
		RecordSets: records,
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

func (m *MockClient) getRecordSet(w http.ResponseWriter, zoneID, recordSetID string) {
	if record, ok := m.recordSets[recordSetID]; ok {
		if record.ZoneID != zoneID {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		respB, err := json.Marshal(record)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %+v", record))
		}
		_, err = w.Write(respB)
		if err != nil {
			panic("failed to write body")
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
