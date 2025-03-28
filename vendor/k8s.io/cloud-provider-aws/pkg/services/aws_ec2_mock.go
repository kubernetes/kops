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

package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
)

// MockedEC2ClientV2 mocks EC2ClientV2.
type MockedEC2ClientV2 struct {
	EC2ClientV2
	mock.Mock
}

// DescribeInstanceTopology mocks EC2ClientV2.DescribeInstanceTopology.
func (m *MockedEC2ClientV2) DescribeInstanceTopology(ctx context.Context, params *ec2.DescribeInstanceTopologyInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTopologyOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.DescribeInstanceTopologyOutput), nil
}

// MockedEc2SdkV2 is an implementation of the EC2 v2 interface, backed by aws-sdk-go-v2
type MockedEc2SdkV2 struct {
	Ec2SdkV2
	mock.Mock
}

// DescribeInstanceTopology mocks EC2ClientV2.DescribeInstanceTopology.
func (m *MockedEc2SdkV2) DescribeInstanceTopology(ctx context.Context, request *ec2.DescribeInstanceTopologyInput) ([]types.InstanceTopology, error) {
	args := m.Called(ctx, request)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).([]types.InstanceTopology), nil
}
