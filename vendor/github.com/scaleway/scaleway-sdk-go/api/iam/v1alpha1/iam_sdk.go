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

// API: iAM API
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ListAPIKeysRequestOrderBy string

const (
	// ListAPIKeysRequestOrderByCreatedAtAsc is [insert doc].
	ListAPIKeysRequestOrderByCreatedAtAsc = ListAPIKeysRequestOrderBy("created_at_asc")
	// ListAPIKeysRequestOrderByCreatedAtDesc is [insert doc].
	ListAPIKeysRequestOrderByCreatedAtDesc = ListAPIKeysRequestOrderBy("created_at_desc")
	// ListAPIKeysRequestOrderByUpdatedAtAsc is [insert doc].
	ListAPIKeysRequestOrderByUpdatedAtAsc = ListAPIKeysRequestOrderBy("updated_at_asc")
	// ListAPIKeysRequestOrderByUpdatedAtDesc is [insert doc].
	ListAPIKeysRequestOrderByUpdatedAtDesc = ListAPIKeysRequestOrderBy("updated_at_desc")
	// ListAPIKeysRequestOrderByExpiresAtAsc is [insert doc].
	ListAPIKeysRequestOrderByExpiresAtAsc = ListAPIKeysRequestOrderBy("expires_at_asc")
	// ListAPIKeysRequestOrderByExpiresAtDesc is [insert doc].
	ListAPIKeysRequestOrderByExpiresAtDesc = ListAPIKeysRequestOrderBy("expires_at_desc")
	// ListAPIKeysRequestOrderByAccessKeyAsc is [insert doc].
	ListAPIKeysRequestOrderByAccessKeyAsc = ListAPIKeysRequestOrderBy("access_key_asc")
	// ListAPIKeysRequestOrderByAccessKeyDesc is [insert doc].
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
	// ListApplicationsRequestOrderByCreatedAtAsc is [insert doc].
	ListApplicationsRequestOrderByCreatedAtAsc = ListApplicationsRequestOrderBy("created_at_asc")
	// ListApplicationsRequestOrderByCreatedAtDesc is [insert doc].
	ListApplicationsRequestOrderByCreatedAtDesc = ListApplicationsRequestOrderBy("created_at_desc")
	// ListApplicationsRequestOrderByUpdatedAtAsc is [insert doc].
	ListApplicationsRequestOrderByUpdatedAtAsc = ListApplicationsRequestOrderBy("updated_at_asc")
	// ListApplicationsRequestOrderByUpdatedAtDesc is [insert doc].
	ListApplicationsRequestOrderByUpdatedAtDesc = ListApplicationsRequestOrderBy("updated_at_desc")
	// ListApplicationsRequestOrderByNameAsc is [insert doc].
	ListApplicationsRequestOrderByNameAsc = ListApplicationsRequestOrderBy("name_asc")
	// ListApplicationsRequestOrderByNameDesc is [insert doc].
	ListApplicationsRequestOrderByNameDesc = ListApplicationsRequestOrderBy("name_desc")
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
	// ListGroupsRequestOrderByCreatedAtAsc is [insert doc].
	ListGroupsRequestOrderByCreatedAtAsc = ListGroupsRequestOrderBy("created_at_asc")
	// ListGroupsRequestOrderByCreatedAtDesc is [insert doc].
	ListGroupsRequestOrderByCreatedAtDesc = ListGroupsRequestOrderBy("created_at_desc")
	// ListGroupsRequestOrderByUpdatedAtAsc is [insert doc].
	ListGroupsRequestOrderByUpdatedAtAsc = ListGroupsRequestOrderBy("updated_at_asc")
	// ListGroupsRequestOrderByUpdatedAtDesc is [insert doc].
	ListGroupsRequestOrderByUpdatedAtDesc = ListGroupsRequestOrderBy("updated_at_desc")
	// ListGroupsRequestOrderByNameAsc is [insert doc].
	ListGroupsRequestOrderByNameAsc = ListGroupsRequestOrderBy("name_asc")
	// ListGroupsRequestOrderByNameDesc is [insert doc].
	ListGroupsRequestOrderByNameDesc = ListGroupsRequestOrderBy("name_desc")
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

type ListPermissionSetsRequestOrderBy string

const (
	// ListPermissionSetsRequestOrderByNameAsc is [insert doc].
	ListPermissionSetsRequestOrderByNameAsc = ListPermissionSetsRequestOrderBy("name_asc")
	// ListPermissionSetsRequestOrderByNameDesc is [insert doc].
	ListPermissionSetsRequestOrderByNameDesc = ListPermissionSetsRequestOrderBy("name_desc")
	// ListPermissionSetsRequestOrderByCreatedAtAsc is [insert doc].
	ListPermissionSetsRequestOrderByCreatedAtAsc = ListPermissionSetsRequestOrderBy("created_at_asc")
	// ListPermissionSetsRequestOrderByCreatedAtDesc is [insert doc].
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
	// ListPoliciesRequestOrderByPolicyNameAsc is [insert doc].
	ListPoliciesRequestOrderByPolicyNameAsc = ListPoliciesRequestOrderBy("policy_name_asc")
	// ListPoliciesRequestOrderByPolicyNameDesc is [insert doc].
	ListPoliciesRequestOrderByPolicyNameDesc = ListPoliciesRequestOrderBy("policy_name_desc")
	// ListPoliciesRequestOrderByCreatedAtAsc is [insert doc].
	ListPoliciesRequestOrderByCreatedAtAsc = ListPoliciesRequestOrderBy("created_at_asc")
	// ListPoliciesRequestOrderByCreatedAtDesc is [insert doc].
	ListPoliciesRequestOrderByCreatedAtDesc = ListPoliciesRequestOrderBy("created_at_desc")
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

type ListSSHKeysRequestOrderBy string

const (
	// ListSSHKeysRequestOrderByCreatedAtAsc is [insert doc].
	ListSSHKeysRequestOrderByCreatedAtAsc = ListSSHKeysRequestOrderBy("created_at_asc")
	// ListSSHKeysRequestOrderByCreatedAtDesc is [insert doc].
	ListSSHKeysRequestOrderByCreatedAtDesc = ListSSHKeysRequestOrderBy("created_at_desc")
	// ListSSHKeysRequestOrderByUpdatedAtAsc is [insert doc].
	ListSSHKeysRequestOrderByUpdatedAtAsc = ListSSHKeysRequestOrderBy("updated_at_asc")
	// ListSSHKeysRequestOrderByUpdatedAtDesc is [insert doc].
	ListSSHKeysRequestOrderByUpdatedAtDesc = ListSSHKeysRequestOrderBy("updated_at_desc")
	// ListSSHKeysRequestOrderByNameAsc is [insert doc].
	ListSSHKeysRequestOrderByNameAsc = ListSSHKeysRequestOrderBy("name_asc")
	// ListSSHKeysRequestOrderByNameDesc is [insert doc].
	ListSSHKeysRequestOrderByNameDesc = ListSSHKeysRequestOrderBy("name_desc")
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
	// ListUsersRequestOrderByCreatedAtAsc is [insert doc].
	ListUsersRequestOrderByCreatedAtAsc = ListUsersRequestOrderBy("created_at_asc")
	// ListUsersRequestOrderByCreatedAtDesc is [insert doc].
	ListUsersRequestOrderByCreatedAtDesc = ListUsersRequestOrderBy("created_at_desc")
	// ListUsersRequestOrderByUpdatedAtAsc is [insert doc].
	ListUsersRequestOrderByUpdatedAtAsc = ListUsersRequestOrderBy("updated_at_asc")
	// ListUsersRequestOrderByUpdatedAtDesc is [insert doc].
	ListUsersRequestOrderByUpdatedAtDesc = ListUsersRequestOrderBy("updated_at_desc")
	// ListUsersRequestOrderByEmailAsc is [insert doc].
	ListUsersRequestOrderByEmailAsc = ListUsersRequestOrderBy("email_asc")
	// ListUsersRequestOrderByEmailDesc is [insert doc].
	ListUsersRequestOrderByEmailDesc = ListUsersRequestOrderBy("email_desc")
	// ListUsersRequestOrderByLastLoginAsc is [insert doc].
	ListUsersRequestOrderByLastLoginAsc = ListUsersRequestOrderBy("last_login_asc")
	// ListUsersRequestOrderByLastLoginDesc is [insert doc].
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
	// PermissionSetScopeTypeUnknownScopeType is [insert doc].
	PermissionSetScopeTypeUnknownScopeType = PermissionSetScopeType("unknown_scope_type")
	// PermissionSetScopeTypeProjects is [insert doc].
	PermissionSetScopeTypeProjects = PermissionSetScopeType("projects")
	// PermissionSetScopeTypeOrganization is [insert doc].
	PermissionSetScopeTypeOrganization = PermissionSetScopeType("organization")
	// PermissionSetScopeTypeAccountRootUser is [insert doc].
	PermissionSetScopeTypeAccountRootUser = PermissionSetScopeType("account_root_user")
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
	// UserStatusUnknownStatus is [insert doc].
	UserStatusUnknownStatus = UserStatus("unknown_status")
	// UserStatusInvitationPending is [insert doc].
	UserStatusInvitationPending = UserStatus("invitation_pending")
	// UserStatusActivated is [insert doc].
	UserStatusActivated = UserStatus("activated")
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
	// UserTypeUnknownType is [insert doc].
	UserTypeUnknownType = UserType("unknown_type")
	// UserTypeGuest is [insert doc].
	UserTypeGuest = UserType("guest")
	// UserTypeOwner is [insert doc].
	UserTypeOwner = UserType("owner")
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

// APIKey: api key
type APIKey struct {
	// AccessKey: access key of API key
	AccessKey string `json:"access_key"`
	// SecretKey: secret key of API Key
	SecretKey *string `json:"secret_key"`
	// ApplicationID: ID of application bearer
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// UserID: ID of user bearer
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// Description: description of API key
	Description string `json:"description"`
	// CreatedAt: creation date and time of API key
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: last update date and time of API key
	UpdatedAt *time.Time `json:"updated_at"`
	// ExpiresAt: expiration date and time of API key
	ExpiresAt *time.Time `json:"expires_at"`
	// DefaultProjectID: the default project ID specified for this API key
	DefaultProjectID string `json:"default_project_id"`
	// Editable: whether or not the API key is editable
	Editable bool `json:"editable"`
	// CreationIP: IP Address of the device which created the API key
	CreationIP string `json:"creation_ip"`
}

// Application: application
type Application struct {
	// ID: ID of application
	ID string `json:"id"`
	// Name: name of application
	Name string `json:"name"`
	// Description: description of the application
	Description string `json:"description"`
	// CreatedAt: creation date of application
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: last update date of application
	UpdatedAt *time.Time `json:"updated_at"`
	// OrganizationID: ID of organization
	OrganizationID string `json:"organization_id"`
	// Editable: whether or not the application is editable
	Editable bool `json:"editable"`
	// NbAPIKeys: number of API keys owned by the application
	NbAPIKeys uint32 `json:"nb_api_keys"`
}

// Group: group
type Group struct {
	// ID: ID of group
	ID string `json:"id"`
	// CreatedAt: creation date and time of group
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: last update date and time of group
	UpdatedAt *time.Time `json:"updated_at"`
	// OrganizationID: ID of organization linked to the group
	OrganizationID string `json:"organization_id"`
	// Name: name of group
	Name string `json:"name"`
	// Description: description of the group
	Description string `json:"description"`
	// UserIDs: iDs of users attached to this group
	UserIDs []string `json:"user_ids"`
	// ApplicationIDs: iDs of applications attached to this group
	ApplicationIDs []string `json:"application_ids"`
}

// ListAPIKeysResponse: list api keys response
type ListAPIKeysResponse struct {
	// APIKeys: list of API keys
	APIKeys []*APIKey `json:"api_keys"`
	// TotalCount: total count of API Keys
	TotalCount uint32 `json:"total_count"`
}

// ListApplicationsResponse: list applications response
type ListApplicationsResponse struct {
	// Applications: list of applications
	Applications []*Application `json:"applications"`
	// TotalCount: total count of applications
	TotalCount uint32 `json:"total_count"`
}

// ListGroupsResponse: list groups response
type ListGroupsResponse struct {
	// Groups: list of groups
	Groups []*Group `json:"groups"`
	// TotalCount: total count of groups
	TotalCount uint32 `json:"total_count"`
}

// ListPermissionSetsResponse: list permission sets response
type ListPermissionSetsResponse struct {
	// PermissionSets: list of permission sets
	PermissionSets []*PermissionSet `json:"permission_sets"`
	// TotalCount: total count of permission sets
	TotalCount uint32 `json:"total_count"`
}

// ListPoliciesResponse: list policies response
type ListPoliciesResponse struct {
	// Policies: list of policies
	Policies []*Policy `json:"policies"`
	// TotalCount: total count of policies
	TotalCount uint32 `json:"total_count"`
}

// ListRulesResponse: list rules response
type ListRulesResponse struct {
	// Rules: rules of the policy
	Rules []*Rule `json:"rules"`
	// TotalCount: total count of rules
	TotalCount uint32 `json:"total_count"`
}

// ListSSHKeysResponse: list ssh keys response
type ListSSHKeysResponse struct {
	// SSHKeys: list of SSH keys
	SSHKeys []*SSHKey `json:"ssh_keys"`
	// TotalCount: total count of SSH keys
	TotalCount uint32 `json:"total_count"`
}

// ListUsersResponse: list users response
type ListUsersResponse struct {
	// Users: list of users
	Users []*User `json:"users"`
	// TotalCount: total count of users
	TotalCount uint32 `json:"total_count"`
}

// PermissionSet: permission set
type PermissionSet struct {
	// ID: id of permission set
	ID string `json:"id"`
	// Name: name of permission set
	Name string `json:"name"`
	// ScopeType: scope of permission set
	//
	// Default value: unknown_scope_type
	ScopeType PermissionSetScopeType `json:"scope_type"`
	// Description: description of permission set
	Description string `json:"description"`
	// Categories: categories of permission set
	Categories *[]string `json:"categories"`
}

// Policy: policy
type Policy struct {
	// ID: id of policy
	ID string `json:"id"`
	// Name: name of policy
	Name string `json:"name"`
	// Description: description of policy
	Description string `json:"description"`
	// OrganizationID: organization ID of policy
	OrganizationID string `json:"organization_id"`
	// CreatedAt: creation date and time of policy
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: last update date and time of policy
	UpdatedAt *time.Time `json:"updated_at"`
	// Editable: editable status of policy
	Editable bool `json:"editable"`
	// NbRules: number of rules of policy
	NbRules uint32 `json:"nb_rules"`
	// NbScopes: number of scopes of policy
	NbScopes uint32 `json:"nb_scopes"`
	// NbPermissionSets: number of permission sets of policy
	NbPermissionSets uint32 `json:"nb_permission_sets"`
	// UserID: ID of user, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// GroupID: ID of group, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	GroupID *string `json:"group_id,omitempty"`
	// ApplicationID: ID of application, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// NoPrincipal: true when the policy do not belong to any principal
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// Rule: rule
type Rule struct {
	// ID: id of rule
	ID string `json:"id"`
	// PermissionSetNames: names of permission sets bound to the rule
	PermissionSetNames *[]string `json:"permission_set_names"`
	// PermissionSetsScopeType: permission_set_names have the same scope_type
	//
	// Default value: unknown_scope_type
	PermissionSetsScopeType PermissionSetScopeType `json:"permission_sets_scope_type"`
	// ProjectIDs: list of project IDs scoped to the rule
	// Precisely one of AccountRootUserID, OrganizationID, ProjectIDs must be set.
	ProjectIDs *[]string `json:"project_ids,omitempty"`
	// OrganizationID: ID of organization scoped to the rule
	// Precisely one of AccountRootUserID, OrganizationID, ProjectIDs must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// AccountRootUserID: ID of account root user scoped to the rule
	// Precisely one of AccountRootUserID, OrganizationID, ProjectIDs must be set.
	AccountRootUserID *string `json:"account_root_user_id,omitempty"`
}

// RuleSpecs: rule specs
type RuleSpecs struct {
	// PermissionSetNames: names of permission sets bound to the rule
	PermissionSetNames *[]string `json:"permission_set_names"`
	// ProjectIDs: list of project IDs scoped to the rule
	// Precisely one of OrganizationID, ProjectIDs must be set.
	ProjectIDs *[]string `json:"project_ids,omitempty"`
	// OrganizationID: ID of organization scoped to the rule
	// Precisely one of OrganizationID, ProjectIDs must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
}

// SSHKey: ssh key
type SSHKey struct {
	// ID: ID of SSH key
	ID string `json:"id"`
	// Name: name of SSH key
	Name string `json:"name"`
	// PublicKey: public key of SSH key
	PublicKey string `json:"public_key"`
	// Fingerprint: fingerprint of SSH key
	Fingerprint string `json:"fingerprint"`
	// CreatedAt: creation date of SSH key
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: last update date of SSH key
	UpdatedAt *time.Time `json:"updated_at"`
	// OrganizationID: ID of organization linked to the SSH key
	OrganizationID string `json:"organization_id"`
	// ProjectID: ID of project linked to the SSH key
	ProjectID string `json:"project_id"`
	// Disabled: SSH key status
	Disabled bool `json:"disabled"`
}

// SetRulesResponse: set rules response
type SetRulesResponse struct {
	// Rules: rules of policy
	Rules []*Rule `json:"rules"`
}

// User: user
type User struct {
	// ID: ID of user
	ID string `json:"id"`
	// Email: email of user
	Email string `json:"email"`
	// CreatedAt: creation date of user
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: last update date of user
	UpdatedAt *time.Time `json:"updated_at"`
	// OrganizationID: ID of organization
	OrganizationID string `json:"organization_id"`
	// Deletable: deletion status of user. Owner user cannot be deleted
	Deletable bool `json:"deletable"`
	// LastLoginAt: last login date
	LastLoginAt *time.Time `json:"last_login_at"`
	// Type: type of the user
	//
	// Default value: unknown_type
	Type UserType `json:"type"`
	// TwoFactorEnabled: 2FA enabled
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	// Status: status of invitation for the user
	//
	// Default value: unknown_status
	Status UserStatus `json:"status"`
}

// Service API

type ListSSHKeysRequest struct {
	// OrderBy: sort order of SSH keys
	//
	// Default value: created_at_asc
	OrderBy ListSSHKeysRequestOrderBy `json:"-"`
	// Page: requested page number. Value must be greater or equals to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// PageSize: number of items per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// OrganizationID: filter by organization ID
	OrganizationID *string `json:"-"`
	// Name: name of group to find
	Name *string `json:"-"`
	// ProjectID: filter by project ID
	ProjectID *string `json:"-"`
	// Disabled: filter out disabled SSH keys or not
	Disabled *bool `json:"-"`
}

// ListSSHKeys: list SSH keys
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
	// Name: the name of the SSH key. Max length is 1000
	Name string `json:"name"`
	// PublicKey: SSH public key. Currently ssh-rsa, ssh-dss (DSA), ssh-ed25519 and ecdsa keys with NIST curves are supported. Max length is 65000
	PublicKey string `json:"public_key"`
	// ProjectID: project owning the resource
	ProjectID string `json:"project_id"`
}

// CreateSSHKey: create an SSH key
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
	// SSHKeyID: the ID of the SSH key
	SSHKeyID string `json:"-"`
}

// GetSSHKey: get an SSH key
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
	// Name: name of the SSH key. Max length is 1000
	Name *string `json:"name"`
	// Disabled: enable or disable the SSH key
	Disabled *bool `json:"disabled"`
}

// UpdateSSHKey: update an SSH key
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

// DeleteSSHKey: delete an SSH key
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
	// OrderBy: criteria for sorting results
	//
	// Default value: created_at_asc
	OrderBy ListUsersRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: number of page. Value must be greater or equals to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: ID of organization to filter
	OrganizationID *string `json:"-"`
	// UserIDs: filter out by a list of ID
	UserIDs []string `json:"-"`
}

