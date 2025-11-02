package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// ISO represents an ISO image in the Hetzner Cloud.
type ISO struct {
	ID           int64
	Name         string
	Description  string
	Type         ISOType
	Architecture *Architecture
	// Deprecated: Use [ISO.Deprecation] instead.
	Deprecated time.Time
	DeprecatableResource
}

// ISOType specifies the type of an ISO image.
type ISOType string

const (
	// ISOTypePublic is the type of a public ISO image.
	ISOTypePublic ISOType = "public"

	// ISOTypePrivate is the type of a private ISO image.
	ISOTypePrivate ISOType = "private"
)

// ISOClient is a client for the ISO API.
type ISOClient struct {
	client *Client
}

// GetByID retrieves an ISO by its ID.
func (c *ISOClient) GetByID(ctx context.Context, id int64) (*ISO, *Response, error) {
	const opPath = "/isos/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.ISOGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return ISOFromSchema(respBody.ISO), resp, nil
}

// GetByName retrieves an ISO by its name.
func (c *ISOClient) GetByName(ctx context.Context, name string) (*ISO, *Response, error) {
	return firstByName(name, func() ([]*ISO, *Response, error) {
		return c.List(ctx, ISOListOpts{Name: name})
	})
}

// Get retrieves an ISO by its ID if the input can be parsed as an integer, otherwise it retrieves an ISO by its name.
func (c *ISOClient) Get(ctx context.Context, idOrName string) (*ISO, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// ISOListOpts specifies options for listing isos.
type ISOListOpts struct {
	ListOpts
	Name string
	Sort []string
	// Architecture filters the ISOs by Architecture. Note that custom ISOs do not have any architecture set, and you
	// must use IncludeWildcardArchitecture to include them.
	Architecture []Architecture
	// IncludeWildcardArchitecture must be set to also return custom ISOs that have no architecture set, if you are
	// also setting the Architecture field.
	//
	// Deprecated: Use [ISOListOpts.IncludeArchitectureWildcard] instead.
	IncludeWildcardArchitecture bool
	// IncludeWildcardArchitecture must be set to also return custom ISOs that have no architecture set, if you are
	// also setting the Architecture field.
	IncludeArchitectureWildcard bool
}

func (l ISOListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	for _, arch := range l.Architecture {
		vals.Add("architecture", string(arch))
	}
	if l.IncludeArchitectureWildcard || l.IncludeWildcardArchitecture {
		vals.Add("include_architecture_wildcard", "true")
	}
	return vals
}

// List returns a list of ISOs for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ISOClient) List(ctx context.Context, opts ISOListOpts) ([]*ISO, *Response, error) {
	const opPath = "/isos?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.ISOListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.ISOs, ISOFromSchema), resp, nil
}

// All returns all ISOs.
func (c *ISOClient) All(ctx context.Context) ([]*ISO, error) {
	return c.AllWithOpts(ctx, ISOListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all ISOs for the given options.
func (c *ISOClient) AllWithOpts(ctx context.Context, opts ISOListOpts) ([]*ISO, error) {
	return iterPages(func(page int) ([]*ISO, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}
