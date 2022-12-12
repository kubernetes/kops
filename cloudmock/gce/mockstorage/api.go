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

package mockstorage

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	option "google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
	"k8s.io/klog/v2"
)

// mockStorageService represents a mocked storage client.
type mockStorageService struct {
	svc *storage.Service

	buckets buckets
}

// New creates a new mock IAM client.
func New() *storage.Service {
	ctx := context.Background()

	s := &mockStorageService{}

	s.buckets.Init()

	httpClient := &http.Client{Transport: s}
	svc, err := storage.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		klog.Fatalf("failed to build mock storage service: %v", err)
	}
	s.svc = svc
	return svc
}

func (s *mockStorageService) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := request.Context()

	url := request.URL
	if url.Host != "storage.googleapis.com" {
		return nil, fmt.Errorf("unexpected host in request %#v", request)
	}

	pathTokens := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")

	if len(pathTokens) >= 3 && pathTokens[0] == "storage" && pathTokens[1] == "v1" && pathTokens[2] == "b" {
		if len(pathTokens) == 3 {
			if request.Method == "POST" {
				return s.buckets.createBucket(request)
			}
		}

		if len(pathTokens) == 4 {
			bucketName := pathTokens[3]
			if request.Method == "GET" {
				return s.buckets.getBucket(bucketName, request)
			}
		}

		if len(pathTokens) == 5 && pathTokens[4] == "iam" {
			bucketName := pathTokens[3]
			if request.Method == "GET" {
				return s.buckets.getIAMPolicy(bucketName, request)
			}

			if request.Method == "PUT" {
				return s.buckets.setIAMPolicy(bucketName, request)
			}
		}

		if len(pathTokens) >= 6 && pathTokens[4] == "o" {
			bucketName := pathTokens[3]
			objectName := strings.Join(pathTokens[5:], "/")
			if request.Method == "GET" {
				return s.buckets.getObject(ctx, bucketName, objectName, request)
			}
		}

	}

	if len(pathTokens) == 6 && pathTokens[0] == "upload" && pathTokens[1] == "storage" && pathTokens[2] == "v1" && pathTokens[3] == "b" && pathTokens[5] == "o" {
		bucket := pathTokens[4]

		if request.Method == "POST" {
			return s.buckets.createObject(ctx, bucket, request)
		}
	}

	klog.Warningf("request: %s %s %#v", request.Method, request.URL, request)
	return nil, fmt.Errorf("unhandled request (pathTokens=%v) %#v", pathTokens, request)
}
