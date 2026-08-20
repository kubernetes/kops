// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package iam provides methods and message types of the iam v1alpha1 API.
package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/marshaler"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/parameter"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

type BearerType string

const (
	// Unknown bearer type.
	BearerTypeUnknownBearerType = BearerType("unknown_bearer_type")
	// User.
	BearerTypeUser = BearerType("user")
	// Application.
	BearerTypeApplication = BearerType("application")
)

func (enum BearerType) String() string {
	if enum == "" {
		// return default value if empty
		return string(BearerTypeUnknownBearerType)
	}
	return string(enum)
}

func (enum BearerType) Values() []BearerType {
	return []BearerType{
		"unknown_bearer_type",
		"user",
		"application",
	}
}

func (enum BearerType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *BearerType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = BearerType(BearerType(tmp).String())
	return nil
}

type GracePeriodType string

const (
	// Unknown grace period type.
	GracePeriodTypeUnknownGracePeriodType = GracePeriodType("unknown_grace_period_type")
	// Password should be updated.
	GracePeriodTypeUpdatePassword = GracePeriodType("update_password")
	// MFA should be configured.
	GracePeriodTypeSetMfa = GracePeriodType("set_mfa")
)

func (enum GracePeriodType) String() string {
	if enum == "" {
		// return default value if empty
		return string(GracePeriodTypeUnknownGracePeriodType)
	}
	return string(enum)
}

func (enum GracePeriodType) Values() []GracePeriodType {
	return []GracePeriodType{
		"unknown_grace_period_type",
		"update_password",
		"set_mfa",
	}
}

func (enum GracePeriodType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *GracePeriodType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = GracePeriodType(GracePeriodType(tmp).String())
	return nil
}

type ListAPIKeysRequestOrderBy string

const (
	// Creation date ascending.
	ListAPIKeysRequestOrderByCreatedAtAsc = ListAPIKeysRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListAPIKeysRequestOrderByCreatedAtDesc = ListAPIKeysRequestOrderBy("created_at_desc")
	// Update date ascending.
	ListAPIKeysRequestOrderByUpdatedAtAsc = ListAPIKeysRequestOrderBy("updated_at_asc")
	// Update date descending.
	ListAPIKeysRequestOrderByUpdatedAtDesc = ListAPIKeysRequestOrderBy("updated_at_desc")
	// Expiration date ascending.
	ListAPIKeysRequestOrderByExpiresAtAsc = ListAPIKeysRequestOrderBy("expires_at_asc")
	// Expiration date descending.
	ListAPIKeysRequestOrderByExpiresAtDesc = ListAPIKeysRequestOrderBy("expires_at_desc")
	// Access key ascending.
	ListAPIKeysRequestOrderByAccessKeyAsc = ListAPIKeysRequestOrderBy("access_key_asc")
	// Access key descending.
	ListAPIKeysRequestOrderByAccessKeyDesc = ListAPIKeysRequestOrderBy("access_key_desc")
)

func (enum ListAPIKeysRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListAPIKeysRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListAPIKeysRequestOrderBy) Values() []ListAPIKeysRequestOrderBy {
	return []ListAPIKeysRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
		"expires_at_asc",
		"expires_at_desc",
		"access_key_asc",
		"access_key_desc",
	}
}

func (enum ListAPIKeysRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListAPIKeysRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListAPIKeysRequestOrderBy(ListAPIKeysRequestOrderBy(tmp).String())
	return nil
}

type ListApplicationsRequestOrderBy string

const (
	// Creation date ascending.
	ListApplicationsRequestOrderByCreatedAtAsc = ListApplicationsRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListApplicationsRequestOrderByCreatedAtDesc = ListApplicationsRequestOrderBy("created_at_desc")
	// Update date ascending.
	ListApplicationsRequestOrderByUpdatedAtAsc = ListApplicationsRequestOrderBy("updated_at_asc")
	// Update date descending.
	ListApplicationsRequestOrderByUpdatedAtDesc = ListApplicationsRequestOrderBy("updated_at_desc")
	// Name ascending.
	ListApplicationsRequestOrderByNameAsc = ListApplicationsRequestOrderBy("name_asc")
	// Name descending.
	ListApplicationsRequestOrderByNameDesc = ListApplicationsRequestOrderBy("name_desc")
)

func (enum ListApplicationsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListApplicationsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListApplicationsRequestOrderBy) Values() []ListApplicationsRequestOrderBy {
	return []ListApplicationsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
		"name_asc",
		"name_desc",
	}
}

func (enum ListApplicationsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListApplicationsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListApplicationsRequestOrderBy(ListApplicationsRequestOrderBy(tmp).String())
	return nil
}

type ListGroupsRequestOrderBy string

const (
	// Creation date ascending.
	ListGroupsRequestOrderByCreatedAtAsc = ListGroupsRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListGroupsRequestOrderByCreatedAtDesc = ListGroupsRequestOrderBy("created_at_desc")
	// Update date ascending.
	ListGroupsRequestOrderByUpdatedAtAsc = ListGroupsRequestOrderBy("updated_at_asc")
	// Update date descending.
	ListGroupsRequestOrderByUpdatedAtDesc = ListGroupsRequestOrderBy("updated_at_desc")
	// Name ascending.
	ListGroupsRequestOrderByNameAsc = ListGroupsRequestOrderBy("name_asc")
	// Name descending.
	ListGroupsRequestOrderByNameDesc = ListGroupsRequestOrderBy("name_desc")
)

func (enum ListGroupsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListGroupsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListGroupsRequestOrderBy) Values() []ListGroupsRequestOrderBy {
	return []ListGroupsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
		"name_asc",
		"name_desc",
	}
}

func (enum ListGroupsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListGroupsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListGroupsRequestOrderBy(ListGroupsRequestOrderBy(tmp).String())
	return nil
}

type ListJWTsRequestOrderBy string

const (
	// Creation date ascending.
	ListJWTsRequestOrderByCreatedAtAsc = ListJWTsRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListJWTsRequestOrderByCreatedAtDesc = ListJWTsRequestOrderBy("created_at_desc")
	// Update date ascending.
	ListJWTsRequestOrderByUpdatedAtAsc = ListJWTsRequestOrderBy("updated_at_asc")
	// Update date descending.
	ListJWTsRequestOrderByUpdatedAtDesc = ListJWTsRequestOrderBy("updated_at_desc")
)

func (enum ListJWTsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListJWTsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListJWTsRequestOrderBy) Values() []ListJWTsRequestOrderBy {
	return []ListJWTsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
	}
}

func (enum ListJWTsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListJWTsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListJWTsRequestOrderBy(ListJWTsRequestOrderBy(tmp).String())
	return nil
}

type ListLogsRequestOrderBy string

const (
	// Creation date ascending.
	ListLogsRequestOrderByCreatedAtAsc = ListLogsRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListLogsRequestOrderByCreatedAtDesc = ListLogsRequestOrderBy("created_at_desc")
)

func (enum ListLogsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListLogsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListLogsRequestOrderBy) Values() []ListLogsRequestOrderBy {
	return []ListLogsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
	}
}

func (enum ListLogsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListLogsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListLogsRequestOrderBy(ListLogsRequestOrderBy(tmp).String())
	return nil
}

type ListPermissionSetsRequestOrderBy string

const (
	// Name ascending.
	ListPermissionSetsRequestOrderByNameAsc = ListPermissionSetsRequestOrderBy("name_asc")
	// Name descending.
	ListPermissionSetsRequestOrderByNameDesc = ListPermissionSetsRequestOrderBy("name_desc")
	// Creation date ascending.
	ListPermissionSetsRequestOrderByCreatedAtAsc = ListPermissionSetsRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListPermissionSetsRequestOrderByCreatedAtDesc = ListPermissionSetsRequestOrderBy("created_at_desc")
)

func (enum ListPermissionSetsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListPermissionSetsRequestOrderByNameAsc)
	}
	return string(enum)
}

func (enum ListPermissionSetsRequestOrderBy) Values() []ListPermissionSetsRequestOrderBy {
	return []ListPermissionSetsRequestOrderBy{
		"name_asc",
		"name_desc",
		"created_at_asc",
		"created_at_desc",
	}
}

func (enum ListPermissionSetsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPermissionSetsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPermissionSetsRequestOrderBy(ListPermissionSetsRequestOrderBy(tmp).String())
	return nil
}

type ListPoliciesRequestOrderBy string

const (
	// Policy name ascending.
	ListPoliciesRequestOrderByPolicyNameAsc = ListPoliciesRequestOrderBy("policy_name_asc")
	// Policy name descending.
	ListPoliciesRequestOrderByPolicyNameDesc = ListPoliciesRequestOrderBy("policy_name_desc")
	// Creation date ascending.
	ListPoliciesRequestOrderByCreatedAtAsc = ListPoliciesRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListPoliciesRequestOrderByCreatedAtDesc = ListPoliciesRequestOrderBy("created_at_desc")
)

func (enum ListPoliciesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListPoliciesRequestOrderByPolicyNameAsc)
	}
	return string(enum)
}

func (enum ListPoliciesRequestOrderBy) Values() []ListPoliciesRequestOrderBy {
	return []ListPoliciesRequestOrderBy{
		"policy_name_asc",
		"policy_name_desc",
		"created_at_asc",
		"created_at_desc",
	}
}

func (enum ListPoliciesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPoliciesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPoliciesRequestOrderBy(ListPoliciesRequestOrderBy(tmp).String())
	return nil
}

type ListQuotaRequestOrderBy string

const (
	// Name ascending.
	ListQuotaRequestOrderByNameAsc = ListQuotaRequestOrderBy("name_asc")
	// Name descending.
	ListQuotaRequestOrderByNameDesc = ListQuotaRequestOrderBy("name_desc")
)

func (enum ListQuotaRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListQuotaRequestOrderByNameAsc)
	}
	return string(enum)
}

func (enum ListQuotaRequestOrderBy) Values() []ListQuotaRequestOrderBy {
	return []ListQuotaRequestOrderBy{
		"name_asc",
		"name_desc",
	}
}

func (enum ListQuotaRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListQuotaRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListQuotaRequestOrderBy(ListQuotaRequestOrderBy(tmp).String())
	return nil
}

type ListSSHKeysRequestOrderBy string

const (
	// Creation date ascending.
	ListSSHKeysRequestOrderByCreatedAtAsc = ListSSHKeysRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListSSHKeysRequestOrderByCreatedAtDesc = ListSSHKeysRequestOrderBy("created_at_desc")
	// Update date ascending.
	ListSSHKeysRequestOrderByUpdatedAtAsc = ListSSHKeysRequestOrderBy("updated_at_asc")
	// Update date descending.
	ListSSHKeysRequestOrderByUpdatedAtDesc = ListSSHKeysRequestOrderBy("updated_at_desc")
	// Name ascending.
	ListSSHKeysRequestOrderByNameAsc = ListSSHKeysRequestOrderBy("name_asc")
	// Name descending.
	ListSSHKeysRequestOrderByNameDesc = ListSSHKeysRequestOrderBy("name_desc")
)

func (enum ListSSHKeysRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListSSHKeysRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListSSHKeysRequestOrderBy) Values() []ListSSHKeysRequestOrderBy {
	return []ListSSHKeysRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
		"name_asc",
		"name_desc",
	}
}

func (enum ListSSHKeysRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListSSHKeysRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListSSHKeysRequestOrderBy(ListSSHKeysRequestOrderBy(tmp).String())
	return nil
}

type ListScimTokensRequestOrderBy string

const (
	ListScimTokensRequestOrderByCreatedAtAsc  = ListScimTokensRequestOrderBy("created_at_asc")
	ListScimTokensRequestOrderByCreatedAtDesc = ListScimTokensRequestOrderBy("created_at_desc")
)

func (enum ListScimTokensRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListScimTokensRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListScimTokensRequestOrderBy) Values() []ListScimTokensRequestOrderBy {
	return []ListScimTokensRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
	}
}

