package webhooks

import (
	"github.com/mitchellh/mapstructure"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

type webhookResult struct {
	gophercloud.Result
}

// Extract interprets any webhookResult as a Webhook, if possible.
func (r webhookResult) Extract() (*Webhook, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	var response struct {
		Webhook Webhook `mapstructure:"webhook"`
	}

	err := mapstructure.Decode(r.Body, &response)

	return &response.Webhook, err
}

// CreateResult represents the result of a create operation.
type CreateResult struct {
	webhookResult
}

// Extract extracts a slice of Webhooks from a CreateResult.  Multiple webhooks
// can be created in a single operation, so the result of a create is always a
// list of webhooks.
func (res CreateResult) Extract() ([]Webhook, error) {
	if res.Err != nil {
		return nil, res.Err
	}

	return commonExtractWebhooks(res.Body)
}

// GetResult temporarily contains the response from a Get call.
type GetResult struct {
	webhookResult
}

// UpdateResult represents the result of an update operation.
type UpdateResult struct {
	gophercloud.ErrResult
}

// DeleteResult represents the result of a delete operation.
type DeleteResult struct {
	gophercloud.ErrResult
}

// Webhook represents a webhook associted with a scaling policy.
type Webhook struct {
	// UUID for the webhook.
	ID string `mapstructure:"id" json:"id"`

	// Name of the webhook.
	Name string `mapstructure:"name" json:"name"`

	// Links associated with the webhook, including the capability URL.
	Links []gophercloud.Link `mapstructure:"links" json:"links"`

	// Metadata associated with the webhook.
	Metadata map[string]string `mapstructure:"metadata" json:"metadata"`
}

// WebhookPage is the page returned by a pager when traversing over a collection
// of webhooks.
type WebhookPage struct {
	pagination.SinglePageBase
}

// IsEmpty returns true if a page contains no Webhook results.
func (page WebhookPage) IsEmpty() (bool, error) {
	hooks, err := ExtractWebhooks(page)

	if err != nil {
		return true, err
	}

	return len(hooks) == 0, nil
}

// ExtractWebhooks interprets the results of a single page from a List() call,
// producing a slice of Webhooks.
func ExtractWebhooks(page pagination.Page) ([]Webhook, error) {
	return commonExtractWebhooks(page.(WebhookPage).Body)
}

func commonExtractWebhooks(body interface{}) ([]Webhook, error) {
	var response struct {
		Webhooks []Webhook `mapstructure:"webhooks"`
	}

	err := mapstructure.Decode(body, &response)

	if err != nil {
		return nil, err
	}

	return response.Webhooks, err
}
