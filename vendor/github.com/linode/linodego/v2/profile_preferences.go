package linodego

import (
	"context"
	"encoding/json"
)

// ProfilePreferences represents the user's preferences.
// The user preferences endpoints allow consumers of the API to store arbitrary JSON data,
// such as a user's font size preference or preferred display name.
type ProfilePreferences map[string]any

// UnmarshalJSON implements the json.Unmarshaler interface
func (p *ProfilePreferences) UnmarshalJSON(b []byte) error {
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	*p = data

	return nil
}

// GetProfilePreferences retrieves the user preferences for the current User
func (c *Client) GetProfilePreferences(ctx context.Context) (*ProfilePreferences, error) {
	return doGETRequest[ProfilePreferences](ctx, c, "profile/preferences")
}

// UpdateProfilePreferences updates the user's preferences with the provided data
func (c *Client) UpdateProfilePreferences(ctx context.Context, opts ProfilePreferences) (*ProfilePreferences, error) {
	return doPUTRequest[ProfilePreferences](ctx, c, "profile/preferences", opts)
}
