/*
Copyright 2021 The Kubernetes Authors.

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

package mockiam

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"google.golang.org/api/iam/v1"
	"k8s.io/kops/cloudmock/gce/gcphttp"
)

// serviceAccounts manages the ServiceAccount resources.
type serviceAccounts struct {
	mutex sync.Mutex

	serviceAccountsByEmail map[string]*iam.ServiceAccount
}

func (s *serviceAccounts) Init() {
	s.serviceAccountsByEmail = make(map[string]*iam.ServiceAccount)
}

func (s *serviceAccounts) Get(projectID string, serviceAccount string, request *http.Request) (*http.Response, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sa := s.serviceAccountsByEmail[serviceAccount]
	if sa == nil {
		return gcphttp.ErrorNotFound("Unknown service account")
	}

	return gcphttp.OKResponse(sa)
}

func (s *serviceAccounts) Create(projectID string, request *http.Request) (*http.Response, error) {
	b, err := io.ReadAll(request.Body)
	if err != nil {
		return gcphttp.ErrorBadRequest("")
	}

	req := &iam.CreateServiceAccountRequest{}
	if err := json.Unmarshal(b, &req); err != nil {
		return gcphttp.ErrorBadRequest("")
	}

	if req.AccountId == "" {
		return gcphttp.ErrorBadRequest("")
	}

	sa := &iam.ServiceAccount{
		Email:       req.AccountId + "@" + projectID + ".iam.gserviceaccount.com",
		Description: req.ServiceAccount.Description,
		DisplayName: req.ServiceAccount.DisplayName,
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	existing := s.serviceAccountsByEmail[sa.Email]
	if existing != nil {
		// TODO: details
		// 	"details": [
		//   {
		//     "@type": "type.googleapis.com/google.rpc.ResourceInfo",
		//     "resourceName": "projects/testproject/serviceAccounts/testaccount@testproject.iam.gserviceaccount.com"
		//   }

		return gcphttp.ErrorAlreadyExists("Service account %s already exists within project projects/%s.", req.AccountId, projectID)
	}

	s.serviceAccountsByEmail[sa.Email] = sa

	return gcphttp.OKResponse(sa)
}
