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

package mockdns

import (
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// MockClient represents a mocked DNS client.
type MockClient struct {
	managedZoneClient       *managedZoneClient
	resourceRecordSetClient *resourceRecordSetClient
	changeClient            *changeClient
}

var _ gce.DNSClient = &MockClient{}

// NewMockClient creates a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{
		managedZoneClient:       newManagedZoneClient(),
		resourceRecordSetClient: newResourceRecordSetClient(),
		changeClient:            newChangeClient(),
	}
}

func (c *MockClient) ManagedZones() gce.ManagedZoneClient {
	return c.managedZoneClient
}

func (c *MockClient) ResourceRecordSets() gce.ResourceRecordSetClient {
	return c.resourceRecordSetClient
}

func (c *MockClient) Changes() gce.ChangeClient {
	return c.changeClient
}
