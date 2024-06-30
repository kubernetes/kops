package hcloud

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"
)

type ActionWaiter interface {
	WaitForFunc(ctx context.Context, handleUpdate func(update *Action) error, actions ...*Action) error
	WaitFor(ctx context.Context, actions ...*Action) error
}

var _ ActionWaiter = (*ActionClient)(nil)

// WaitForFunc waits until all actions are completed by polling the API at the interval
// defined by [WithPollBackoffFunc]. An action is considered as complete when its status is
// either [ActionStatusSuccess] or [ActionStatusError].
//
// The handleUpdate callback is called every time an action is updated.
func (c *ActionClient) WaitForFunc(ctx context.Context, handleUpdate func(update *Action) error, actions ...*Action) error {
	// Filter out nil actions
	actions = slices.DeleteFunc(actions, func(a *Action) bool { return a == nil })

	running := make(map[int]struct{}, len(actions))
	for _, action := range actions {
		if action.Status == ActionStatusRunning {
			running[action.ID] = struct{}{}
		} else if handleUpdate != nil {
			// We filter out already completed actions from the API polling loop; while
			// this isn't a real update, the caller should be notified about the new
			// state.
			if err := handleUpdate(action); err != nil {
				return err
			}
		}
	}

	retries := 0
	for {
		if len(running) == 0 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.action.client.pollBackoffFunc(retries)):
			retries++
		}

		opts := ActionListOpts{
			Sort: []string{"status", "id"},
			ID:   make([]int, 0, len(running)),
		}
		for actionID := range running {
			opts.ID = append(opts.ID, actionID)
		}
		slices.Sort(opts.ID)

		updates, err := c.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}

		if len(updates) != len(running) {
			// Some actions may not exist in the API, also fail early to prevent an
			// infinite loop when updates == 0.

			notFound := maps.Clone(running)
			for _, update := range updates {
				delete(notFound, update.ID)
			}
			notFoundIDs := make([]int, 0, len(notFound))
			for unknownID := range notFound {
				notFoundIDs = append(notFoundIDs, unknownID)
			}

			return fmt.Errorf("actions not found: %v", notFoundIDs)
		}

		for _, update := range updates {
			if update.Status != ActionStatusRunning {
				delete(running, update.ID)
			}

			if handleUpdate != nil {
				if err := handleUpdate(update); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// WaitFor waits until all actions succeed by polling the API at the interval defined by
// [WithPollBackoffFunc]. An action is considered as succeeded when its status is either
// [ActionStatusSuccess].
//
// If a single action fails, the function will stop waiting and the error set in the
// action will be returned as an [ActionError].
//
// For more flexibility, see the [ActionClient.WaitForFunc] function.
func (c *ActionClient) WaitFor(ctx context.Context, actions ...*Action) error {
	return c.WaitForFunc(
		ctx,
		func(update *Action) error {
			if update.Status == ActionStatusError {
				return update.Error()
			}
			return nil
		},
		actions...,
	)
}
