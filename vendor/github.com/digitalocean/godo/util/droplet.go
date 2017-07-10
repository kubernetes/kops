package util

import (
	"fmt"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/context"
)

const (
	// activeFailure is the amount of times we can fail before deciding
	// the check for active is a total failure. This can help account
	// for servers randomly not answering.
	activeFailure = 3
)

// WaitForActive waits for a droplet to become active
func WaitForActive(ctx context.Context, client *godo.Client, monitorURI string) error {
	if len(monitorURI) == 0 {
		return fmt.Errorf("create had no monitor uri")
	}

	completed := false
	failCount := 0
	for !completed {
		action, _, err := client.DropletActions.GetByURI(ctx, monitorURI)

		if err != nil {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			if failCount <= activeFailure {
				failCount++
				continue
			}
			return err
		}

		switch action.Status {
		case godo.ActionInProgress:
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return err
			}
		case godo.ActionCompleted:
			completed = true
		default:
			return fmt.Errorf("unknown status: [%s]", action.Status)
		}
	}

	return nil
}
