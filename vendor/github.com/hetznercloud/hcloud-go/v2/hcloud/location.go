package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// Location represents a location in the Hetzner Cloud.
type Location struct {
	ID          int64
	Name        string
	Description string
	Country     string
	City        string
	Latitude    float64
	Longitude   float64
	NetworkZone NetworkZone
}

// LocationClient is a client for the location API.
type LocationClient struct {
	client *Client
}

// GetByID retrieves a location by its ID. If the location does not exist, nil is returned.
func (c *LocationClient) GetByID(ctx context.Context, id int64) (*Location, *Response, error) {
	const opPath = "/locations/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.LocationGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return LocationFromSchema(respBody.Location), resp, nil
}

// GetByName retrieves an location by its name. If the location does not exist, nil is returned.
func (c *LocationClient) GetByName(ctx context.Context, name string) (*Location, *Response, error) {
	return firstByName(name, func() ([]*Location, *Response, error) {
		return c.List(ctx, LocationListOpts{Name: name})
	})
}

// Get retrieves a location by its ID if the input can be parsed as an integer, otherwise it
// retrieves a location by its name. If the location does not exist, nil is returned.
func (c *LocationClient) Get(ctx context.Context, idOrName string) (*Location, *Response, error) {
	if id, err := strconv.ParseInt(idOrName, 10, 64); err == nil {
		return c.GetByID(ctx, id)
	}
	return c.GetByName(ctx, idOrName)
}

// LocationListOpts specifies options for listing location.
type LocationListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l LocationListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of locations for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *LocationClient) List(ctx context.Context, opts LocationListOpts) ([]*Location, *Response, error) {
	const opPath = "/locations?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.LocationListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Locations, LocationFromSchema), resp, nil
}

// All returns all locations.
func (c *LocationClient) All(ctx context.Context) ([]*Location, error) {
	return c.AllWithOpts(ctx, LocationListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all locations for the given options.
func (c *LocationClient) AllWithOpts(ctx context.Context, opts LocationListOpts) ([]*Location, error) {
	return iterPages(func(page int) ([]*Location, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}
