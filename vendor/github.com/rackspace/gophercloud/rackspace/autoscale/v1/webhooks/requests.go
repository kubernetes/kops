package webhooks

import (
	"errors"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

// Validation errors returned by create or update operations.
var (
	ErrNoName     = errors.New("Webhook name cannot by empty.")
	ErrNoMetadata = errors.New("Webhook metadata cannot be nil.")
)

// List returns all webhooks for a scaling policy.
func List(client *gophercloud.ServiceClient, groupID, policyID string) pagination.Pager {
	url := listURL(client, groupID, policyID)

	createPageFn := func(r pagination.PageResult) pagination.Page {
		return WebhookPage{pagination.SinglePageBase(r)}
	}

	return pagination.NewPager(client, url, createPageFn)
}

// CreateOptsBuilder is the interface responsible for generating the JSON
// for a Create operation.
type CreateOptsBuilder interface {
	ToWebhookCreateMap() ([]map[string]interface{}, error)
}

// CreateOpts is a slice of CreateOpt structs, that allow the user to create
// multiple webhooks in a single operation.
type CreateOpts []CreateOpt

// CreateOpt represents the options to create a webhook.
type CreateOpt struct {
	// Name [required] is a name for the webhook.
	Name string

	// Metadata [optional] is user-provided key-value metadata.
	// Maximum length for keys and values is 256 characters.
	Metadata map[string]string
}

// ToWebhookCreateMap converts a slice of CreateOpt structs into a map for use
// in the request body of a Create operation.
func (opts CreateOpts) ToWebhookCreateMap() ([]map[string]interface{}, error) {
	var webhooks []map[string]interface{}

	for _, o := range opts {
		if o.Name == "" {
			return nil, ErrNoName
		}

		hook := make(map[string]interface{})

		hook["name"] = o.Name

		if o.Metadata != nil {
			hook["metadata"] = o.Metadata
		}

		webhooks = append(webhooks, hook)
	}

	return webhooks, nil
}

// Create requests a new webhook be created and associated with the given group
// and scaling policy.
func Create(client *gophercloud.ServiceClient, groupID, policyID string, opts CreateOptsBuilder) CreateResult {
	var res CreateResult

	reqBody, err := opts.ToWebhookCreateMap()

	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = client.Post(createURL(client, groupID, policyID), reqBody, &res.Body, nil)

	return res
}

// Get requests the details of a single webhook with the given ID.
func Get(client *gophercloud.ServiceClient, groupID, policyID, webhookID string) GetResult {
	var result GetResult

	_, result.Err = client.Get(getURL(client, groupID, policyID, webhookID), &result.Body, nil)

	return result
}

// UpdateOptsBuilder is the interface responsible for generating the map
// structure for producing JSON for an Update operation.
type UpdateOptsBuilder interface {
	ToWebhookUpdateMap() (map[string]interface{}, error)
}

// UpdateOpts represents the options for updating an existing webhook.
//
// Update operations completely replace the configuration being updated. Empty
// values in the update are accepted and overwrite previously specified
// parameters.
type UpdateOpts struct {
	// Name of the webhook.
	Name string `mapstructure:"name" json:"name"`

	// Metadata associated with the webhook.
	Metadata map[string]string `mapstructure:"metadata" json:"metadata"`
}

// ToWebhookUpdateMap converts an UpdateOpts struct into a map for use as the
// request body in an Update request.
func (opts UpdateOpts) ToWebhookUpdateMap() (map[string]interface{}, error) {
	if opts.Name == "" {
		return nil, ErrNoName
	}

	if opts.Metadata == nil {
		return nil, ErrNoMetadata
	}

	hook := make(map[string]interface{})

	hook["name"] = opts.Name
	hook["metadata"] = opts.Metadata

	return hook, nil
}

// Update requests the configuration of the given webhook be updated.
func Update(client *gophercloud.ServiceClient, groupID, policyID, webhookID string, opts UpdateOptsBuilder) UpdateResult {
	var result UpdateResult

	url := updateURL(client, groupID, policyID, webhookID)
	reqBody, err := opts.ToWebhookUpdateMap()

	if err != nil {
		result.Err = err
		return result
	}

	_, result.Err = client.Put(url, reqBody, nil, &gophercloud.RequestOpts{
		OkCodes: []int{204},
	})

	return result
}

// Delete requests the given webhook be permanently deleted.
func Delete(client *gophercloud.ServiceClient, groupID, policyID, webhookID string) DeleteResult {
	var result DeleteResult

	url := deleteURL(client, groupID, policyID, webhookID)
	_, result.Err = client.Delete(url, &gophercloud.RequestOpts{
		OkCodes: []int{204},
	})

	return result
}
