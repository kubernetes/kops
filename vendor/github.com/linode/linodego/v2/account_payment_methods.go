package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// PaymentMethod represents a PaymentMethod object
type PaymentMethod struct {
	// The unique ID of the Payment Method.
	ID int `json:"id"`

	// When the Payment Method was created.
	Created *time.Time `json:"created"`

	// Whether this Payment Method is the default method for automatically processing service charges.
	IsDefault bool `json:"is_default"`

	// The type of Payment Method.
	Type string `json:"type"`

	// The detailed data for the Payment Method, which can be of varying types.
	Data any `json:"data"`
}

// PaymentMethodDataCreditCard represents a PaymentMethodDataCreditCard object
type PaymentMethodDataCreditCard struct {
	// The type of credit card.
	CardType string `json:"card_type"`

	// The expiration month and year of the credit card.
	Expiry string `json:"expiry"`

	// The last four digits of the credit card number.
	LastFour string `json:"last_four"`
}

// PaymentMethodDataGooglePay represents a PaymentMethodDataGooglePay object
type PaymentMethodDataGooglePay struct {
	// The type of credit card.
	CardType string `json:"card_type"`

	// The expiration month and year of the credit card.
	Expiry string `json:"expiry"`

	// The last four digits of the credit card number.
	LastFour string `json:"last_four"`
}

// PaymentMethodDataPaypal represents a PaymentMethodDataPaypal object
type PaymentMethodDataPaypal struct {
	// The email address associated with your PayPal account.
	Email string `json:"email"`

	// PayPal Merchant ID associated with your PayPal account.
	PaypalID string `json:"paypal_id"`
}

// PaymentMethodCreateOptions fields are those accepted by CreatePaymentMethod
type PaymentMethodCreateOptions struct {
	// Whether this Payment Method is the default method for automatically processing service charges.
	IsDefault bool `json:"is_default"`

	// The type of Payment Method. Alternative payment methods including Google Pay and PayPal can be added using the Cloud Manager.
	Type string `json:"type"`

	// An object representing the credit card information you have on file with Linode to make Payments against your Account.
	Data *PaymentMethodCreateOptionsData `json:"data"`
}

type PaymentMethodCreateOptionsData struct {
	// Your credit card number. No spaces or hyphens (-) allowed.
	CardNumber string `json:"card_number"`

	// CVV (Card Verification Value) of the credit card, typically found on the back of the card.
	CVV string `json:"cvv"`

	// A value from 1-12 representing the expiration month of your credit card.
	ExpiryMonth int `json:"expiry_month"`

	// A four-digit integer representing the expiration year of your credit card.
	ExpiryYear int `json:"expiry_year"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *PaymentMethod) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "{}" || string(b) == "null" {
		return nil
	}

	type Mask PaymentMethod

	pm := &struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Data    json.RawMessage              `json:"data"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &pm); err != nil {
		return err
	}

	// Process Data based on the Type field
	switch i.Type {
	case "credit_card":
		var creditCardData PaymentMethodDataCreditCard
		if err := json.Unmarshal(pm.Data, &creditCardData); err != nil {
			return err
		}

		i.Data = creditCardData
	case "google_pay":
		var googlePayData PaymentMethodDataGooglePay
		if err := json.Unmarshal(pm.Data, &googlePayData); err != nil {
			return err
		}

		i.Data = googlePayData
	case "paypal":
		var paypalData PaymentMethodDataPaypal
		if err := json.Unmarshal(pm.Data, &paypalData); err != nil {
			return err
		}

		i.Data = paypalData
	default:
		return fmt.Errorf("unknown payment method type: %s", i.Type)
	}

	i.Created = (*time.Time)(pm.Created)

	return nil
}

// ListPaymentMethods lists PaymentMethods
func (c *Client) ListPaymentMethods(ctx context.Context, opts *ListOptions) ([]PaymentMethod, error) {
	return getPaginatedResults[PaymentMethod](ctx, c, "account/payment-methods", opts)
}

// GetPaymentMethod gets the payment method with the provided ID
func (c *Client) GetPaymentMethod(ctx context.Context, paymentMethodID int) (*PaymentMethod, error) {
	e := formatAPIPath("account/payment-methods/%d", paymentMethodID)
	return doGETRequest[PaymentMethod](ctx, c, e)
}

// DeletePaymentMethod deletes the payment method with the provided ID
func (c *Client) DeletePaymentMethod(ctx context.Context, paymentMethodID int) error {
	e := formatAPIPath("account/payment-methods/%d", paymentMethodID)
	return doDELETERequest(ctx, c, e)
}

// AddPaymentMethod adds the provided payment method to the account
func (c *Client) AddPaymentMethod(ctx context.Context, opts PaymentMethodCreateOptions) error {
	return doPOSTRequestNoResponseBody(ctx, c, "account/payment-methods", opts)
}

// SetDefaultPaymentMethod sets the payment method with the provided ID as the default
func (c *Client) SetDefaultPaymentMethod(ctx context.Context, paymentMethodID int) error {
	e := formatAPIPath("account/payment-methods/%d", paymentMethodID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}
