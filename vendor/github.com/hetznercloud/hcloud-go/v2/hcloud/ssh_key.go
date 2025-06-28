package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// SSHKey represents a SSH key in the Hetzner Cloud.
type SSHKey struct {
	ID          int64
	Name        string
	Fingerprint string
	PublicKey   string
	Labels      map[string]string
	Created     time.Time
}

// SSHKeyClient is a client for the SSH keys API.
type SSHKeyClient struct {
	client *Client
}

// GetByID retrieves a SSH key by its ID. If the SSH key does not exist, nil is returned.
func (c *SSHKeyClient) GetByID(ctx context.Context, id int64) (*SSHKey, *Response, error) {
	const opPath = "/ssh_keys/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.SSHKeyGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return SSHKeyFromSchema(respBody.SSHKey), resp, nil
}

// GetByName retrieves a SSH key by its name. If the SSH key does not exist, nil is returned.
func (c *SSHKeyClient) GetByName(ctx context.Context, name string) (*SSHKey, *Response, error) {
	return firstByName(name, func() ([]*SSHKey, *Response, error) {
		return c.List(ctx, SSHKeyListOpts{Name: name})
	})
}

// GetByFingerprint retreives a SSH key by its fingerprint. If the SSH key does not exist, nil is returned.
func (c *SSHKeyClient) GetByFingerprint(ctx context.Context, fingerprint string) (*SSHKey, *Response, error) {
	return firstBy(func() ([]*SSHKey, *Response, error) {
		return c.List(ctx, SSHKeyListOpts{Fingerprint: fingerprint})
	})
}

// Get retrieves a SSH key by its ID if the input can be parsed as an integer, otherwise it
// retrieves a SSH key by its name. If the SSH key does not exist, nil is returned.
func (c *SSHKeyClient) Get(ctx context.Context, idOrName string) (*SSHKey, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// SSHKeyListOpts specifies options for listing SSH keys.
type SSHKeyListOpts struct {
	ListOpts
	Name        string
	Fingerprint string
	Sort        []string
}

func (l SSHKeyListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	if l.Fingerprint != "" {
		vals.Add("fingerprint", l.Fingerprint)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of SSH keys for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *SSHKeyClient) List(ctx context.Context, opts SSHKeyListOpts) ([]*SSHKey, *Response, error) {
	const opPath = "/ssh_keys?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.SSHKeyListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.SSHKeys, SSHKeyFromSchema), resp, nil
}

// All returns all SSH keys.
func (c *SSHKeyClient) All(ctx context.Context) ([]*SSHKey, error) {
	return c.AllWithOpts(ctx, SSHKeyListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all SSH keys with the given options.
func (c *SSHKeyClient) AllWithOpts(ctx context.Context, opts SSHKeyListOpts) ([]*SSHKey, error) {
	return iterPages(func(page int) ([]*SSHKey, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// SSHKeyCreateOpts specifies parameters for creating a SSH key.
type SSHKeyCreateOpts struct {
	Name      string
	PublicKey string
	Labels    map[string]string
}

// Validate checks if options are valid.
func (o SSHKeyCreateOpts) Validate() error {
	if o.Name == "" {
		return missingField(o, "Name")
	}
	if o.PublicKey == "" {
		return missingField(o, "PublicKey")
	}
	return nil
}

// Create creates a new SSH key with the given options.
func (c *SSHKeyClient) Create(ctx context.Context, opts SSHKeyCreateOpts) (*SSHKey, *Response, error) {
	const opPath = "/ssh_keys"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := opPath

	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}
	reqBody := schema.SSHKeyCreateRequest{
		Name:      opts.Name,
		PublicKey: opts.PublicKey,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := postRequest[schema.SSHKeyCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return SSHKeyFromSchema(respBody.SSHKey), resp, nil
}

// Delete deletes a SSH key.
func (c *SSHKeyClient) Delete(ctx context.Context, sshKey *SSHKey) (*Response, error) {
	const opPath = "/ssh_keys/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, sshKey.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

// SSHKeyUpdateOpts specifies options for updating a SSH key.
type SSHKeyUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a SSH key.
func (c *SSHKeyClient) Update(ctx context.Context, sshKey *SSHKey, opts SSHKeyUpdateOpts) (*SSHKey, *Response, error) {
	const opPath = "/ssh_keys/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, sshKey.ID)

	reqBody := schema.SSHKeyUpdateRequest{
		Name: opts.Name,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.SSHKeyUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return SSHKeyFromSchema(respBody.SSHKey), resp, nil
}
