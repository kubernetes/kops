package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// ConsumerImageShareGroup represents an ImageShareGroup that the consumer is a member of.
type ConsumerImageShareGroup struct {
	ID          int        `json:"id"`
	UUID        string     `json:"uuid"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	IsSuspended bool       `json:"is_suspended"`
	Created     *time.Time `json:"-"`
	Updated     *time.Time `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (isg *ConsumerImageShareGroup) UnmarshalJSON(b []byte) error {
	type Mask ConsumerImageShareGroup

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(isg),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	isg.Created = (*time.Time)(p.Created)
	isg.Updated = (*time.Time)(p.Updated)

	return nil
}

// ImageShareGroupToken contains information about a token created by a consumer.
// The token itself is only visible once upon creation.
type ImageShareGroupToken struct {
	TokenUUID              string     `json:"token_uuid"`
	Status                 string     `json:"status"`
	Label                  string     `json:"label"`
	ValidForShareGroupUUID string     `json:"valid_for_sharegroup_uuid"`
	Created                *time.Time `json:"-"`
	Updated                *time.Time `json:"-"`
	Expiry                 *time.Time `json:"-"`
	ShareGroupUUID         *string    `json:"sharegroup_uuid"`
	ShareGroupLabel        *string    `json:"sharegroup_label"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *ImageShareGroupToken) UnmarshalJSON(b []byte) error {
	type Mask ImageShareGroupToken

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
	}{
		Mask: (*Mask)(t),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	t.Created = (*time.Time)(p.Created)
	t.Updated = (*time.Time)(p.Updated)
	t.Expiry = (*time.Time)(p.Expiry)

	return nil
}

// ImageShareGroupCreateTokenResponse represents the response when the consumer
// creates a single-use ImageShareGroup membership token.
// The token itself is only provided upon creation, and must be given to the producer
// via an outside medium for the consumer to be added as a member of the producer's ImageShareGroup.
type ImageShareGroupCreateTokenResponse struct {
	Token                  string     `json:"token"`
	TokenUUID              string     `json:"token_uuid"`
	Status                 string     `json:"status"`
	Label                  string     `json:"label"`
	ValidForShareGroupUUID string     `json:"valid_for_sharegroup_uuid"`
	Created                *time.Time `json:"-"`
	Updated                *time.Time `json:"-"`
	Expiry                 *time.Time `json:"-"`
	ShareGroupUUID         *string    `json:"sharegroup_uuid"`
	ShareGroupLabel        *string    `json:"sharegroup_label"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *ImageShareGroupCreateTokenResponse) UnmarshalJSON(b []byte) error {
	type Mask ImageShareGroupCreateTokenResponse

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
	}{
		Mask: (*Mask)(t),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	t.Created = (*time.Time)(p.Created)
	t.Updated = (*time.Time)(p.Updated)
	t.Expiry = (*time.Time)(p.Expiry)

	return nil
}

// ImageShareGroupCreateTokenOptions fields are those accepted by ImageShareGroupCreateToken
type ImageShareGroupCreateTokenOptions struct {
	Label                  *string `json:"label,omitzero"`
	ValidForShareGroupUUID string  `json:"valid_for_sharegroup_uuid"`
}

// ImageShareGroupUpdateTokenOptions fields are those accepted by ImageShareGroupUpdateToken
type ImageShareGroupUpdateTokenOptions struct {
	Label string `json:"label"`
}

// ImageShareGroupListTokens lists information about all the ImageShareGroupTokens created by the user.
// The tokens themselves are only visible once upon creation.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupListTokens(ctx context.Context, opts *ListOptions) ([]ImageShareGroupToken, error) {
	return getPaginatedResults[ImageShareGroupToken](
		ctx,
		c,
		"images/sharegroups/tokens",
		opts,
	)
}

// ImageShareGroupGetToken gets information about the specified ImageShareGroupToken created by the user.
// The tokens themselves are only visible once upon creation.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupGetToken(ctx context.Context, tokenUUID string) (*ImageShareGroupToken, error) {
	return doGETRequest[ImageShareGroupToken](
		ctx,
		c,
		formatAPIPath("images/sharegroups/tokens/%s", tokenUUID),
	)
}

// ImageShareGroupCreateToken allows the consumer to create a single-use ImageShareGroup membership
// token for a specific ImageShareGroup owned by the producer.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupCreateToken(ctx context.Context, opts ImageShareGroupCreateTokenOptions) (*ImageShareGroupCreateTokenResponse, error) {
	return doPOSTRequest[ImageShareGroupCreateTokenResponse](
		ctx,
		c,
		formatAPIPath("images/sharegroups/tokens"),
		opts,
	)
}

// ImageShareGroupUpdateToken allows the consumer to update an ImageShareGroupToken's label.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupUpdateToken(ctx context.Context, tokenUUID string, opts ImageShareGroupUpdateTokenOptions) (*ImageShareGroupToken, error) {
	return doPUTRequest[ImageShareGroupToken](
		ctx,
		c,
		formatAPIPath("images/sharegroups/tokens/%s", tokenUUID),
		opts,
	)
}

// ImageShareGroupRemoveToken allows the consumer to remove an individual ImageShareGroupToken from an ImageShareGroup
// this token has been accepted into.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupRemoveToken(ctx context.Context, tokenUUID string) error {
	return doDELETERequest(
		ctx,
		c,
		formatAPIPath("images/sharegroups/tokens/%s", tokenUUID),
	)
}

// ImageShareGroupGetByToken gets information about the ImageShareGroup that the
// consumer's specified token has been accepted into.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupGetByToken(ctx context.Context, tokenUUID string) (*ConsumerImageShareGroup, error) {
	return doGETRequest[ConsumerImageShareGroup](
		ctx,
		c,
		formatAPIPath("images/sharegroups/tokens/%s/sharegroup", tokenUUID),
	)
}

// ImageShareGroupGetImageShareEntriesByToken lists the shared image entries in the ImageShareGroup that the
// consumer's specified token has been accepted into.
// NOTE: May not currently be available to all users and can only be used with v4beta.
func (c *Client) ImageShareGroupGetImageShareEntriesByToken(ctx context.Context, tokenUUID string, opts *ListOptions) ([]ImageShareEntry, error) {
	return getPaginatedResults[ImageShareEntry](
		ctx,
		c,
		formatAPIPath("images/sharegroups/tokens/%s/sharegroup/images", tokenUUID),
		opts,
	)
}
