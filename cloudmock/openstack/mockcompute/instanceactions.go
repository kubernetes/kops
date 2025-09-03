/*
Copyright 2025 The Kubernetes Authors.

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
	"net/http"
)

type instanceActionsResponse struct {
	InstanceActions []interface{} `json:"instanceActions"`
}

func (m *MockClient) mockInstanceActions() {
	// re := regexp.MustCompile(`/servers/(.*?)/os-instance-actions/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := instanceActionsResponse{
			InstanceActions: make([]interface{}, 0),
		}
		respB, err := json.Marshal(resp)
		if err != nil {
			panic("failed to marshal response")
		}
		_, err = w.Write(respB)
		if err != nil {
			panic("failed to write body")
		}
	}
	m.Mux.HandleFunc("/servers/{server_id}/os-instance-actions/", handler)
}
