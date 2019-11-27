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

func (m *MockIAM) GetInstanceProfile(request *iam.GetInstanceProfileInput) (*iam.GetInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

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
}
func (m *MockIAM) GetInstanceProfileRequest(*iam.GetInstanceProfileInput) (*request.Request, *iam.GetInstanceProfileOutput) {
	panic("Not implemented")
}

func (m *MockIAM) CreateInstanceProfile(request *iam.CreateInstanceProfileInput) (*iam.CreateInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateInstanceProfile: %v", request)

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
	// 	// about IDs, see IAM Identifiers (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html)
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
}
func (m *MockIAM) CreateInstanceProfileRequest(*iam.CreateInstanceProfileInput) (*request.Request, *iam.CreateInstanceProfileOutput) {
	panic("Not implemented")
}
func (m *MockIAM) AddRoleToInstanceProfile(request *iam.AddRoleToInstanceProfileInput) (*iam.AddRoleToInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AddRoleToInstanceProfile: %v", request)

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
}

func (m *MockIAM) AddRoleToInstanceProfileRequest(*iam.AddRoleToInstanceProfileInput) (*request.Request, *iam.AddRoleToInstanceProfileOutput) {
	panic("Not implemented")
}

func (m *MockIAM) RemoveRoleFromInstanceProfile(request *iam.RemoveRoleFromInstanceProfileInput) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("RemoveRoleFromInstanceProfile: %v", request)

	ip := m.InstanceProfiles[aws.StringValue(request.InstanceProfileName)]
	if ip == nil {
		return nil, fmt.Errorf("InstanceProfile not found")
	}

	found := false
	var newRoles []*iam.Role
	for _, role := range ip.Roles {
		if aws.StringValue(role.RoleName) == aws.StringValue(request.RoleName) {
			found = true
			continue
		}
		newRoles = append(newRoles, role)
	}

	if !found {
		return nil, fmt.Errorf("Role not found")
	}
	ip.Roles = newRoles

	return &iam.RemoveRoleFromInstanceProfileOutput{}, nil
}

func (m *MockIAM) RemoveRoleFromInstanceProfileWithContext(aws.Context, *iam.RemoveRoleFromInstanceProfileInput, ...request.Option) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) RemoveRoleFromInstanceProfileRequest(*iam.RemoveRoleFromInstanceProfileInput) (*request.Request, *iam.RemoveRoleFromInstanceProfileOutput) {
	panic("Not implemented")
}

func (m *MockIAM) ListInstanceProfiles(request *iam.ListInstanceProfilesInput) (*iam.ListInstanceProfilesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListInstanceProfiles: %v", request)

	if request.PathPrefix != nil {
		klog.Fatalf("MockIAM ListInstanceProfiles PathPrefix not implemented")
	}

	var instanceProfiles []*iam.InstanceProfile

	for _, ip := range m.InstanceProfiles {
		copy := *ip
		instanceProfiles = append(instanceProfiles, &copy)
	}

	response := &iam.ListInstanceProfilesOutput{
		InstanceProfiles: instanceProfiles,
	}

	return response, nil
}

func (m *MockIAM) ListInstanceProfilesWithContext(aws.Context, *iam.ListInstanceProfilesInput, ...request.Option) (*iam.ListInstanceProfilesOutput, error) {
	panic("Not implemented")
}

func (m *MockIAM) ListInstanceProfilesRequest(*iam.ListInstanceProfilesInput) (*request.Request, *iam.ListInstanceProfilesOutput) {
	panic("Not implemented")
}

func (m *MockIAM) ListInstanceProfilesPages(request *iam.ListInstanceProfilesInput, callback func(*iam.ListInstanceProfilesOutput, bool) bool) error {
	// For the mock, we just send everything in one page
	page, err := m.ListInstanceProfiles(request)
	if err != nil {
		return err
	}

	callback(page, false)

	return nil
}

func (m *MockIAM) ListInstanceProfilesPagesWithContext(aws.Context, *iam.ListInstanceProfilesInput, func(*iam.ListInstanceProfilesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockIAM) DeleteInstanceProfile(request *iam.DeleteInstanceProfileInput) (*iam.DeleteInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteInstanceProfile: %v", request)

	id := aws.StringValue(request.InstanceProfileName)
	o := m.InstanceProfiles[id]
	if o == nil {
		return nil, fmt.Errorf("InstanceProfile %q not found", id)
	}
	delete(m.InstanceProfiles, id)

	return &iam.DeleteInstanceProfileOutput{}, nil
}

func (m *MockIAM) DeleteInstanceProfileWithContext(aws.Context, *iam.DeleteInstanceProfileInput, ...request.Option) (*iam.DeleteInstanceProfileOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) DeleteInstanceProfileRequest(*iam.DeleteInstanceProfileInput) (*request.Request, *iam.DeleteInstanceProfileOutput) {
	panic("Not implemented")
}
