package linodego

import (
	"context"
)

// LishAuthMethod constants start with AuthMethod and include Linode API Lish Authentication Methods
type LishAuthMethod string

// LishAuthMethod constants are the methods of authentication allowed when connecting via Lish
const (
	AuthMethodPasswordKeys LishAuthMethod = "password_keys"
	AuthMethodKeysOnly     LishAuthMethod = "keys_only"
	AuthMethodDisabled     LishAuthMethod = "disabled"
)

// ProfileReferrals represent a User's status in the Referral Program
type ProfileReferrals struct {
	Total     int     `json:"total"`
	Completed int     `json:"completed"`
	Pending   int     `json:"pending"`
	Credit    float64 `json:"credit"`
	Code      string  `json:"code"`
	URL       string  `json:"url"`
}

// Profile represents a Profile object
type Profile struct {
	UID                 int              `json:"uid"`
	Username            string           `json:"username"`
	Email               string           `json:"email"`
	Timezone            string           `json:"timezone"`
	EmailNotifications  bool             `json:"email_notifications"`
	IPWhitelistEnabled  bool             `json:"ip_whitelist_enabled"`
	TwoFactorAuth       bool             `json:"two_factor_auth"`
	Restricted          bool             `json:"restricted"`
	LishAuthMethod      LishAuthMethod   `json:"lish_auth_method"`
	Referrals           ProfileReferrals `json:"referrals"`
	AuthorizedKeys      []string         `json:"authorized_keys"`
	AuthenticationType  string           `json:"authentication_type"`
	VerifiedPhoneNumber string           `json:"verified_phone_number,omitzero"`
}

// ProfileUpdateOptions fields are those accepted by UpdateProfile
type ProfileUpdateOptions struct {
	Email              string         `json:"email,omitzero"`
	Timezone           string         `json:"timezone,omitzero"`
	EmailNotifications *bool          `json:"email_notifications,omitzero"`
	IPWhitelistEnabled *bool          `json:"ip_whitelist_enabled,omitzero"`
	LishAuthMethod     LishAuthMethod `json:"lish_auth_method,omitzero"`
	AuthorizedKeys     []string       `json:"authorized_keys,omitzero"`
	TwoFactorAuth      *bool          `json:"two_factor_auth,omitzero"`
	Restricted         *bool          `json:"restricted,omitzero"`
}

// GetUpdateOptions converts a Profile to ProfileUpdateOptions for use in UpdateProfile
func (i Profile) GetUpdateOptions() (o ProfileUpdateOptions) {
	o.Email = i.Email
	o.Timezone = i.Timezone
	o.EmailNotifications = copyBool(&i.EmailNotifications)
	o.IPWhitelistEnabled = copyBool(&i.IPWhitelistEnabled)
	o.LishAuthMethod = i.LishAuthMethod
	authorizedKeys := make([]string, len(i.AuthorizedKeys))
	copy(authorizedKeys, i.AuthorizedKeys)
	o.AuthorizedKeys = authorizedKeys
	o.TwoFactorAuth = copyBool(&i.TwoFactorAuth)
	o.Restricted = copyBool(&i.Restricted)

	return o
}

// GetProfile returns the Profile of the authenticated user
func (c *Client) GetProfile(ctx context.Context) (*Profile, error) {
	return doGETRequest[Profile](ctx, c, "profile")
}

// UpdateProfile updates the Profile with the specified id
func (c *Client) UpdateProfile(ctx context.Context, opts ProfileUpdateOptions) (*Profile, error) {
	return doPUTRequest[Profile](ctx, c, "profile", opts)
}
