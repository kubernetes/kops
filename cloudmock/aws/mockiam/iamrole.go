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

func (m *MockIAM) GetRole(request *iam.GetRoleInput) (*iam.GetRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	role := m.Roles[aws.StringValue(request.RoleName)]
	if role == nil {
		return nil, awserr.New("NoSuchEntity", "No such entity", nil)
	}
	response := &iam.GetRoleOutput{
		Role: role,
	}
	return response, nil
}

func (m *MockIAM) GetRoleWithContext(aws.Context, *iam.GetRoleInput, ...request.Option) (*iam.GetRoleOutput, error) {
	panic("Not implemented")
}

func (m *MockIAM) GetRoleRequest(*iam.GetRoleInput) (*request.Request, *iam.GetRoleOutput) {
	panic("Not implemented")
}

func (m *MockIAM) CreateRole(request *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateRole: %v", request)

	roleID := m.createID()
	r := &iam.Role{
		AssumeRolePolicyDocument: request.AssumeRolePolicyDocument,
		Description:              request.Description,
		Path:                     request.Path,
		RoleName:                 request.RoleName,
		RoleId:                   &roleID,
	}

	if m.Roles == nil {
		m.Roles = make(map[string]*iam.Role)
	}
	m.Roles[*r.RoleName] = r

	copy := *r
	return &iam.CreateRoleOutput{Role: &copy}, nil
}

func (m *MockIAM) CreateRoleWithContext(aws.Context, *iam.CreateRoleInput, ...request.Option) (*iam.CreateRoleOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) CreateRoleRequest(*iam.CreateRoleInput) (*request.Request, *iam.CreateRoleOutput) {
	panic("Not implemented")
}

func (m *MockIAM) ListRoles(request *iam.ListRolesInput) (*iam.ListRolesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListRoles: %v", request)

	if request.PathPrefix != nil {
		klog.Fatalf("MockIAM ListRoles PathPrefix not implemented")
	}

	var roles []*iam.Role

	for _, r := range m.Roles {
		copy := *r
		roles = append(roles, &copy)
	}

	response := &iam.ListRolesOutput{
		Roles: roles,
	}

	return response, nil
}

func (m *MockIAM) ListRolesWithContext(aws.Context, *iam.ListRolesInput, ...request.Option) (*iam.ListRolesOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) ListRolesRequest(*iam.ListRolesInput) (*request.Request, *iam.ListRolesOutput) {
	panic("Not implemented")
}

func (m *MockIAM) ListRolesPages(request *iam.ListRolesInput, callback func(*iam.ListRolesOutput, bool) bool) error {
	// For the mock, we just send everything in one page
	page, err := m.ListRoles(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockIAM) ListRolesPagesWithContext(aws.Context, *iam.ListRolesInput, func(*iam.ListRolesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockIAM) DeleteRole(request *iam.DeleteRoleInput) (*iam.DeleteRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteRole: %v", request)

	id := aws.StringValue(request.RoleName)
	o := m.Roles[id]
	if o == nil {
		return nil, fmt.Errorf("Role %q not found", id)
	}
	delete(m.Roles, id)

	return &iam.DeleteRoleOutput{}, nil
}
func (m *MockIAM) DeleteRoleWithContext(aws.Context, *iam.DeleteRoleInput, ...request.Option) (*iam.DeleteRoleOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) DeleteRoleRequest(*iam.DeleteRoleInput) (*request.Request, *iam.DeleteRoleOutput) {
	panic("Not implemented")
}
