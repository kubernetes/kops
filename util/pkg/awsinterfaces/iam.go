/*
Copyright 2024 The Kubernetes Authors.

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

package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IAMAPI interface {
	AddClientIDToOpenIDConnectProvider(ctx context.Context, params *iam.AddClientIDToOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)
	AddRoleToInstanceProfile(ctx context.Context, params *iam.AddRoleToInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.AddRoleToInstanceProfileOutput, error)
	AttachRolePolicy(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
	CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	CreateInstanceProfile(ctx context.Context, params *iam.CreateInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.CreateInstanceProfileOutput, error)
	CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	DeleteInstanceProfile(ctx context.Context, params *iam.DeleteInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.DeleteInstanceProfileOutput, error)
	DeleteOpenIDConnectProvider(ctx context.Context, params *iam.DeleteOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error)
	DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	DeleteRolePermissionsBoundary(ctx context.Context, params *iam.DeleteRolePermissionsBoundaryInput, optFns ...func(*iam.Options)) (*iam.DeleteRolePermissionsBoundaryOutput, error)
	DeleteRolePolicy(ctx context.Context, params *iam.DeleteRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DeleteRolePolicyOutput, error)
	DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error)
	GetInstanceProfile(ctx context.Context, params *iam.GetInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.GetInstanceProfileOutput, error)
	GetOpenIDConnectProvider(ctx context.Context, params *iam.GetOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error)
	GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	GetRolePolicy(ctx context.Context, params *iam.GetRolePolicyInput, optFns ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error)
	ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	ListInstanceProfiles(ctx context.Context, params *iam.ListInstanceProfilesInput, optFns ...func(*iam.Options)) (*iam.ListInstanceProfilesOutput, error)
	ListOpenIDConnectProviders(ctx context.Context, params *iam.ListOpenIDConnectProvidersInput, optFns ...func(*iam.Options)) (*iam.ListOpenIDConnectProvidersOutput, error)
	ListRolePolicies(ctx context.Context, params *iam.ListRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListRolePoliciesOutput, error)
	ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error)
	PutRolePermissionsBoundary(ctx context.Context, params *iam.PutRolePermissionsBoundaryInput, optFns ...func(*iam.Options)) (*iam.PutRolePermissionsBoundaryOutput, error)
	PutRolePolicy(ctx context.Context, params *iam.PutRolePolicyInput, optFns ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error)
	RemoveClientIDFromOpenIDConnectProvider(ctx context.Context, params *iam.RemoveClientIDFromOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error)
	RemoveRoleFromInstanceProfile(ctx context.Context, params *iam.RemoveRoleFromInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.RemoveRoleFromInstanceProfileOutput, error)
	TagInstanceProfile(ctx context.Context, params *iam.TagInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.TagInstanceProfileOutput, error)
	TagOpenIDConnectProvider(ctx context.Context, params *iam.TagOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.TagOpenIDConnectProviderOutput, error)
	TagRole(ctx context.Context, params *iam.TagRoleInput, optFns ...func(*iam.Options)) (*iam.TagRoleOutput, error)
	UntagInstanceProfile(ctx context.Context, params *iam.UntagInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.UntagInstanceProfileOutput, error)
	UntagOpenIDConnectProvider(ctx context.Context, params *iam.UntagOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.UntagOpenIDConnectProviderOutput, error)
	UpdateOpenIDConnectProviderThumbprint(ctx context.Context, params *iam.UpdateOpenIDConnectProviderThumbprintInput, optFns ...func(*iam.Options)) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error)
	UntagRole(ctx context.Context, params *iam.UntagRoleInput, optFns ...func(*iam.Options)) (*iam.UntagRoleOutput, error)
	UpdateAssumeRolePolicy(ctx context.Context, params *iam.UpdateAssumeRolePolicyInput, optFns ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error)
}
