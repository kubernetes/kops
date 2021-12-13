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

package mockcloudresourcemanager

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"google.golang.org/api/cloudresourcemanager/v1"
	"k8s.io/kops/cloudmock/gce/gcphttp"
)

type projects struct {
	mutex sync.Mutex

	projectBindings map[string]*cloudresourcemanager.Policy
}

func (s *projects) Init() {
	s.projectBindings = make(map[string]*cloudresourcemanager.Policy)
}

func (s *projects) getIAMPolicy(projectID string, request *http.Request) (*http.Response, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bindings := s.projectBindings[projectID]
	if bindings == nil {
		bindings = &cloudresourcemanager.Policy{
			Etag: nextEtag(""),
		}
	}

	return gcphttp.OKResponse(bindings)
}

func nextEtag(etag string) string {
	hash := sha256.Sum256([]byte(etag))
	nextEtag := hex.EncodeToString(hash[:])
	return nextEtag
}

func (s *projects) setIAMPolicy(projectID string, request *http.Request) (*http.Response, error) {
	b, err := io.ReadAll(request.Body)
	if err != nil {
		return gcphttp.ErrorBadRequest("")
	}

	req := &cloudresourcemanager.SetIamPolicyRequest{}
	if err := json.Unmarshal(b, &req); err != nil {
		return gcphttp.ErrorBadRequest("")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldBindings := s.projectBindings[projectID]
	if oldBindings == nil {
		oldBindings = &cloudresourcemanager.Policy{
			Etag: nextEtag(""),
		}
	}

	newBindings := req.Policy

	if oldBindings.Etag != newBindings.Etag {
		// TODO: What is the actual error?
		return gcphttp.ErrorNotFound("etag")
	}

	newBindings.Etag = nextEtag(oldBindings.Etag)
	s.projectBindings[projectID] = newBindings

	return gcphttp.OKResponse(newBindings)
}