// ListUsers: list users of an organization
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
	// UserID: ID of user to find
	UserID string `json:"-"`
}

// GetUser: retrieve a user from its ID
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
	// UserID: ID of user to delete
	UserID string `json:"-"`
}

// DeleteUser: delete a guest user from an organization
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
	// OrderBy: criteria for sorting results
	//
	// Default value: created_at_asc
	OrderBy ListApplicationsRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: number of page. Value must be greater to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// Name: name of application to filter
	Name *string `json:"-"`
	// OrganizationID: ID of organization to filter
	OrganizationID *string `json:"-"`
	// Editable: filter out editable applications or not
	Editable *bool `json:"-"`
	// ApplicationIDs: filter out by a list of ID
	ApplicationIDs []string `json:"-"`
}

// ListApplications: list applications of an organization
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
	// Name: name of application to create (max length is 64 chars)
	Name string `json:"name"`
	// OrganizationID: ID of organization
	OrganizationID string `json:"organization_id"`
	// Description: description of application (max length is 200 chars)
	Description string `json:"description"`
}

// CreateApplication: create a new application
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
	// ApplicationID: ID of application to find
	ApplicationID string `json:"-"`
}

// GetApplication: get an existing application
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
	// ApplicationID: ID of application to update
	ApplicationID string `json:"-"`
	// Name: new name of application (max length is 64 chars)
	Name *string `json:"name"`
	// Description: new description of application (max length is 200 chars)
	Description *string `json:"description"`
}

