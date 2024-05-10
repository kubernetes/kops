package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// Action represents an action in the Hetzner Cloud.
type Action struct {
	ID           int
	Status       ActionStatus
	Command      string
	Progress     int
	Started      time.Time
	Finished     time.Time
	ErrorCode    string
	ErrorMessage string
	Resources    []*ActionResource
}

// ActionStatus represents an action's status.
type ActionStatus string

// List of action statuses.
const (
	ActionStatusRunning ActionStatus = "running"
	ActionStatusSuccess ActionStatus = "success"
	ActionStatusError   ActionStatus = "error"
)

// ActionResource references other resources from an action.
type ActionResource struct {
	ID   int
	Type ActionResourceType
}

// ActionResourceType represents an action's resource reference type.
type ActionResourceType string

// List of action resource reference types.
const (
	ActionResourceTypeServer     ActionResourceType = "server"
	ActionResourceTypeImage      ActionResourceType = "image"
	ActionResourceTypeISO        ActionResourceType = "iso"
	ActionResourceTypeFloatingIP ActionResourceType = "floating_ip"
	ActionResourceTypeVolume     ActionResourceType = "volume"
)

// ActionError is the error of an action.
type ActionError struct {
	Code    string
	Message string
}

func (e ActionError) Error() string {
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

func (a *Action) Error() error {
	if a.ErrorCode != "" && a.ErrorMessage != "" {
		return ActionError{
			Code:    a.ErrorCode,
			Message: a.ErrorMessage,
		}
	}
	return nil
}

// ActionClient is a client for the actions API.
type ActionClient struct {
	action *ResourceActionClient
}

// GetByID retrieves an action by its ID. If the action does not exist, nil is returned.
func (c *ActionClient) GetByID(ctx context.Context, id int) (*Action, *Response, error) {
	return c.action.GetByID(ctx, id)
}

// ActionListOpts specifies options for listing actions.
type ActionListOpts struct {
	ListOpts
	ID     []int
	Status []ActionStatus
	Sort   []string
}

func (l ActionListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	for _, id := range l.ID {
		vals.Add("id", fmt.Sprintf("%d", id))
	}
	for _, status := range l.Status {
		vals.Add("status", string(status))
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of actions for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ActionClient) List(ctx context.Context, opts ActionListOpts) ([]*Action, *Response, error) {
	return c.action.List(ctx, opts)
}

// All returns all actions.
func (c *ActionClient) All(ctx context.Context) ([]*Action, error) {
	return c.action.All(ctx, ActionListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all actions for the given options.
func (c *ActionClient) AllWithOpts(ctx context.Context, opts ActionListOpts) ([]*Action, error) {
	return c.action.All(ctx, opts)
}

// ResourceActionClient is a client for the actions API exposed by the resource.
type ResourceActionClient struct {
	resource string
	client   *Client
}

func (c *ResourceActionClient) getBaseURL() string {
	if c.resource == "" {
		return ""
	}

	return "/" + c.resource
}

// GetByID retrieves an action by its ID. If the action does not exist, nil is returned.
func (c *ResourceActionClient) GetByID(ctx context.Context, id int) (*Action, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("%s/actions/%d", c.getBaseURL(), id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ActionGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return ActionFromSchema(body.Action), resp, nil
}

// List returns a list of actions for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ResourceActionClient) List(ctx context.Context, opts ActionListOpts) ([]*Action, *Response, error) {
	req, err := c.client.NewRequest(
		ctx,
		"GET",
		fmt.Sprintf("%s/actions?%s", c.getBaseURL(), opts.values().Encode()),
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	var body schema.ActionListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	actions := make([]*Action, 0, len(body.Actions))
	for _, i := range body.Actions {
		actions = append(actions, ActionFromSchema(i))
	}
	return actions, resp, nil
}

// All returns all actions for the given options.
func (c *ResourceActionClient) All(ctx context.Context, opts ActionListOpts) ([]*Action, error) {
	allActions := []*Action{}

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		actions, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allActions = append(allActions, actions...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allActions, nil
}
