package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// Promotion represents a Promotion object
type Promotion struct {
	// The amount available to spend per month.
	CreditMonthlyCap string `json:"credit_monthly_cap"`

	// The total amount of credit left for this promotion.
	CreditRemaining string `json:"credit_remaining"`

	// A detailed description of this promotion.
	Description string `json:"description"`

	// When this promotion's credits expire.
	ExpirationDate *time.Time `json:"-"`

	// The location of an image for this promotion.
	ImageURL string `json:"image_url"`

	// The service to which this promotion applies.
	ServiceType string `json:"service_type"`

	// Short details of this promotion.
	Summary string `json:"summary"`

	// The amount of credit left for this month for this promotion.
	ThisMonthCreditRemaining string `json:"this_month_credit_remaining"`
}

// PromoCodeCreateOptions fields are those accepted by AddPromoCode
type PromoCodeCreateOptions struct {
	// The Promo Code.
	PromoCode string `json:"promo_code"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Promotion) UnmarshalJSON(b []byte) error {
	type Mask Promotion

	p := struct {
		*Mask

		ExpirationDate *parseabletime.ParseableTime `json:"date"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.ExpirationDate = (*time.Time)(p.ExpirationDate)

	return nil
}

// AddPromoCode adds the provided promo code to the account
func (c *Client) AddPromoCode(ctx context.Context, opts PromoCodeCreateOptions) (*Promotion, error) {
	return doPOSTRequest[Promotion](ctx, c, "account/promo-codes", opts)
}
