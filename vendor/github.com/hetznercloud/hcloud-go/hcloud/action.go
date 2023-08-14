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
	client *Client
}

// GetByID retrieves an action by its ID. If the action does not exist, nil is returned.
func (c *ActionClient) GetByID(ctx context.Context, id int) (*Action, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/actions/%d", id), nil)
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
	path := "/actions?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
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

// All returns all actions.
func (c *ActionClient) All(ctx context.Context) ([]*Action, error) {
	return c.AllWithOpts(ctx, ActionListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all actions for the given options.
func (c *ActionClient) AllWithOpts(ctx context.Context, opts ActionListOpts) ([]*Action, error) {
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

// WatchOverallProgress watches several actions' progress until they complete
// with success or error. This watching happens in a goroutine and updates are
// provided through the two returned channels:
//
//   - The first channel receives percentage updates of the progress, based on
//     the number of completed versus total watched actions. The return value
//     is an int between 0 and 100.
//   - The second channel returned receives errors for actions that did not
//     complete successfully, as well as any errors that happened while
//     querying the API.
//
// By default, the method keeps watching until all actions have finished
// processing. If you want to be able to cancel the method or configure a
// timeout, use the [context.Context]. Once the method has stopped watching,
// both returned channels are closed.
//
// WatchOverallProgress uses the [WithPollBackoffFunc] of the [Client] to wait
// until sending the next request.
func (c *ActionClient) WatchOverallProgress(ctx context.Context, actions []*Action) (<-chan int, <-chan error) {
	errCh := make(chan error, len(actions))
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		successIDs := make([]int, 0, len(actions))
		watchIDs := make(map[int]struct{}, len(actions))
		for _, action := range actions {
			watchIDs[action.ID] = struct{}{}
		}

		retries := 0

		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-time.After(c.client.pollBackoffFunc(retries)):
				retries++
			}

			opts := ActionListOpts{}
			for watchID := range watchIDs {
				opts.ID = append(opts.ID, watchID)
			}

			as, err := c.AllWithOpts(ctx, opts)
			if err != nil {
				errCh <- err
				return
			}
			if len(as) == 0 {
				// No actions returned for the provided IDs, they do not exist in the API.
				// We need to catch and fail early for this, otherwise the loop will continue
				// indefinitely.
				errCh <- fmt.Errorf("failed to wait for actions: remaining actions (%v) are not returned from API", opts.ID)
				return
			}

			for _, a := range as {
				switch a.Status {
				case ActionStatusRunning:
					continue
				case ActionStatusSuccess:
					delete(watchIDs, a.ID)
					successIDs = append(successIDs, a.ID)
					sendProgress(progressCh, int(float64(len(actions)-len(successIDs))/float64(len(actions))*100))
				case ActionStatusError:
					delete(watchIDs, a.ID)
					errCh <- fmt.Errorf("action %d failed: %w", a.ID, a.Error())
				}
			}

			if len(watchIDs) == 0 {
				return
			}
		}
	}()

	return progressCh, errCh
}

// WatchProgress watches one action's progress until it completes with success
// or error. This watching happens in a goroutine and updates are provided
// through the two returned channels:
//
//   - The first channel receives percentage updates of the progress, based on
//     the progress percentage indicated by the API. The return value is an int
//     between 0 and 100.
//   - The second channel receives any errors that happened while querying the
//     API, as well as the error of the action if it did not complete
//     successfully, or nil if it did.
//
// By default, the method keeps watching until the action has finished
// processing. If you want to be able to cancel the method or configure a
// timeout, use the [context.Context]. Once the method has stopped watching,
// both returned channels are closed.
//
// WatchProgress uses the [WithPollBackoffFunc] of the [Client] to wait until
// sending the next request.
func (c *ActionClient) WatchProgress(ctx context.Context, action *Action) (<-chan int, <-chan error) {
	errCh := make(chan error, 1)
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		retries := 0

		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-time.After(c.client.pollBackoffFunc(retries)):
				retries++
			}

			a, _, err := c.GetByID(ctx, action.ID)
			if err != nil {
				errCh <- err
				return
			}
			if a == nil {
				errCh <- fmt.Errorf("failed to wait for action %d: action not returned from API", action.ID)
				return
			}

			switch a.Status {
			case ActionStatusRunning:
				sendProgress(progressCh, a.Progress)
			case ActionStatusSuccess:
				sendProgress(progressCh, 100)
				errCh <- nil
				return
			case ActionStatusError:
				errCh <- a.Error()
				return
			}
		}
	}()

	return progressCh, errCh
}

func sendProgress(progressCh chan int, p int) {
	select {
	case progressCh <- p:
		break
	default:
		break
	}
}
