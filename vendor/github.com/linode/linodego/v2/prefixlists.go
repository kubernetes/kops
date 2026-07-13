package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// PrefixList represents a network prefix list returned by the API.
type PrefixList struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Visibility         string    `json:"visibility"`
	SourcePrefixListID *int      `json:"source_prefixlist_id"`
	IPv4               *[]string `json:"ipv4"`
	IPv6               *[]string `json:"ipv6"`
	Version            int       `json:"version"`

	Created *time.Time `json:"-"`
	Updated *time.Time `json:"-"`
	Deleted *time.Time `json:"-"`
}

// UnmarshalJSON implements custom timestamp parsing for PrefixList values.
func (p *PrefixList) UnmarshalJSON(data []byte) error {
	type Mask PrefixList

	aux := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
		Deleted *parseabletime.ParseableTime `json:"deleted"`
	}{
		Mask: (*Mask)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Created = (*time.Time)(aux.Created)
	p.Updated = (*time.Time)(aux.Updated)
	p.Deleted = (*time.Time)(aux.Deleted)

	return nil
}

// ListPrefixLists returns a paginated collection of Prefix Lists.
func (c *Client) ListPrefixLists(ctx context.Context, opts *ListOptions) ([]PrefixList, error) {
	return getPaginatedResults[PrefixList](ctx, c, "networking/prefixlists", opts)
}

// GetPrefixList fetches a single Prefix List by its ID.
func (c *Client) GetPrefixList(ctx context.Context, id int) (*PrefixList, error) {
	endpoint := formatAPIPath("networking/prefixlists/%d", id)
	return doGETRequest[PrefixList](ctx, c, endpoint)
}