// UpdateApplication: update an existing application
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
	// ApplicationID: ID of application to delete
	ApplicationID string `json:"-"`
}

// DeleteApplication: delete an application
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
	// OrderBy: sort order of groups
	//
	// Default value: created_at_asc
	OrderBy ListGroupsRequestOrderBy `json:"-"`
	// Page: requested page number. Value must be greater or equals to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// PageSize: number of items per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// OrganizationID: filter by organization ID
	OrganizationID *string `json:"-"`
	// Name: name of group to find
	Name *string `json:"-"`
	// ApplicationIDs: filter out by a list of application ID
	ApplicationIDs []string `json:"-"`
	// UserIDs: filter out by a list of user ID
	UserIDs []string `json:"-"`
	// GroupIDs: filter out by a list of group ID
	GroupIDs []string `json:"-"`
}

// ListGroups: list groups
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
	// OrganizationID: ID of organization linked to the group
	OrganizationID string `json:"organization_id"`
	// Name: name of the group to create (max length is 64 chars). MUST be unique inside an organization
	Name string `json:"name"`
	// Description: description of the group to create (max length is 200 chars)
	Description string `json:"description"`
}

// CreateGroup: create a new group
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
	// GroupID: ID of group
	GroupID string `json:"-"`
}

// GetGroup: get a group
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
	// GroupID: ID of group to update
	GroupID string `json:"-"`
	// Name: new name for the group (max length is 64 chars). MUST be unique inside an organization
	Name *string `json:"name"`
	// Description: new description for the group (max length is 200 chars)
	Description *string `json:"description"`
}

