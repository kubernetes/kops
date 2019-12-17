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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func (m *MockEC2) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	klog.Warningf("MockEc2::DescribeInstances is stub-implemented")
	return &ec2.DescribeInstancesOutput{}, nil
}

func (m *MockEC2) DescribeInstancesWithContext(aws.Context, *ec2.DescribeInstancesInput, ...request.Option) (*ec2.DescribeInstancesOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeInstancesRequest(*ec2.DescribeInstancesInput) (*request.Request, *ec2.DescribeInstancesOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeInstancesPages(request *ec2.DescribeInstancesInput, callback func(*ec2.DescribeInstancesOutput, bool) bool) error {
	// For the mock, we just send everything in one page
	page, err := m.DescribeInstances(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}
func (m *MockEC2) DescribeInstancesPagesWithContext(aws.Context, *ec2.DescribeInstancesInput, func(*ec2.DescribeInstancesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}
