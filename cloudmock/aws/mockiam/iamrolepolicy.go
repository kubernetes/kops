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

package mockiam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog"
)

type rolePolicy struct {
	PolicyDocument string
	PolicyName     string
	RoleName       string
}

func (m *MockIAM) GetRolePolicy(request *iam.GetRolePolicyInput) (*iam.GetRolePolicyOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, rp := range m.RolePolicies {
		if rp.PolicyName != aws.StringValue(request.PolicyName) {
			// TODO: check regex?
			continue
		}
		if rp.RoleName != aws.StringValue(request.RoleName) {
			// TODO: check regex?
			continue
		}

		response := &iam.GetRolePolicyOutput{
			RoleName:       aws.String(rp.RoleName),
			PolicyDocument: aws.String(rp.PolicyDocument),
			PolicyName:     aws.String(rp.PolicyName),
		}
		return response, nil
	}
	return nil, awserr.New("NoSuchEntity", "No such entity", nil)
}
func (m *MockIAM) GetRolePolicyWithContext(aws.Context, *iam.GetRolePolicyInput, ...request.Option) (*iam.GetRolePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) GetRolePolicyRequest(*iam.GetRolePolicyInput) (*request.Request, *iam.GetRolePolicyOutput) {
	panic("Not implemented")
}

func (m *MockIAM) PutRolePolicy(request *iam.PutRolePolicyInput) (*iam.PutRolePolicyOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("PutRolePolicy: %v", request)

	for _, rp := range m.RolePolicies {
		if rp.PolicyName != aws.StringValue(request.PolicyName) {
			// TODO: check regex?
			continue
		}
		if rp.RoleName != aws.StringValue(request.RoleName) {
			// TODO: check regex?
			continue
		}

		rp.PolicyDocument = aws.StringValue(request.PolicyDocument)
		return &iam.PutRolePolicyOutput{}, nil
	}

	m.RolePolicies = append(m.RolePolicies, &rolePolicy{
		PolicyDocument: aws.StringValue(request.PolicyDocument),
		PolicyName:     aws.StringValue(request.PolicyName),
		RoleName:       aws.StringValue(request.RoleName),
	})

	return &iam.PutRolePolicyOutput{}, nil
}
func (m *MockIAM) PutRolePolicyWithContext(aws.Context, *iam.PutRolePolicyInput, ...request.Option) (*iam.PutRolePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) PutRolePolicyRequest(*iam.PutRolePolicyInput) (*request.Request, *iam.PutRolePolicyOutput) {
	panic("Not implemented")
}

func (m *MockIAM) ListRolePolicies(request *iam.ListRolePoliciesInput) (*iam.ListRolePoliciesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListRolePolicies: %v", request)

	var policyNames []string

	for _, r := range m.RolePolicies {
		if request.RoleName != nil {
			if r.RoleName != aws.StringValue(request.RoleName) {
				continue
			}
		}
		policyNames = append(policyNames, r.PolicyName)
	}

	response := &iam.ListRolePoliciesOutput{
		PolicyNames: aws.StringSlice(policyNames),
	}

	return response, nil
}

func (m *MockIAM) ListRolePoliciesWithContext(aws.Context, *iam.ListRolePoliciesInput, ...request.Option) (*iam.ListRolePoliciesOutput, error) {
	panic("Not implemented")
}

func (m *MockIAM) ListRolePoliciesRequest(*iam.ListRolePoliciesInput) (*request.Request, *iam.ListRolePoliciesOutput) {
	panic("Not implemented")
}

func (m *MockIAM) ListRolePoliciesPages(request *iam.ListRolePoliciesInput, callback func(*iam.ListRolePoliciesOutput, bool) bool) error {
	// For the mock, we just send everything in one page
	page, err := m.ListRolePolicies(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockIAM) ListRolePoliciesPagesWithContext(aws.Context, *iam.ListRolePoliciesInput, func(*iam.ListRolePoliciesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockIAM) DeleteRolePolicy(request *iam.DeleteRolePolicyInput) (*iam.DeleteRolePolicyOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteRolePolicy: %v", request)

	found := false
	var newRolePolicies []*rolePolicy
	for _, rp := range m.RolePolicies {
		if rp.PolicyName == aws.StringValue(request.PolicyName) && rp.RoleName == aws.StringValue(request.RoleName) {
			found = true
			continue
		}
		newRolePolicies = append(newRolePolicies, rp)
	}
	if !found {
		return nil, fmt.Errorf("RolePolicy not found")
	}
	m.RolePolicies = newRolePolicies

	return &iam.DeleteRolePolicyOutput{}, nil
}

func (m *MockIAM) DeleteRolePolicyWithContext(aws.Context, *iam.DeleteRolePolicyInput, ...request.Option) (*iam.DeleteRolePolicyOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) DeleteRolePolicyRequest(*iam.DeleteRolePolicyInput) (*request.Request, *iam.DeleteRolePolicyOutput) {
	panic("Not implemented")
}
