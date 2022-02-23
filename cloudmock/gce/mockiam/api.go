/*
Copyright 2022 The Kubernetes Authors.

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
	"google.golang.org/api/googleapi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// MockClient represents a mocked IAM client.
type MockClient struct {
	serviceAccounts *serviceAccountClient
}

var _ gce.IamClient = &MockClient{}

// NewMockClient creates a new mock client.
func NewMockClient(project string) *MockClient {
	return &MockClient{
		serviceAccounts: newServiceAccounts(project),
	}
}

func (c *MockClient) ServiceAccounts() gce.ServiceAccountClient {
	return c.serviceAccounts
}

func notFoundError() error {
	return &googleapi.Error{
		Code: 404,
	}
}
