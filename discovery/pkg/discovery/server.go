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

	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/discovery/apis/discovery.kops.k8s.io/v1alpha1"
)

type Server struct {
	Store Store
	mux   *http.ServeMux
}

func NewServer(store Store) *Server {
	s := &Server{
		Store: store,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Public OIDC Discovery
	s.mux.HandleFunc("GET /{universe}/.well-known/openid-configuration", s.handleOIDCDiscovery)
	s.mux.HandleFunc("GET /{universe}/openid/v1/jwks", s.handleOIDCJWKS)

	// Authenticated Routes
	// 1. Root Discovery (/apis)
	s.mux.HandleFunc("GET /{universe}/apis", s.withAuth(s.handleAPIGroupList))

	// Discovering resources in group
	s.mux.HandleFunc("GET /{universe}/apis/discovery.kops.k8s.io/v1alpha1", s.withAuth(s.handleAPIResourceList))

	// Listing DiscoveryEndpoints (All namespaces)
	s.mux.HandleFunc("GET /{universe}/apis/discovery.kops.k8s.io/v1alpha1/discoveryendpoints", s.withAuth(s.handleListDiscoveryEndpoints))

	// Listing DiscoveryEndpoints (Specific namespace)
	s.mux.HandleFunc("GET /{universe}/apis/discovery.kops.k8s.io/v1alpha1/namespaces/{namespace}/discoveryendpoints", s.withAuth(s.handleListDiscoveryEndpoints))

	// Create DiscoveryEndpoint
	s.mux.HandleFunc("POST /{universe}/apis/discovery.kops.k8s.io/v1alpha1/namespaces/{namespace}/discoveryendpoints", s.withAuth(s.handleCreateDiscoveryEndpoint))

	// Get DiscoveryEndpoint
	s.mux.HandleFunc("GET /{universe}/apis/discovery.kops.k8s.io/v1alpha1/namespaces/{namespace}/discoveryendpoints/{name}", s.withAuth(s.handleGetDiscoveryEndpoint))

	// Apply (Patch) DiscoveryEndpoint
	s.mux.HandleFunc("PATCH /{universe}/apis/discovery.kops.k8s.io/v1alpha1/namespaces/{namespace}/discoveryendpoints/{name}", s.withAuth(s.handleApplyDiscoveryEndpoint))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) withAuth(next func(http.ResponseWriter, *http.Request, *UserInfo)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		universeID := r.PathValue("universe")
		userInfo, err := AuthenticateClientToUniverse(r, universeID)
		if err != nil {
			klog.Warningf("Unauthorized access attempt to universe %s: %v", universeID, err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r, userInfo)
	}
}

// Handlers

func (s *Server) handleAPIGroupList(w http.ResponseWriter, r *http.Request, _ *UserInfo) {
	resp := metav1.APIGroupList{
		TypeMeta: metav1.TypeMeta{Kind: "APIGroupList", APIVersion: "v1"},
		Groups: []metav1.APIGroup{
			{
				Name: "discovery.kops.k8s.io",
				Versions: []metav1.GroupVersionForDiscovery{
					{GroupVersion: "discovery.kops.k8s.io/v1alpha1", Version: "v1alpha1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "discovery.kops.k8s.io/v1alpha1",
					Version:      "v1alpha1",
				},
			},
		},
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAPIResourceList(w http.ResponseWriter, r *http.Request, _ *UserInfo) {
	resp := metav1.APIResourceList{
		TypeMeta:     metav1.TypeMeta{Kind: "APIResourceList", APIVersion: "v1"},
		GroupVersion: "discovery.kops.k8s.io/v1alpha1",
		APIResources: []metav1.APIResource{
			{
				Name:         "discoveryendpoints",
				SingularName: "discoveryendpoint",
				Namespaced:   true,
				Kind:         "DiscoveryEndpoint",
				Verbs:        []string{"get", "list", "create", "update", "patch"},
			},
		},
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListDiscoveryEndpoints(w http.ResponseWriter, r *http.Request, _ *UserInfo) {
	universeID := r.PathValue("universe")
	ns := r.PathValue("namespace") // Empty if not in path

	endpoints, err := s.Store.ListDiscoveryEndpoints(r.Context(), universeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing endpoints: %v", err), http.StatusInternalServerError)
		return
	}

	resp := api.DiscoveryEndpointList{
		TypeMeta: metav1.TypeMeta{Kind: "DiscoveryEndpointList", APIVersion: "discovery.kops.k8s.io/v1alpha1"},
	}

	for _, ep := range endpoints {
		if ns == "" || ep.ObjectMeta.Namespace == ns {
			resp.Items = append(resp.Items, *ep)
		}
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateDiscoveryEndpoint(w http.ResponseWriter, r *http.Request, userInfo *UserInfo) {
	universeID := r.PathValue("universe")
	ns := r.PathValue("namespace")

	var input api.DiscoveryEndpoint
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation: ensure the name matches the clientID from the cert
	if input.ObjectMeta.Name != "" && input.ObjectMeta.Name != userInfo.ClientID {
		http.Error(w, fmt.Sprintf("Forbidden: cannot register node name '%s' with client cert '%s'", input.ObjectMeta.Name, userInfo.ClientID), http.StatusForbidden)
		return
	}

	// Validation: ensure the namespace in body matches the URL
	if input.ObjectMeta.Namespace != ns {
		http.Error(w, "Forbidden: namespace does not match", http.StatusForbidden)
		return
	}

	if err := s.Store.UpsertDiscoveryEndpoint(r.Context(), universeID, &input); err != nil {
		http.Error(w, fmt.Sprintf("Error creating endpoint: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the created object
	s.writeJSON(w, http.StatusCreated, input)
}

func (s *Server) handleApplyDiscoveryEndpoint(w http.ResponseWriter, r *http.Request, userInfo *UserInfo) {
	ctx := r.Context()
	log := klog.FromContext(ctx)

	universeID := r.PathValue("universe")
	ns := r.PathValue("namespace")
	name := r.PathValue("name")

	var input api.DiscoveryEndpoint
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Info("invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation: ensure the name matches the clientID from the cert
	if input.ObjectMeta.Name != "" && input.ObjectMeta.Name != userInfo.ClientID {
		log.Info("Forbidden: cannot register node name", "name", input.ObjectMeta.Name, "clientID", userInfo.ClientID)
		http.Error(w, fmt.Sprintf("Forbidden: cannot register node name '%s' with client cert '%s'", input.ObjectMeta.Name, userInfo.ClientID), http.StatusForbidden)
		return
	}

	// Validation: ensure the name in body matches the URL
	if input.ObjectMeta.Name != name {
		log.Info("Forbidden: name does not match", "name", input.ObjectMeta.Name, "expected", name)
		http.Error(w, "Forbidden: name does not match", http.StatusForbidden)
		return
	}

	// Validation: ensure the namespace in body matches the URL
	if input.ObjectMeta.Namespace != ns {
		log.Info("Forbidden: namespace does not match", "namespace", input.ObjectMeta.Namespace, "expected", ns)
		http.Error(w, "Forbidden: namespace does not match", http.StatusForbidden)
		return
	}

	if err := s.Store.UpsertDiscoveryEndpoint(r.Context(), universeID, &input); err != nil {
		log.Error(err, "error applying endpoint")
		http.Error(w, fmt.Sprintf("Error applying endpoint: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the created object
	s.writeJSON(w, http.StatusCreated, input)

	log.Info("Applied endpoint", "namespace", input.ObjectMeta.Namespace, "name", input.ObjectMeta.Name, "universe", universeID)
}

func (s *Server) handleGetDiscoveryEndpoint(w http.ResponseWriter, r *http.Request, _ *UserInfo) {
	universeID := r.PathValue("universe")
	ns := r.PathValue("namespace")
	name := r.PathValue("name")

	found, err := s.Store.GetDiscoveryEndpoint(r.Context(), universeID, ns, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting endpoint: %v", err), http.StatusInternalServerError)
		return
	}

	if found == nil {
		http.NotFound(w, r)
		return
	}

	s.writeJSON(w, http.StatusOK, found)
}
