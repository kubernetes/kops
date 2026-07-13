package linodego

import "context"

// UserRolePermissions are the account and entity permissions for the User
type UserRolePermissions struct {
	AccountAccess []string     `json:"account_access"`
	EntityAccess  []UserAccess `json:"entity_access"`
}

// GetUpdateOptions converts UserRolePermissions for use in UpdateUserRolePermissions
func (p *UserRolePermissions) GetUpdateOptions() UserRolePermissionsUpdateOptions {
	return UserRolePermissionsUpdateOptions{
		AccountAccess: p.AccountAccess,
		EntityAccess:  p.EntityAccess,
	}
}

// UserRolePermissionsUpdateOptions are fields accepted by UpdateUserRolePermissions
type UserRolePermissionsUpdateOptions struct {
	AccountAccess []string     `json:"account_access"`
	EntityAccess  []UserAccess `json:"entity_access"`
}

// UserAccess is the breakdown of entities Roles
type UserAccess struct {
	ID    int      `json:"id"`
	Type  string   `json:"type"`
	Roles []string `json:"roles"`
}

// AccountRolePermissions are the account and entity roles for the Account
type AccountRolePermissions struct {
	AccountAccess []AccountAccess `json:"account_access"`
	EntityAccess  []AccountAccess `json:"entity_access"`
}

// AccountAccess is the Roles for each Type for the Account
type AccountAccess struct {
	Type  string `json:"type"`
	Roles []Role `json:"roles"`
}

// Role is the IAM Role and its Permissions
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// GetUserRolePermissions returns any role permissions for username
func (c *Client) GetUserRolePermissions(ctx context.Context, username string) (*UserRolePermissions, error) {
	return doGETRequest[UserRolePermissions](ctx, c,
		formatAPIPath("iam/users/%s/role-permissions", username),
	)
}

// UpdateUserRolePermissions updates any role permissions for username
func (c *Client) UpdateUserRolePermissions(ctx context.Context, username string, opts UserRolePermissionsUpdateOptions) (*UserRolePermissions, error) {
	return doPUTRequest[UserRolePermissions](ctx, c,
		formatAPIPath("iam/users/%s/role-permissions", username),
		opts,
	)
}

// GetAccountRolePermissions returns the role permissions for this Account
func (c *Client) GetAccountRolePermissions(ctx context.Context) (*AccountRolePermissions, error) {
	return doGETRequest[AccountRolePermissions](ctx, c, "iam/role-permissions")
}

// GetUserAccountPermissions returns the account permissions for username
func (c *Client) GetUserAccountPermissions(ctx context.Context, username string) ([]string, error) {
	perms, err := doGETRequest[[]string](ctx, c,
		formatAPIPath("iam/users/%s/permissions/account", username))
	if err != nil || perms == nil {
		return nil, err
	}

	return (*perms), err
}
