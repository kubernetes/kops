package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// PlacementGroup represents a Placement Group in the Hetzner Cloud.
type PlacementGroup struct {
	ID      int64
	Name    string
	Labels  map[string]string
	Created time.Time
	Servers []int64
	Type    PlacementGroupType
}

// PlacementGroupType specifies the type of a Placement Group.
type PlacementGroupType string

const (
	// PlacementGroupTypeSpread spreads all servers in the group on different vhosts.
	PlacementGroupTypeSpread PlacementGroupType = "spread"
)

// PlacementGroupClient is a client for the Placement Groups API.
type PlacementGroupClient struct {
	client *Client
}

// GetByID retrieves a PlacementGroup by its ID. If the PlacementGroup does not exist, nil is returned.
func (c *PlacementGroupClient) GetByID(ctx context.Context, id int64) (*PlacementGroup, *Response, error) {
	const opPath = "/placement_groups/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.PlacementGroupGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return PlacementGroupFromSchema(respBody.PlacementGroup), resp, nil
}

// GetByName retrieves a PlacementGroup by its name. If the PlacementGroup does not exist, nil is returned.
func (c *PlacementGroupClient) GetByName(ctx context.Context, name string) (*PlacementGroup, *Response, error) {
	return firstByName(name, func() ([]*PlacementGroup, *Response, error) {
		return c.List(ctx, PlacementGroupListOpts{Name: name})
	})
}

// Get retrieves a PlacementGroup by its ID if the input can be parsed as an integer, otherwise it
// retrieves a PlacementGroup by its name. If the PlacementGroup does not exist, nil is returned.
func (c *PlacementGroupClient) Get(ctx context.Context, idOrName string) (*PlacementGroup, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// PlacementGroupListOpts specifies options for listing PlacementGroup.
type PlacementGroupListOpts struct {
	ListOpts
	Name string
	Type PlacementGroupType
	Sort []string
}

func (l PlacementGroupListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	if l.Type != "" {
		vals.Add("type", string(l.Type))
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of PlacementGroups for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *PlacementGroupClient) List(ctx context.Context, opts PlacementGroupListOpts) ([]*PlacementGroup, *Response, error) {
	const opPath = "/placement_groups?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.PlacementGroupListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.PlacementGroups, PlacementGroupFromSchema), resp, nil
}

// All returns all PlacementGroups.
func (c *PlacementGroupClient) All(ctx context.Context) ([]*PlacementGroup, error) {
	opts := PlacementGroupListOpts{
		ListOpts: ListOpts{
			PerPage: 50,
		},
	}

	return c.AllWithOpts(ctx, opts)
}

// AllWithOpts returns all PlacementGroups for the given options.
func (c *PlacementGroupClient) AllWithOpts(ctx context.Context, opts PlacementGroupListOpts) ([]*PlacementGroup, error) {
	return iterPages(func(page int) ([]*PlacementGroup, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// PlacementGroupCreateOpts specifies options for creating a new PlacementGroup.
type PlacementGroupCreateOpts struct {
	Name   string
	Labels map[string]string
	Type   PlacementGroupType
}

// Validate checks if options are valid.
func (o PlacementGroupCreateOpts) Validate() error {
	if o.Name == "" {
		return missingField(o, "Name")
	}
	return nil
}

// PlacementGroupCreateResult is the result of a create PlacementGroup call.
type PlacementGroupCreateResult struct {
	PlacementGroup *PlacementGroup
	Action         *Action
}

// Create creates a new PlacementGroup.
func (c *PlacementGroupClient) Create(ctx context.Context, opts PlacementGroupCreateOpts) (PlacementGroupCreateResult, *Response, error) {
	const opPath = "/placement_groups"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := PlacementGroupCreateResult{}

	reqPath := opPath

	if err := opts.Validate(); err != nil {
		return result, nil, err
	}

	reqBody := placementGroupCreateOptsToSchema(opts)

	respBody, resp, err := postRequest[schema.PlacementGroupCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.PlacementGroup = PlacementGroupFromSchema(respBody.PlacementGroup)
	if respBody.Action != nil {
		result.Action = ActionFromSchema(*respBody.Action)
	}

	return result, resp, nil
}

// PlacementGroupUpdateOpts specifies options for updating a PlacementGroup.
type PlacementGroupUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a PlacementGroup.
func (c *PlacementGroupClient) Update(ctx context.Context, placementGroup *PlacementGroup, opts PlacementGroupUpdateOpts) (*PlacementGroup, *Response, error) {
	const opPath = "/placement_groups/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, placementGroup.ID)

	reqBody := schema.PlacementGroupUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.PlacementGroupUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return PlacementGroupFromSchema(respBody.PlacementGroup), resp, nil
}

// Delete deletes a PlacementGroup.
func (c *PlacementGroupClient) Delete(ctx context.Context, placementGroup *PlacementGroup) (*Response, error) {
	const opPath = "/placement_groups/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, placementGroup.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}