func (enum ListScimTokensRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListScimTokensRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListScimTokensRequestOrderBy(ListScimTokensRequestOrderBy(tmp).String())
	return nil
}

type ListUsersRequestOrderBy string

const (
	// Creation date ascending.
	ListUsersRequestOrderByCreatedAtAsc = ListUsersRequestOrderBy("created_at_asc")
	// Creation date descending.
	ListUsersRequestOrderByCreatedAtDesc = ListUsersRequestOrderBy("created_at_desc")
	// Update date ascending.
	ListUsersRequestOrderByUpdatedAtAsc = ListUsersRequestOrderBy("updated_at_asc")
	// Update date descending.
	ListUsersRequestOrderByUpdatedAtDesc = ListUsersRequestOrderBy("updated_at_desc")
	// Email ascending.
	ListUsersRequestOrderByEmailAsc = ListUsersRequestOrderBy("email_asc")
	// Email descending.
	ListUsersRequestOrderByEmailDesc = ListUsersRequestOrderBy("email_desc")
	// Last login ascending.
	ListUsersRequestOrderByLastLoginAsc = ListUsersRequestOrderBy("last_login_asc")
	// Last login descending.
	ListUsersRequestOrderByLastLoginDesc = ListUsersRequestOrderBy("last_login_desc")
	// Username ascending.
	ListUsersRequestOrderByUsernameAsc = ListUsersRequestOrderBy("username_asc")
	// Username descending.
	ListUsersRequestOrderByUsernameDesc = ListUsersRequestOrderBy("username_desc")
)

func (enum ListUsersRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListUsersRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListUsersRequestOrderBy) Values() []ListUsersRequestOrderBy {
	return []ListUsersRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
		"email_asc",
		"email_desc",
		"last_login_asc",
		"last_login_desc",
		"username_asc",
		"username_desc",
	}
}

func (enum ListUsersRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListUsersRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListUsersRequestOrderBy(ListUsersRequestOrderBy(tmp).String())
	return nil
}

type LocalityType string

const (
	LocalityTypeGlobal = LocalityType("global")
	LocalityTypeRegion = LocalityType("region")
	LocalityTypeZone   = LocalityType("zone")
)

func (enum LocalityType) String() string {
	if enum == "" {
		// return default value if empty
		return string(LocalityTypeGlobal)
	}
	return string(enum)
}

func (enum LocalityType) Values() []LocalityType {
	return []LocalityType{
		"global",
		"region",
		"zone",
	}
}

func (enum LocalityType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LocalityType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LocalityType(LocalityType(tmp).String())
	return nil
}

type LogAction string

const (
	// Unknown action.
	LogActionUnknownAction = LogAction("unknown_action")
	// Created.
	LogActionCreated = LogAction("created")
	// Updated.
	LogActionUpdated = LogAction("updated")
	// Deleted.
	LogActionDeleted = LogAction("deleted")
)

func (enum LogAction) String() string {
	if enum == "" {
		// return default value if empty
		return string(LogActionUnknownAction)
	}
	return string(enum)
}

func (enum LogAction) Values() []LogAction {
	return []LogAction{
		"unknown_action",
		"created",
		"updated",
		"deleted",
	}
}

func (enum LogAction) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LogAction) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LogAction(LogAction(tmp).String())
	return nil
}

type LogResourceType string

const (
	// Unknown resource type.
	LogResourceTypeUnknownResourceType = LogResourceType("unknown_resource_type")
	// API Key.
	LogResourceTypeAPIKey = LogResourceType("api_key")
	// User.
	LogResourceTypeUser = LogResourceType("user")
	// Application.
	LogResourceTypeApplication = LogResourceType("application")
	// Group.
	LogResourceTypeGroup = LogResourceType("group")
	// Policy.
	LogResourceTypePolicy = LogResourceType("policy")
)

func (enum LogResourceType) String() string {
	if enum == "" {
		// return default value if empty
		return string(LogResourceTypeUnknownResourceType)
	}
	return string(enum)
}

func (enum LogResourceType) Values() []LogResourceType {
	return []LogResourceType{
		"unknown_resource_type",
		"api_key",
		"user",
		"application",
		"group",
		"policy",
	}
}

func (enum LogResourceType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LogResourceType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LogResourceType(LogResourceType(tmp).String())
	return nil
}

type PermissionSetScopeType string

const (
	// Unknown scope type.
	PermissionSetScopeTypeUnknownScopeType = PermissionSetScopeType("unknown_scope_type")
	// Projects.
	PermissionSetScopeTypeProjects = PermissionSetScopeType("projects")
	// Organization.
	PermissionSetScopeTypeOrganization = PermissionSetScopeType("organization")
	// Account root user.
	PermissionSetScopeTypeAccountRootUser = PermissionSetScopeType("account_root_user")
)

func (enum PermissionSetScopeType) String() string {
	if enum == "" {
		// return default value if empty
		return string(PermissionSetScopeTypeUnknownScopeType)
	}
	return string(enum)
}

func (enum PermissionSetScopeType) Values() []PermissionSetScopeType {
	return []PermissionSetScopeType{
		"unknown_scope_type",
		"projects",
		"organization",
		"account_root_user",
	}
}

func (enum PermissionSetScopeType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PermissionSetScopeType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PermissionSetScopeType(PermissionSetScopeType(tmp).String())
	return nil
}

type SamlCertificateOrigin string

const (
	// Unknown certificate origin.
	SamlCertificateOriginUnknownCertificateOrigin = SamlCertificateOrigin("unknown_certificate_origin")
	// Certificate from Scaleway.
	SamlCertificateOriginScaleway = SamlCertificateOrigin("scaleway")
	// Certificate from Identity Provider.
	SamlCertificateOriginIdentityProvider = SamlCertificateOrigin("identity_provider")
)

func (enum SamlCertificateOrigin) String() string {
	if enum == "" {
		// return default value if empty
		return string(SamlCertificateOriginUnknownCertificateOrigin)
	}
	return string(enum)
}

func (enum SamlCertificateOrigin) Values() []SamlCertificateOrigin {
	return []SamlCertificateOrigin{
		"unknown_certificate_origin",
		"scaleway",
		"identity_provider",
	}
}

func (enum SamlCertificateOrigin) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SamlCertificateOrigin) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SamlCertificateOrigin(SamlCertificateOrigin(tmp).String())
	return nil
}

type SamlCertificateType string

const (
	// Unknown certificate type.
	SamlCertificateTypeUnknownCertificateType = SamlCertificateType("unknown_certificate_type")
	// Signing certificate.
	SamlCertificateTypeSigning = SamlCertificateType("signing")
	// Encryption certificate.
	SamlCertificateTypeEncryption = SamlCertificateType("encryption")
)

func (enum SamlCertificateType) String() string {
	if enum == "" {
		// return default value if empty
		return string(SamlCertificateTypeUnknownCertificateType)
	}
	return string(enum)
}

func (enum SamlCertificateType) Values() []SamlCertificateType {
	return []SamlCertificateType{
		"unknown_certificate_type",
		"signing",
		"encryption",
	}
}

func (enum SamlCertificateType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SamlCertificateType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SamlCertificateType(SamlCertificateType(tmp).String())
	return nil
}

type SamlStatus string

const (
	SamlStatusUnknownSamlStatus      = SamlStatus("unknown_saml_status")
	SamlStatusValid                  = SamlStatus("valid")
	SamlStatusMissingCertificate     = SamlStatus("missing_certificate")
	SamlStatusMissingEntityID        = SamlStatus("missing_entity_id")
	SamlStatusMissingSingleSignOnURL = SamlStatus("missing_single_sign_on_url")
)

func (enum SamlStatus) String() string {
	if enum == "" {
		// return default value if empty
		return string(SamlStatusUnknownSamlStatus)
	}
	return string(enum)
}

func (enum SamlStatus) Values() []SamlStatus {
	return []SamlStatus{
		"unknown_saml_status",
		"valid",
		"missing_certificate",
		"missing_entity_id",
		"missing_single_sign_on_url",
	}
}

func (enum SamlStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SamlStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SamlStatus(SamlStatus(tmp).String())
	return nil
}

type UserStatus string

const (
	// Unknown status.
	UserStatusUnknownStatus = UserStatus("unknown_status")
	// Invitation pending.
	UserStatusInvitationPending = UserStatus("invitation_pending")
	// Activated.
	UserStatusActivated = UserStatus("activated")
)

func (enum UserStatus) String() string {
	if enum == "" {
		// return default value if empty
		return string(UserStatusUnknownStatus)
	}
	return string(enum)
}

func (enum UserStatus) Values() []UserStatus {
	return []UserStatus{
		"unknown_status",
		"invitation_pending",
		"activated",
	}
}

func (enum UserStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *UserStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = UserStatus(UserStatus(tmp).String())
	return nil
}

type UserType string

const (
	// Unknown type.
	UserTypeUnknownType = UserType("unknown_type")
	// Owner.
	UserTypeOwner  = UserType("owner")
	UserTypeMember = UserType("member")
)

func (enum UserType) String() string {
	if enum == "" {
		// return default value if empty
		return string(UserTypeUnknownType)
	}
	return string(enum)
}

func (enum UserType) Values() []UserType {
	return []UserType{
		"unknown_type",
		"owner",
		"member",
	}
}

func (enum UserType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *UserType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = UserType(UserType(tmp).String())
	return nil
}

// ConnectionConnectedOrganization: connection connected organization.
type ConnectionConnectedOrganization struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Locked bool `json:"locked"`
}

// ConnectionConnectedUser: connection connected user.
type ConnectionConnectedUser struct {
	ID string `json:"id"`

	Username string `json:"username"`

	// Type: default value: unknown_type
	Type UserType `json:"type"`
}

// QuotumLimit: quotum limit.
type QuotumLimit struct {
	// Global: whether or not the limit is applied globally.
	// Precisely one of Global, Region, Zone must be set.
	Global *bool `json:"global,omitempty"`

	// Region: the region on which the limit is applied.
	// Precisely one of Global, Region, Zone must be set.
	Region *scw.Region `json:"region,omitempty"`

	// Zone: the zone on which the limit is applied.
	// Precisely one of Global, Region, Zone must be set.
	Zone *scw.Zone `json:"zone,omitempty"`

	// Limit: maximum locality limit.
	// Precisely one of Limit, Unlimited must be set.
	Limit *uint64 `json:"limit,omitempty"`

	// Unlimited: whether or not the quota per locality is unlimited.
	// Precisely one of Limit, Unlimited must be set.
	Unlimited *bool `json:"unlimited,omitempty"`
}

// JWT: jwt.
type JWT struct {
	// Jti: jWT ID.
	Jti string `json:"jti"`

	// IssuerID: ID of the user who issued the JWT.
	IssuerID string `json:"issuer_id"`

	// AudienceID: ID of the user targeted by the JWT.
	AudienceID string `json:"audience_id"`

	// CreatedAt: creation date of the JWT.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: last update date of the JWT.
	UpdatedAt *time.Time `json:"updated_at"`

	// ExpiresAt: expiration date of the JWT.
	ExpiresAt *time.Time `json:"expires_at"`

	// IP: IP address used during the creation of the JWT.
	IP net.IP `json:"ip"`

	// UserAgent: user-agent used during the creation of the JWT.
	UserAgent string `json:"user_agent"`
}

// RuleSpecs: rule specs.
type RuleSpecs struct {
	// PermissionSetNames: names of permission sets bound to the rule.
	PermissionSetNames *[]string `json:"permission_set_names"`

	// Condition: condition expression to evaluate.
	Condition string `json:"condition"`

	// ProjectIDs: list of Project IDs the rule is scoped to.
	// Precisely one of ProjectIDs, OrganizationID must be set.
	ProjectIDs *[]string `json:"project_ids,omitempty"`

	// OrganizationID: ID of Organization the rule is scoped to.
	// Precisely one of ProjectIDs, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
}

