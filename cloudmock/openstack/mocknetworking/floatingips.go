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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

type floatingIPListResponse struct {
	FloatingIPs []floatingips.FloatingIP `json:"floatingips"`
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
