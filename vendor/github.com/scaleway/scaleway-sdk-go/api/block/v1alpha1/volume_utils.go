package block

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultTimeout       = 5 * time.Minute
	defaultRetryInterval = 5 * time.Second
)

// WaitForVolumeRequest is used by WaitForVolume method.
type WaitForVolumeRequest struct {
	VolumeID      string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration

	// If set, will wait until this specific status has been reached or the
	// volume has an error status. This is useful when we need to wait for
	// the volume to transition from "in_use" to "available".
	TerminalStatus *VolumeStatus
}

// WaitForVolume waits for the volume to be in a "terminal state" before returning.
func (s *API) WaitForVolume(req *WaitForVolumeRequest, opts ...scw.RequestOption) (*Volume, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[VolumeStatus]struct{}{
		VolumeStatusError:   {},
		VolumeStatusLocked:  {},
		VolumeStatusDeleted: {},
	}

	if req.TerminalStatus != nil {
		terminalStatus[*req.TerminalStatus] = struct{}{}
	} else {
		terminalStatus[VolumeStatusAvailable] = struct{}{}
		terminalStatus[VolumeStatusInUse] = struct{}{}
	}

	volume, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetVolume(&GetVolumeRequest{
				VolumeID: req.VolumeID,
				Zone:     req.Zone,
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
		return nil, errors.Wrap(err, "waiting for volume failed")
	}
	return volume.(*Volume), nil
}

// WaitForVolumeAndReferencesRequest is used by WaitForVolumeAndReferences method.
type WaitForVolumeAndReferencesRequest struct {
	VolumeID      string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration

	VolumeTerminalStatus    *VolumeStatus
	ReferenceTerminalStatus *ReferenceStatus
}

// WaitForVolumeAndReferences waits for the volume and its references to be in a "terminal state" before returning.
func (s *API) WaitForVolumeAndReferences(req *WaitForVolumeAndReferencesRequest, opts ...scw.RequestOption) (*Volume, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[VolumeStatus]struct{}{
		VolumeStatusError:   {},
		VolumeStatusLocked:  {},
		VolumeStatusDeleted: {},
	}
	if req.VolumeTerminalStatus != nil {
		terminalStatus[*req.VolumeTerminalStatus] = struct{}{}
	} else {
		terminalStatus[VolumeStatusAvailable] = struct{}{}
		terminalStatus[VolumeStatusInUse] = struct{}{}
	}

	referenceTerminalStatus := map[ReferenceStatus]struct{}{
		ReferenceStatusError: {},
	}
	if req.ReferenceTerminalStatus != nil {
		referenceTerminalStatus[*req.ReferenceTerminalStatus] = struct{}{}
	} else {
		referenceTerminalStatus[ReferenceStatusAttached] = struct{}{}
		referenceTerminalStatus[ReferenceStatusDetached] = struct{}{}
	}

	volume, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			volume, err := s.GetVolume(&GetVolumeRequest{
				VolumeID: req.VolumeID,
				Zone:     req.Zone,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			referencesAreTerminal := true

			for _, reference := range volume.References {
				_, referenceIsTerminal := referenceTerminalStatus[reference.Status]
				referencesAreTerminal = referencesAreTerminal && referenceIsTerminal
			}

			_, isTerminal := terminalStatus[volume.Status]

			return volume, isTerminal && referencesAreTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for Volume failed")
	}

	return volume.(*Volume), nil
}
