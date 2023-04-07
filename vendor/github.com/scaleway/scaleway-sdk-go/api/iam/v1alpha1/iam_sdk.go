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

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
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

// API: iAM API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type BearerType string

const (
	BearerTypeUnknownBearerType = BearerType("unknown_bearer_type")
	BearerTypeUser              = BearerType("user")
	BearerTypeApplication       = BearerType("application")
)

func (enum BearerType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_bearer_type"
	}
	return string(enum)
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

type ListAPIKeysRequestOrderBy string

const (
	ListAPIKeysRequestOrderByCreatedAtAsc  = ListAPIKeysRequestOrderBy("created_at_asc")
	ListAPIKeysRequestOrderByCreatedAtDesc = ListAPIKeysRequestOrderBy("created_at_desc")
	ListAPIKeysRequestOrderByUpdatedAtAsc  = ListAPIKeysRequestOrderBy("updated_at_asc")
	ListAPIKeysRequestOrderByUpdatedAtDesc = ListAPIKeysRequestOrderBy("updated_at_desc")
	ListAPIKeysRequestOrderByExpiresAtAsc  = ListAPIKeysRequestOrderBy("expires_at_asc")
	ListAPIKeysRequestOrderByExpiresAtDesc = ListAPIKeysRequestOrderBy("expires_at_desc")
	ListAPIKeysRequestOrderByAccessKeyAsc  = ListAPIKeysRequestOrderBy("access_key_asc")
	ListAPIKeysRequestOrderByAccessKeyDesc = ListAPIKeysRequestOrderBy("access_key_desc")
)

func (enum ListAPIKeysRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
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
	ListApplicationsRequestOrderByCreatedAtAsc  = ListApplicationsRequestOrderBy("created_at_asc")
	ListApplicationsRequestOrderByCreatedAtDesc = ListApplicationsRequestOrderBy("created_at_desc")
	ListApplicationsRequestOrderByUpdatedAtAsc  = ListApplicationsRequestOrderBy("updated_at_asc")
	ListApplicationsRequestOrderByUpdatedAtDesc = ListApplicationsRequestOrderBy("updated_at_desc")
	ListApplicationsRequestOrderByNameAsc       = ListApplicationsRequestOrderBy("name_asc")
	ListApplicationsRequestOrderByNameDesc      = ListApplicationsRequestOrderBy("name_desc")
)

func (enum ListApplicationsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
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
	ListGroupsRequestOrderByCreatedAtAsc  = ListGroupsRequestOrderBy("created_at_asc")
	ListGroupsRequestOrderByCreatedAtDesc = ListGroupsRequestOrderBy("created_at_desc")
	ListGroupsRequestOrderByUpdatedAtAsc  = ListGroupsRequestOrderBy("updated_at_asc")
	ListGroupsRequestOrderByUpdatedAtDesc = ListGroupsRequestOrderBy("updated_at_desc")
	ListGroupsRequestOrderByNameAsc       = ListGroupsRequestOrderBy("name_asc")
	ListGroupsRequestOrderByNameDesc      = ListGroupsRequestOrderBy("name_desc")
)

func (enum ListGroupsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
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
	ListJWTsRequestOrderByCreatedAtAsc  = ListJWTsRequestOrderBy("created_at_asc")
	ListJWTsRequestOrderByCreatedAtDesc = ListJWTsRequestOrderBy("created_at_desc")
	ListJWTsRequestOrderByUpdatedAtAsc  = ListJWTsRequestOrderBy("updated_at_asc")
	ListJWTsRequestOrderByUpdatedAtDesc = ListJWTsRequestOrderBy("updated_at_desc")
)

func (enum ListJWTsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
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

type ListPermissionSetsRequestOrderBy string

const (
	ListPermissionSetsRequestOrderByNameAsc       = ListPermissionSetsRequestOrderBy("name_asc")
	ListPermissionSetsRequestOrderByNameDesc      = ListPermissionSetsRequestOrderBy("name_desc")
	ListPermissionSetsRequestOrderByCreatedAtAsc  = ListPermissionSetsRequestOrderBy("created_at_asc")
	ListPermissionSetsRequestOrderByCreatedAtDesc = ListPermissionSetsRequestOrderBy("created_at_desc")
)

func (enum ListPermissionSetsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "name_asc"
	}
	return string(enum)
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
	ListPoliciesRequestOrderByPolicyNameAsc  = ListPoliciesRequestOrderBy("policy_name_asc")
	ListPoliciesRequestOrderByPolicyNameDesc = ListPoliciesRequestOrderBy("policy_name_desc")
	ListPoliciesRequestOrderByCreatedAtAsc   = ListPoliciesRequestOrderBy("created_at_asc")
	ListPoliciesRequestOrderByCreatedAtDesc  = ListPoliciesRequestOrderBy("created_at_desc")
)

func (enum ListPoliciesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "policy_name_asc"
	}
	return string(enum)
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
	ListQuotaRequestOrderByNameAsc  = ListQuotaRequestOrderBy("name_asc")
	ListQuotaRequestOrderByNameDesc = ListQuotaRequestOrderBy("name_desc")
)

