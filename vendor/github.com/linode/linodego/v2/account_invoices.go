package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// Invoice structs reflect an invoice for billable activity on the account.
type Invoice struct {
	ID            int                 `json:"id"`
	Label         string              `json:"label"`
	Total         float32             `json:"total"`
	Date          *time.Time          `json:"-"`
	Tax           float32             `json:"tax"`
	Subtotal      float32             `json:"subtotal"`
	BillingSource string              `json:"billing_source"`
	TaxSummary    []InvoiceTaxSummary `json:"tax_summary"`
}

type InvoiceTaxSummary struct {
	Tax  float32 `json:"tax"`
	Name string  `json:"name"`
}

// InvoiceItem structs reflect a single billable activity associate with an Invoice
type InvoiceItem struct {
	Label     string     `json:"label"`
	Type      string     `json:"type"`
	UnitPrice string     `json:"unit_price"`
	Quantity  int        `json:"quantity"`
	Amount    float32    `json:"amount"`
	Tax       float32    `json:"tax"`
	Region    *string    `json:"region"`
	From      *time.Time `json:"-"`
	To        *time.Time `json:"-"`
	Total     float32    `json:"total"`
}

// ListInvoices gets a paginated list of Invoices against the Account
func (c *Client) ListInvoices(ctx context.Context, opts *ListOptions) ([]Invoice, error) {
	return getPaginatedResults[Invoice](ctx, c, "account/invoices", opts)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Invoice) UnmarshalJSON(b []byte) error {
	type Mask Invoice

	p := struct {
		*Mask

		Date *parseabletime.ParseableTime `json:"date"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Date = (*time.Time)(p.Date)

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *InvoiceItem) UnmarshalJSON(b []byte) error {
	type Mask InvoiceItem

	p := struct {
		*Mask

		From *parseabletime.ParseableTime `json:"from"`
		To   *parseabletime.ParseableTime `json:"to"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.From = (*time.Time)(p.From)
	i.To = (*time.Time)(p.To)

	return nil
}

// GetInvoice gets a single Invoice matching the provided ID
func (c *Client) GetInvoice(ctx context.Context, invoiceID int) (*Invoice, error) {
	e := formatAPIPath("account/invoices/%d", invoiceID)
	return doGETRequest[Invoice](ctx, c, e)
}

// ListInvoiceItems gets the invoice items associated with a specific Invoice
func (c *Client) ListInvoiceItems(ctx context.Context, invoiceID int, opts *ListOptions) ([]InvoiceItem, error) {
	return getPaginatedResults[InvoiceItem](ctx, c, formatAPIPath("account/invoices/%d/items", invoiceID), opts)
}
