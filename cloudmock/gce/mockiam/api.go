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
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/iam/v1"
	option "google.golang.org/api/option"
	"k8s.io/klog/v2"
)

// mockIAMService represents a mocked IAM client.
type mockIAMService struct {
	svc *iam.Service

	serviceAccounts serviceAccounts
}

// New creates a new mock IAM client.
func New(project string) *iam.Service {
	ctx := context.Background()

	s := &mockIAMService{}

	s.serviceAccounts.Init()

	httpClient := &http.Client{Transport: s}
	svc, err := iam.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		klog.Fatalf("failed to build mock iam service: %v", err)
	}
	s.svc = svc
	return svc
}

func (s *mockIAMService) RoundTrip(request *http.Request) (*http.Response, error) {
	url := request.URL
	if url.Host != "iam.googleapis.com" {
		return nil, fmt.Errorf("unexpected host in request %#v", request)
	}

	pathTokens := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
	if len(pathTokens) >= 1 && pathTokens[0] == "v1" {
		if len(pathTokens) >= 3 && pathTokens[1] == "projects" {
			projectID := pathTokens[2]
			if len(pathTokens) >= 5 && pathTokens[3] == "serviceAccounts" {
				serviceAccount := pathTokens[4]
				if len(pathTokens) == 5 && request.Method == "GET" {
					return s.serviceAccounts.Get(projectID, serviceAccount, request)
				}
			}

			if len(pathTokens) == 4 && pathTokens[3] == "serviceAccounts" && request.Method == "POST" {
				return s.serviceAccounts.Create(projectID, request)
			}
		}
	}

	klog.Warningf("request: %s %s %#v", request.Method, request.URL, request)
	return nil, fmt.Errorf("unhandled request %#v", request)
}
