package keypairs

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// CreateOptsExt adds a KeyPair option to the base CreateOpts.
type CreateOptsExt struct {
	servers.CreateOptsBuilder

	// KeyName is the name of the key pair.
	KeyName string `json:"key_name,omitempty"`
}

// ToServerCreateMap adds the key_name to the base server creation options.
func (opts CreateOptsExt) ToServerCreateMap() (map[string]any, error) {
	base, err := opts.CreateOptsBuilder.ToServerCreateMap()
	if err != nil {
		return nil, err
	}

	if opts.KeyName == "" {
		return base, nil
	}

	serverMap := base["server"].(map[string]any)
	serverMap["key_name"] = opts.KeyName

	return base, nil
}

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToKeyPairListQuery() (string, error)
}

// ListOpts enables listing KeyPairs based on specific attributes.
type ListOpts struct {
	// UserID is the user ID that owns the key pair.
	// This requires microversion 2.10 or higher.
	UserID string `q:"user_id"`
}

// ToKeyPairListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToKeyPairListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// List returns a Pager that allows you to iterate over a collection of KeyPairs.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := listURL(client)
	if opts != nil {
		query, err := opts.ToKeyPairListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return KeyPairPage{pagination.SinglePageBase(r)}
	})
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToKeyPairCreateMap() (map[string]any, error)
}

// CreateOpts specifies KeyPair creation or import parameters.
type CreateOpts struct {
	// Name is a friendly name to refer to this KeyPair in other services.
	Name string `json:"name" required:"true"`

	// UserID [optional] is the user_id for a keypair.
	// This allows administrative users to upload keys for other users than themselves.
	// This requires microversion 2.10 or higher.
	UserID string `json:"user_id,omitempty"`

	// The type of the keypair. Allowed values are ssh or x509
	// This requires microversion 2.2 or higher.
	Type string `json:"type,omitempty"`

	// PublicKey [optional] is a pregenerated OpenSSH-formatted public key.
	// If provided, this key will be imported and no new key will be created.
	PublicKey string `json:"public_key,omitempty"`
}

// ToKeyPairCreateMap constructs a request body from CreateOpts.
func (opts CreateOpts) ToKeyPairCreateMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "keypair")
}

// Create requests the creation of a new KeyPair on the server, or to import a
// pre-existing keypair.
func Create(ctx context.Context, client *gophercloud.ServiceClient, opts CreateOptsBuilder) (r CreateResult) {
	b, err := opts.ToKeyPairCreateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, createURL(client), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// GetOptsBuilder allows extensions to add additional parameters to the
// Get request.
type GetOptsBuilder interface {
	ToKeyPairGetQuery() (string, error)
}

// GetOpts enables retrieving KeyPairs based on specific attributes.
type GetOpts struct {
	// UserID is the user ID that owns the key pair.
	// This requires microversion 2.10 or higher.
	UserID string `q:"user_id"`
}

// ToKeyPairGetQuery formats a GetOpts into a query string.
func (opts GetOpts) ToKeyPairGetQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// Get returns public data about a previously uploaded KeyPair.
func Get(ctx context.Context, client *gophercloud.ServiceClient, name string, opts GetOptsBuilder) (r GetResult) {
	url := getURL(client, name)
	if opts != nil {
		query, err := opts.ToKeyPairGetQuery()
		if err != nil {
			r.Err = err
			return
		}
		url += query
	}

	resp, err := client.Get(ctx, url, &r.Body, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// DeleteOptsBuilder allows extensions to add additional parameters to the
// Delete request.
type DeleteOptsBuilder interface {
	ToKeyPairDeleteQuery() (string, error)
}

// DeleteOpts enables deleting KeyPairs based on specific attributes.
type DeleteOpts struct {
	// UserID is the user ID of the user that owns the key pair.
	// This requires microversion 2.10 or higher.
	UserID string `q:"user_id"`
}

// ToKeyPairDeleteQuery formats a DeleteOpts into a query string.
func (opts DeleteOpts) ToKeyPairDeleteQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// Delete requests the deletion of a previous stored KeyPair from the server.
func Delete(ctx context.Context, client *gophercloud.ServiceClient, name string, opts DeleteOptsBuilder) (r DeleteResult) {
	url := deleteURL(client, name)
	if opts != nil {
		query, err := opts.ToKeyPairDeleteQuery()
		if err != nil {
			r.Err = err
			return
		}
		url += query
	}

	resp, err := client.Delete(ctx, url, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