// ScimToken: scim token.
type ScimToken struct {
	ID string `json:"id"`

	ScimID string `json:"scim_id"`

	CreatedAt *time.Time `json:"created_at"`

	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateUserRequestMember: create user request member.
type CreateUserRequestMember struct {
	// Email: email of the user to create.
	Email string `json:"email"`

	// SendPasswordEmail: whether or not to send an email containing the member's password.
	SendPasswordEmail bool `json:"send_password_email"`

	// SendWelcomeEmail: whether or not to send a welcome email that includes onboarding information.
	SendWelcomeEmail bool `json:"send_welcome_email"`

	// Username: the member's username.
	Username string `json:"username"`

	// Password: the member's password.
	Password string `json:"password"`

	// FirstName: the member's first name.
	FirstName string `json:"first_name"`

	// LastName: the member's last name.
	LastName string `json:"last_name"`

	// PhoneNumber: the member's phone number.
	PhoneNumber string `json:"phone_number"`

	// Locale: the member's locale.
	Locale string `json:"locale"`
}

// Connection: connection.
type Connection struct {
	// Organization: information about the connected organization.
	Organization *ConnectionConnectedOrganization `json:"organization"`

	// User: information about the connected user.
	User *ConnectionConnectedUser `json:"user"`
}

// APIKey: api key.
type APIKey struct {
	// AccessKey: access key of the API key.
	AccessKey string `json:"access_key"`

	// SecretKey: secret key of the API Key.
	SecretKey *string `json:"secret_key"`

	// ApplicationID: ID of application that bears the API key.
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`

	// UserID: ID of user that bears the API key.
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`

	// Description: description of API key.
	Description string `json:"description"`

	// CreatedAt: date and time of API key creation.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: date and time of last API key update.
	UpdatedAt *time.Time `json:"updated_at"`

	// ExpiresAt: date and time of API key expiration.
	ExpiresAt *time.Time `json:"expires_at"`

	// DefaultProjectID: default Project ID specified for this API key.
	DefaultProjectID string `json:"default_project_id"`

	// Editable: defines whether or not the API key is editable.
	Editable bool `json:"editable"`

	// Deletable: defines whether or not the API key is deletable.
	Deletable bool `json:"deletable"`

	// Managed: defines whether or not the API key is managed.
	Managed bool `json:"managed"`

	// CreationIP: IP address of the device that created the API key.
	CreationIP string `json:"creation_ip"`
}

// Application: application.
type Application struct {
	// ID: ID of the application.
	ID string `json:"id"`

	// Name: name of the application.
	Name string `json:"name"`

	// Description: description of the application.
	Description string `json:"description"`

	// CreatedAt: date and time application was created.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: date and time of last application update.
	UpdatedAt *time.Time `json:"updated_at"`

	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`

	// Editable: defines whether or not the application is editable.
	Editable bool `json:"editable"`

	// Deletable: defines whether or not the application is deletable.
	Deletable bool `json:"deletable"`

	// Managed: defines whether or not the application is managed.
	Managed bool `json:"managed"`

	// NbAPIKeys: number of API keys attributed to the application.
	NbAPIKeys uint32 `json:"nb_api_keys"`

	// Tags: tags associated with the user.
	Tags []string `json:"tags"`
}

// GracePeriod: grace period.
type GracePeriod struct {
	// Type: type of grace period.
	// Default value: unknown_grace_period_type
	Type GracePeriodType `json:"type"`

	// CreatedAt: date and time the grace period was created.
	CreatedAt *time.Time `json:"created_at"`

	// ExpiresAt: date and time the grace period expires.
	ExpiresAt *time.Time `json:"expires_at"`
}

// Group: group.
type Group struct {
	// ID: ID of the group.
	ID string `json:"id"`

	// CreatedAt: date and time of group creation.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: date and time of last group update.
	UpdatedAt *time.Time `json:"updated_at"`

	// OrganizationID: ID of Organization linked to the group.
	OrganizationID string `json:"organization_id"`

	// Name: name of the group.
	Name string `json:"name"`

	// Description: description of the group.
	Description string `json:"description"`

	// UserIDs: iDs of users attached to this group.
	UserIDs []string `json:"user_ids"`

	// ApplicationIDs: iDs of applications attached to this group.
	ApplicationIDs []string `json:"application_ids"`

	// Tags: tags associated to the group.
	Tags []string `json:"tags"`

	// Editable: defines whether or not the group is editable.
	Editable bool `json:"editable"`

	// Deletable: defines whether or not the group is deletable.
	Deletable bool `json:"deletable"`

	// Managed: defines whether or not the group is managed.
	Managed bool `json:"managed"`
}

// Log: log.
type Log struct {
	// ID: log ID.
	ID string `json:"id"`

	// CreatedAt: creation date of the log.
	CreatedAt *time.Time `json:"created_at"`

	// IP: IP address of the HTTP request linked to the log.
	IP net.IP `json:"ip"`

	// UserAgent: user-Agent of the HTTP request linked to the log.
	UserAgent string `json:"user_agent"`

	// Action: action linked to the log.
	// Default value: unknown_action
	Action LogAction `json:"action"`

	// BearerID: ID of the principal at the origin of the log.
	BearerID string `json:"bearer_id"`

	// OrganizationID: ID of Organization linked to the log.
	OrganizationID string `json:"organization_id"`

	// ResourceType: type of the resource linked to the log.
	// Default value: unknown_resource_type
	ResourceType LogResourceType `json:"resource_type"`

	// ResourceID: ID of the resource linked  to the log.
	ResourceID string `json:"resource_id"`
}

// PermissionSet: permission set.
type PermissionSet struct {
	// ID: id of the permission set.
	ID string `json:"id"`

	// Name: name of the permission set.
	Name string `json:"name"`

	// ScopeType: scope of the permission set.
	// Default value: unknown_scope_type
	ScopeType PermissionSetScopeType `json:"scope_type"`

	// Description: description of the permission set.
	Description string `json:"description"`

	// Categories: categories of the permission set.
	Categories *[]string `json:"categories"`
}

// Policy: policy.
type Policy struct {
	// ID: id of the policy.
	ID string `json:"id"`

	// Name: name of the policy.
	Name string `json:"name"`

	// Description: description of the policy.
	Description string `json:"description"`

	// OrganizationID: organization ID of the policy.
	OrganizationID string `json:"organization_id"`

	// CreatedAt: date and time of policy creation.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: date and time of last policy update.
	UpdatedAt *time.Time `json:"updated_at"`

	// Editable: defines whether or not a policy is editable.
	Editable bool `json:"editable"`

	// Deletable: defines whether or not a policy is deletable.
	Deletable bool `json:"deletable"`

	// Managed: defines whether or not a policy is managed.
	Managed bool `json:"managed"`

	// NbRules: number of rules of the policy.
	NbRules uint32 `json:"nb_rules"`

	// NbScopes: number of policy scopes.
	NbScopes uint32 `json:"nb_scopes"`

	// NbPermissionSets: number of permission sets of the policy.
	NbPermissionSets uint32 `json:"nb_permission_sets"`

	// Tags: tags associated with the policy.
	Tags []string `json:"tags"`

	// UserID: ID of the user attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	UserID *string `json:"user_id,omitempty"`

	// GroupID: ID of the group attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	GroupID *string `json:"group_id,omitempty"`

	// ApplicationID: ID of the application attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	ApplicationID *string `json:"application_id,omitempty"`

	// NoPrincipal: defines whether or not a policy is attributed to a principal.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// Quotum: quotum.
type Quotum struct {
	// Name: name of the quota.
	Name string `json:"name"`

	// Deprecated: Limit: maximum limit of the quota.
	// Precisely one of Limit, Unlimited must be set.
	Limit *uint64 `json:"limit,omitempty"`

	// Deprecated: Unlimited: defines whether or not the quota is unlimited.
	// Precisely one of Limit, Unlimited must be set.
	Unlimited *bool `json:"unlimited,omitempty"`

	// PrettyName: a human-readable name for the quota.
	PrettyName string `json:"pretty_name"`

	// Unit: the unit in which the quota is expressed.
	Unit string `json:"unit"`

	// Description: details about the quota.
	Description string `json:"description"`

	// LocalityType: whether this quotum is applied on at the zone level, region level, or globally.
	// Default value: global
	LocalityType LocalityType `json:"locality_type"`

	// Limits: limits per locality.
	Limits []*QuotumLimit `json:"limits"`
}

// Rule: rule.
type Rule struct {
	// ID: id of rule.
	ID string `json:"id"`

	// PermissionSetNames: names of permission sets bound to the rule.
	PermissionSetNames *[]string `json:"permission_set_names"`

	// PermissionSetsScopeType: permission_set_names have the same scope_type.
	// Default value: unknown_scope_type
	PermissionSetsScopeType PermissionSetScopeType `json:"permission_sets_scope_type"`

	// Condition: condition expression to evaluate.
	Condition string `json:"condition"`

	// ProjectIDs: list of Project IDs the rule is scoped to.
	// Precisely one of ProjectIDs, OrganizationID, AccountRootUserID must be set.
	ProjectIDs *[]string `json:"project_ids,omitempty"`

	// OrganizationID: ID of Organization the rule is scoped to.
	// Precisely one of ProjectIDs, OrganizationID, AccountRootUserID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// AccountRootUserID: ID of account root user the rule is scoped to.
	// Precisely one of ProjectIDs, OrganizationID, AccountRootUserID must be set.
	AccountRootUserID *string `json:"account_root_user_id,omitempty"`
}

// SSHKey: ssh key.
type SSHKey struct {
	// ID: ID of SSH key.
	ID string `json:"id"`

	// Name: name of SSH key.
	Name string `json:"name"`

	// PublicKey: public key of SSH key.
	PublicKey string `json:"public_key"`

	// Fingerprint: fingerprint of the SSH key.
	Fingerprint string `json:"fingerprint"`

	// CreatedAt: creation date of SSH key.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: last update date of SSH key.
	UpdatedAt *time.Time `json:"updated_at"`

	// OrganizationID: ID of Organization linked to the SSH key.
	OrganizationID string `json:"organization_id"`

	// ProjectID: ID of Project linked to the SSH key.
	ProjectID string `json:"project_id"`

	// Disabled: SSH key status.
	Disabled bool `json:"disabled"`
}

// SamlCertificate: saml certificate.
type SamlCertificate struct {
	// ID: ID of the SAML certificate.
	ID string `json:"id"`

	// Type: type of the SAML certificate.
	// Default value: unknown_certificate_type
	Type SamlCertificateType `json:"type"`

	// Origin: origin of the SAML certificate.
	// Default value: unknown_certificate_origin
	Origin SamlCertificateOrigin `json:"origin"`

	// Content: content of the SAML certificate.
	Content string `json:"content"`

	// ExpiresAt: date and time of the SAML certificate expiration.
	ExpiresAt *time.Time `json:"expires_at"`
}

// User: user.
type User struct {
	// ID: ID of user.
	ID string `json:"id"`

	// Email: email of user.
	Email string `json:"email"`

	// Username: user identifier unique to the Organization.
	Username string `json:"username"`

	// FirstName: first name of the user.
	FirstName string `json:"first_name"`

	// LastName: last name of the user.
	LastName string `json:"last_name"`

	// PhoneNumber: phone number of the user.
	PhoneNumber string `json:"phone_number"`

	// Locale: locale of the user.
	Locale string `json:"locale"`

	// CreatedAt: date user was created.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: date of last user update.
	UpdatedAt *time.Time `json:"updated_at"`

	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`

	// Deletable: deletion status of user. Owners cannot be deleted.
	Deletable bool `json:"deletable"`

	// LastLoginAt: date of the last login.
	LastLoginAt *time.Time `json:"last_login_at"`

	// Type: type of user.
	// Default value: unknown_type
	Type UserType `json:"type"`

	// Deprecated: TwoFactorEnabled: deprecated, use "mfa" instead.
	TwoFactorEnabled *bool `json:"two_factor_enabled,omitempty"`

	// Deprecated: Status: status of user invitation.
	// Default value: unknown_status
	Status *UserStatus `json:"status,omitempty"`

	// Mfa: defines whether MFA is enabled.
	Mfa bool `json:"mfa"`

	// AccountRootUserID: ID of the account root user associated with the user.
	AccountRootUserID string `json:"account_root_user_id"`

	// Tags: tags associated with the user.
	Tags []string `json:"tags"`

	// Locked: defines whether the user is locked.
	Locked bool `json:"locked"`
}

// SamlServiceProvider: saml service provider.
type SamlServiceProvider struct {
	EntityID string `json:"entity_id"`

	AssertionConsumerServiceURL string `json:"assertion_consumer_service_url"`
}

// AddGroupMemberRequest: add group member request.
type AddGroupMemberRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`

	// UserID: ID of the user to add.
	// Precisely one of UserID, ApplicationID must be set.
	UserID *string `json:"user_id,omitempty"`

	// ApplicationID: ID of the application to add.
	// Precisely one of UserID, ApplicationID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
}

// AddGroupMembersRequest: add group members request.
type AddGroupMembersRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`

	// UserIDs: iDs of the users to add.
	UserIDs []string `json:"user_ids"`

	// ApplicationIDs: iDs of the applications to add.
	ApplicationIDs []string `json:"application_ids"`
}

