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

package mockkms

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/kms"
)

func (_ *MockKMS) CancelKeyDeletion(*kms.CancelKeyDeletionInput) (*kms.CancelKeyDeletionOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CancelKeyDeletionWithContext(aws.Context, *kms.CancelKeyDeletionInput, ...request.Option) (*kms.CancelKeyDeletionOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CancelKeyDeletionRequest(*kms.CancelKeyDeletionInput) (*request.Request, *kms.CancelKeyDeletionOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) CreateAliasWithContext(aws.Context, *kms.CreateAliasInput, ...request.Option) (*kms.CreateAliasOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CreateAliasRequest(*kms.CreateAliasInput) (*request.Request, *kms.CreateAliasOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) CreateGrant(*kms.CreateGrantInput) (*kms.CreateGrantOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CreateGrantWithContext(aws.Context, *kms.CreateGrantInput, ...request.Option) (*kms.CreateGrantOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CreateGrantRequest(*kms.CreateGrantInput) (*request.Request, *kms.CreateGrantOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CreateKeyWithContext(aws.Context, *kms.CreateKeyInput, ...request.Option) (*kms.CreateKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) CreateKeyRequest(*kms.CreateKeyInput) (*request.Request, *kms.CreateKeyOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) Decrypt(*kms.DecryptInput) (*kms.DecryptOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DecryptWithContext(aws.Context, *kms.DecryptInput, ...request.Option) (*kms.DecryptOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DecryptRequest(*kms.DecryptInput) (*request.Request, *kms.DecryptOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) DeleteAlias(*kms.DeleteAliasInput) (*kms.DeleteAliasOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DeleteAliasWithContext(aws.Context, *kms.DeleteAliasInput, ...request.Option) (*kms.DeleteAliasOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DeleteAliasRequest(*kms.DeleteAliasInput) (*request.Request, *kms.DeleteAliasOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DeleteImportedKeyMaterial(*kms.DeleteImportedKeyMaterialInput) (*kms.DeleteImportedKeyMaterialOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DeleteImportedKeyMaterialWithContext(aws.Context, *kms.DeleteImportedKeyMaterialInput, ...request.Option) (*kms.DeleteImportedKeyMaterialOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DeleteImportedKeyMaterialRequest(*kms.DeleteImportedKeyMaterialInput) (*request.Request, *kms.DeleteImportedKeyMaterialOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DescribeKeyWithContext(aws.Context, *kms.DescribeKeyInput, ...request.Option) (*kms.DescribeKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DescribeKeyRequest(*kms.DescribeKeyInput) (*request.Request, *kms.DescribeKeyOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DisableKeyWithContext(aws.Context, *kms.DisableKeyInput, ...request.Option) (*kms.DisableKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DisableKeyRequest(*kms.DisableKeyInput) (*request.Request, *kms.DisableKeyOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DisableKeyRotation(*kms.DisableKeyRotationInput) (*kms.DisableKeyRotationOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DisableKeyRotationWithContext(aws.Context, *kms.DisableKeyRotationInput, ...request.Option) (*kms.DisableKeyRotationOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) DisableKeyRotationRequest(*kms.DisableKeyRotationInput) (*request.Request, *kms.DisableKeyRotationOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EnableKey(*kms.EnableKeyInput) (*kms.EnableKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EnableKeyWithContext(aws.Context, *kms.EnableKeyInput, ...request.Option) (*kms.EnableKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EnableKeyRequest(*kms.EnableKeyInput) (*request.Request, *kms.EnableKeyOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EnableKeyRotation(*kms.EnableKeyRotationInput) (*kms.EnableKeyRotationOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EnableKeyRotationWithContext(aws.Context, *kms.EnableKeyRotationInput, ...request.Option) (*kms.EnableKeyRotationOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EnableKeyRotationRequest(*kms.EnableKeyRotationInput) (*request.Request, *kms.EnableKeyRotationOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EncryptWithContext(aws.Context, *kms.EncryptInput, ...request.Option) (*kms.EncryptOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) EncryptRequest(*kms.EncryptInput) (*request.Request, *kms.EncryptOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateDataKey(*kms.GenerateDataKeyInput) (*kms.GenerateDataKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateDataKeyWithContext(aws.Context, *kms.GenerateDataKeyInput, ...request.Option) (*kms.GenerateDataKeyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateDataKeyRequest(*kms.GenerateDataKeyInput) (*request.Request, *kms.GenerateDataKeyOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateDataKeyWithoutPlaintext(*kms.GenerateDataKeyWithoutPlaintextInput) (*kms.GenerateDataKeyWithoutPlaintextOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateDataKeyWithoutPlaintextWithContext(aws.Context, *kms.GenerateDataKeyWithoutPlaintextInput, ...request.Option) (*kms.GenerateDataKeyWithoutPlaintextOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateDataKeyWithoutPlaintextRequest(*kms.GenerateDataKeyWithoutPlaintextInput) (*request.Request, *kms.GenerateDataKeyWithoutPlaintextOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateRandom(*kms.GenerateRandomInput) (*kms.GenerateRandomOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateRandomWithContext(aws.Context, *kms.GenerateRandomInput, ...request.Option) (*kms.GenerateRandomOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GenerateRandomRequest(*kms.GenerateRandomInput) (*request.Request, *kms.GenerateRandomOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetKeyPolicy(*kms.GetKeyPolicyInput) (*kms.GetKeyPolicyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetKeyPolicyWithContext(aws.Context, *kms.GetKeyPolicyInput, ...request.Option) (*kms.GetKeyPolicyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetKeyPolicyRequest(*kms.GetKeyPolicyInput) (*request.Request, *kms.GetKeyPolicyOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetKeyRotationStatus(*kms.GetKeyRotationStatusInput) (*kms.GetKeyRotationStatusOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetKeyRotationStatusWithContext(aws.Context, *kms.GetKeyRotationStatusInput, ...request.Option) (*kms.GetKeyRotationStatusOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetKeyRotationStatusRequest(*kms.GetKeyRotationStatusInput) (*request.Request, *kms.GetKeyRotationStatusOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetParametersForImport(*kms.GetParametersForImportInput) (*kms.GetParametersForImportOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetParametersForImportWithContext(aws.Context, *kms.GetParametersForImportInput, ...request.Option) (*kms.GetParametersForImportOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) GetParametersForImportRequest(*kms.GetParametersForImportInput) (*request.Request, *kms.GetParametersForImportOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ImportKeyMaterial(*kms.ImportKeyMaterialInput) (*kms.ImportKeyMaterialOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ImportKeyMaterialWithContext(aws.Context, *kms.ImportKeyMaterialInput, ...request.Option) (*kms.ImportKeyMaterialOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ImportKeyMaterialRequest(*kms.ImportKeyMaterialInput) (*request.Request, *kms.ImportKeyMaterialOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListAliases(*kms.ListAliasesInput) (*kms.ListAliasesOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListAliasesWithContext(aws.Context, *kms.ListAliasesInput, ...request.Option) (*kms.ListAliasesOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListAliasesRequest(*kms.ListAliasesInput) (*request.Request, *kms.ListAliasesOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListAliasesPages(*kms.ListAliasesInput, func(*kms.ListAliasesOutput, bool) bool) error {
	panic("Not implemented")
	return nil
}
func (_ *MockKMS) ListAliasesPagesWithContext(aws.Context, *kms.ListAliasesInput, func(*kms.ListAliasesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
	return nil
}
func (_ *MockKMS) ListGrants(*kms.ListGrantsInput) (*kms.ListGrantsResponse, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListGrantsWithContext(aws.Context, *kms.ListGrantsInput, ...request.Option) (*kms.ListGrantsResponse, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListGrantsRequest(*kms.ListGrantsInput) (*request.Request, *kms.ListGrantsResponse) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListGrantsPages(*kms.ListGrantsInput, func(*kms.ListGrantsResponse, bool) bool) error {
	panic("Not implemented")
	return nil
}
func (_ *MockKMS) ListGrantsPagesWithContext(aws.Context, *kms.ListGrantsInput, func(*kms.ListGrantsResponse, bool) bool, ...request.Option) error {
	panic("Not implemented")
	return nil
}
func (_ *MockKMS) ListKeyPolicies(*kms.ListKeyPoliciesInput) (*kms.ListKeyPoliciesOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListKeyPoliciesWithContext(aws.Context, *kms.ListKeyPoliciesInput, ...request.Option) (*kms.ListKeyPoliciesOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListKeyPoliciesRequest(*kms.ListKeyPoliciesInput) (*request.Request, *kms.ListKeyPoliciesOutput) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListKeyPoliciesPages(*kms.ListKeyPoliciesInput, func(*kms.ListKeyPoliciesOutput, bool) bool) error {
	panic("Not implemented")
	return nil
}
func (_ *MockKMS) ListKeyPoliciesPagesWithContext(aws.Context, *kms.ListKeyPoliciesInput, func(*kms.ListKeyPoliciesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
	return nil
}

func (_ *MockKMS) ListKeys(*kms.ListKeysInput) (*kms.ListKeysOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListKeysWithContext(aws.Context, *kms.ListKeysInput, ...request.Option) (*kms.ListKeysOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListKeysRequest(*kms.ListKeysInput) (*request.Request, *kms.ListKeysOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) ListKeysPages(*kms.ListKeysInput, func(*kms.ListKeysOutput, bool) bool) error {
	panic("Not implemented")
	return nil
}
func (_ *MockKMS) ListKeysPagesWithContext(aws.Context, *kms.ListKeysInput, func(*kms.ListKeysOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
	return nil
}

func (_ *MockKMS) ListResourceTags(*kms.ListResourceTagsInput) (*kms.ListResourceTagsOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListResourceTagsWithContext(aws.Context, *kms.ListResourceTagsInput, ...request.Option) (*kms.ListResourceTagsOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListResourceTagsRequest(*kms.ListResourceTagsInput) (*request.Request, *kms.ListResourceTagsOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) ListRetirableGrants(*kms.ListRetirableGrantsInput) (*kms.ListGrantsResponse, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListRetirableGrantsWithContext(aws.Context, *kms.ListRetirableGrantsInput, ...request.Option) (*kms.ListGrantsResponse, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ListRetirableGrantsRequest(*kms.ListRetirableGrantsInput) (*request.Request, *kms.ListGrantsResponse) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) PutKeyPolicy(*kms.PutKeyPolicyInput) (*kms.PutKeyPolicyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) PutKeyPolicyWithContext(aws.Context, *kms.PutKeyPolicyInput, ...request.Option) (*kms.PutKeyPolicyOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) PutKeyPolicyRequest(*kms.PutKeyPolicyInput) (*request.Request, *kms.PutKeyPolicyOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) ReEncrypt(*kms.ReEncryptInput) (*kms.ReEncryptOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ReEncryptWithContext(aws.Context, *kms.ReEncryptInput, ...request.Option) (*kms.ReEncryptOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ReEncryptRequest(*kms.ReEncryptInput) (*request.Request, *kms.ReEncryptOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) RetireGrant(*kms.RetireGrantInput) (*kms.RetireGrantOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) RetireGrantWithContext(aws.Context, *kms.RetireGrantInput, ...request.Option) (*kms.RetireGrantOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) RetireGrantRequest(*kms.RetireGrantInput) (*request.Request, *kms.RetireGrantOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) RevokeGrant(*kms.RevokeGrantInput) (*kms.RevokeGrantOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) RevokeGrantWithContext(aws.Context, *kms.RevokeGrantInput, ...request.Option) (*kms.RevokeGrantOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) RevokeGrantRequest(*kms.RevokeGrantInput) (*request.Request, *kms.RevokeGrantOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) ScheduleKeyDeletion(*kms.ScheduleKeyDeletionInput) (*kms.ScheduleKeyDeletionOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ScheduleKeyDeletionWithContext(aws.Context, *kms.ScheduleKeyDeletionInput, ...request.Option) (*kms.ScheduleKeyDeletionOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) ScheduleKeyDeletionRequest(*kms.ScheduleKeyDeletionInput) (*request.Request, *kms.ScheduleKeyDeletionOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) TagResource(*kms.TagResourceInput) (*kms.TagResourceOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) TagResourceWithContext(aws.Context, *kms.TagResourceInput, ...request.Option) (*kms.TagResourceOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) TagResourceRequest(*kms.TagResourceInput) (*request.Request, *kms.TagResourceOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) UntagResource(*kms.UntagResourceInput) (*kms.UntagResourceOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) UntagResourceWithContext(aws.Context, *kms.UntagResourceInput, ...request.Option) (*kms.UntagResourceOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) UntagResourceRequest(*kms.UntagResourceInput) (*request.Request, *kms.UntagResourceOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) UpdateAlias(*kms.UpdateAliasInput) (*kms.UpdateAliasOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) UpdateAliasWithContext(aws.Context, *kms.UpdateAliasInput, ...request.Option) (*kms.UpdateAliasOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) UpdateAliasRequest(*kms.UpdateAliasInput) (*request.Request, *kms.UpdateAliasOutput) {
	panic("Not implemented")
	return nil, nil
}

func (_ *MockKMS) UpdateKeyDescription(*kms.UpdateKeyDescriptionInput) (*kms.UpdateKeyDescriptionOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) UpdateKeyDescriptionWithContext(aws.Context, *kms.UpdateKeyDescriptionInput, ...request.Option) (*kms.UpdateKeyDescriptionOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (_ *MockKMS) UpdateKeyDescriptionRequest(*kms.UpdateKeyDescriptionInput) (*request.Request, *kms.UpdateKeyDescriptionOutput) {
	panic("Not implemented")
	return nil, nil
}
