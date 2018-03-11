/*
Copyright 2018 The Kubernetes Authors.

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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
)

func (m *MockIAM) GetRole(request *iam.GetRoleInput) (*iam.GetRoleOutput, error) {
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
	return nil, nil
}

func (m *MockIAM) GetRoleRequest(*iam.GetRoleInput) (*request.Request, *iam.GetRoleOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockIAM) CreateRole(request *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	glog.Infof("CreateRole: %v", request)

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
	return nil, nil
}
func (m *MockIAM) CreateRoleRequest(*iam.CreateRoleInput) (*request.Request, *iam.CreateRoleOutput) {
	panic("Not implemented")
	return nil, nil
}