// UpdateGroup: update a group
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

// SetGroupMembers: overwrite users and applications of a group
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
	// GroupID: ID of group
	GroupID string `json:"-"`
	// UserID: ID of the user to add
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// ApplicationID: ID of the application to add
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
}

// AddGroupMember: add a user of an application to a group
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
	// GroupID: ID of group
	GroupID string `json:"-"`
	// UserID: ID of the user to remove
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// ApplicationID: ID of the application to remove
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
}

// RemoveGroupMember: remove a user or an application from a group
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
	// GroupID: ID of group to delete
	GroupID string `json:"-"`
}

// DeleteGroup: delete a group
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
	// OrderBy: criteria for sorting results
	//
	// Default value: created_at_asc
	OrderBy ListPoliciesRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: number of page. Value must be greater to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: ID of organization to filter
	OrganizationID *string `json:"-"`
	// Editable: filter out editable policies or not
	Editable *bool `json:"-"`
	// UserIDs: filter out by a list of user ID
	UserIDs []string `json:"-"`
	// GroupIDs: filter out by a list of group ID
	GroupIDs []string `json:"-"`
	// ApplicationIDs: filter out by a list of application ID
	ApplicationIDs []string `json:"-"`
	// NoPrincipal: true when the policy do not belong to any principal
	NoPrincipal *bool `json:"-"`
	// PolicyName: name of policy to fetch
	PolicyName *string `json:"-"`
}

