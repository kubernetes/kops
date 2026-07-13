package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

type Login struct {
	ID         int        `json:"id"`
	Datetime   *time.Time `json:"datetime"`
	IP         string     `json:"ip"`
	Restricted bool       `json:"restricted"`
	Username   string     `json:"username"`
	Status     string     `json:"status"`
}

func (c *Client) ListLogins(ctx context.Context, opts *ListOptions) ([]Login, error) {
	return getPaginatedResults[Login](ctx, c, "account/logins", opts)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Login) UnmarshalJSON(b []byte) error {
	type Mask Login

	l := struct {
		*Mask

		Datetime *parseabletime.ParseableTime `json:"datetime"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &l); err != nil {
		return err
	}

	i.Datetime = (*time.Time)(l.Datetime)

	return nil
}

func (c *Client) GetLogin(ctx context.Context, loginID int) (*Login, error) {
	e := formatAPIPath("account/logins/%d", loginID)
	return doGETRequest[Login](ctx, c, e)
}
