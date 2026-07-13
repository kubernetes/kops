package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// ProfileApp represents a ProfileApp object
type ProfileApp struct {
	// When this app was authorized.
	Created *time.Time `json:"-"`

	// When the app's access to your account expires.
	Expiry *time.Time `json:"-"`

	// This authorization's ID, used for revoking access.
	ID int `json:"id"`

	// The name of the application you've authorized.
	Label string `json:"label"`

	// The OAuth scopes this app was authorized with.
	Scopes string `json:"scopes"`

	// The URL at which this app's thumbnail may be accessed.
	ThumbnailURL string `json:"thumbnail_url"`

	// The website where you can get more information about this app.
	Website string `json:"website"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (pa *ProfileApp) UnmarshalJSON(b []byte) error {
	type Mask ProfileApp

	l := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
	}{
		Mask: (*Mask)(pa),
	}

	if err := json.Unmarshal(b, &l); err != nil {
		return err
	}

	pa.Created = (*time.Time)(l.Created)
	pa.Expiry = (*time.Time)(l.Expiry)

	return nil
}

// GetProfileApp returns the ProfileApp with the provided id
func (c *Client) GetProfileApp(ctx context.Context, appID int) (*ProfileApp, error) {
	e := formatAPIPath("profile/apps/%d", appID)
	return doGETRequest[ProfileApp](ctx, c, e)
}

// ListProfileApps lists ProfileApps that have access to the Account
func (c *Client) ListProfileApps(ctx context.Context, opts *ListOptions) ([]ProfileApp, error) {
	return getPaginatedResults[ProfileApp](ctx, c, "profile/apps", opts)
}

// DeleteProfileApp revokes the given ProfileApp's access to the account
func (c *Client) DeleteProfileApp(ctx context.Context, appID int) error {
	e := formatAPIPath("profile/apps/%d", appID)
	return doDELETERequest(ctx, c, e)
}
