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

func (m *MockIAM) GetRole(ctx context.Context, request *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	role := m.Roles[aws.ToString(request.RoleName)]
	if role == nil {
		return nil, &iamtypes.NoSuchEntityException{}
	}
	response := &iam.GetRoleOutput{
		Role: role,
	}
	return response, nil
}

func (m *MockIAM) CreateRole(ctx context.Context, request *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	roleID := m.createID()
	r := iamtypes.Role{
		AssumeRolePolicyDocument: request.AssumeRolePolicyDocument,
		Description:              request.Description,
		Path:                     request.Path,
		PermissionsBoundary: &iamtypes.AttachedPermissionsBoundary{
			PermissionsBoundaryArn: request.PermissionsBoundary,
		},
		RoleName: request.RoleName,
		RoleId:   &roleID,
		Tags:     request.Tags,
	}

	if m.Roles == nil {
		m.Roles = make(map[string]*iamtypes.Role)
	}
	m.Roles[*r.RoleName] = &r

	copy := r
	return &iam.CreateRoleOutput{Role: &copy}, nil
}

func (m *MockIAM) ListRoles(ctx context.Context, request *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListRoles: %v", request)

	if request.PathPrefix != nil {
		klog.Fatalf("MockIAM ListRoles PathPrefix not implemented")
	}

	var roles []iamtypes.Role

	for _, r := range m.Roles {
		copy := *r
		roles = append(roles, copy)
	}

	response := &iam.ListRolesOutput{
		Roles: roles,
	}

	return response, nil
}

func (m *MockIAM) DeleteRole(ctx context.Context, request *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteRole: %v", request)

	id := aws.ToString(request.RoleName)
	o := m.Roles[id]
	if o == nil {
		return nil, fmt.Errorf("role %q not found", id)
	}
	delete(m.Roles, id)

	return &iam.DeleteRoleOutput{}, nil
}

func (m *MockIAM) ListAttachedRolePolicies(ctx context.Context, request *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListAttachedRolePolicies: %s", aws.ToString(request.RoleName))

	for _, r := range m.Roles {
		if r.RoleName == request.RoleName {
			role := aws.ToString(r.RoleName)

			return &iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: m.AttachedPolicies[role],
			}, nil
		}
	}

	return &iam.ListAttachedRolePoliciesOutput{}, nil
}
