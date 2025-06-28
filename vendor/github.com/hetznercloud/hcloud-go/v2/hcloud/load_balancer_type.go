package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// LoadBalancerType represents a LoadBalancer type in the Hetzner Cloud.
type LoadBalancerType struct {
	ID                      int64
	Name                    string
	Description             string
	MaxConnections          int
	MaxServices             int
	MaxTargets              int
	MaxAssignedCertificates int
	Pricings                []LoadBalancerTypeLocationPricing
	Deprecated              *string
}

// LoadBalancerTypeClient is a client for the Load Balancer types API.
type LoadBalancerTypeClient struct {
	client *Client
}

// GetByID retrieves a Load Balancer type by its ID. If the Load Balancer type does not exist, nil is returned.
func (c *LoadBalancerTypeClient) GetByID(ctx context.Context, id int64) (*LoadBalancerType, *Response, error) {
	const opPath = "/load_balancer_types/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.LoadBalancerTypeGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return LoadBalancerTypeFromSchema(respBody.LoadBalancerType), resp, nil
}

// GetByName retrieves a Load Balancer type by its name. If the Load Balancer type does not exist, nil is returned.
func (c *LoadBalancerTypeClient) GetByName(ctx context.Context, name string) (*LoadBalancerType, *Response, error) {
	return firstByName(name, func() ([]*LoadBalancerType, *Response, error) {
		return c.List(ctx, LoadBalancerTypeListOpts{Name: name})
	})
}

// Get retrieves a Load Balancer type by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Load Balancer type by its name. If the Load Balancer type does not exist, nil is returned.
func (c *LoadBalancerTypeClient) Get(ctx context.Context, idOrName string) (*LoadBalancerType, *Response, error) {
	if id, err := strconv.ParseInt(idOrName, 10, 64); err == nil {
		return c.GetByID(ctx, id)
	}
	return c.GetByName(ctx, idOrName)
}

// LoadBalancerTypeListOpts specifies options for listing Load Balancer types.
type LoadBalancerTypeListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l LoadBalancerTypeListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of Load Balancer types for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *LoadBalancerTypeClient) List(ctx context.Context, opts LoadBalancerTypeListOpts) ([]*LoadBalancerType, *Response, error) {
	const opPath = "/load_balancer_types?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.LoadBalancerTypeListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.LoadBalancerTypes, LoadBalancerTypeFromSchema), resp, nil
}

// All returns all Load Balancer types.
func (c *LoadBalancerTypeClient) All(ctx context.Context) ([]*LoadBalancerType, error) {
	return c.AllWithOpts(ctx, LoadBalancerTypeListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Load Balancer types for the given options.
func (c *LoadBalancerTypeClient) AllWithOpts(ctx context.Context, opts LoadBalancerTypeListOpts) ([]*LoadBalancerType, error) {
	return iterPages(func(page int) ([]*LoadBalancerType, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}
