/*
Copyright 2024 The Kubernetes Authors.

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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
)

// MockedInstanceTopologyManager creates an InstanceTopologyManager mock.
type MockedInstanceTopologyManager struct {
	InstanceTopologyManager
	mock.Mock
}

// GetNodeTopology mocks InstanceTopologyManager.GetNodeTopology.
func (m *MockedInstanceTopologyManager) GetNodeTopology(ctx context.Context, instanceType string, region string, instanceID string) (*types.InstanceTopology, error) {
	args := m.Called(ctx, instanceType, region, instanceID)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*types.InstanceTopology), nil
}

// DoesInstanceTypeRequireResponse mocks InstanceTopologyManager.DoesInstanceTypeRequireResponse.
func (m *MockedInstanceTopologyManager) DoesInstanceTypeRequireResponse(instanceType string) bool {
	args := m.Called(instanceType)
	return args.Get(0).(bool)
}
