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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"k8s.io/klog/v2"
)

type rolePolicy struct {
	PolicyDocument string
	PolicyName     string
	RoleName       string
}

func (m *MockIAM) GetRolePolicy(ctx context.Context, request *iam.GetRolePolicyInput, optFns ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, rp := range m.RolePolicies {
		if rp.PolicyName != aws.ToString(request.PolicyName) {
			// TODO: check regex?
			continue
		}
		if rp.RoleName != aws.ToString(request.RoleName) {
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
	return nil, &iamtypes.NoSuchEntityException{}
}

func (m *MockIAM) PutRolePolicy(ctx context.Context, request *iam.PutRolePolicyInput, optFns ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("PutRolePolicy: %v", request)

	for _, rp := range m.RolePolicies {
		if rp.PolicyName != aws.ToString(request.PolicyName) {
			// TODO: check regex?
			continue
		}
		if rp.RoleName != aws.ToString(request.RoleName) {
			// TODO: check regex?
			continue
		}

		rp.PolicyDocument = aws.ToString(request.PolicyDocument)
		return &iam.PutRolePolicyOutput{}, nil
	}

	m.RolePolicies = append(m.RolePolicies, &rolePolicy{
		PolicyDocument: aws.ToString(request.PolicyDocument),
		PolicyName:     aws.ToString(request.PolicyName),
		RoleName:       aws.ToString(request.RoleName),
	})

	return &iam.PutRolePolicyOutput{}, nil
}

func (m *MockIAM) ListRolePolicies(ctx context.Context, request *iam.ListRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListRolePoliciesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListRolePolicies: %v", request)

	var policyNames []string

	for _, r := range m.RolePolicies {
		if request.RoleName != nil {
			if r.RoleName != aws.ToString(request.RoleName) {
				continue
			}
		}
		policyNames = append(policyNames, r.PolicyName)
	}

	response := &iam.ListRolePoliciesOutput{
		PolicyNames: policyNames,
	}

	return response, nil
}

func (m *MockIAM) DeleteRolePolicy(ctx context.Context, request *iam.DeleteRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DeleteRolePolicyOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteRolePolicy: %v", request)

	found := false
	var newRolePolicies []*rolePolicy
	for _, rp := range m.RolePolicies {
		if rp.PolicyName == aws.ToString(request.PolicyName) && rp.RoleName == aws.ToString(request.RoleName) {
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
