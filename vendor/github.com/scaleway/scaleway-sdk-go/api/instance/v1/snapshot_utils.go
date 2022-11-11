package instance

import (
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForImageRequest is used by WaitForImage method.
type WaitForSnapshotRequest struct {
	SnapshotID    string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
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

	terminalStatus := map[SnapshotState]struct{}{
		SnapshotStateAvailable: {},
		SnapshotStateError:     {},
	}

	snapshot, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := s.GetSnapshot(&GetSnapshotRequest{
				SnapshotID: req.SnapshotID,
				Zone:       req.Zone,
			}, opts...)

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Snapshot.State]

			return res.Snapshot, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for snapshot failed")
	}
	return snapshot.(*Snapshot), nil
}

type UpdateSnapshotRequest struct {
	Zone       scw.Zone
	SnapshotID string
	Name       *string   `json:"name,omitempty"`
	Tags       *[]string `json:"tags,omitempty"`
}

type UpdateSnapshotResponse struct {
	Snapshot *Snapshot
}

func (s *API) UpdateSnapshot(req *UpdateSnapshotRequest, opts ...scw.RequestOption) (*UpdateSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return nil, errors.New("field SnapshotID cannot be empty in request")
	}

	getSnapshotResponse, err := s.GetSnapshot(&GetSnapshotRequest{
		Zone:       req.Zone,
		SnapshotID: req.SnapshotID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	setRequest := &setSnapshotRequest{
		SnapshotID:       getSnapshotResponse.Snapshot.ID,
		Zone:             getSnapshotResponse.Snapshot.Zone,
		ID:               getSnapshotResponse.Snapshot.ID,
		Name:             getSnapshotResponse.Snapshot.Name,
		CreationDate:     getSnapshotResponse.Snapshot.CreationDate,
		ModificationDate: getSnapshotResponse.Snapshot.ModificationDate,
		Organization:     getSnapshotResponse.Snapshot.Organization,
		Project:          getSnapshotResponse.Snapshot.Project,
	}

	// Override the values that need to be updated
	if req.Name != nil {
		setRequest.Name = *req.Name
	}

	if req.Tags != nil {
		setRequest.Tags = req.Tags
	}

	setRes, err := s.setSnapshot(setRequest, opts...)
	if err != nil {
		return nil, err
	}

	return &UpdateSnapshotResponse{
		Snapshot: setRes.Snapshot,
	}, nil
}
