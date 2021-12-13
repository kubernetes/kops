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
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v1"
	option "google.golang.org/api/option"
	"k8s.io/klog/v2"
)

// mockCloudResourceManagerService represents a mocked cloudresourcemanager client.
type mockCloudResourceManagerService struct {
	svc *cloudresourcemanager.Service

	projects projects
}

// New creates a new mock cloudresourcemanager client.
func New() *cloudresourcemanager.Service {
	ctx := context.Background()

	s := &mockCloudResourceManagerService{}

	s.projects.Init()

	httpClient := &http.Client{Transport: s}
	svc, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		klog.Fatalf("failed to build mock cloudresourcemanager service: %v", err)
	}
	s.svc = svc
	return svc
}

func (s *mockCloudResourceManagerService) RoundTrip(request *http.Request) (*http.Response, error) {
	url := request.URL
	if url.Host != "cloudresourcemanager.googleapis.com" {
		return nil, fmt.Errorf("unexpected host in request %#v", request)
	}

	pathTokens := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
	if len(pathTokens) >= 1 && pathTokens[0] == "v1" {
		if len(pathTokens) >= 3 && pathTokens[1] == "projects" {
			projectTokens := strings.Split(pathTokens[2], ":")
			if len(projectTokens) == 2 {
				projectID := projectTokens[0]
				verb := projectTokens[1]

				if request.Method == "POST" && verb == "getIamPolicy" {
					return s.projects.getIAMPolicy(projectID, request)
				}

				if request.Method == "POST" && verb == "setIamPolicy" {
					return s.projects.setIAMPolicy(projectID, request)
				}
			}
		}
	}

	// klog.Warningf("request: %s %s %#v", request.Method, request.URL, request)
	return nil, fmt.Errorf("unhandled request %#v", request)
}
