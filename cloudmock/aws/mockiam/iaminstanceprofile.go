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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
)

func (m *MockIAM) GetInstanceProfile(request *iam.GetInstanceProfileInput) (*iam.GetInstanceProfileOutput, error) {
	ip := m.InstanceProfiles[aws.StringValue(request.InstanceProfileName)]
	if ip == nil {
		return nil, awserr.New("NoSuchEntity", "No such entity", nil)
	}
	response := &iam.GetInstanceProfileOutput{
		InstanceProfile: ip,
	}
	return response, nil
}
func (m *MockIAM) GetInstanceProfileWithContext(aws.Context, *iam.GetInstanceProfileInput, ...request.Option) (*iam.GetInstanceProfileOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockIAM) GetInstanceProfileRequest(*iam.GetInstanceProfileInput) (*request.Request, *iam.GetInstanceProfileOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockIAM) CreateInstanceProfile(request *iam.CreateInstanceProfileInput) (*iam.CreateInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	glog.Infof("CreateInstanceProfile: %v", request)

	p := &iam.InstanceProfile{
		InstanceProfileName: request.InstanceProfileName,
		// Arn:                 request.Arn,
		// InstanceProfileId:   request.InstanceProfileId,
		Path: request.Path,
		// Roles:               request.Roles,
	}

	// TODO: Some fields
	// // The date when the instance profile was created.
	// //
	// // CreateDate is a required field
	// CreateDate *time.Time `type:"timestamp" timestampFormat:"iso8601" required:"true"`
	// // The stable and unique string identifying the instance profile. For more information
	// 	// about IDs, see IAM Identifiers (http://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html)
	// 	// in the Using IAM guide.
	// 	//
	// 	// InstanceProfileId is a required field
	// 	InstanceProfileId *string `min:"16" type:"string" required:"true"`

	if m.InstanceProfiles == nil {
		m.InstanceProfiles = make(map[string]*iam.InstanceProfile)
	}
	m.InstanceProfiles[*p.InstanceProfileName] = p

	copy := *p
	return &iam.CreateInstanceProfileOutput{InstanceProfile: &copy}, nil
}
func (m *MockIAM) CreateInstanceProfileWithContext(aws.Context, *iam.CreateInstanceProfileInput, ...request.Option) (*iam.CreateInstanceProfileOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockIAM) CreateInstanceProfileRequest(*iam.CreateInstanceProfileInput) (*request.Request, *iam.CreateInstanceProfileOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockIAM) AddRoleToInstanceProfile(request *iam.AddRoleToInstanceProfileInput) (*iam.AddRoleToInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	glog.Infof("AddRoleToInstanceProfile: %v", request)

	ip := m.InstanceProfiles[aws.StringValue(request.InstanceProfileName)]
	if ip == nil {
		return nil, fmt.Errorf("InstanceProfile not found")
	}
	r := m.Roles[aws.StringValue(request.RoleName)]
	if r == nil {
		return nil, fmt.Errorf("Role not found")
	}

	ip.Roles = append(ip.Roles, r)

	return &iam.AddRoleToInstanceProfileOutput{}, nil
}
func (m *MockIAM) AddRoleToInstanceProfileWithContext(aws.Context, *iam.AddRoleToInstanceProfileInput, ...request.Option) (*iam.AddRoleToInstanceProfileOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockIAM) AddRoleToInstanceProfileRequest(*iam.AddRoleToInstanceProfileInput) (*request.Request, *iam.AddRoleToInstanceProfileOutput) {
	panic("Not implemented")
	return nil, nil
}
