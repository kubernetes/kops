package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// AccountServiceTransferStatus constants start with AccountServiceTransfer and
// include Linode API Account Service Transfer Status values.
type AccountServiceTransferStatus string

// AccountServiceTransferStatus constants reflect the current status of an AccountServiceTransfer
const (
	AccountServiceTransferAccepted  AccountServiceTransferStatus = "accepted"
	AccountServiceTransferCanceled  AccountServiceTransferStatus = "canceled"
	AccountServiceTransferCompleted AccountServiceTransferStatus = "completed"
	AccountServiceTransferFailed    AccountServiceTransferStatus = "failed"
	AccountServiceTransferPending   AccountServiceTransferStatus = "pending"
	AccountServiceTransferStale     AccountServiceTransferStatus = "stale"
)

// AccountServiceTransfer represents a request to transfer a service on an Account
type AccountServiceTransfer struct {
	Created  *time.Time                   `json:"-"`
	Entities AccountServiceTransferEntity `json:"entities"`
	Expiry   *time.Time                   `json:"-"`
	IsSender bool                         `json:"is_sender"`
	Status   AccountServiceTransferStatus `json:"status"`
	Token    string                       `json:"token"`
	Updated  *time.Time                   `json:"-"`
}

// AccountServiceTransferEntity represents a collection of the services to include
// in a transfer request, separated by type.
// Note: At this time, only Linodes can be transferred.
type AccountServiceTransferEntity struct {
	Linodes []int `json:"linodes"`
}

type AccountServiceTransferRequestOptions struct {
	Entities AccountServiceTransferEntity `json:"entities"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (ast *AccountServiceTransfer) UnmarshalJSON(b []byte) error {
	type Mask AccountServiceTransfer

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(ast),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	ast.Created = (*time.Time)(p.Created)
	ast.Expiry = (*time.Time)(p.Expiry)
	ast.Updated = (*time.Time)(p.Updated)

	return nil
}

// ListAccountServiceTransfer gets a paginated list of AccountServiceTransfer for the Account.
func (c *Client) ListAccountServiceTransfer(ctx context.Context, opts *ListOptions) ([]AccountServiceTransfer, error) {
	return getPaginatedResults[AccountServiceTransfer](ctx, c, "account/service-transfers", opts)
}

// GetAccountServiceTransfer gets the details of the AccountServiceTransfer for the provided token.
func (c *Client) GetAccountServiceTransfer(ctx context.Context, token string) (*AccountServiceTransfer, error) {
	e := formatAPIPath("account/service-transfers/%s", token)
	return doGETRequest[AccountServiceTransfer](ctx, c, e)
}

// RequestAccountServiceTransfer creates a transfer request for the specified services.
func (c *Client) RequestAccountServiceTransfer(ctx context.Context, opts AccountServiceTransferRequestOptions) (*AccountServiceTransfer, error) {
	return doPOSTRequest[AccountServiceTransfer](ctx, c, "account/service-transfers", opts)
}

// AcceptAccountServiceTransfer accepts an AccountServiceTransfer for the provided token to
// receive the services included in the transfer to the Account.
func (c *Client) AcceptAccountServiceTransfer(ctx context.Context, token string) error {
	e := formatAPIPath("account/service-transfers/%s/accept", token)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// CancelAccountServiceTransfer cancels the AccountServiceTransfer for the provided token.
func (c *Client) CancelAccountServiceTransfer(ctx context.Context, token string) error {
	e := formatAPIPath("account/service-transfers/%s", token)
	return doDELETERequest(ctx, c, e)
}
