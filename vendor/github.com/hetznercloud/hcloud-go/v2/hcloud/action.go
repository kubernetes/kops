package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// Action represents an action in the Hetzner Cloud.
type Action struct {
	ID           int64
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
	ID   int64
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

	action *Action
}

// Action returns the [Action] that triggered the error if available.
func (e ActionError) Action() *Action {
	return e.action
}

func (e ActionError) Error() string {
	action := e.Action()
	if action != nil {
		// For easier debugging, the error string contains the Action ID.
		return fmt.Sprintf("%s (%s, %d)", e.Message, e.Code, action.ID)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

func (a *Action) Error() error {
	if a.ErrorCode != "" && a.ErrorMessage != "" {
		return ActionError{
			Code:    a.ErrorCode,
			Message: a.ErrorMessage,
			action:  a,
		}
	}
	return nil
}

// ActionClient is a client for the actions API.
type ActionClient struct {
	action *ResourceActionClient
}

// GetByID retrieves an action by its ID. If the action does not exist, nil is returned.
func (c *ActionClient) GetByID(ctx context.Context, id int64) (*Action, *Response, error) {
	return c.action.GetByID(ctx, id)
}

// ActionListOpts specifies options for listing actions.
type ActionListOpts struct {
	ListOpts
	ID     []int64
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
//
// Deprecated: It is required to pass in a list of IDs since 30 January 2025. Please use [ActionClient.AllWithOpts] instead.
func (c *ActionClient) All(ctx context.Context) ([]*Action, error) {
	return c.action.All(ctx, ActionListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all actions for the given options.
//
// It is required to set [ActionListOpts.ID]. Any other fields set in the opts are ignored.
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
func (c *ResourceActionClient) GetByID(ctx context.Context, id int64) (*Action, *Response, error) {
	opPath := c.getBaseURL() + "/actions/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.ActionGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// List returns a list of actions for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ResourceActionClient) List(ctx context.Context, opts ActionListOpts) ([]*Action, *Response, error) {
	opPath := c.getBaseURL() + "/actions?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.ActionListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Actions, ActionFromSchema), resp, nil
}

// All returns all actions for the given options.
func (c *ResourceActionClient) All(ctx context.Context, opts ActionListOpts) ([]*Action, error) {
	return iterPages(func(page int) ([]*Action, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}
