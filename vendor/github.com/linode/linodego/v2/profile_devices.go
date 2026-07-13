package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// ProfileDevice represents a ProfileDevice object
type ProfileDevice struct {
	// When this Remember Me session was started.
	Created *time.Time `json:"-"`

	// When this TrustedDevice session expires. Sessions typically last 30 days.
	Expiry *time.Time `json:"-"`

	// The unique ID for this TrustedDevice.
	ID int `json:"id"`

	// he last time this TrustedDevice was successfully used to authenticate to login.linode.com
	LastAuthenticated *time.Time `json:"-"`

	// The last IP Address to successfully authenticate with this TrustedDevice.
	LastRemoteAddr string `json:"last_remote_addr"`

	// The User Agent of the browser that created this TrustedDevice session.
	UserAgent string `json:"user_agent"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (pd *ProfileDevice) UnmarshalJSON(b []byte) error {
	type Mask ProfileDevice

	l := struct {
		*Mask

		Created           *parseabletime.ParseableTime `json:"created"`
		Expiry            *parseabletime.ParseableTime `json:"expiry"`
		LastAuthenticated *parseabletime.ParseableTime `json:"last_authenticated"`
	}{
		Mask: (*Mask)(pd),
	}

	if err := json.Unmarshal(b, &l); err != nil {
		return err
	}

	pd.Created = (*time.Time)(l.Created)
	pd.Expiry = (*time.Time)(l.Expiry)
	pd.LastAuthenticated = (*time.Time)(l.LastAuthenticated)

	return nil
}

// GetProfileDevice returns the ProfileDevice with the provided id
func (c *Client) GetProfileDevice(ctx context.Context, deviceID int) (*ProfileDevice, error) {
	e := formatAPIPath("profile/devices/%d", deviceID)
	return doGETRequest[ProfileDevice](ctx, c, e)
}

// ListProfileDevices lists ProfileDevices for the User
func (c *Client) ListProfileDevices(ctx context.Context, opts *ListOptions) ([]ProfileDevice, error) {
	return getPaginatedResults[ProfileDevice](ctx, c, "profile/devices", opts)
}

// DeleteProfileDevice revokes the given ProfileDevice's status as a trusted device
func (c *Client) DeleteProfileDevice(ctx context.Context, deviceID int) error {
	e := formatAPIPath("profile/devices/%d", deviceID)
	return doDELETERequest(ctx, c, e)
}
