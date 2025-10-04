package block

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForSnapshotRequest is used by WaitForSnapshot method.
type WaitForSnapshotRequest struct {
	SnapshotID    string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration

	// If set, will wait until this specific status has been reached or the
	// snapshot has an error status. This is useful when we need to wait for
	// the snapshot to transition from "in_use" to "available".
	TerminalStatus *SnapshotStatus
}

// WaitForSnapshot wait for the snapshot to be in a "terminal state" before returning.
func (s *API) WaitForSnapshot(req *WaitForSnapshotRequest, opts ...scw.RequestOption) (*Snapshot, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[SnapshotStatus]struct{}{
		SnapshotStatusError:   {},
		SnapshotStatusLocked:  {},
		SnapshotStatusDeleted: {},
	}

	if req.TerminalStatus != nil {
		terminalStatus[*req.TerminalStatus] = struct{}{}
	} else {
		terminalStatus[SnapshotStatusAvailable] = struct{}{}
		terminalStatus[SnapshotStatusInUse] = struct{}{}
	}

	snapshot, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetSnapshot(&GetSnapshotRequest{
				SnapshotID: req.SnapshotID,
				Zone:       req.Zone,
			}, opts...)
			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for snapshot failed")
	}
	return snapshot.(*Snapshot), nil
}