// AddSamlCertificateRequest: add saml certificate request.
type AddSamlCertificateRequest struct {
	// SamlID: ID of the SAML configuration.
	SamlID string `json:"-"`

	// Type: type of the SAML certificate.
	// Default value: unknown_certificate_type
	Type SamlCertificateType `json:"type"`

	// Content: content of the SAML certificate.
	Content string `json:"content"`
}

// ClonePolicyRequest: clone policy request.
type ClonePolicyRequest struct {
	PolicyID string `json:"-"`
}

// CreateAPIKeyRequest: create api key request.
type CreateAPIKeyRequest struct {
	// ApplicationID: ID of the application.
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`

	// UserID: ID of the user.
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`

	// ExpiresAt: expiration date of the API key.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// DefaultProjectID: default Project ID to use with Object Storage.
	DefaultProjectID *string `json:"default_project_id,omitempty"`

	// Description: description of the API key (max length is 200 characters).
	Description string `json:"description"`
}

// CreateApplicationRequest: create application request.
type CreateApplicationRequest struct {
	// Name: name of the application to create (max length is 64 characters).
	Name string `json:"name"`

	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`

	// Description: description of the application (max length is 200 characters).
	Description string `json:"description"`

	// Tags: tags associated with the application (maximum of 10 tags).
	Tags []string `json:"tags"`
}

// CreateGroupRequest: create group request.
type CreateGroupRequest struct {
	// OrganizationID: ID of Organization linked to the group.
	OrganizationID string `json:"organization_id"`

	// Name: name of the group to create (max length is 64 chars). MUST be unique inside an Organization.
	Name string `json:"name"`

	// Description: description of the group to create (max length is 200 chars).
	Description string `json:"description"`

	// Tags: tags associated with the group (maximum of 10 tags).
	Tags []string `json:"tags"`
}

// CreateJWTRequest: create jwt request.
type CreateJWTRequest struct {
	// UserID: ID of the user the JWT will be created for.
	UserID string `json:"user_id"`

	// Referrer: referrer of the JWT.
	Referrer string `json:"referrer"`
}

// CreatePolicyRequest: create policy request.
type CreatePolicyRequest struct {
	// Name: name of the policy to create (max length is 64 characters).
	Name string `json:"name"`

	// Description: description of the policy to create (max length is 200 characters).
	Description string `json:"description"`

	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`

	// Rules: rules of the policy to create.
	Rules []*RuleSpecs `json:"rules"`

	// Tags: tags associated with the policy (maximum of 10 tags).
	Tags []string `json:"tags"`

	// UserID: ID of user attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	UserID *string `json:"user_id,omitempty"`

	// GroupID: ID of group attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	GroupID *string `json:"group_id,omitempty"`

	// ApplicationID: ID of application attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	ApplicationID *string `json:"application_id,omitempty"`

	// NoPrincipal: defines whether or not a policy is attributed to a principal.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// CreateSSHKeyRequest: create ssh key request.
type CreateSSHKeyRequest struct {
	// Name: name of the SSH key. Max length is 1000.
	Name string `json:"name"`

	// PublicKey: SSH public key. Currently only the ssh-rsa, ssh-dss (DSA), ssh-ed25519 and ecdsa keys with NIST curves are supported. Max length is 65000.
	PublicKey string `json:"public_key"`

	// ProjectID: project the resource is attributed to.
	ProjectID string `json:"project_id"`
}

// CreateScimTokenRequest: create scim token request.
type CreateScimTokenRequest struct {
	// ScimID: ID of the SCIM configuration.
	ScimID string `json:"-"`
}

// CreateScimTokenResponse: create scim token response.
type CreateScimTokenResponse struct {
	// Token: the SCIM token metadata.
	Token *ScimToken `json:"token"`

	// BearerToken: the Bearer Token to use to authenticate to SCIM endpoints.
	BearerToken string `json:"bearer_token"`
}

// CreateUserMFAOTPRequest: create user mfaotp request.
type CreateUserMFAOTPRequest struct {
	// UserID: user ID of the MFA OTP.
	UserID string `json:"-"`
}

// CreateUserRequest: create user request.
type CreateUserRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`

	// Email: email of the user.
	// Precisely one of Email, Member must be set.
	Email *string `json:"email,omitempty"`

	// Tags: tags associated with the user.
	Tags []string `json:"tags"`

	// Member: details of IAM member.
	// Precisely one of Email, Member must be set.
	Member *CreateUserRequestMember `json:"member,omitempty"`
}

// DeleteAPIKeyRequest: delete api key request.
type DeleteAPIKeyRequest struct {
	// AccessKey: access key to delete.
	AccessKey string `json:"-"`
}

// DeleteApplicationRequest: delete application request.
type DeleteApplicationRequest struct {
	// ApplicationID: ID of the application to delete.
	ApplicationID string `json:"-"`
}

// DeleteGroupRequest: delete group request.
type DeleteGroupRequest struct {
	// GroupID: ID of the group to delete.
	GroupID string `json:"-"`
}

// DeleteJWTRequest: delete jwt request.
type DeleteJWTRequest struct {
	// Jti: jWT ID of the JWT to delete.
	Jti string `json:"-"`
}

// DeletePolicyRequest: delete policy request.
type DeletePolicyRequest struct {
	// PolicyID: id of policy to delete.
	PolicyID string `json:"-"`
}

// DeleteSSHKeyRequest: delete ssh key request.
type DeleteSSHKeyRequest struct {
	SSHKeyID string `json:"-"`
}

// DeleteSamlCertificateRequest: delete saml certificate request.
type DeleteSamlCertificateRequest struct {
	// CertificateID: ID of the certificate to delete.
	CertificateID string `json:"-"`
}

// DeleteSamlRequest: delete saml request.
type DeleteSamlRequest struct {
	// SamlID: ID of the SAML configuration.
	SamlID string `json:"-"`
}

// DeleteScimRequest: delete scim request.
type DeleteScimRequest struct {
	// ScimID: ID of the SCIM configuration.
	ScimID string `json:"-"`
}

// DeleteScimTokenRequest: delete scim token request.
type DeleteScimTokenRequest struct {
	// TokenID: the SCIM token ID.
	TokenID string `json:"-"`
}

// DeleteUserMFAOTPRequest: delete user mfaotp request.
type DeleteUserMFAOTPRequest struct {
	// UserID: user ID of the MFA OTP.
	UserID string `json:"-"`
}

// DeleteUserRequest: delete user request.
type DeleteUserRequest struct {
	// UserID: ID of the user to delete.
	UserID string `json:"-"`
}

// EnableOrganizationSamlRequest: enable organization saml request.
type EnableOrganizationSamlRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// EnableOrganizationScimRequest: enable organization scim request.
type EnableOrganizationScimRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// EncodedJWT: encoded jwt.
type EncodedJWT struct {
	// Jwt: the renewed JWT.
	Jwt *JWT `json:"jwt"`

	// Token: the encoded token of the renewed JWT.
	Token string `json:"token"`

	// RenewToken: the encoded renew token. This token is necessary to renew the JWT.
	RenewToken string `json:"renew_token"`
}

// GetAPIKeyRequest: get api key request.
type GetAPIKeyRequest struct {
	// AccessKey: access key to search for.
	AccessKey string `json:"-"`
}

// GetApplicationRequest: get application request.
type GetApplicationRequest struct {
	// ApplicationID: ID of the application to find.
	ApplicationID string `json:"-"`
}

// GetGroupRequest: get group request.
type GetGroupRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`
}

// GetJWTRequest: get jwt request.
type GetJWTRequest struct {
	// Jti: jWT ID of the JWT to get.
	Jti string `json:"-"`
}

// GetLogRequest: get log request.
type GetLogRequest struct {
	// LogID: ID of the log.
	LogID string `json:"-"`
}

// GetOrganizationRequest: get organization request.
type GetOrganizationRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// GetOrganizationSamlRequest: get organization saml request.
type GetOrganizationSamlRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// GetOrganizationScimRequest: get organization scim request.
type GetOrganizationScimRequest struct {
	OrganizationID string `json:"-"`
}

// GetOrganizationSecuritySettingsRequest: get organization security settings request.
type GetOrganizationSecuritySettingsRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// GetPolicyRequest: get policy request.
type GetPolicyRequest struct {
	// PolicyID: id of policy to search.
	PolicyID string `json:"-"`
}

// GetQuotumRequest: get quotum request.
type GetQuotumRequest struct {
	// QuotumName: name of the quota to get.
	QuotumName string `json:"-"`

	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// GetSSHKeyRequest: get ssh key request.
type GetSSHKeyRequest struct {
	// SSHKeyID: ID of the SSH key.
	SSHKeyID string `json:"-"`
}

// GetUserConnectionsRequest: get user connections request.
type GetUserConnectionsRequest struct {
	// UserID: ID of the user to list connections for.
	UserID string `json:"-"`
}

// GetUserConnectionsResponse: get user connections response.
type GetUserConnectionsResponse struct {
	// Connections: list of connections.
	Connections []*Connection `json:"connections"`
}

// GetUserRequest: get user request.
type GetUserRequest struct {
	// UserID: ID of the user to find.
	UserID string `json:"-"`
}

// InitiateUserConnectionRequest: initiate user connection request.
type InitiateUserConnectionRequest struct {
	// UserID: ID of the user that will be added to your connection.
	UserID string `json:"-"`
}

// InitiateUserConnectionResponse: initiate user connection response.
type InitiateUserConnectionResponse struct {
	// Token: token to be used in JoinUserConnection.
	Token string `json:"token"`
}

// JoinUserConnectionRequest: join user connection request.
type JoinUserConnectionRequest struct {
	// UserID: user ID.
	UserID string `json:"-"`

	// Token: a token returned by InitiateUserConnection.
	Token string `json:"token"`
}

// ListAPIKeysRequest: list api keys request.
type ListAPIKeysRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListAPIKeysRequestOrderBy `json:"-"`

	// Page: page number. Value must be greater or equal to 1.
	Page *int32 `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// OrganizationID: ID of Organization.
	OrganizationID *string `json:"-"`

	// Deprecated: ApplicationID: ID of application that bears the API key.
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`

	// Deprecated: UserID: ID of user that bears the API key.
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`

	// Editable: defines whether to filter out editable API keys or not.
	Editable *bool `json:"-"`

	// Expired: defines whether to filter out expired API keys or not.
	Expired *bool `json:"-"`

	// Deprecated: AccessKey: filter by access key (deprecated in favor of `access_keys`).
	AccessKey *string `json:"-"`

	// Description: filter by description.
	Description *string `json:"-"`

	// BearerID: filter by bearer ID.
	BearerID *string `json:"-"`

	// BearerType: filter by type of bearer.
	// Default value: unknown_bearer_type
	BearerType BearerType `json:"-"`

	// AccessKeys: filter by a list of access keys.
	AccessKeys []string `json:"-"`
}

// ListAPIKeysResponse: list api keys response.
type ListAPIKeysResponse struct {
	// APIKeys: list of API keys.
	APIKeys []*APIKey `json:"api_keys"`

	// TotalCount: total count of API Keys.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListAPIKeysResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListAPIKeysResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListAPIKeysResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.APIKeys = append(r.APIKeys, results.APIKeys...)
	r.TotalCount += uint32(len(results.APIKeys))
	return uint32(len(results.APIKeys)), nil
}

// ListApplicationsRequest: list applications request.
type ListApplicationsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListApplicationsRequestOrderBy `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater than 1.
	Page *int32 `json:"-"`

	// Name: name of the application to filter.
	Name *string `json:"-"`

	// OrganizationID: ID of the Organization to filter.
	OrganizationID string `json:"-"`

	// Editable: defines whether to filter out editable applications or not.
	Editable *bool `json:"-"`

	// ApplicationIDs: filter by list of IDs.
	ApplicationIDs []string `json:"-"`

	// Tag: filter by tags containing a given string.
	Tag *string `json:"-"`
}

