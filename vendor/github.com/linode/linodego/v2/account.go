package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// Account associated with the token in use.
type Account struct {
	FirstName         string      `json:"first_name"`
	LastName          string      `json:"last_name"`
	Email             string      `json:"email"`
	Company           string      `json:"company"`
	Address1          string      `json:"address_1"`
	Address2          string      `json:"address_2"`
	Balance           float32     `json:"balance"`
	BalanceUninvoiced float32     `json:"balance_uninvoiced"`
	City              string      `json:"city"`
	State             string      `json:"state"`
	Zip               string      `json:"zip"`
	Country           string      `json:"country"`
	TaxID             string      `json:"tax_id"`
	Phone             string      `json:"phone"`
	CreditCard        *CreditCard `json:"credit_card"`
	EUUID             string      `json:"euuid"`
	BillingSource     string      `json:"billing_source"`
	Capabilities      []string    `json:"capabilities"`
	ActiveSince       *time.Time  `json:"active_since"`
	ActivePromotions  []Promotion `json:"active_promotions"`
}

// AccountUpdateOptions fields are those accepted by UpdateAccount
type AccountUpdateOptions struct {
	Address1  string `json:"address_1,omitzero"`
	Address2  string `json:"address_2,omitzero"`
	City      string `json:"city,omitzero"`
	Company   string `json:"company,omitzero"`
	Country   string `json:"country,omitzero"`
	Email     string `json:"email,omitzero"`
	FirstName string `json:"first_name,omitzero"`
	LastName  string `json:"last_name,omitzero"`
	Phone     string `json:"phone,omitzero"`
	State     string `json:"state,omitzero"`
	TaxID     string `json:"tax_id,omitzero"`
	Zip       string `json:"zip,omitzero"`
}

// GetUpdateOptions converts an Account to AccountUpdateOptions for use in UpdateAccount
func (i Account) GetUpdateOptions() (o AccountUpdateOptions) {
	o.Address1 = i.Address1
	o.Address2 = i.Address2
	o.City = i.City
	o.Company = i.Company
	o.Country = i.Country
	o.Email = i.Email
	o.FirstName = i.FirstName
	o.LastName = i.LastName
	o.Phone = i.Phone
	o.State = i.State
	o.TaxID = i.TaxID
	o.Zip = i.Zip

	return o
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Account) UnmarshalJSON(b []byte) error {
	type Mask Account

	p := struct {
		*Mask

		ActiveSince *parseabletime.ParseableTime `json:"active_since"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.ActiveSince = (*time.Time)(p.ActiveSince)

	return nil
}

// CreditCard information associated with the Account.
type CreditCard struct {
	LastFour string `json:"last_four"`
	Expiry   string `json:"expiry"`
}

// GetAccount gets the contact and billing information related to the Account.
func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	return doGETRequest[Account](ctx, c, "account")
}

// UpdateAccount updates the Account
func (c *Client) UpdateAccount(ctx context.Context, opts AccountUpdateOptions) (*Account, error) {
	return doPUTRequest[Account](ctx, c, "account", opts)
}
