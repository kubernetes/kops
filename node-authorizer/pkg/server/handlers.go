/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
	"k8s.io/kops/node-authorizer/pkg/utils"

	"github.com/gorilla/mux"
)

// authorizeHandler is responsible for handling the authorization requests
func (n *NodeAuthorizer) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		// @check we have a body to read in
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}

		// @step: read in the request body
		content, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			return err
		}

		address, err := getClientAddress(r.RemoteAddr)
		if err != nil {
			return err
		}

		// @step: construct the node registration request
		req := &NodeRegistration{
			Spec: NodeRegistrationSpec{
				NodeName:   mux.Vars(r)["name"],
				RemoteAddr: address,
				Request:    content,
			},
		}

		// @step: attempt to authorise the request
		if err := n.authorizeNodeRequest(r.Context(), req); err != nil {
			return err
		}

		// @check if the node was denied and if so, 403 it
		if !req.Status.Allowed {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}

		return json.NewEncoder(w).Encode(req)
	}()
	if err != nil {
		utils.Logger.Info("failed to handle node request", zap.Error(err))

		w.WriteHeader(http.StatusInternalServerError)
	}
}

// healthHandler is responsible for providing health
func (n *NodeAuthorizer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Node-Authorizer-Version", Version)
	w.Write([]byte("OK\n"))
}