func (enum ListQuotaRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "name_asc"
	}
	return string(enum)
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
	ListSSHKeysRequestOrderByCreatedAtAsc  = ListSSHKeysRequestOrderBy("created_at_asc")
	ListSSHKeysRequestOrderByCreatedAtDesc = ListSSHKeysRequestOrderBy("created_at_desc")
	ListSSHKeysRequestOrderByUpdatedAtAsc  = ListSSHKeysRequestOrderBy("updated_at_asc")
	ListSSHKeysRequestOrderByUpdatedAtDesc = ListSSHKeysRequestOrderBy("updated_at_desc")
	ListSSHKeysRequestOrderByNameAsc       = ListSSHKeysRequestOrderBy("name_asc")
	ListSSHKeysRequestOrderByNameDesc      = ListSSHKeysRequestOrderBy("name_desc")
)

func (enum ListSSHKeysRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
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

type ListUsersRequestOrderBy string

const (
	ListUsersRequestOrderByCreatedAtAsc  = ListUsersRequestOrderBy("created_at_asc")
	ListUsersRequestOrderByCreatedAtDesc = ListUsersRequestOrderBy("created_at_desc")
	ListUsersRequestOrderByUpdatedAtAsc  = ListUsersRequestOrderBy("updated_at_asc")
	ListUsersRequestOrderByUpdatedAtDesc = ListUsersRequestOrderBy("updated_at_desc")
	ListUsersRequestOrderByEmailAsc      = ListUsersRequestOrderBy("email_asc")
	ListUsersRequestOrderByEmailDesc     = ListUsersRequestOrderBy("email_desc")
	ListUsersRequestOrderByLastLoginAsc  = ListUsersRequestOrderBy("last_login_asc")
	ListUsersRequestOrderByLastLoginDesc = ListUsersRequestOrderBy("last_login_desc")
)

func (enum ListUsersRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
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

type PermissionSetScopeType string

const (
	PermissionSetScopeTypeUnknownScopeType = PermissionSetScopeType("unknown_scope_type")
	PermissionSetScopeTypeProjects         = PermissionSetScopeType("projects")
	PermissionSetScopeTypeOrganization     = PermissionSetScopeType("organization")
	PermissionSetScopeTypeAccountRootUser  = PermissionSetScopeType("account_root_user")
)

func (enum PermissionSetScopeType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_scope_type"
	}
	return string(enum)
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

type UserStatus string

const (
	UserStatusUnknownStatus     = UserStatus("unknown_status")
	UserStatusInvitationPending = UserStatus("invitation_pending")
	UserStatusActivated         = UserStatus("activated")
)

func (enum UserStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_status"
	}
	return string(enum)
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
	UserTypeUnknownType = UserType("unknown_type")
	UserTypeGuest       = UserType("guest")
	UserTypeOwner       = UserType("owner")
)

func (enum UserType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_type"
	}
	return string(enum)
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
	// DefaultProjectID: the default Project ID specified for this API key.
	DefaultProjectID string `json:"default_project_id"`
	// Editable: whether or not the API key is editable.
	Editable bool `json:"editable"`
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
	// Editable: whether or not the application is editable.
	Editable bool `json:"editable"`
	// NbAPIKeys: number of API keys attributed to the application.
	NbAPIKeys uint32 `json:"nb_api_keys"`
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

// ListAPIKeysResponse: list api keys response.
type ListAPIKeysResponse struct {
	// APIKeys: list of API keys.
	APIKeys []*APIKey `json:"api_keys"`
	// TotalCount: total count of API Keys.
	TotalCount uint32 `json:"total_count"`
}

// ListApplicationsResponse: list applications response.
type ListApplicationsResponse struct {
	// Applications: list of applications.
	Applications []*Application `json:"applications"`
	// TotalCount: total count of applications.
	TotalCount uint32 `json:"total_count"`
}

// ListGroupsResponse: list groups response.
type ListGroupsResponse struct {
	// Groups: list of groups.
	Groups []*Group `json:"groups"`
	// TotalCount: total count of groups.
	TotalCount uint32 `json:"total_count"`
}

type ListJWTsResponse struct {
	Jwts []*JWT `json:"jwts"`

	TotalCount uint64 `json:"total_count"`
}

// ListPermissionSetsResponse: list permission sets response.
type ListPermissionSetsResponse struct {
	// PermissionSets: list of permission sets.
	PermissionSets []*PermissionSet `json:"permission_sets"`
	// TotalCount: total count of permission sets.
	TotalCount uint32 `json:"total_count"`
}

// ListPoliciesResponse: list policies response.
type ListPoliciesResponse struct {
	// Policies: list of policies.
	Policies []*Policy `json:"policies"`
	// TotalCount: total count of policies.
	TotalCount uint32 `json:"total_count"`
}

// ListQuotaResponse: list quota response.
type ListQuotaResponse struct {
	// Quota: list of quota.
	Quota []*Quotum `json:"quota"`
	// TotalCount: total count of quota.
	TotalCount uint64 `json:"total_count"`
}

// ListRulesResponse: list rules response.
type ListRulesResponse struct {
	// Rules: rules of the policy.
	Rules []*Rule `json:"rules"`
	// TotalCount: total count of rules.
	TotalCount uint32 `json:"total_count"`
}

// ListSSHKeysResponse: list ssh keys response.
type ListSSHKeysResponse struct {
	// SSHKeys: list of SSH keys.
	SSHKeys []*SSHKey `json:"ssh_keys"`
	// TotalCount: total count of SSH keys.
	TotalCount uint32 `json:"total_count"`
}

// ListUsersResponse: list users response.
type ListUsersResponse struct {
	// Users: list of users.
	Users []*User `json:"users"`
	// TotalCount: total count of users.
	TotalCount uint32 `json:"total_count"`
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
	// Editable: whether or not a policy is editable.
	Editable bool `json:"editable"`
	// NbRules: number of rules of the policy.
	NbRules uint32 `json:"nb_rules"`
	// NbScopes: number of policy scopes.
	NbScopes uint32 `json:"nb_scopes"`
	// NbPermissionSets: number of permission sets of the policy.
	NbPermissionSets uint32 `json:"nb_permission_sets"`
	// UserID: ID of the user attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// GroupID: ID of the group attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	GroupID *string `json:"group_id,omitempty"`
	// ApplicationID: ID of the application attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// NoPrincipal: whether or not a policy is attributed to a principal.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// Quotum: quotum.
type Quotum struct {
	// Name: name of the quota.
	Name string `json:"name"`
	// Limit: maximum limit of the quota.
	// Precisely one of Limit, Unlimited must be set.
	Limit *uint64 `json:"limit,omitempty"`
	// Unlimited: whether or not the quota is unlimited.
	// Precisely one of Limit, Unlimited must be set.
	Unlimited *bool `json:"unlimited,omitempty"`
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
	// ProjectIDs: list of Project IDs the rule is scoped to.
	// Precisely one of AccountRootUserID, OrganizationID, ProjectIDs must be set.
	ProjectIDs *[]string `json:"project_ids,omitempty"`
	// OrganizationID: ID of Organization the rule is scoped to.
	// Precisely one of AccountRootUserID, OrganizationID, ProjectIDs must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// AccountRootUserID: ID of account root user the rule is scoped to.
	// Precisely one of AccountRootUserID, OrganizationID, ProjectIDs must be set.
	AccountRootUserID *string `json:"account_root_user_id,omitempty"`
}

// RuleSpecs: rule specs.
type RuleSpecs struct {
	// PermissionSetNames: names of permission sets bound to the rule.
	PermissionSetNames *[]string `json:"permission_set_names"`
	// ProjectIDs: list of Project IDs the rule is scoped to.
	// Precisely one of OrganizationID, ProjectIDs must be set.
	ProjectIDs *[]string `json:"project_ids,omitempty"`
	// OrganizationID: ID of Organization the rule is scoped to.
	// Precisely one of OrganizationID, ProjectIDs must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
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

// SetRulesResponse: set rules response.
type SetRulesResponse struct {
	// Rules: rules of the policy.
	Rules []*Rule `json:"rules"`
}

// User: user.
type User struct {
	// ID: ID of user.
	ID string `json:"id"`
	// Email: email of user.
	Email string `json:"email"`
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
	// TwoFactorEnabled: whether MFA is enabled.
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	// Status: status of user invitation.
	// Default value: unknown_status
	Status UserStatus `json:"status"`
}

// Service API

type ListSSHKeysRequest struct {
	// OrderBy: sort order of the SSH keys.
	// Default value: created_at_asc
	OrderBy ListSSHKeysRequestOrderBy `json:"-"`
	// Page: requested page number. Value must be greater or equal to 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// PageSize: number of items per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// OrganizationID: filter by Organization ID.
	OrganizationID *string `json:"-"`
	// Name: name of group to find.
	Name *string `json:"-"`
	// ProjectID: filter by Project ID.
	ProjectID *string `json:"-"`
	// Disabled: whether to include disabled SSH keys or not.
	Disabled *bool `json:"-"`
}

// ListSSHKeys: list SSH keys.
// List SSH keys. By default, the SSH keys listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You can define additional parameters for your query such as `organization_id`, `name`, `project_id` and `disabled`.
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
		Method:  "GET",
		Path:    "/iam/v1alpha1/ssh-keys",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSSHKeysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateSSHKeyRequest struct {
	// Name: the name of the SSH key. Max length is 1000.
	Name string `json:"name"`
	// PublicKey: SSH public key. Currently only the ssh-rsa, ssh-dss (DSA), ssh-ed25519 and ecdsa keys with NIST curves are supported. Max length is 65000.
	PublicKey string `json:"public_key"`
	// ProjectID: project the resource is attributed to.
	ProjectID string `json:"project_id"`
}

// CreateSSHKey: create an SSH key.
// Add a new SSH key to a Scaleway Project. You must specify the `name`, `public_key` and `project_id`.
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
		Method:  "POST",
		Path:    "/iam/v1alpha1/ssh-keys",
		Headers: http.Header{},
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

type GetSSHKeyRequest struct {
	// SSHKeyID: the ID of the SSH key.
	SSHKeyID string `json:"-"`
}

// GetSSHKey: get an SSH key.
// Retrieve information about a given SSH key, specified by the `ssh_key_id` parameter. The SSH key's full details, including `id`, `name`, `public_key`, and `project_id` are returned in the response.
func (s *API) GetSSHKey(req *GetSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return nil, errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/ssh-keys/" + fmt.Sprint(req.SSHKeyID) + "",
		Headers: http.Header{},
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateSSHKeyRequest struct {
	SSHKeyID string `json:"-"`
	// Name: name of the SSH key. Max length is 1000.
	Name *string `json:"name"`
	// Disabled: enable or disable the SSH key.
	Disabled *bool `json:"disabled"`
}

// UpdateSSHKey: update an SSH key.
// Update the parameters of an SSH key, including `name` and `disable`.
func (s *API) UpdateSSHKey(req *UpdateSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return nil, errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/iam/v1alpha1/ssh-keys/" + fmt.Sprint(req.SSHKeyID) + "",
		Headers: http.Header{},
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

type DeleteSSHKeyRequest struct {
	SSHKeyID string `json:"-"`
}

// DeleteSSHKey: delete an SSH key.
// Delete a given SSH key, specified by the `ssh_key_id`. Deleting an SSH is permanent, and cannot be undone. Note that you might need to update any configurations that used the SSH key.
func (s *API) DeleteSSHKey(req *DeleteSSHKeyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/ssh-keys/" + fmt.Sprint(req.SSHKeyID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListUsersRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListUsersRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater or equal to 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: ID of the Organization to filter.
	OrganizationID *string `json:"-"`
	// UserIDs: filter by list of IDs.
	UserIDs []string `json:"-"`
}

// ListUsers: list users of an Organization.
// List the users of an Organization. By default, the users listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request. You can also define additional parameters for your query such as `user_ids`.
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

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/users",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListUsersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetUserRequest struct {
	// UserID: ID of the user to find.
	UserID string `json:"-"`
}

// GetUser: get a given user.
// Retrieve information about a user, specified by the `user_id` parameter. The user's full details, including `id`, `email`, `organization_id`, `status` and `two_factor_enabled` are returned in the response.
func (s *API) GetUser(req *GetUserRequest, opts ...scw.RequestOption) (*User, error) {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return nil, errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "",
		Headers: http.Header{},
	}

	var resp User

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteUserRequest struct {
	// UserID: ID of the user to delete.
	UserID string `json:"-"`
}

// DeleteUser: delete a guest user from an Organization.
// Remove a user from an Organization in which they are a guest. You must define the `user_id` in your request. Note that removing a user from an Organization automatically deletes their API keys, and any policies directly attached to them become orphaned.
func (s *API) DeleteUser(req *DeleteUserRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.UserID) == "" {
		return errors.New("field UserID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/users/" + fmt.Sprint(req.UserID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListApplicationsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListApplicationsRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater than 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// Name: name of the application to filter.
	Name *string `json:"-"`
	// OrganizationID: ID of the Organization to filter.
	OrganizationID *string `json:"-"`
	// Editable: whether to filter out editable applications or not.
	Editable *bool `json:"-"`
	// ApplicationIDs: filter by list of IDs.
	ApplicationIDs []string `json:"-"`
}

// ListApplications: list applications of an Organization.
// List the applications of an Organization. By default, the applications listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request. You can also define additional parameters for your query such as `application_ids`.
func (s *API) ListApplications(req *ListApplicationsRequest, opts ...scw.RequestOption) (*ListApplicationsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "editable", req.Editable)
	parameter.AddToQuery(query, "application_ids", req.ApplicationIDs)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/applications",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListApplicationsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateApplicationRequest struct {
	// Name: name of the application to create (max length is 64 characters).
	Name string `json:"name"`
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`
	// Description: description of the application (max length is 200 characters).
	Description string `json:"description"`
}

// CreateApplication: create a new application.
// Create a new application. You must define the `name` parameter in the request.
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
		Method:  "POST",
		Path:    "/iam/v1alpha1/applications",
		Headers: http.Header{},
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

type GetApplicationRequest struct {
	// ApplicationID: ID of the application to find.
	ApplicationID string `json:"-"`
}

// GetApplication: get a given application.
// Retrieve information about an application, specified by the `application_id` parameter. The application's full details, including `id`, `email`, `organization_id`, `status` and `two_factor_enabled` are returned in the response.
func (s *API) GetApplication(req *GetApplicationRequest, opts ...scw.RequestOption) (*Application, error) {
	var err error

	if fmt.Sprint(req.ApplicationID) == "" {
		return nil, errors.New("field ApplicationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/applications/" + fmt.Sprint(req.ApplicationID) + "",
		Headers: http.Header{},
	}

	var resp Application

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateApplicationRequest struct {
	// ApplicationID: ID of the application to update.
	ApplicationID string `json:"-"`
	// Name: new name for the application (max length is 64 chars).
	Name *string `json:"name"`
	// Description: new description for the application (max length is 200 chars).
	Description *string `json:"description"`
}

// UpdateApplication: update an application.
// Update the parameters of an application, including `name` and `description`.
func (s *API) UpdateApplication(req *UpdateApplicationRequest, opts ...scw.RequestOption) (*Application, error) {
	var err error

	if fmt.Sprint(req.ApplicationID) == "" {
		return nil, errors.New("field ApplicationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/iam/v1alpha1/applications/" + fmt.Sprint(req.ApplicationID) + "",
		Headers: http.Header{},
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

type DeleteApplicationRequest struct {
	// ApplicationID: ID of the application to delete.
	ApplicationID string `json:"-"`
}

// DeleteApplication: delete an application.
// Delete an application. Note that this action is irreversible and will automatically delete the application's API keys. Policies attached to users and applications via this group will no longer apply.
func (s *API) DeleteApplication(req *DeleteApplicationRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.ApplicationID) == "" {
		return errors.New("field ApplicationID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/applications/" + fmt.Sprint(req.ApplicationID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListGroupsRequest struct {
	// OrderBy: sort order of groups.
	// Default value: created_at_asc
	OrderBy ListGroupsRequestOrderBy `json:"-"`
	// Page: requested page number. Value must be greater or equal to 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// PageSize: number of items per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// OrganizationID: filter by Organization ID.
	OrganizationID *string `json:"-"`
	// Name: name of group to find.
	Name *string `json:"-"`
	// ApplicationIDs: filter by a list of application IDs.
	ApplicationIDs []string `json:"-"`
	// UserIDs: filter by a list of user IDs.
	UserIDs []string `json:"-"`
	// GroupIDs: filter by a list of group IDs.
	GroupIDs []string `json:"-"`
}

// ListGroups: list groups.
// List groups. By default, the groups listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You can define additional parameters to filter your query. Use `user_ids` or `application_ids` to list all groups certain users or applications belong to.
func (s *API) ListGroups(req *ListGroupsRequest, opts ...scw.RequestOption) (*ListGroupsResponse, error) {
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
	parameter.AddToQuery(query, "application_ids", req.ApplicationIDs)
	parameter.AddToQuery(query, "user_ids", req.UserIDs)
	parameter.AddToQuery(query, "group_ids", req.GroupIDs)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/groups",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListGroupsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateGroupRequest struct {
	// OrganizationID: ID of Organization linked to the group.
	OrganizationID string `json:"organization_id"`
	// Name: name of the group to create (max length is 64 chars). MUST be unique inside an Organization.
	Name string `json:"name"`
	// Description: description of the group to create (max length is 200 chars).
	Description string `json:"description"`
}

// CreateGroup: create a group.
// Create a new group. You must define the `name` and `organization_id` parameters in the request.
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
		Method:  "POST",
		Path:    "/iam/v1alpha1/groups",
		Headers: http.Header{},
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

type GetGroupRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`
}

// GetGroup: get a group.
// Retrive information about a given group, specified by the `group_id` parameter. The group's full details, including `user_ids` and `application_ids` are returned in the response.
func (s *API) GetGroup(req *GetGroupRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "",
		Headers: http.Header{},
	}

	var resp Group

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateGroupRequest struct {
	// GroupID: ID of the group to update.
	GroupID string `json:"-"`
	// Name: new name for the group (max length is 64 chars). MUST be unique inside an Organization.
	Name *string `json:"name"`
	// Description: new description for the group (max length is 200 chars).
	Description *string `json:"description"`
}

// UpdateGroup: update a group.
// Update the parameters of group, including `name` and `description`.
func (s *API) UpdateGroup(req *UpdateGroupRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "",
		Headers: http.Header{},
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

type SetGroupMembersRequest struct {
	GroupID string `json:"-"`

	UserIDs []string `json:"user_ids"`

	ApplicationIDs []string `json:"application_ids"`
}

// SetGroupMembers: overwrite users and applications of a group.
// Overwrite users and applications configuration in a group. Any information that you add using this command will overwrite the previous configuration.
func (s *API) SetGroupMembers(req *SetGroupMembersRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/members",
		Headers: http.Header{},
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

type AddGroupMemberRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`
	// UserID: ID of the user to add.
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// ApplicationID: ID of the application to add.
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
}

// AddGroupMember: add a user or an application to a group.
// Add a user or an application to a group. You can specify a `user_id` and and `application_id` in the body of your request. Note that you can only add one of each per request.
func (s *API) AddGroupMember(req *AddGroupMemberRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/add-member",
		Headers: http.Header{},
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

type RemoveGroupMemberRequest struct {
	// GroupID: ID of the group.
	GroupID string `json:"-"`
	// UserID: ID of the user to remove.
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// ApplicationID: ID of the application to remove.
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
}

// RemoveGroupMember: remove a user or an application from a group.
// Remove a user or an application from a group. You can specify a `user_id` and and `application_id` in the body of your request. Note that you can only remove one of each per request. Removing a user from a group means that any permissions given to them via the group (i.e. from an attached policy) will no longer apply. Be sure you want to remove these permissions from the user before proceeding.
func (s *API) RemoveGroupMember(req *RemoveGroupMemberRequest, opts ...scw.RequestOption) (*Group, error) {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return nil, errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "/remove-member",
		Headers: http.Header{},
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

type DeleteGroupRequest struct {
	// GroupID: ID of the group to delete.
	GroupID string `json:"-"`
}

// DeleteGroup: delete a group.
// Delete a group. Note that this action is irreversible and could delete permissions for group members. Policies attached to users and applications via this group will no longer apply.
func (s *API) DeleteGroup(req *DeleteGroupRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.GroupID) == "" {
		return errors.New("field GroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/groups/" + fmt.Sprint(req.GroupID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListPoliciesRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListPoliciesRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater than 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: ID of the Organization to filter.
	OrganizationID *string `json:"-"`
	// Editable: whether or not filter out editable policies.
	Editable *bool `json:"-"`
	// UserIDs: whether or not to filter by list of user IDs.
	UserIDs []string `json:"-"`
	// GroupIDs: whether or not to filter by list of group IDs.
	GroupIDs []string `json:"-"`
	// ApplicationIDs: filter by a list of application IDs.
	ApplicationIDs []string `json:"-"`
	// NoPrincipal: whether or not the policy is attributed to a principal.
	NoPrincipal *bool `json:"-"`
	// PolicyName: name of the policy to fetch.
	PolicyName *string `json:"-"`
}

// ListPolicies: list policies of an Organization.
// List the policies of an Organization. By default, the policies listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request. You can also define additional parameters to filter your query, such as `user_ids`, `groups_ids`, `application_ids`, and `policy_name`.
func (s *API) ListPolicies(req *ListPoliciesRequest, opts ...scw.RequestOption) (*ListPoliciesResponse, error) {
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
	parameter.AddToQuery(query, "editable", req.Editable)
	parameter.AddToQuery(query, "user_ids", req.UserIDs)
	parameter.AddToQuery(query, "group_ids", req.GroupIDs)
	parameter.AddToQuery(query, "application_ids", req.ApplicationIDs)
	parameter.AddToQuery(query, "no_principal", req.NoPrincipal)
	parameter.AddToQuery(query, "policy_name", req.PolicyName)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/policies",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListPoliciesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreatePolicyRequest struct {
	// Name: name of the policy to create (max length is 64 characters).
	Name string `json:"name"`
	// Description: description of the policy to create (max length is 200 characters).
	Description string `json:"description"`
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"organization_id"`
	// Rules: rules of the policy to create.
	Rules []*RuleSpecs `json:"rules"`
	// UserID: ID of user attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// GroupID: ID of group attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	GroupID *string `json:"group_id,omitempty"`
	// ApplicationID: ID of application attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// NoPrincipal: whether or not a policy is attributed to a principal.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// CreatePolicy: create a new policy.
// Create a new application. You must define the `name` parameter in the request. You can specify parameters such as `user_id`, `groups_id`, `application_id`, `no_principal`, `rules` and its child attributes.
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
		Method:  "POST",
		Path:    "/iam/v1alpha1/policies",
		Headers: http.Header{},
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

type GetPolicyRequest struct {
	// PolicyID: id of policy to search.
	PolicyID string `json:"-"`
}

// GetPolicy: get an existing policy.
// Retrieve information about a policy, speficified by the `policy_id` parameter. The policy's full details, including `id`, `name`, `organization_id`, `nb_rules` and `nb_scopes`, `nb_permission_sets` are returned in the response.
func (s *API) GetPolicy(req *GetPolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return nil, errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "",
		Headers: http.Header{},
	}

	var resp Policy

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdatePolicyRequest struct {
	// PolicyID: id of policy to update.
	PolicyID string `json:"-"`
	// Name: new name for the policy (max length is 64 characters).
	Name *string `json:"name"`
	// Description: new description of policy (max length is 200 characters).
	Description *string `json:"description"`
	// UserID: new ID of user attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// GroupID: new ID of group attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	GroupID *string `json:"group_id,omitempty"`
	// ApplicationID: new ID of application attributed to the policy.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// NoPrincipal: whether or not the policy is attributed to a principal.
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// UpdatePolicy: update an existing policy.
// Update the parameters of a policy, including `name`, `description`, `user_id`, `group_id`, `application_id` and `no_principal`.
func (s *API) UpdatePolicy(req *UpdatePolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return nil, errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "",
		Headers: http.Header{},
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

type DeletePolicyRequest struct {
	// PolicyID: id of policy to delete.
	PolicyID string `json:"-"`
}

// DeletePolicy: delete a policy.
// Delete a policy. You must define specify the `policy_id` parameter in your request. Note that when deleting a policy, all permissions it gives to its principal (user, group or application) will be revoked.
func (s *API) DeletePolicy(req *DeletePolicyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ClonePolicyRequest struct {
	PolicyID string `json:"-"`
}

func (s *API) ClonePolicy(req *ClonePolicyRequest, opts ...scw.RequestOption) (*Policy, error) {
	var err error

	if fmt.Sprint(req.PolicyID) == "" {
		return nil, errors.New("field PolicyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/iam/v1alpha1/policies/" + fmt.Sprint(req.PolicyID) + "/clone",
		Headers: http.Header{},
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

type SetRulesRequest struct {
	// PolicyID: id of policy to update.
	PolicyID string `json:"policy_id"`
	// Rules: rules of the policy to set.
	Rules []*RuleSpecs `json:"rules"`
}

// SetRules: set rules of a given policy.
// Overwrite the rules of a given policy. Any information that you add using this command will overwrite the previous configuration. If you include some of the rules you already had in your previous configuration in your new one, but you change their order, the new order of display will apply. While policy rules are ordered, they have no impact on the access logic of IAM because rules are allow-only.
func (s *API) SetRules(req *SetRulesRequest, opts ...scw.RequestOption) (*SetRulesResponse, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/iam/v1alpha1/rules",
		Headers: http.Header{},
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

type ListRulesRequest struct {
	// PolicyID: id of policy to search.
	PolicyID *string `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater than 1.
	// Default value: 1
	Page *int32 `json:"-"`
}

// ListRules: list rules of a given policy.
// List the rules of a given policy. By default, the rules listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `policy_id` in the query path of your request.
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
		Method:  "GET",
		Path:    "/iam/v1alpha1/rules",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListPermissionSetsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListPermissionSetsRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater than 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: filter by Organization ID.
	OrganizationID string `json:"-"`
}

// ListPermissionSets: list permission sets.
// List permission sets available for given Organization. You must define the `organization_id` in the query path of your request.
func (s *API) ListPermissionSets(req *ListPermissionSetsRequest, opts ...scw.RequestOption) (*ListPermissionSetsResponse, error) {
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
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/permission-sets",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListPermissionSetsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListAPIKeysRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListAPIKeysRequestOrderBy `json:"-"`
	// Page: page number. Value must be greater or equal to 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// OrganizationID: ID of Organization.
	OrganizationID *string `json:"-"`
	// Deprecated: ApplicationID: ID of application that bears the API key.
	ApplicationID *string `json:"-"`
	// Deprecated: UserID: ID of user that bears the API key.
	UserID *string `json:"-"`
	// Editable: whether to filter out editable API keys or not.
	Editable *bool `json:"-"`
	// Expired: whether to filter out expired API keys or not.
	Expired *bool `json:"-"`
	// AccessKey: filter by access key.
	AccessKey *string `json:"-"`
	// Description: filter by description.
	Description *string `json:"-"`
	// BearerID: filter by bearer ID.
	BearerID *string `json:"-"`
	// BearerType: filter by type of bearer.
	// Default value: unknown_bearer_type
	BearerType BearerType `json:"-"`
}

// ListAPIKeys: list API keys.
// List API keys. By default, the API keys listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You can define additional parameters for your query such as `editable`, `expired`, `access_key` and `bearer_id`.
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
	parameter.AddToQuery(query, "application_id", req.ApplicationID)
	parameter.AddToQuery(query, "user_id", req.UserID)
	parameter.AddToQuery(query, "editable", req.Editable)
	parameter.AddToQuery(query, "expired", req.Expired)
	parameter.AddToQuery(query, "access_key", req.AccessKey)
	parameter.AddToQuery(query, "description", req.Description)
	parameter.AddToQuery(query, "bearer_id", req.BearerID)
	parameter.AddToQuery(query, "bearer_type", req.BearerType)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/api-keys",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListAPIKeysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateAPIKeyRequest struct {
	// ApplicationID: ID of the application.
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// UserID: ID of the user.
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// ExpiresAt: expiration date of the API key.
	ExpiresAt *time.Time `json:"expires_at"`
	// DefaultProjectID: the default Project ID to use with Object Storage.
	DefaultProjectID *string `json:"default_project_id"`
	// Description: the description of the API key (max length is 200 characters).
	Description string `json:"description"`
}

// CreateAPIKey: create an API key.
// Create an API key. You must specify the `application_id` or the `user_id` and the description. You can also specify the `default_project_id` which is the Project ID of your preferred Project, to use with Object Storage. The `access_key` and `secret_key` values are returned in the response. Note that he secret key is only showed once. Make sure that you copy and store both keys somewhere safe.
func (s *API) CreateAPIKey(req *CreateAPIKeyRequest, opts ...scw.RequestOption) (*APIKey, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/iam/v1alpha1/api-keys",
		Headers: http.Header{},
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

type GetAPIKeyRequest struct {
	// AccessKey: access key to search for.
	AccessKey string `json:"-"`
}

// GetAPIKey: get an API key.
// Retrive information about an API key, specified by the `access_key` parameter. The API key's details, including either the `user_id` or `application_id` of its bearer are returned in the response. Note that the string value for the `secret_key` is nullable, and therefore is not displayed in the response. The `secret_key` value is only displayed upon API key creation.
func (s *API) GetAPIKey(req *GetAPIKeyRequest, opts ...scw.RequestOption) (*APIKey, error) {
	var err error

	if fmt.Sprint(req.AccessKey) == "" {
		return nil, errors.New("field AccessKey cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/api-keys/" + fmt.Sprint(req.AccessKey) + "",
		Headers: http.Header{},
	}

	var resp APIKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateAPIKeyRequest struct {
	// AccessKey: access key to update.
	AccessKey string `json:"-"`
	// DefaultProjectID: the new default Project ID to set.
	DefaultProjectID *string `json:"default_project_id"`
	// Description: the new description to update.
	Description *string `json:"description"`
}

// UpdateAPIKey: update an API key.
// Update the parameters of an API key, including `default_project_id` and `description`.
func (s *API) UpdateAPIKey(req *UpdateAPIKeyRequest, opts ...scw.RequestOption) (*APIKey, error) {
	var err error

	if fmt.Sprint(req.AccessKey) == "" {
		return nil, errors.New("field AccessKey cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/iam/v1alpha1/api-keys/" + fmt.Sprint(req.AccessKey) + "",
		Headers: http.Header{},
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

type DeleteAPIKeyRequest struct {
	// AccessKey: access key to delete.
	AccessKey string `json:"-"`
}

// DeleteAPIKey: delete an API key.
// Delete an API key. Note that this action is irreversible and cannot be undone. Make sure you update any configurations using the API keys you delete.
func (s *API) DeleteAPIKey(req *DeleteAPIKeyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.AccessKey) == "" {
		return errors.New("field AccessKey cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/api-keys/" + fmt.Sprint(req.AccessKey) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListQuotaRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: name_asc
	OrderBy ListQuotaRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater than 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: filter by Organization ID.
	OrganizationID string `json:"-"`
}

// ListQuota: list all quotas in the Organization.
// List all product and features quota for an Organization, with their associated limits. By default, the quota listed are ordered by creation date in ascending order. This can be modified via the `order_by` field. You must define the `organization_id` in the query path of your request.
func (s *API) ListQuota(req *ListQuotaRequest, opts ...scw.RequestOption) (*ListQuotaResponse, error) {
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
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/quota",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListQuotaResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetQuotumRequest struct {
	// QuotumName: name of the quota to get.
	QuotumName string `json:"-"`
	// OrganizationID: ID of the Organization.
	OrganizationID string `json:"-"`
}

// GetQuotum: get a quota in the Organization.
// Retrieve information about a resource quota, speficified by the `quotum_name` parameter. The quota's `limit`, or whether it is unlimited, is returned in the response.
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
		Method:  "GET",
		Path:    "/iam/v1alpha1/quota/" + fmt.Sprint(req.QuotumName) + "",
		Query:   query,
		Headers: http.Header{},
	}

	var resp Quotum

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListJWTsRequest struct {
	// OrderBy: criteria for sorting results.
	// Default value: created_at_asc
	OrderBy ListJWTsRequestOrderBy `json:"-"`
	// AudienceID: ID of the user to search.
	AudienceID *string `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100.
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: page number. Value must be greater to 1.
	// Default value: 1
	Page *int32 `json:"-"`
	// Expired: filter out expired JWTs or not.
	Expired *bool `json:"-"`
}

// ListJWTs: list JWTs.
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
		Method:  "GET",
		Path:    "/iam/v1alpha1/jwts",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListJWTsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetJWTRequest struct {
	// Jti: jWT ID of the JWT to get.
	Jti string `json:"-"`
}

// GetJWT: get a JWT.
func (s *API) GetJWT(req *GetJWTRequest, opts ...scw.RequestOption) (*JWT, error) {
	var err error

	if fmt.Sprint(req.Jti) == "" {
		return nil, errors.New("field Jti cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/iam/v1alpha1/jwts/" + fmt.Sprint(req.Jti) + "",
		Headers: http.Header{},
	}

	var resp JWT

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteJWTRequest struct {
	// Jti: jWT ID of the JWT to delete.
	Jti string `json:"-"`
}

// DeleteJWT: delete a JWT.
func (s *API) DeleteJWT(req *DeleteJWTRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.Jti) == "" {
		return errors.New("field Jti cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/iam/v1alpha1/jwts/" + fmt.Sprint(req.Jti) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSSHKeysResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSSHKeysResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListSSHKeysResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.SSHKeys = append(r.SSHKeys, results.SSHKeys...)
	r.TotalCount += uint32(len(results.SSHKeys))
	return uint32(len(results.SSHKeys)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListUsersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListUsersResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListUsersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Users = append(r.Users, results.Users...)
	r.TotalCount += uint32(len(results.Users))
	return uint32(len(results.Users)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListApplicationsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListApplicationsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListApplicationsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Applications = append(r.Applications, results.Applications...)
	r.TotalCount += uint32(len(results.Applications))
	return uint32(len(results.Applications)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListGroupsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListGroupsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListGroupsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Groups = append(r.Groups, results.Groups...)
	r.TotalCount += uint32(len(results.Groups))
	return uint32(len(results.Groups)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPoliciesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPoliciesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListPoliciesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Policies = append(r.Policies, results.Policies...)
	r.TotalCount += uint32(len(results.Policies))
	return uint32(len(results.Policies)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListRulesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListRulesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListRulesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Rules = append(r.Rules, results.Rules...)
	r.TotalCount += uint32(len(results.Rules))
	return uint32(len(results.Rules)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPermissionSetsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPermissionSetsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListPermissionSetsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PermissionSets = append(r.PermissionSets, results.PermissionSets...)
	r.TotalCount += uint32(len(results.PermissionSets))
	return uint32(len(results.PermissionSets)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListAPIKeysResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListAPIKeysResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListAPIKeysResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.APIKeys = append(r.APIKeys, results.APIKeys...)
	r.TotalCount += uint32(len(results.APIKeys))
	return uint32(len(results.APIKeys)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListQuotaResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListQuotaResponse) UnsafeAppend(res interface{}) (uint64, error) {
	results, ok := res.(*ListQuotaResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Quota = append(r.Quota, results.Quota...)
	r.TotalCount += uint64(len(results.Quota))
	return uint64(len(results.Quota)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListJWTsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListJWTsResponse) UnsafeAppend(res interface{}) (uint64, error) {
	results, ok := res.(*ListJWTsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Jwts = append(r.Jwts, results.Jwts...)
	r.TotalCount += uint64(len(results.Jwts))
	return uint64(len(results.Jwts)), nil
}
