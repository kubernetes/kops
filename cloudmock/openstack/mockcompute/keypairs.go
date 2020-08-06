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

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"k8s.io/kops/pkg/pki"
)

type keyPairListResponse struct {
	KeyPairs []keyPairGetResponse `json:"keypairs"`
}

type keyPairGetResponse struct {
	KeyPair keypairs.KeyPair `json:"keypair"`
}

type keyPairCreateRequest struct {
	KeyPair keypairs.CreateOpts `json:"keypair"`
}

func (m *MockClient) mockKeyPairs() {
	re := regexp.MustCompile(`/os-keypairs/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")
		kpID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if kpID == "" {
				m.listKeyPairs(w)
			} else {
				m.getKeyPair(w, kpID)
			}
		case http.MethodPost:
			m.createKeyPair(w, r)
		case http.MethodDelete:
			m.deleteKeyPair(w, kpID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/os-keypairs/", handler)
	m.Mux.HandleFunc("/os-keypairs", handler)
}

func (m *MockClient) listKeyPairs(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)

	keypairs := make([]keyPairGetResponse, 0)
	for _, k := range m.keyPairs {
		keypairs = append(keypairs, keyPairGetResponse{KeyPair: k})
	}

	resp := keyPairListResponse{
		KeyPairs: keypairs,
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

func (m *MockClient) getKeyPair(w http.ResponseWriter, keyPairID string) {
	if keyPair, ok := m.keyPairs[keyPairID]; ok {
		resp := keyPairGetResponse{
			KeyPair: keyPair,
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

func (m *MockClient) deleteKeyPair(w http.ResponseWriter, keyPairID string) {
	if _, ok := m.keyPairs[keyPairID]; ok {
		delete(m.keyPairs, keyPairID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createKeyPair(w http.ResponseWriter, r *http.Request) {
	var create keyPairCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create keyPair request")
	}

	w.WriteHeader(http.StatusCreated)

	fp, err := pki.ComputeOpenSSHKeyFingerprint(create.KeyPair.PublicKey)
	if err != nil {
		panic("error computing public key fingerprint")
	}
	keyPair := keypairs.KeyPair{
		Name:        create.KeyPair.Name,
		PublicKey:   create.KeyPair.PublicKey,
		Fingerprint: fp,
	}
	m.keyPairs[keyPair.Name] = keyPair

	resp := keyPairGetResponse{
		KeyPair: keyPair,
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