// ListApplicationsResponse: list applications response.
type ListApplicationsResponse struct {
	// Applications: list of applications.
	Applications []*Application `json:"applications"`

	// TotalCount: total count of applications.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListApplicationsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListApplicationsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListApplicationsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Applications = append(r.Applications, results.Applications...)
	r.TotalCount += uint32(len(results.Applications))
	return uint32(len(results.Applications)), nil
}

// ListGracePeriodsRequest: list grace periods request.
type ListGracePeriodsRequest struct {
	// UserID: ID of the user to list grace periods for.
	UserID *string `json:"-"`
}

// ListGracePeriodsResponse: list grace periods response.
type ListGracePeriodsResponse struct {
	// GracePeriods: list of grace periods.
	GracePeriods []*GracePeriod `json:"grace_periods"`
}

// ListGroupsRequest: list groups request.
type ListGroupsRequest struct {
	// OrderBy: sort order of groups.
	// Default value: created_at_asc
	OrderBy ListGroupsRequestOrderBy `json:"-"`

	// Page: requested page number. Value must be greater or equal to 1.
	Page *int32 `json:"-"`

	// PageSize: number of items per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID string `json:"-"`

	// Name: name of group to find.
	Name *string `json:"-"`

	// ApplicationIDs: filter by a list of application IDs.
	ApplicationIDs []string `json:"-"`

	// UserIDs: filter by a list of user IDs.
	UserIDs []string `json:"-"`

	// GroupIDs: filter by a list of group IDs.
	GroupIDs []string `json:"-"`

	// Tag: filter by tags containing a given string.
	Tag *string `json:"-"`
}

// ListGroupsResponse: list groups response.
type ListGroupsResponse struct {
	// Groups: list of groups.
	Groups []*Group `json:"groups"`

	// TotalCount: total count of groups.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListGroupsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListGroupsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListGroupsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Groups = append(r.Groups, results.Groups...)
	r.TotalCount += uint32(len(results.Groups))
	return uint32(len(results.Groups)), nil
}

// ListJWTsRequest: list jw ts request.
type ListJWTsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListJWTsRequestOrderBy `json:"-"`

	// AudienceID: ID of the user to search.
	AudienceID string `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater to 1.
	Page *int32 `json:"-"`

	// Expired: filter out expired JWTs or not.
	Expired *bool `json:"-"`
}

// ListJWTsResponse: list jw ts response.
type ListJWTsResponse struct {
	Jwts []*JWT `json:"jwts"`

	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListJWTsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListJWTsResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListJWTsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Jwts = append(r.Jwts, results.Jwts...)
	r.TotalCount += uint64(len(results.Jwts))
	return uint64(len(results.Jwts)), nil
}

// ListLogsRequest: list logs request.
type ListLogsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListLogsRequestOrderBy `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID string `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater to 1.
	Page *int32 `json:"-"`

	// CreatedAfter: defined whether or not to filter out logs created after this timestamp.
	CreatedAfter *time.Time `json:"-"`

	// CreatedBefore: defined whether or not to filter out logs created before this timestamp.
	CreatedBefore *time.Time `json:"-"`

	// Action: defined whether or not to filter out by a specific action.
	// Default value: unknown_action
	Action LogAction `json:"-"`

	// ResourceType: defined whether or not to filter out by a specific type of resource.
	// Default value: unknown_resource_type
	ResourceType LogResourceType `json:"-"`

	// Search: defined whether or not to filter out log by bearer ID or resource ID.
	Search *string `json:"-"`
}

// ListLogsResponse: list logs response.
type ListLogsResponse struct {
	// Logs: list of logs.
	Logs []*Log `json:"logs"`

	// TotalCount: total count of logs.
	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLogsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLogsResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListLogsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Logs = append(r.Logs, results.Logs...)
	r.TotalCount += uint64(len(results.Logs))
	return uint64(len(results.Logs)), nil
}

// ListPermissionSetsRequest: list permission sets request.
type ListPermissionSetsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: name_asc
	OrderBy ListPermissionSetsRequestOrderBy `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater than 1.
	Page *int32 `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID string `json:"-"`
}

// ListPermissionSetsResponse: list permission sets response.
type ListPermissionSetsResponse struct {
	// PermissionSets: list of permission sets.
	PermissionSets []*PermissionSet `json:"permission_sets"`

	// TotalCount: total count of permission sets.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPermissionSetsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPermissionSetsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListPermissionSetsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PermissionSets = append(r.PermissionSets, results.PermissionSets...)
	r.TotalCount += uint32(len(results.PermissionSets))
	return uint32(len(results.PermissionSets)), nil
}

// ListPoliciesRequest: list policies request.
type ListPoliciesRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: policy_name_asc
	OrderBy ListPoliciesRequestOrderBy `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater than 1.
	Page *int32 `json:"-"`

	// OrganizationID: ID of the Organization to filter.
	OrganizationID string `json:"-"`

	// Editable: defines whether or not filter out editable policies.
	Editable *bool `json:"-"`

	// UserIDs: defines whether or not to filter by list of user IDs.
	UserIDs []string `json:"-"`

	// GroupIDs: defines whether or not to filter by list of group IDs.
	GroupIDs []string `json:"-"`

	// ApplicationIDs: filter by a list of application IDs.
	ApplicationIDs []string `json:"-"`

	// NoPrincipal: defines whether or not the policy is attributed to a principal.
	NoPrincipal *bool `json:"-"`

	// PolicyName: name of the policy to fetch.
	PolicyName *string `json:"-"`

	// Tag: filter by tags containing a given string.
	Tag *string `json:"-"`

	// PolicyIDs: filter by a list of IDs.
	PolicyIDs []string `json:"-"`
}

// ListPoliciesResponse: list policies response.
type ListPoliciesResponse struct {
	// Policies: list of policies.
	Policies []*Policy `json:"policies"`

	// TotalCount: total count of policies.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPoliciesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPoliciesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListPoliciesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Policies = append(r.Policies, results.Policies...)
	r.TotalCount += uint32(len(results.Policies))
	return uint32(len(results.Policies)), nil
}

// ListQuotaRequest: list quota request.
type ListQuotaRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: name_asc
	OrderBy ListQuotaRequestOrderBy `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater than 1.
	Page *int32 `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID string `json:"-"`

	// QuotumNames: list of quotum names to filter from.
	QuotumNames []string `json:"-"`
}

// ListQuotaResponse: list quota response.
type ListQuotaResponse struct {
	// Quota: list of quota.
	Quota []*Quotum `json:"quota"`

	// TotalCount: total count of quota.
	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListQuotaResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListQuotaResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListQuotaResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Quota = append(r.Quota, results.Quota...)
	r.TotalCount += uint64(len(results.Quota))
	return uint64(len(results.Quota)), nil
}

// ListRulesRequest: list rules request.
type ListRulesRequest struct {
	// PolicyID: id of policy to search.
	PolicyID string `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater than 1.
	Page *int32 `json:"-"`
}

// ListRulesResponse: list rules response.
type ListRulesResponse struct {
	// Rules: rules of the policy.
	Rules []*Rule `json:"rules"`

	// TotalCount: total count of rules.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListRulesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListRulesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListRulesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Rules = append(r.Rules, results.Rules...)
	r.TotalCount += uint32(len(results.Rules))
	return uint32(len(results.Rules)), nil
}

// ListSSHKeysRequest: list ssh keys request.
type ListSSHKeysRequest struct {
	// OrderBy: sort order of the SSH keys.
	// Default value: created_at_asc
	OrderBy ListSSHKeysRequestOrderBy `json:"-"`

	// Page: requested page number. Value must be greater or equal to 1.
	Page *int32 `json:"-"`

	// PageSize: number of items per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID *string `json:"-"`

	// Name: name of group to find.
	Name *string `json:"-"`

	// ProjectID: filter by Project ID.
	ProjectID *string `json:"-"`

	// Disabled: defines whether to include disabled SSH keys or not.
	Disabled *bool `json:"-"`
}

// ListSSHKeysResponse: list ssh keys response.
type ListSSHKeysResponse struct {
	// SSHKeys: list of SSH keys.
	SSHKeys []*SSHKey `json:"ssh_keys"`

	// TotalCount: total count of SSH keys.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSSHKeysResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSSHKeysResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListSSHKeysResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.SSHKeys = append(r.SSHKeys, results.SSHKeys...)
	r.TotalCount += uint32(len(results.SSHKeys))
	return uint32(len(results.SSHKeys)), nil
}

// ListSamlCertificatesRequest: list saml certificates request.
type ListSamlCertificatesRequest struct {
	// SamlID: ID of the SAML configuration.
	SamlID string `json:"-"`
}

// ListSamlCertificatesResponse: list saml certificates response.
type ListSamlCertificatesResponse struct {
	// Certificates: list of SAML certificates.
	Certificates []*SamlCertificate `json:"certificates"`
}

// ListScimTokensRequest: list scim tokens request.
type ListScimTokensRequest struct {
	// ScimID: ID of the SCIM configuration.
	ScimID string `json:"-"`

	// OrderBy: sort order of SCIM tokens.
	// Default value: created_at_asc
	OrderBy ListScimTokensRequestOrderBy `json:"-"`

	// Page: requested page number. Value must be greater or equal to 1.
	Page *int32 `json:"-"`

	// PageSize: number of items per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`
}

// ListScimTokensResponse: list scim tokens response.
type ListScimTokensResponse struct {
	// ScimTokens: list of SCIM tokens.
	ScimTokens []*ScimToken `json:"scim_tokens"`

	// TotalCount: total count of SCIM tokens.
	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListScimTokensResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListScimTokensResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListScimTokensResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.ScimTokens = append(r.ScimTokens, results.ScimTokens...)
	r.TotalCount += uint64(len(results.ScimTokens))
	return uint64(len(results.ScimTokens)), nil
}

// ListUsersRequest: list users request.
type ListUsersRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListUsersRequestOrderBy `json:"-"`

	// PageSize: number of results per page. Value must be between 1 and 100.
	PageSize *uint32 `json:"-"`

	// Page: page number. Value must be greater or equal to 1.
	Page *int32 `json:"-"`

	// OrganizationID: ID of the Organization to filter.
	OrganizationID *string `json:"-"`

	// UserIDs: filter by list of IDs.
	UserIDs []string `json:"-"`

	// Mfa: filter by MFA status.
	Mfa *bool `json:"-"`

	// Tag: filter by tags containing a given string.
	Tag *string `json:"-"`

	// Type: filter by user type.
	// Default value: unknown_type
	Type UserType `json:"-"`
}

