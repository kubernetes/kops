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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"k8s.io/klog/v2"
)

func (m *MockIAM) GetInstanceProfile(ctx context.Context, request *iam.GetInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.GetInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ip, ok := m.InstanceProfiles[aws.ToString(request.InstanceProfileName)]
	if !ok || strings.Contains(aws.ToString(ip.InstanceProfileName), "__no_entity__") {
		return nil, &iamtypes.NoSuchEntityException{}
	}
	response := &iam.GetInstanceProfileOutput{
		InstanceProfile: ip,
	}
	return response, nil
}

func (m *MockIAM) CreateInstanceProfile(ctx context.Context, request *iam.CreateInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.CreateInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateInstanceProfile: %v", request)

	p := iamtypes.InstanceProfile{
		InstanceProfileName: request.InstanceProfileName,
		// Arn:                 request.Arn,
		// InstanceProfileId:   request.InstanceProfileId,
		Path: request.Path,
		Tags: request.Tags,
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
		m.InstanceProfiles = make(map[string]*iamtypes.InstanceProfile)
	}
	m.InstanceProfiles[*p.InstanceProfileName] = &p

	copy := p
	return &iam.CreateInstanceProfileOutput{InstanceProfile: &copy}, nil
}

func (m *MockIAM) TagInstanceProfile(ctx context.Context, request *iam.TagInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.TagInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateInstanceProfile: %v", request)

	ip, ok := m.InstanceProfiles[aws.ToString(request.InstanceProfileName)]
	if !ok {
		return nil, fmt.Errorf("InstanceProfile not found")
	}

	for _, tag := range request.Tags {
		key := *tag.Key
		overwritten := false
		for _, existingTag := range ip.Tags {
			if *existingTag.Key == key {
				existingTag.Value = tag.Value
				overwritten = true
				break
			}
		}
		if !overwritten {
			ip.Tags = append(ip.Tags, tag)
		}
	}
	return &iam.TagInstanceProfileOutput{}, nil
}

func (m *MockIAM) AddRoleToInstanceProfile(ctx context.Context, request *iam.AddRoleToInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.AddRoleToInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AddRoleToInstanceProfile: %v", request)

	ip, ok := m.InstanceProfiles[aws.ToString(request.InstanceProfileName)]
	if !ok {
		return nil, fmt.Errorf("InstanceProfile not found")
	}
	r, ok := m.Roles[aws.ToString(request.RoleName)]
	if !ok {
		return nil, fmt.Errorf("Role not found")
	}

	ip.Roles = append(ip.Roles, *r)

	return &iam.AddRoleToInstanceProfileOutput{}, nil
}

func (m *MockIAM) RemoveRoleFromInstanceProfile(ctx context.Context, request *iam.RemoveRoleFromInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("RemoveRoleFromInstanceProfile: %v", request)

	ip, ok := m.InstanceProfiles[aws.ToString(request.InstanceProfileName)]
	if !ok {
		return nil, fmt.Errorf("InstanceProfile not found")
	}

	found := false
	var newRoles []iamtypes.Role
	for _, role := range ip.Roles {
		if aws.ToString(role.RoleName) == aws.ToString(request.RoleName) {
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

func (m *MockIAM) ListInstanceProfiles(ctx context.Context, request *iam.ListInstanceProfilesInput, optFns ...func(*iam.Options)) (*iam.ListInstanceProfilesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListInstanceProfiles: %v", request)

	if request.PathPrefix != nil {
		klog.Fatalf("MockIAM ListInstanceProfiles PathPrefix not implemented")
	}

	var instanceProfiles []iamtypes.InstanceProfile

	for _, ip := range m.InstanceProfiles {
		copy := *ip
		instanceProfiles = append(instanceProfiles, copy)
	}

	response := &iam.ListInstanceProfilesOutput{
		InstanceProfiles: instanceProfiles,
	}

	return response, nil
}

func (m *MockIAM) DeleteInstanceProfile(ctx context.Context, request *iam.DeleteInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.DeleteInstanceProfileOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteInstanceProfile: %v", request)

	id := aws.ToString(request.InstanceProfileName)
	_, ok := m.InstanceProfiles[id]
	if !ok {
		return nil, fmt.Errorf("InstanceProfile %q not found", id)
	}
	delete(m.InstanceProfiles, id)

	return &iam.DeleteInstanceProfileOutput{}, nil
}
