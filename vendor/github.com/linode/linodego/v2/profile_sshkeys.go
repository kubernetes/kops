package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// SSHKey represents a SSHKey object
type SSHKey struct {
	ID      int        `json:"id"`
	Label   string     `json:"label"`
	SSHKey  string     `json:"ssh_key"`
	Created *time.Time `json:"-"`
}

// SSHKeyCreateOptions fields are those accepted by CreateSSHKey
type SSHKeyCreateOptions struct {
	Label  string `json:"label"`
	SSHKey string `json:"ssh_key"`
}

// SSHKeyUpdateOptions fields are those accepted by UpdateSSHKey
type SSHKeyUpdateOptions struct {
	Label string `json:"label"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *SSHKey) UnmarshalJSON(b []byte) error {
	type Mask SSHKey

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)

	return nil
}

// GetCreateOptions converts a SSHKey to SSHKeyCreateOptions for use in CreateSSHKey
func (i SSHKey) GetCreateOptions() (o SSHKeyCreateOptions) {
	o.Label = i.Label
	o.SSHKey = i.SSHKey

	return o
}

// GetUpdateOptions converts a SSHKey to SSHKeyCreateOptions for use in UpdateSSHKey
func (i SSHKey) GetUpdateOptions() (o SSHKeyUpdateOptions) {
	o.Label = i.Label
	return o
}

// ListSSHKeys lists SSHKeys
func (c *Client) ListSSHKeys(ctx context.Context, opts *ListOptions) ([]SSHKey, error) {
	return getPaginatedResults[SSHKey](ctx, c, "profile/sshkeys", opts)
}

// GetSSHKey gets the sshkey with the provided ID
func (c *Client) GetSSHKey(ctx context.Context, keyID int) (*SSHKey, error) {
	e := formatAPIPath("profile/sshkeys/%d", keyID)
	return doGETRequest[SSHKey](ctx, c, e)
}

// CreateSSHKey creates a SSHKey
func (c *Client) CreateSSHKey(ctx context.Context, opts SSHKeyCreateOptions) (*SSHKey, error) {
	return doPOSTRequest[SSHKey](ctx, c, "profile/sshkeys", opts)
}

// UpdateSSHKey updates the SSHKey with the specified id
func (c *Client) UpdateSSHKey(ctx context.Context, keyID int, opts SSHKeyUpdateOptions) (*SSHKey, error) {
	e := formatAPIPath("profile/sshkeys/%d", keyID)
	return doPUTRequest[SSHKey](ctx, c, e, opts)
}

// DeleteSSHKey deletes the SSHKey with the specified id
func (c *Client) DeleteSSHKey(ctx context.Context, keyID int) error {
	e := formatAPIPath("profile/sshkeys/%d", keyID)
	return doDELETERequest(ctx, c, e)
}