// ListUsersResponse: list users response.
type ListUsersResponse struct {
	// Users: list of users.
	Users []*User `json:"users"`

	// TotalCount: total count of users.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListUsersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListUsersResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListUsersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Users = append(r.Users, results.Users...)
	r.TotalCount += uint32(len(results.Users))
	return uint32(len(results.Users)), nil
}

// LockUserRequest: lock user request.
type LockUserRequest struct {
	// UserID: ID of the user to lock.
	UserID string `json:"-"`
}

// MFAOTP: mfaotp.
type MFAOTP struct {
	Secret string `json:"secret"`
}

// Organization: organization.
type Organization struct {
	// ID: ID of the Organization.
	ID string `json:"id"`

	// Name: name of the Organization.
	Name string `json:"name"`

	// Alias: alias of the Organization.
	Alias string `json:"alias"`

	// LoginPasswordEnabled: defines whether login with a password is enabled for the Organization.
	LoginPasswordEnabled bool `json:"login_password_enabled"`

	// LoginMagicCodeEnabled: defines whether login with an authentication code is enabled for the Organization.
	LoginMagicCodeEnabled bool `json:"login_magic_code_enabled"`

	// LoginOauth2Enabled: defines whether login through OAuth2 is enabled for the Organization.
	LoginOauth2Enabled bool `json:"login_oauth2_enabled"`

	// LoginSamlEnabled: defines whether login through SAML is enabled for the Organization.
	LoginSamlEnabled bool `json:"login_saml_enabled"`
}

// OrganizationSecuritySettings: organization security settings.
type OrganizationSecuritySettings struct {
	// EnforcePasswordRenewal: defines whether password renewal is enforced during first login.
	EnforcePasswordRenewal bool `json:"enforce_password_renewal"`

	// GracePeriodDuration: duration of the grace period to renew password or enable MFA.
	GracePeriodDuration *scw.Duration `json:"grace_period_duration"`

	// LoginAttemptsBeforeLocked: number of login attempts before the account is locked.
	LoginAttemptsBeforeLocked uint32 `json:"login_attempts_before_locked"`

	// MaxLoginSessionDuration: maximum duration a login session will stay active before needing to relogin.
	MaxLoginSessionDuration *scw.Duration `json:"max_login_session_duration"`

	// MaxAPIKeyExpirationDuration: maximum duration the `expires_at` field of an API key can represent. A value of 0 means there is no maximum duration.
	MaxAPIKeyExpirationDuration *scw.Duration `json:"max_api_key_expiration_duration"`
}

// ParseSamlMetadataRequest: parse saml metadata request.
type ParseSamlMetadataRequest struct {
	File scw.File `json:"file"`
}

// ParseSamlMetadataResponse: parse saml metadata response.
type ParseSamlMetadataResponse struct {
	SingleSignOnURL string `json:"single_sign_on_url"`

	EntityID string `json:"entity_id"`

	SigningCertificates []string `json:"signing_certificates"`
}

// RemoveGroupMemberRequest: remove group member request.
type RemoveGroupMemberRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`

	// UserID: ID of the user to remove.
	// Precisely one of UserID, ApplicationID must be set.
	UserID *string `json:"user_id,omitempty"`

	// ApplicationID: ID of the application to remove.
	// Precisely one of UserID, ApplicationID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
}

// RemoveUserConnectionRequest: remove user connection request.
type RemoveUserConnectionRequest struct {
	// UserID: ID of the user you want to manage the connection for.
	UserID string `json:"-"`

	// TargetUserID: ID of the user you want to remove from your connection.
	TargetUserID string `json:"target_user_id"`
}

// Saml: saml.
type Saml struct {
	// ID: ID of the SAML configuration.
	ID string `json:"id"`

	// Status: status of the SAML configuration.
	// Default value: unknown_saml_status
	Status SamlStatus `json:"status"`

	// ServiceProvider: service Provider information.
	ServiceProvider *SamlServiceProvider `json:"service_provider"`

	// EntityID: entity ID of the SAML Identity Provider.
	EntityID string `json:"entity_id"`

	// SingleSignOnURL: single Sign-On URL of the SAML Identity Provider.
	SingleSignOnURL string `json:"single_sign_on_url"`
}

// Scim: scim.
type Scim struct {
	// ID: ID of the SCIM configuration.
	ID string `json:"id"`

	// CreatedAt: date and time of SCIM configuration creation.
	CreatedAt *time.Time `json:"created_at"`
}

// SetGroupMembersRequest: set group members request.
type SetGroupMembersRequest struct {
	GroupID string `json:"-"`

	UserIDs []string `json:"user_ids"`

	ApplicationIDs []string `json:"application_ids"`
}

// SetOrganizationAliasRequest: set organization alias request.
type SetOrganizationAliasRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`

	// Alias: alias of the Organization.
	Alias string `json:"alias"`
}

// SetRulesRequest: set rules request.
type SetRulesRequest struct {
	// PolicyID: id of policy to update.
	PolicyID string `json:"policy_id"`

	// Rules: rules of the policy to set.
	Rules []*RuleSpecs `json:"rules"`
}

// SetRulesResponse: set rules response.
type SetRulesResponse struct {
	// Rules: rules of the policy.
	Rules []*Rule `json:"rules"`
}

// UnlockUserRequest: unlock user request.
type UnlockUserRequest struct {
	// UserID: ID of the user to unlock.
	UserID string `json:"-"`
}

// UpdateAPIKeyRequest: update api key request.
type UpdateAPIKeyRequest struct {
	// AccessKey: access key to update.
	AccessKey string `json:"-"`

	// DefaultProjectID: new default Project ID to set.
	DefaultProjectID *string `json:"default_project_id,omitempty"`

	// Description: new description to update.
	Description *string `json:"description,omitempty"`

	// ExpiresAt: new expiration date of the API key.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpdateApplicationRequest: update application request.
type UpdateApplicationRequest struct {
	// ApplicationID: ID of the application to update.
	ApplicationID string `json:"-"`

	// Name: new name for the application (max length is 64 chars).
	Name *string `json:"name,omitempty"`

	// Description: new description for the application (max length is 200 chars).
	Description *string `json:"description,omitempty"`

	// Tags: new tags for the application (maximum of 10 tags).
	Tags *[]string `json:"tags,omitempty"`
}

// UpdateGroupRequest: update group request.
type UpdateGroupRequest struct {
	// GroupID: ID of the group to update.
	GroupID string `json:"-"`

	// Name: new name for the group (max length is 64 chars). MUST be unique inside an Organization.
	Name *string `json:"name,omitempty"`

	// Description: new description for the group (max length is 200 chars).
	Description *string `json:"description,omitempty"`

	// Tags: new tags for the group (maximum of 10 tags).
	Tags *[]string `json:"tags,omitempty"`
}

// UpdateOrganizationLoginMethodsRequest: update organization login methods request.
type UpdateOrganizationLoginMethodsRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`

	// LoginPasswordEnabled: defines whether login with a password is enabled for the Organization.
	LoginPasswordEnabled *bool `json:"login_password_enabled,omitempty"`

	// LoginOauth2Enabled: defines whether login through OAuth2 is enabled for the Organization.
	LoginOauth2Enabled *bool `json:"login_oauth2_enabled,omitempty"`

	// LoginMagicCodeEnabled: defines whether login with an authentication code is enabled for the Organization.
	LoginMagicCodeEnabled *bool `json:"login_magic_code_enabled,omitempty"`

	// LoginSamlEnabled: defines whether login through SAML is enabled for the Organization.
	LoginSamlEnabled *bool `json:"login_saml_enabled,omitempty"`
}

// UpdateOrganizationSecuritySettingsRequest: update organization security settings request.
type UpdateOrganizationSecuritySettingsRequest struct {
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`

	// EnforcePasswordRenewal: defines whether password renewal is enforced during first login.
	EnforcePasswordRenewal *bool `json:"enforce_password_renewal,omitempty"`

	// GracePeriodDuration: duration of the grace period to renew password or enable MFA.
	GracePeriodDuration *scw.Duration `json:"grace_period_duration,omitempty"`

	// LoginAttemptsBeforeLocked: number of login attempts before the account is locked.
	LoginAttemptsBeforeLocked *uint32 `json:"login_attempts_before_locked,omitempty"`

	// MaxLoginSessionDuration: maximum duration a login session will stay active before needing to relogin.
	MaxLoginSessionDuration *scw.Duration `json:"max_login_session_duration,omitempty"`

	// MaxAPIKeyExpirationDuration: maximum duration the `expires_at` field of an API key can represent. A value of 0 means there is no maximum duration.
	MaxAPIKeyExpirationDuration *scw.Duration `json:"max_api_key_expiration_duration,omitempty"`
}

// UpdatePolicyRequest: update policy request.
type UpdatePolicyRequest struct {
	// PolicyID: id of policy to update.
	PolicyID string `json:"-"`

	// Name: new name for the policy (max length is 64 characters).
	Name *string `json:"name,omitempty"`

	// Description: new description of policy (max length is 200 characters).
	Description *string `json:"description,omitempty"`

	// Tags: new tags for the policy (maximum of 10 tags).
	Tags *[]string `json:"tags,omitempty"`

	// UserID: new ID of user attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	UserID *string `json:"user_id,omitempty"`

	// GroupID: new ID of group attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	GroupID *string `json:"group_id,omitempty"`

	// ApplicationID: new ID of application attributed to the policy.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	ApplicationID *string `json:"application_id,omitempty"`

	// NoPrincipal: defines whether or not the policy is attributed to a principal.
	// Precisely one of UserID, GroupID, ApplicationID, NoPrincipal must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// UpdateSSHKeyRequest: update ssh key request.
type UpdateSSHKeyRequest struct {
	SSHKeyID string `json:"-"`

	// Name: name of the SSH key. Max length is 1000.
	Name *string `json:"name,omitempty"`

	// Disabled: enable or disable the SSH key.
	Disabled *bool `json:"disabled,omitempty"`
}

// UpdateSamlRequest: update saml request.
type UpdateSamlRequest struct {
	// SamlID: ID of the SAML configuration.
	SamlID string `json:"-"`

	// EntityID: entity ID of the SAML Identity Provider.
	EntityID *string `json:"entity_id,omitempty"`

	// SingleSignOnURL: single Sign-On URL of the SAML Identity Provider.
	SingleSignOnURL *string `json:"single_sign_on_url,omitempty"`
}

// UpdateUserPasswordRequest: update user password request.
type UpdateUserPasswordRequest struct {
	// UserID: ID of the user to update.
	UserID string `json:"-"`

	// Password: the new password.
	Password string `json:"password"`
}

// UpdateUserRequest: update user request.
type UpdateUserRequest struct {
	// UserID: ID of the user to update.
	UserID string `json:"-"`

	// Tags: new tags for the user (maximum of 10 tags).
	Tags *[]string `json:"tags,omitempty"`

	// Email: iAM member email.
	Email *string `json:"email,omitempty"`

	// FirstName: iAM member first name.
	FirstName *string `json:"first_name,omitempty"`

	// LastName: iAM member last name.
	LastName *string `json:"last_name,omitempty"`

	// PhoneNumber: iAM member phone number.
	PhoneNumber *string `json:"phone_number,omitempty"`

	// Locale: iAM member locale.
	Locale *string `json:"locale,omitempty"`
}

// UpdateUserUsernameRequest: update user username request.
type UpdateUserUsernameRequest struct {
	// UserID: ID of the user to update.
	UserID string `json:"-"`

	// Username: the new username.
	Username string `json:"username"`
}

// ValidateUserMFAOTPRequest: validate user mfaotp request.
type ValidateUserMFAOTPRequest struct {
	// UserID: user ID of the MFA OTP.
	UserID string `json:"-"`

	// OneTimePassword: a password generated using the OTP.
	OneTimePassword string `json:"one_time_password"`
}

// ValidateUserMFAOTPResponse: validate user mfaotp response.
type ValidateUserMFAOTPResponse struct {
	// RecoveryCodes: list of recovery codes usable for this OTP method.
	RecoveryCodes []string `json:"recovery_codes"`
}

// This API allows you to manage Identity and Access Management (IAM) across your Scaleway Organizations, Projects and resources.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

// ListSSHKeys: List SSH keys. By default, the SSH keys listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You can define additional parameters for your query such as `organization_id`, `name`, `project_id` and `disabled`.
func (s *API) ListSSHKeys(req *ListSSHKeysRequest, opts ...scw.RequestOption) (*ListSSHKeysResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "disabled", req.Disabled)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/ssh-keys",
		Query:  query,
	}

	var resp ListSSHKeysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSSHKey: Add a new SSH key to a Scaleway Project. You must specify the `name`, `public_key` and `project_id`.
