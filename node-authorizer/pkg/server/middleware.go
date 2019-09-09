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
	"fmt"
	"net/http"

	"k8s.io/kops/node-authorizer/pkg/utils"

	"go.uber.org/zap"
)

// recovery is responsible for ensuring we don't exit on a panic
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				utils.Logger.Error("failed to handle request, threw exception",
					zap.String("error", fmt.Sprintf("%v", err)))
			}
		}()

		next.ServeHTTP(w, req)
	})
}

// authorized is responsible for validating the client certificate
func authorized(next http.HandlerFunc, commonName string, requireAuth bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requireAuth {
			found := func() bool {
				for _, x := range r.TLS.PeerCertificates {
					if x.Subject.CommonName == commonName {
						return true
					}
				}

				return false
			}()
			if !found {
				utils.Logger.Error("invalid client certificate",
					zap.String("client", r.RemoteAddr))

				w.WriteHeader(http.StatusForbidden)

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