// ListPolicies: list policies of an organization
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
	// Name: name of policy to create (max length is 64 chars)
	Name string `json:"name"`
	// Description: description of policy to create (max length is 200 chars)
	Description string `json:"description"`
	// OrganizationID: ID of organization
	OrganizationID string `json:"organization_id"`
	// Rules: rules of the policy to create
	Rules []*RuleSpecs `json:"rules"`
	// UserID: ID of user, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// GroupID: ID of group, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	GroupID *string `json:"group_id,omitempty"`
	// ApplicationID: ID of application, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// NoPrincipal: true when the policy do not belong to any principal
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// CreatePolicy: create a new policy
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
	// PolicyID: id of policy to search
	PolicyID string `json:"-"`
}

// GetPolicy: get an existing policy
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
	// PolicyID: id of policy to update
	PolicyID string `json:"-"`
	// Name: new name of policy (max length is 64 chars)
	Name *string `json:"name"`
	// Description: new description of policy (max length is 200 chars)
	Description *string `json:"description"`
	// UserID: new ID of user, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// GroupID: new ID of group, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	GroupID *string `json:"group_id,omitempty"`
	// ApplicationID: new ID of application, owner of the policy
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// NoPrincipal: true when the policy do not belong to any principal
	// Precisely one of ApplicationID, GroupID, NoPrincipal, UserID must be set.
	NoPrincipal *bool `json:"no_principal,omitempty"`
}

