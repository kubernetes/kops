/*
Copyright 2019 The Kubernetes Authors.

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

package mockec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

func (m *MockEC2) DescribeInstances(ctx context.Context, request *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	klog.Warningf("MockEc2::DescribeInstances is stub-implemented")
	return &ec2.DescribeInstancesOutput{}, nil
}

func (m *MockEC2) DescribeInstanceTypes(ctx context.Context, request *ec2.DescribeInstanceTypesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTypesOutput, error) {
	klog.Warningf("MockEc2::DescribeInstanceTypes is stub-implemented")
	return &ec2.DescribeInstanceTypesOutput{}, nil
}

func (m *MockEC2) GetInstanceTypesFromInstanceRequirements(ctx context.Context, request *ec2.GetInstanceTypesFromInstanceRequirementsInput, optFns ...func(*ec2.Options)) (*ec2.GetInstanceTypesFromInstanceRequirementsOutput, error) {
	return &ec2.GetInstanceTypesFromInstanceRequirementsOutput{
		InstanceTypes: []ec2types.InstanceTypeInfoFromInstanceRequirements{
			{
				InstanceType: aws.String("c5.large"),
			},
		},
	}, nil
}