func (s *API) CreateSSHKey(req *CreateSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("key")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/ssh-keys",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSSHKey: Retrieve information about a given SSH key, specified by the `ssh_key_id` parameter. The SSH key's full details, including `id`, `name`, `public_key`, and `project_id` are returned in the response.
func (s *API) GetSSHKey(req *GetSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return nil, errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/ssh-keys/" + fmt.Sprint(req.SSHKeyID) + "",
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSSHKey: Update the parameters of an SSH key, including `name` and `disable`.
func (s *API) UpdateSSHKey(req *UpdateSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return nil, errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/ssh-keys/" + fmt.Sprint(req.SSHKeyID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSSHKey: Delete a given SSH key, specified by the `ssh_key_id`. Deleting an SSH is permanent, and cannot be undone. Note that you might need to update any configurations that used the SSH key.
func (s *API) DeleteSSHKey(req *DeleteSSHKeyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/ssh-keys/" + fmt.Sprint(req.SSHKeyID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListUsers: List the users of an Organization. By default, the users listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request. You can also define additional parameters for your query such as `user_ids`.
func (s *API) ListUsers(req *ListUsersRequest, opts ...scw.RequestOption) (*ListUsersResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "user_ids", req.UserIDs)
	parameter.AddToQuery(query, "mfa", req.Mfa)
	parameter.AddToQuery(query, "tag", req.Tag)
	parameter.AddToQuery(query, "type", req.Type)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/users",
		Query:  query,
	}

	var resp ListUsersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUser: Retrieve information about a user, specified by the `user_id` parameter. The user's full details, including `id`, `email`, `organization_id`, `status` and `mfa` are returned in the response.
func (s *API) GetUser(req *GetUserRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "",
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUser: Update the parameters of a user, including `tags`.
func (s *API) UpdateUser(req *UpdateUserRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteUser: Remove a user from an Organization in which they are a guest. You must define the `user_id` in your request. Note that removing a user from an Organization automatically deletes their API keys, and any policies directly attached to them become orphaned.
func (s *API) DeleteUser(req *DeleteUserRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// CreateUser: Create a new user. You must define the `organization_id` in your request. If you are adding a member, enter the member's details. If you are adding a guest, you must define the `email` and not add the member attribute.
func (s *API) CreateUser(req *CreateUserRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserUsername: Update an user's username.
func (s *API) UpdateUserUsername(req *UpdateUserUsernameRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/update-username",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserPassword: Update an user's password.
func (s *API) UpdateUserPassword(req *UpdateUserPasswordRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/update-password",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateUserMFAOTP: Create a MFA OTP.
func (s *API) CreateUserMFAOTP(req *CreateUserMFAOTPRequest, opts ...scw.RequestOption) (*MFAOTP, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/mfa-otp",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp MFAOTP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ValidateUserMFAOTP: Validate a MFA OTP.
func (s *API) ValidateUserMFAOTP(req *ValidateUserMFAOTPRequest, opts ...scw.RequestOption) (*ValidateUserMFAOTPResponse, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/validate-mfa-otp",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ValidateUserMFAOTPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteUserMFAOTP: Delete a MFA OTP.
func (s *API) DeleteUserMFAOTP(req *DeleteUserMFAOTPRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/mfa-otp",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// LockUser: Lock a member. A locked member cannot log in or use API keys until the locked status is removed.
func (s *API) LockUser(req *LockUserRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/lock",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnlockUser: Unlock a member.
func (s *API) UnlockUser(req *UnlockUserRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/unlock",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListGracePeriods: List the grace periods of a member.
func (s *API) ListGracePeriods(req *ListGracePeriodsRequest, opts ...scw.RequestOption) (*ListGracePeriodsResponse, error) {
	var err error

	query := url.Values{}
	parameter.AddToQuery(query, "user_id", req.UserID)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/grace-periods",
		Query:  query,
	}

	var resp ListGracePeriodsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserConnections:
func (s *API) GetUserConnections(req *GetUserConnectionsRequest, opts ...scw.RequestOption) (*GetUserConnectionsResponse, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/connections",
	}

	var resp GetUserConnectionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// InitiateUserConnection:
func (s *API) InitiateUserConnection(req *InitiateUserConnectionRequest, opts ...scw.RequestOption) (*InitiateUserConnectionResponse, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/initiate-connection",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp InitiateUserConnectionResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// JoinUserConnection:
func (s *API) JoinUserConnection(req *JoinUserConnectionRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/join-connection",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// RemoveUserConnection:
func (s *API) RemoveUserConnection(req *RemoveUserConnectionRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "/remove-connection",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListApplications: List the applications of an Organization. By default, the applications listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request. You can also define additional parameters for your query such as `application_ids`.
func (s *API) ListApplications(req *ListApplicationsRequest, opts ...scw.RequestOption) (*ListApplicationsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "editable", req.Editable)
	parameter.AddToQuery(query, "application_ids", req.ApplicationIDs)
	parameter.AddToQuery(query, "tag", req.Tag)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/applications",
		Query:  query,
	}

	var resp ListApplicationsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateApplication: Create a new application. You must define the `name` parameter in the request.
func (s *API) CreateApplication(req *CreateApplicationRequest, opts ...scw.RequestOption) (*Application, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("app")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/applications",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Application

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApplication: Retrieve information about an application, specified by the `application_id` parameter. The application's full details, including `id`, `email`, `organization_id`, `status` and `two_factor_enabled` are returned in the response.
func (s *API) GetApplication(req *GetApplicationRequest, opts ...scw.RequestOption) (*Application, error) {
	var err error

	if fmt.Sprint(req.ApplicationID) == "" {
		return nil, errors.New("field ApplicationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/applications/" + fmt.Sprint(req.ApplicationID) + "",
	}

	var resp Application

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateApplication: Update the parameters of an application, including `name` and `description`.
func (s *API) UpdateApplication(req *UpdateApplicationRequest, opts ...scw.RequestOption) (*Application, error) {
	var err error

	if fmt.Sprint(req.ApplicationID) == "" {
		return nil, errors.New("field ApplicationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/applications/" + fmt.Sprint(req.ApplicationID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Application

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteApplication: Delete an application. Note that this action is irreversible and will automatically delete the application's API keys. Policies attached to users and applications via this group will no longer apply.
func (s *API) DeleteApplication(req *DeleteApplicationRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.ApplicationID) == "" {
		return errors.New("field ApplicationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/applications/" + fmt.Sprint(req.ApplicationID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListGroups: List groups. By default, the groups listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You can define additional parameters to filter your query. Use `user_ids` or `application_ids` to list all groups certain users or applications belong to.
func (s *API) ListGroups(req *ListGroupsRequest, opts ...scw.RequestOption) (*ListGroupsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "application_ids", req.ApplicationIDs)
	parameter.AddToQuery(query, "user_ids", req.UserIDs)
	parameter.AddToQuery(query, "group_ids", req.GroupIDs)
	parameter.AddToQuery(query, "tag", req.Tag)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/groups",
		Query:  query,
	}

	var resp ListGroupsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateGroup: Create a new group. You must define the `name` and `organization_id` parameters in the request.
func (s *API) CreateGroup(req *CreateGroupRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("grp")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/groups",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGroup: Retrieve information about a given group, specified by the `group_id` parameter. The group's full details, including `user_ids` and `application_ids` are returned in the response.
func (s *API) GetGroup(req *GetGroupRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "",
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateGroup: Update the parameters of group, including `name` and `description`.
func (s *API) UpdateGroup(req *UpdateGroupRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetGroupMembers: Overwrite users and applications configuration in a group. Any information that you add using this command will overwrite the previous configuration.
func (s *API) SetGroupMembers(req *SetGroupMembersRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/members",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddGroupMember: Add a user or an application to a group. You can specify a `user_id` and `application_id` in the body of your request. Note that you can only add one of each per request.
func (s *API) AddGroupMember(req *AddGroupMemberRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/add-member",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddGroupMembers: Add multiple users and applications to a group in a single call. You can specify an array of `user_id`s and `application_id`s. Note that any existing users and applications in the group will remain. To add new users/applications and delete pre-existing ones, use the [Overwrite users and applications of a group](#path-groups-overwrite-users-and-applications-of-a-group) method.
func (s *API) AddGroupMembers(req *AddGroupMembersRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/add-members",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemoveGroupMember: Remove a user or an application from a group. You can specify a `user_id` and `application_id` in the body of your request. Note that you can only remove one of each per request. Removing a user from a group means that any permissions given to them via the group (i.e. from an attached policy) will no longer apply. Be sure you want to remove these permissions from the user before proceeding.
func (s *API) RemoveGroupMember(req *RemoveGroupMemberRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/remove-member",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteGroup: Delete a group. Note that this action is irreversible and could delete permissions for group members. Policies attached to users and applications via this group will no longer apply.
func (s *API) DeleteGroup(req *DeleteGroupRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListPolicies: List the policies of an Organization. By default, the policies listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request. You can also define additional parameters to filter your query, such as `user_ids`, `groups_ids`, `application_ids`, and `policy_name`.
func (s *API) ListPolicies(req *ListPoliciesRequest, opts ...scw.RequestOption) (*ListPoliciesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "editable", req.Editable)
	parameter.AddToQuery(query, "user_ids", req.UserIDs)
	parameter.AddToQuery(query, "group_ids", req.GroupIDs)
	parameter.AddToQuery(query, "application_ids", req.ApplicationIDs)
	parameter.AddToQuery(query, "no_principal", req.NoPrincipal)
	parameter.AddToQuery(query, "policy_name", req.PolicyName)
	parameter.AddToQuery(query, "tag", req.Tag)
	parameter.AddToQuery(query, "policy_ids", req.PolicyIDs)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/policies",
		Query:  query,
	}

	var resp ListPoliciesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePolicy: Create a new application. You must define the `name` parameter in the request. You can specify parameters such as `user_id`, `groups_id`, `application_id`, `no_principal`, `rules` and its child attributes.
func (s *API) CreatePolicy(req *CreatePolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("pol")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/policies",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Policy

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPolicy: Retrieve information about a policy, specified by the `policy_id` parameter. The policy's full details, including `id`, `name`, `organization_id`, `nb_rules` and `nb_scopes`, `nb_permission_sets` are returned in the response.
func (s *API) GetPolicy(req *GetPolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return nil, errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "",
	}

	var resp Policy

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdatePolicy: Update the parameters of a policy, including `name`, `description`, `user_id`, `group_id`, `application_id` and `no_principal`.
func (s *API) UpdatePolicy(req *UpdatePolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return nil, errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Policy

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeletePolicy: Delete a policy. You must define specify the `policy_id` parameter in your request. Note that when deleting a policy, all permissions it gives to its principal (user, group or application) will be revoked.
func (s *API) DeletePolicy(req *DeletePolicyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ClonePolicy: Clone a policy. You must define specify the `policy_id` parameter in your request.
func (s *API) ClonePolicy(req *ClonePolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return nil, errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "/clone",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Policy

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetRules: Overwrite the rules of a given policy. Any information that you add using this command will overwrite the previous configuration. If you include some of the rules you already had in your previous configuration in your new one, but you change their order, the new order of display will apply. While policy rules are ordered, they have no impact on the access logic of IAM because rules are allow-only.
func (s *API) SetRules(req *SetRulesRequest, opts ...scw.RequestOption) (*SetRulesResponse, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/iam/v1alpha1/rules",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRules: List the rules of a given policy. By default, the rules listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `policy_id` in the query path of your request.
func (s *API) ListRules(req *ListRulesRequest, opts ...scw.RequestOption) (*ListRulesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "policy_id", req.PolicyID)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/rules",
		Query:  query,
	}

	var resp ListRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPermissionSets: List permission sets available for given Organization. You must define the `organization_id` in the query path of your request.
func (s *API) ListPermissionSets(req *ListPermissionSetsRequest, opts ...scw.RequestOption) (*ListPermissionSetsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/permission-sets",
		Query:  query,
	}

	var resp ListPermissionSetsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListAPIKeys: List API keys. By default, the API keys listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You can define additional parameters for your query such as `editable`, `expired`, `access_key` and `bearer_id`.
func (s *API) ListAPIKeys(req *ListAPIKeysRequest, opts ...scw.RequestOption) (*ListAPIKeysResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "editable", req.Editable)
	parameter.AddToQuery(query, "expired", req.Expired)
	parameter.AddToQuery(query, "access_key", req.AccessKey)
	parameter.AddToQuery(query, "description", req.Description)
	parameter.AddToQuery(query, "bearer_id", req.BearerID)
	parameter.AddToQuery(query, "bearer_type", req.BearerType)
	parameter.AddToQuery(query, "access_keys", req.AccessKeys)
	parameter.AddToQuery(query, "application_id", req.ApplicationID)
	parameter.AddToQuery(query, "user_id", req.UserID)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/api-keys",
		Query:  query,
	}

	var resp ListAPIKeysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateAPIKey: Create an API key. You must specify the `application_id` or the `user_id` and the description. You can also specify the `default_project_id`, which is the Project ID of your preferred Project, to use with Object Storage. The `access_key` and `secret_key` values are returned in the response. Note that the secret key is only shown once. Make sure that you copy and store both keys somewhere safe.
func (s *API) CreateAPIKey(req *CreateAPIKeyRequest, opts ...scw.RequestOption) (*APIKey, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/api-keys",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp APIKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAPIKey: Retrieve information about an API key, specified by the `access_key` parameter. The API key's details, including either the `user_id` or `application_id` of its bearer are returned in the response. Note that the string value for the `secret_key` is nullable, and therefore is not displayed in the response. The `secret_key` value is only displayed upon API key creation.
func (s *API) GetAPIKey(req *GetAPIKeyRequest, opts ...scw.RequestOption) (*APIKey, error) {
	var err error

	if fmt.Sprint(req.AccessKey) == "" {
		return nil, errors.New("field AccessKey cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/api-keys/" + fmt.Sprint(req.AccessKey) + "",
	}

	var resp APIKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateAPIKey: Update the parameters of an API key, including `default_project_id` and `description`.
func (s *API) UpdateAPIKey(req *UpdateAPIKeyRequest, opts ...scw.RequestOption) (*APIKey, error) {
	var err error

	if fmt.Sprint(req.AccessKey) == "" {
		return nil, errors.New("field AccessKey cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/api-keys/" + fmt.Sprint(req.AccessKey) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp APIKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteAPIKey: Delete an API key. Note that this action is irreversible and cannot be undone. Make sure you update any configurations using the API keys you delete.
func (s *API) DeleteAPIKey(req *DeleteAPIKeyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.AccessKey) == "" {
		return errors.New("field AccessKey cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/api-keys/" + fmt.Sprint(req.AccessKey) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListQuota: List all product and features quota for an Organization, with their associated limits. By default, the quota listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request.
func (s *API) ListQuota(req *ListQuotaRequest, opts ...scw.RequestOption) (*ListQuotaResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "quotum_names", req.QuotumNames)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/quota",
		Query:  query,
	}

	var resp ListQuotaResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetQuotum: Retrieve information about a resource quota, specified by the `quotum_name` parameter. The quota's `limit`, or whether it is unlimited, is returned in the response.
func (s *API) GetQuotum(req *GetQuotumRequest, opts ...scw.RequestOption) (*Quotum, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	if fmt.Sprint(req.QuotumName) == "" {
		return nil, errors.New("field QuotumName cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/quota/" + fmt.Sprint(req.QuotumName) + "",
		Query:  query,
	}

	var resp Quotum

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListJWTs: List JWTs.
func (s *API) ListJWTs(req *ListJWTsRequest, opts ...scw.RequestOption) (*ListJWTsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "audience_id", req.AudienceID)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "expired", req.Expired)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/jwts",
		Query:  query,
	}

	var resp ListJWTsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateJWT: Create a JWT.
func (s *API) CreateJWT(req *CreateJWTRequest, opts ...scw.RequestOption) (*EncodedJWT, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/jwts",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp EncodedJWT

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetJWT: Get a JWT.
func (s *API) GetJWT(req *GetJWTRequest, opts ...scw.RequestOption) (*JWT, error) {
	var err error

	if fmt.Sprint(req.Jti) == "" {
		return nil, errors.New("field Jti cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/jwts/" + fmt.Sprint(req.Jti) + "",
	}

	var resp JWT

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteJWT: Delete a JWT.
func (s *API) DeleteJWT(req *DeleteJWTRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.Jti) == "" {
		return errors.New("field Jti cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/jwts/" + fmt.Sprint(req.Jti) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListLogs: List logs available for given Organization. You must define the `organization_id` in the query path of your request.
func (s *API) ListLogs(req *ListLogsRequest, opts ...scw.RequestOption) (*ListLogsResponse, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "created_after", req.CreatedAfter)
	parameter.AddToQuery(query, "created_before", req.CreatedBefore)
	parameter.AddToQuery(query, "action", req.Action)
	parameter.AddToQuery(query, "resource_type", req.ResourceType)
	parameter.AddToQuery(query, "search", req.Search)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/logs",
		Query:  query,
	}

	var resp ListLogsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLog: Retrieve information about a log, specified by the `log_id` parameter. The log's full details, including `id`, `ip`, `user_agent`, `action`, `bearer_id`, `resource_type` and `resource_id` are returned in the response.
func (s *API) GetLog(req *GetLogRequest, opts ...scw.RequestOption) (*Log, error) {
	var err error

	if fmt.Sprint(req.LogID) == "" {
		return nil, errors.New("field LogID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/logs/" + fmt.Sprint(req.LogID) + "",
	}

	var resp Log

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrganizationSecuritySettings: Retrieve information about the security settings of an Organization, specified by the `organization_id` parameter.
func (s *API) GetOrganizationSecuritySettings(req *GetOrganizationSecuritySettingsRequest, opts ...scw.RequestOption) (*OrganizationSecuritySettings, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/security-settings",
	}

	var resp OrganizationSecuritySettings

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateOrganizationSecuritySettings: Update the security settings of an Organization.
func (s *API) UpdateOrganizationSecuritySettings(req *UpdateOrganizationSecuritySettingsRequest, opts ...scw.RequestOption) (*OrganizationSecuritySettings, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/security-settings",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp OrganizationSecuritySettings

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetOrganizationAlias: This will fail if an alias has already been defined. Please contact support if you need to change your Organization's alias.
func (s *API) SetOrganizationAlias(req *SetOrganizationAliasRequest, opts ...scw.RequestOption) (*Organization, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/alias",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Organization

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrganization: Get your Organization's IAM information.
func (s *API) GetOrganization(req *GetOrganizationRequest, opts ...scw.RequestOption) (*Organization, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "",
	}

	var resp Organization

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateOrganizationLoginMethods: Set your Organization's allowed login methods.
func (s *API) UpdateOrganizationLoginMethods(req *UpdateOrganizationLoginMethodsRequest, opts ...scw.RequestOption) (*Organization, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/login-methods",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Organization

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrganizationSaml: Get SAML Identity Provider configuration of an Organization.
func (s *API) GetOrganizationSaml(req *GetOrganizationSamlRequest, opts ...scw.RequestOption) (*Saml, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/saml",
	}

	var resp Saml

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// EnableOrganizationSaml: Enable SAML Identity Provider for an Organization.
func (s *API) EnableOrganizationSaml(req *EnableOrganizationSamlRequest, opts ...scw.RequestOption) (*Saml, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/saml",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Saml

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSaml: Update SAML Identity Provider configuration.
func (s *API) UpdateSaml(req *UpdateSamlRequest, opts ...scw.RequestOption) (*Saml, error) {
	var err error

	if fmt.Sprint(req.SamlID) == "" {
		return nil, errors.New("field SamlID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/iam/v1alpha1/saml/" + fmt.Sprint(req.SamlID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Saml

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSaml: Disable SAML Identity Provider for an Organization.
func (s *API) DeleteSaml(req *DeleteSamlRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.SamlID) == "" {
		return errors.New("field SamlID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/saml/" + fmt.Sprint(req.SamlID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ParseSamlMetadata: Parse SAML xml metadata file.
func (s *API) ParseSamlMetadata(req *ParseSamlMetadataRequest, opts ...scw.RequestOption) (*ParseSamlMetadataResponse, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/parse-saml-metadata",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ParseSamlMetadataResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSamlCertificates: List SAML certificates.
func (s *API) ListSamlCertificates(req *ListSamlCertificatesRequest, opts ...scw.RequestOption) (*ListSamlCertificatesResponse, error) {
	var err error

	if fmt.Sprint(req.SamlID) == "" {
		return nil, errors.New("field SamlID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/saml/" + fmt.Sprint(req.SamlID) + "/certificates",
	}

	var resp ListSamlCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddSamlCertificate: Add a SAML certificate.
func (s *API) AddSamlCertificate(req *AddSamlCertificateRequest, opts ...scw.RequestOption) (*SamlCertificate, error) {
	var err error

	if fmt.Sprint(req.SamlID) == "" {
		return nil, errors.New("field SamlID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/saml/" + fmt.Sprint(req.SamlID) + "/certificates",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SamlCertificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSamlCertificate: Delete a SAML certificate.
func (s *API) DeleteSamlCertificate(req *DeleteSamlCertificateRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.CertificateID) == "" {
		return errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/saml-certificates/" + fmt.Sprint(req.CertificateID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GetOrganizationScim: Get SCIM configuration of an Organization.
func (s *API) GetOrganizationScim(req *GetOrganizationScimRequest, opts ...scw.RequestOption) (*Scim, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/scim",
	}

	var resp Scim

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// EnableOrganizationScim: Enable SCIM for an Organization.
func (s *API) EnableOrganizationScim(req *EnableOrganizationScimRequest, opts ...scw.RequestOption) (*Scim, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if fmt.Sprint(req.OrganizationID) == "" {
		return nil, errors.New("field OrganizationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/organizations/" + fmt.Sprint(req.OrganizationID) + "/scim",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Scim

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteScim: Disable SCIM for an Organization.
func (s *API) DeleteScim(req *DeleteScimRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.ScimID) == "" {
		return errors.New("field ScimID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/scim/" + fmt.Sprint(req.ScimID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListScimTokens: List SCIM tokens.
func (s *API) ListScimTokens(req *ListScimTokensRequest, opts ...scw.RequestOption) (*ListScimTokensResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.ScimID) == "" {
		return nil, errors.New("field ScimID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/iam/v1alpha1/scim/" + fmt.Sprint(req.ScimID) + "/tokens",
		Query:  query,
	}

	var resp ListScimTokensResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateScimToken: Create a SCIM token.
func (s *API) CreateScimToken(req *CreateScimTokenRequest, opts ...scw.RequestOption) (*CreateScimTokenResponse, error) {
	var err error

	if fmt.Sprint(req.ScimID) == "" {
		return nil, errors.New("field ScimID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/iam/v1alpha1/scim/" + fmt.Sprint(req.ScimID) + "/tokens",
	}

	var resp CreateScimTokenResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteScimToken: Delete a SCIM token.
func (s *API) DeleteScimToken(req *DeleteScimTokenRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.TokenID) == "" {
		return errors.New("field TokenID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/iam/v1alpha1/scim-tokens/" + fmt.Sprint(req.TokenID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}
