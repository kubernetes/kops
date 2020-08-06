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
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

type subnetListResponse struct {
	Subnets []subnets.Subnet `json:"subnets"`
}

type subnetGetResponse struct {
	Subnet subnets.Subnet `json:"subnet"`
}

type subnetCreateRequest struct {
	Subnet subnets.CreateOpts `json:"subnet"`
}

func (m *MockClient) mockSubnets() {
	re := regexp.MustCompile(`/subnets/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		subnetID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if subnetID == "" {
				r.ParseForm()
				m.listSubnets(w, r.Form)
			} else {
				m.getSubnet(w, subnetID)
			}
		case http.MethodPut:
			m.tagSubnet(w, r)
		case http.MethodPost:
			m.createSubnet(w, r)
		case http.MethodDelete:
			m.deleteSubnet(w, subnetID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/subnets/", handler)
	m.Mux.HandleFunc("/subnets", handler)
}

func (m *MockClient) listSubnets(w http.ResponseWriter, vals url.Values) {
	w.WriteHeader(http.StatusOK)

	subnets := filterSubnets(m.subnets, vals)

	resp := subnetListResponse{
		Subnets: subnets,
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

func (m *MockClient) getSubnet(w http.ResponseWriter, subnetID string) {
	if subnet, ok := m.subnets[subnetID]; ok {
		resp := subnetGetResponse{
			Subnet: subnet,
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

func (m *MockClient) deleteSubnet(w http.ResponseWriter, subnetID string) {
	if _, ok := m.subnets[subnetID]; ok {
		delete(m.subnets, subnetID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createSubnet(w http.ResponseWriter, r *http.Request) {
	var create subnetCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create subnet request")
	}

	w.WriteHeader(http.StatusAccepted)

	subnet := subnets.Subnet{
		ID:             uuid.New().String(),
		Name:           create.Subnet.Name,
		Description:    create.Subnet.Description,
		NetworkID:      create.Subnet.NetworkID,
		CIDR:           create.Subnet.CIDR,
		DNSNameservers: create.Subnet.DNSNameservers,
		EnableDHCP:     *create.Subnet.EnableDHCP,
		IPVersion:      int(create.Subnet.IPVersion),
	}
	m.subnets[subnet.ID] = subnet

	resp := subnetGetResponse{
		Subnet: subnet,
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

func (m *MockClient) tagSubnet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	subnetID := parts[1]
	tag := parts[3]

	if _, ok := m.subnets[subnetID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	subnet := m.subnets[subnetID]
	subnet.Tags = append(subnet.Tags, tag)
	m.subnets[subnetID] = subnet

	w.WriteHeader(http.StatusCreated)
}

func filterSubnets(allSubnets map[string]subnets.Subnet, vals url.Values) []subnets.Subnet {
	subnets := make([]subnets.Subnet, 0)

	idFilter := vals.Get("id")
	nameFilter := vals.Get("name")
	ipVersionFilter := vals.Get("ip_version")
	cidrFilter := vals.Get("cidr")
	enableDHCPFilter := vals.Get("enable_dhcp")
	networkIDFilter := vals.Get("network_id")

	for _, s := range allSubnets {
		if idFilter != "" && s.ID != idFilter {
			continue
		}
		if nameFilter != "" && s.Name != nameFilter {
			continue
		}
		if ipVersionFilter != "" {
			ipVersion, err := strconv.ParseInt(ipVersionFilter, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse ip_version parameter %v", err))
			}
			if int64(s.IPVersion) != ipVersion {
				continue
			}
		}
		if cidrFilter != "" && s.CIDR != cidrFilter {
			continue
		}
		if enableDHCPFilter != "" {
			enableDHCP, err := strconv.ParseBool(enableDHCPFilter)
			if err != nil {
				panic(fmt.Sprintf("failed to parse enable_dhcp parameter %v", err))
			}
			if s.EnableDHCP != enableDHCP {
				continue
			}
		}
		if networkIDFilter != "" && s.NetworkID != networkIDFilter {
			continue
		}
		subnets = append(subnets, s)
	}
	return subnets
}
