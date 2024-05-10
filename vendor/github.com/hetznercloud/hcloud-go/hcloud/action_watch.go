package hcloud

import (
	"context"
	"fmt"
)

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
//
// Deprecated: WatchOverallProgress is deprecated, use [WaitForFunc] instead.
func (c *ActionClient) WatchOverallProgress(ctx context.Context, actions []*Action) (<-chan int, <-chan error) {
	errCh := make(chan error, len(actions))
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		previousGlobalProgress := 0
		progressByAction := make(map[int]int, len(actions))
		err := c.WaitForFunc(ctx, func(update *Action) error {
			switch update.Status {
			case ActionStatusRunning:
				progressByAction[update.ID] = update.Progress
			case ActionStatusSuccess:
				progressByAction[update.ID] = 100
			case ActionStatusError:
				progressByAction[update.ID] = 100
				errCh <- fmt.Errorf("action %d failed: %w", update.ID, update.Error())
			}

			// Compute global progress
			progressSum := 0
			for _, value := range progressByAction {
				progressSum += value
			}
			globalProgress := progressSum / len(actions)

			// Only send progress when it changed
			if globalProgress != 0 && globalProgress != previousGlobalProgress {
				sendProgress(progressCh, globalProgress)
				previousGlobalProgress = globalProgress
			}

			return nil
		}, actions...)

		if err != nil {
			errCh <- err
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
//
// Deprecated: WatchProgress is deprecated, use [WaitForFunc] instead.
func (c *ActionClient) WatchProgress(ctx context.Context, action *Action) (<-chan int, <-chan error) {
	errCh := make(chan error, 1)
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		err := c.WaitForFunc(ctx, func(update *Action) error {
			switch update.Status {
			case ActionStatusRunning:
				sendProgress(progressCh, update.Progress)
			case ActionStatusSuccess:
				sendProgress(progressCh, 100)
			case ActionStatusError:
				// Do not wrap the action error
				return update.Error()
			}

			return nil
		}, action)

		if err != nil {
			errCh <- err
		}
	}()

	return progressCh, errCh
}

// sendProgress allows the user to only read from the error channel and ignore any progress updates.
func sendProgress(progressCh chan int, p int) {
	select {
	case progressCh <- p:
		break
	default:
		break
	}
}
