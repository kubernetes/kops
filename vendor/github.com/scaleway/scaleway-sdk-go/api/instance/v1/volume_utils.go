package instance

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForImageRequest is used by WaitForImage method.
type WaitForVolumeRequest struct {
	VolumeID      string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForSnapshot wait for the snapshot to be in a "terminal state" before returning.
func (s *API) WaitForVolume(req *WaitForVolumeRequest, opts ...scw.RequestOption) (*Volume, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[VolumeState]struct{}{
		VolumeStateAvailable: {},
		VolumeStateError:     {},
	}

	volume, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := s.GetVolume(&GetVolumeRequest{
				VolumeID: req.VolumeID,
				Zone:     req.Zone,
			}, opts...)

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Volume.State]

			return res.Volume, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for volume failed")
	}
	return volume.(*Volume), nil
}
