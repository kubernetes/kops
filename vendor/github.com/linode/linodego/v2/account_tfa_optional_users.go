package linodego

import "context"

// TFAOptionalUser represents an account user exempted from TFA enforcement.
type TFAOptionalUser struct {
	Username string `json:"username"`
	URL      string `json:"url"`
}

// TFAOptionalUsersUpdateOptions contains the replacement list of exempted users.
type TFAOptionalUsersUpdateOptions struct {
	Usernames []string `json:"usernames"`
}

// TFAOptionalUsersUpdateResponse summarizes the updated TFA exemption list.
type TFAOptionalUsersUpdateResponse struct {
	TFAOptionalUsersCount int    `json:"tfa_optional_users_count"`
	URL                   string `json:"url"`
}

// ListTFAOptionalUsers lists exempted users regardless of the account's TFA policy.
func (c *Client) ListTFAOptionalUsers(ctx context.Context, opts *ListOptions) ([]TFAOptionalUser, error) {
	return getPaginatedResults[TFAOptionalUser](ctx, c, "account/tfa-optional-users", opts)
}

// UpdateTFAOptionalUsers replaces the exemption list. An empty slice clears it.
func (c *Client) UpdateTFAOptionalUsers(ctx context.Context, opts TFAOptionalUsersUpdateOptions) (*TFAOptionalUsersUpdateResponse, error) {
	return doPUTRequest[TFAOptionalUsersUpdateResponse](ctx, c, "account/tfa-optional-users", opts)
}
