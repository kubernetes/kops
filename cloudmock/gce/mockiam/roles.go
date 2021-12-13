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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"google.golang.org/api/iam/v1"
)

type roles struct {
	mutex sync.Mutex

	roles map[string]*iam.Role
}

func (s *roles) Init() {
	s.roles = make(map[string]*iam.Role)
}

func (s *roles) Get(projectID string, roleID string, request *http.Request) (*http.Response, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sa := s.roles[roleID]
	if sa == nil {
		return errorResponse(http.StatusNotFound)
	}

	return jsonResponse(sa)
}

func (s *roles) Create(projectID string, request *http.Request) (*http.Response, error) {
	b, err := io.ReadAll(request.Body)
	if err != nil {
		return errorResponse(http.StatusBadRequest)
	}

	req := &iam.CreateRoleRequest{}
	if err := json.Unmarshal(b, &req); err != nil {
		return errorResponse(http.StatusBadRequest)
	}

	if req.RoleId == "" {
		return errorResponse(http.StatusBadRequest)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	existing := s.roles[req.RoleId]
	if existing != nil {
		return errorResponse(http.StatusConflict)
	}

	if existing.Deleted {
		return failedPrecondition("You can't create a role with role_id (%s) where there is an existing role with that role_id in a deleted state.", req.RoleId)
	}
	s.roles[req.RoleId] = req.Role

	return jsonResponse(req.Role)
}

type errorResponseData struct {
	Error errorInfo `json:"error"`
}

type errorInfo struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

func failedPrecondition(message string, args ...interface{}) (*http.Response, error) {
	statusCode := http.StatusBadRequest

	e := errorResponseData{
		Error: errorInfo{
			Code:    statusCode,
			Message: fmt.Sprintf(message, args...),
			Status:  "FAILED_PRECONDITION",
		},
	}

	b, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to build error response: %w", err)
	}
	r := &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
	}
	r.Header.Add("Content-Type", "application/json; charset=UTF-8")
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	return r, nil
}
