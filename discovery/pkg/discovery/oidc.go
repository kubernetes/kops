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

package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
	api "k8s.io/kops/discovery/apis/discovery.kops.k8s.io/v1alpha1"
)

func (s *Server) handleOIDCDiscovery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := klog.FromContext(ctx)

	universeID := r.PathValue("universe")

	host := r.Host
	if host == "" {
		log.Info("Cannot determine host for OIDC discovery")
		http.Error(w, "Cannot determine host", http.StatusBadRequest)
		return
	}

	endpoints, err := s.Store.ListDiscoveryEndpoints(r.Context(), universeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing endpoints: %v", err), http.StatusInternalServerError)
		return
	}

	issuerURL := "https://" + host + "/" + universeID + "/"

	var oidcSpec *api.OIDCSpec
	for _, ep := range endpoints {
		if ep.Spec.OIDC != nil {
			oidcSpec = ep.Spec.OIDC
			break
		}
	}

	if oidcSpec == nil {
		http.NotFound(w, r)
		return
	}

	// Construct minimal OIDC discovery document
	jwksURI := issuerURL
	if !strings.HasSuffix(jwksURI, "/") {
		jwksURI += "/"
	}
	jwksURI += "openid/v1/jwks"

	resp := OIDCDiscoveryResponse{
		Issuer:                           issuerURL,
		JWKSURI:                          jwksURI,
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
	}
	s.writeJSON(w, http.StatusOK, resp)

	log.Info("served OIDC discovery document", "universe", universeID)
}

func (s *Server) handleOIDCJWKS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := klog.FromContext(ctx)

	universeID := r.PathValue("universe")

	endpoints, err := s.Store.ListDiscoveryEndpoints(r.Context(), universeID)
	if err != nil {
		log.Error(err, "error listing endpoints")
		http.Error(w, fmt.Sprintf("Error listing endpoints: %v", err), http.StatusInternalServerError)
		return
	}

	type keyInfo struct {
		key      api.JSONWebKey
		lastSeen string
	}
	keysMap := make(map[string]keyInfo)

	for _, ep := range endpoints {
		if ep.Spec.OIDC == nil {
			continue
		}
		for _, key := range ep.Spec.OIDC.Keys {
			kid := key.KeyID
			if kid == "" {
				continue
			}

			// Conflict resolution: prefer newest LastSeen
			currentLastSeen := ep.Spec.LastSeen
			if existing, exists := keysMap[kid]; exists {
				if currentLastSeen > existing.lastSeen {
					keysMap[kid] = keyInfo{key: key, lastSeen: currentLastSeen}
				}
			} else {
				keysMap[kid] = keyInfo{key: key, lastSeen: currentLastSeen}
			}
		}
	}

	var mergedKeys []api.JSONWebKey
	for _, info := range keysMap {
		mergedKeys = append(mergedKeys, info.key)
	}

	response := map[string]interface{}{
		"keys": mergedKeys,
	}

	s.writeJSON(w, http.StatusOK, response)
	log.Info("served OIDC JWKS", "universe", universeID)
}

type OIDCDiscoveryResponse struct {
	Issuer                           string   `json:"issuer,omitempty"`
	JWKSURI                          string   `json:"jwks_uri,omitempty"`
	ResponseTypesSupported           []string `json:"response_types_supported,omitempty"`
	SubjectTypesSupported            []string `json:"subject_types_supported,omitempty"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported,omitempty"`
}

func (s *Server) writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}
