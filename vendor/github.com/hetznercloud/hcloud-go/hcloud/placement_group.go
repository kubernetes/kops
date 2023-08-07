package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// PlacementGroup represents a Placement Group in the Hetzner Cloud.
type PlacementGroup struct {
	ID      int
	Name    string
	Labels  map[string]string
	Created time.Time
	Servers []int
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
func (c *PlacementGroupClient) GetByID(ctx context.Context, id int) (*PlacementGroup, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/placement_groups/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.PlacementGroupGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return PlacementGroupFromSchema(body.PlacementGroup), resp, nil
}

// GetByName retrieves a PlacementGroup by its name. If the PlacementGroup does not exist, nil is returned.
func (c *PlacementGroupClient) GetByName(ctx context.Context, name string) (*PlacementGroup, *Response, error) {
	if name == "" {
		return nil, nil, nil
	}
	placementGroups, response, err := c.List(ctx, PlacementGroupListOpts{Name: name})
	if len(placementGroups) == 0 {
		return nil, response, err
	}
	return placementGroups[0], response, err
}

// Get retrieves a PlacementGroup by its ID if the input can be parsed as an integer, otherwise it
// retrieves a PlacementGroup by its name. If the PlacementGroup does not exist, nil is returned.
func (c *PlacementGroupClient) Get(ctx context.Context, idOrName string) (*PlacementGroup, *Response, error) {
	if id, err := strconv.Atoi(idOrName); err == nil {
		return c.GetByID(ctx, int(id))
	}
	return c.GetByName(ctx, idOrName)
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
	path := "/placement_groups?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.PlacementGroupListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	placementGroups := make([]*PlacementGroup, 0, len(body.PlacementGroups))
	for _, g := range body.PlacementGroups {
		placementGroups = append(placementGroups, PlacementGroupFromSchema(g))
	}
	return placementGroups, resp, nil
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
	var allPlacementGroups []*PlacementGroup

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		placementGroups, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allPlacementGroups = append(allPlacementGroups, placementGroups...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allPlacementGroups, nil
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
		return errors.New("missing name")
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
	if err := opts.Validate(); err != nil {
		return PlacementGroupCreateResult{}, nil, err
	}
	reqBody := placementGroupCreateOptsToSchema(opts)
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return PlacementGroupCreateResult{}, nil, err
	}
	req, err := c.client.NewRequest(ctx, "POST", "/placement_groups", bytes.NewReader(reqBodyData))
	if err != nil {
		return PlacementGroupCreateResult{}, nil, err
	}

	respBody := schema.PlacementGroupCreateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return PlacementGroupCreateResult{}, nil, err
	}
	result := PlacementGroupCreateResult{
		PlacementGroup: PlacementGroupFromSchema(respBody.PlacementGroup),
	}
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
	reqBody := schema.PlacementGroupUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/placement_groups/%d", placementGroup.ID)
	req, err := c.client.NewRequest(ctx, "PUT", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.PlacementGroupUpdateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}

	return PlacementGroupFromSchema(respBody.PlacementGroup), resp, nil
}

// Delete deletes a PlacementGroup.
func (c *PlacementGroupClient) Delete(ctx context.Context, placementGroup *PlacementGroup) (*Response, error) {
	req, err := c.client.NewRequest(ctx, "DELETE", fmt.Sprintf("/placement_groups/%d", placementGroup.ID), nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req, nil)
}
