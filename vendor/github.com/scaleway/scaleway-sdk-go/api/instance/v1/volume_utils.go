package instance

import (
	goerrors "errors"
	"time"

	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
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
		Get: func() (any, bool, error) {
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

type unknownVolume struct {
	ID       string
	ServerID *string
	Type     VolumeVolumeType
}

type getUnknownVolumeRequest struct {
	Zone          scw.Zone
	VolumeID      string
	IsBlockVolume *bool
}

// getUnknownVolume is used to get a volume that can be either from instance or block API
func (s *API) getUnknownVolume(req *getUnknownVolumeRequest, opts ...scw.RequestOption) (*unknownVolume, error) {
	volume := &unknownVolume{
		ID: req.VolumeID,
	}

	// Try instance API
	if req.IsBlockVolume == nil || !*req.IsBlockVolume {
		getVolumeResponse, err := s.GetVolume(&GetVolumeRequest{
			Zone:     req.Zone,
			VolumeID: req.VolumeID,
		}, opts...)
		notFoundErr := &scw.ResourceNotFoundError{}
		if err != nil && !goerrors.As(err, &notFoundErr) {
			return nil, err
		}

		if getVolumeResponse != nil {
			if getVolumeResponse.Volume != nil && getVolumeResponse.Volume.Server != nil {
				volume.ServerID = &getVolumeResponse.Volume.Server.ID
			}
			volume.Type = getVolumeResponse.Volume.VolumeType
		}
	}

	if volume.Type == "" && (req.IsBlockVolume == nil || *req.IsBlockVolume) {
		getVolumeResponse, err := block.NewAPI(s.client).GetVolume(&block.GetVolumeRequest{
			Zone:     req.Zone,
			VolumeID: req.VolumeID,
		}, opts...)
		if err != nil {
			return nil, err
		}
		for _, reference := range getVolumeResponse.References {
			if reference.ProductResourceType == "instance_server" {
				volume.ServerID = &reference.ProductResourceID
			}
		}
		volume.Type = VolumeVolumeTypeSbsVolume
	}

	return volume, nil
}