// UpdatePolicy: update an existing policy
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
	// PolicyID: id of policy to delete
	PolicyID string `json:"-"`
}

// DeletePolicy: delete a policy
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
	// PolicyID: id of policy to update
	PolicyID string `json:"policy_id"`
	// Rules: rules of the policy to set
	Rules []*RuleSpecs `json:"rules"`
}

// SetRules: set rules of an existing policy
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
	// PolicyID: id of policy to search
	PolicyID *string `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: number of page. Value must be greater to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
}

// ListRules: list rules of an existing policy
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
	// OrderBy: criteria for sorting results
	//
	// Default value: created_at_asc
	OrderBy ListPermissionSetsRequestOrderBy `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// Page: number of page. Value must be greater to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// OrganizationID: filter by organization ID
	OrganizationID string `json:"-"`
}

// ListPermissionSets: list permission sets
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
	// OrderBy: criteria for sorting results
	//
	// Default value: created_at_asc
	OrderBy ListAPIKeysRequestOrderBy `json:"-"`
	// Page: number of page. Value must be greater or equals to 1
	//
	// Default value: 1
	Page *int32 `json:"-"`
	// PageSize: number of results per page. Value must be between 1 and 100
	//
	// Default value: 20
	PageSize *uint32 `json:"-"`
	// OrganizationID: ID of organization
	OrganizationID *string `json:"-"`
	// ApplicationID: ID of an application bearer
	ApplicationID *string `json:"-"`
	// UserID: ID of a user bearer
	UserID *string `json:"-"`
	// Editable: filter out editable API keys or not
	Editable *bool `json:"-"`
}

// ListAPIKeys: list API keys
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
	// ApplicationID: ID of application principal
	// Precisely one of ApplicationID, UserID must be set.
	ApplicationID *string `json:"application_id,omitempty"`
	// UserID: ID of user principal
	// Precisely one of ApplicationID, UserID must be set.
	UserID *string `json:"user_id,omitempty"`
	// ExpiresAt: expiration date of the API key
	ExpiresAt *time.Time `json:"expires_at"`
	// DefaultProjectID: the default project ID to use with object storage
	DefaultProjectID *string `json:"default_project_id"`
	// Description: the description of the API key (max length is 200 chars)
	Description string `json:"description"`
}

// CreateAPIKey: create an API key
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
	// AccessKey: access key to search for
	AccessKey string `json:"-"`
}

// GetAPIKey: get an API key
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
	// AccessKey: access key to update
	AccessKey string `json:"-"`
	// DefaultProjectID: the new default project ID to set
	DefaultProjectID *string `json:"default_project_id"`
	// Description: the new description to update
	Description *string `json:"description"`
}

// UpdateAPIKey: update an API key
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
	// AccessKey: access key to delete
	AccessKey string `json:"-"`
}

// DeleteAPIKey: delete an API key
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
