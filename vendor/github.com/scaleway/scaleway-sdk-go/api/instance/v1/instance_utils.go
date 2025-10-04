package instance

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var resourceLock sync.Map

// lockResource locks a resource from a specific resourceID
func lockResource(resourceID string) *sync.Mutex {
	v, _ := resourceLock.LoadOrStore(resourceID, &sync.Mutex{})
	mutex := v.(*sync.Mutex)
	mutex.Lock()
	return mutex
}

// lockServer locks a server from its zone and its ID
func lockServer(zone scw.Zone, serverID string) *sync.Mutex {
	return lockResource(fmt.Sprint("server", zone, serverID))
}

// AttachIPRequest contains the parameters to attach an IP to a server
//
// Deprecated: UpdateIPRequest should be used instead
type AttachIPRequest struct {
	Zone     scw.Zone `json:"-"`
	IP       string   `json:"-"`
	ServerID string   `json:"server_id"`
}

// AttachIPResponse contains the updated IP after attaching
//
// Deprecated: UpdateIPResponse should be used instead
type AttachIPResponse struct {
	IP *IP
}

// AttachIP attaches an IP to a server.
//
// Deprecated: UpdateIP() should be used instead
func (s *API) AttachIP(req *AttachIPRequest, opts ...scw.RequestOption) (*AttachIPResponse, error) {
	ipResponse, err := s.UpdateIP(&UpdateIPRequest{
		Zone:   req.Zone,
		IP:     req.IP,
		Server: &NullableStringValue{Value: req.ServerID},
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &AttachIPResponse{IP: ipResponse.IP}, nil
}

// DetachIPRequest contains the parameters to detach an IP from a server
//
// Deprecated: UpdateIPRequest should be used instead
type DetachIPRequest struct {
	Zone scw.Zone `json:"-"`
	IP   string   `json:"-"`
}

// DetachIPResponse contains the updated IP after detaching
//
// Deprecated: UpdateIPResponse should be used instead
type DetachIPResponse struct {
	IP *IP
}

// DetachIP detaches an IP from a server.
//
// Deprecated: UpdateIP() should be used instead
func (s *API) DetachIP(req *DetachIPRequest, opts ...scw.RequestOption) (*DetachIPResponse, error) {
	ipResponse, err := s.UpdateIP(&UpdateIPRequest{
		Zone:   req.Zone,
		IP:     req.IP,
		Server: &NullableStringValue{Null: true},
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &DetachIPResponse{IP: ipResponse.IP}, nil
}

// AttachVolumeRequest contains the parameters to attach a volume to a server
// Deprecated by AttachServerVolumeRequest
type AttachVolumeRequest struct {
	Zone     scw.Zone `json:"-"`
	ServerID string   `json:"-"`
	VolumeID string   `json:"-"`
}

// AttachVolumeResponse contains the updated server after attaching a volume
// Deprecated by AttachServerVolumeResponse
type AttachVolumeResponse struct {
	Server *Server `json:"-"`
}

// AttachVolume attaches a volume to a server
//
// Note: Implementation is thread-safe.
// Deprecated by AttachServerVolume provided by instance API
func (s *API) AttachVolume(req *AttachVolumeRequest, opts ...scw.RequestOption) (*AttachVolumeResponse, error) {
	defer lockServer(req.Zone, req.ServerID).Unlock()
	// check where the volume comes from
	volume, err := s.getUnknownVolume(&getUnknownVolumeRequest{
		Zone:     req.Zone,
		VolumeID: req.VolumeID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	attachServerVolumeReq := &AttachServerVolumeRequest{
		Zone:       req.Zone,
		ServerID:   req.ServerID,
		VolumeID:   req.VolumeID,
		VolumeType: AttachServerVolumeRequestVolumeType(volume.Type),
	}

	resp, err := s.AttachServerVolume(attachServerVolumeReq, opts...)
	if err != nil {
		return nil, err
	}

	return &AttachVolumeResponse{Server: resp.Server}, nil
}

// DetachVolumeRequest contains the parameters to detach a volume from a server
// Deprecated by DetachServerVolumeRequest
type DetachVolumeRequest struct {
	Zone     scw.Zone `json:"-"`
	VolumeID string   `json:"-"`
	// IsBlockVolume should be set to true if volume is from block API,
	// can be set to false if volume is from instance API,
	// if left nil both API will be tried
	IsBlockVolume *bool `json:"-"`
}

// DetachVolumeResponse contains the updated server after detaching a volume
// Deprecated by DetachServerVolumeResponse
type DetachVolumeResponse struct {
	Server *Server `json:"-"`
}

// DetachVolume detaches a volume from a server
//
// Note: Implementation is thread-safe.
// Deprecated by DetachServerVolume provided by instance API
func (s *API) DetachVolume(req *DetachVolumeRequest, opts ...scw.RequestOption) (*DetachVolumeResponse, error) {
	volume, err := s.getUnknownVolume(&getUnknownVolumeRequest{
		Zone:     req.Zone,
		VolumeID: req.VolumeID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	if volume.ServerID == nil {
		return nil, errors.New("volume should be attached to a server")
	}

	defer lockServer(req.Zone, *volume.ServerID).Unlock()

	resp, err := s.DetachServerVolume(&DetachServerVolumeRequest{
		Zone:     req.Zone,
		ServerID: *volume.ServerID,
		VolumeID: volume.ID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &DetachVolumeResponse{Server: resp.Server}, nil
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListSecurityGroupRulesResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListSecurityGroupsResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListServersTypesResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

func (v *NullableStringValue) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		v.Null = true
		return nil
	}

	var tmp string
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	v.Null = false
	v.Value = tmp
	return nil
}

func (v *NullableStringValue) MarshalJSON() ([]byte, error) {
	if v.Null {
		return []byte("null"), nil
	}
	return json.Marshal(v.Value)
}

// WaitForPrivateNICRequest is used by WaitForPrivateNIC method.
type WaitForPrivateNICRequest struct {
	ServerID      string
	PrivateNicID  string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForPrivateNIC wait for the private network to be in a "terminal state" before returning.
// This function can be used to wait for the private network to be attached for example.
func (s *API) WaitForPrivateNIC(req *WaitForPrivateNICRequest, opts ...scw.RequestOption) (*PrivateNIC, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[PrivateNICState]struct{}{
		PrivateNICStateAvailable:    {},
		PrivateNICStateSyncingError: {},
	}

	pn, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetPrivateNIC(&GetPrivateNICRequest{
				ServerID:     req.ServerID,
				Zone:         req.Zone,
				PrivateNicID: req.PrivateNicID,
			}, opts...)
			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.PrivateNic.State]

			return res.PrivateNic, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}
	return pn.(*PrivateNIC), nil
}

// WaitForMACAddressRequest is used by WaitForMACAddress method.
type WaitForMACAddressRequest struct {
	ServerID      string
	PrivateNicID  string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForMACAddress wait for the MAC address be assigned on instance before returning.
// This function can be used to wait for the private network to be attached for example.
func (s *API) WaitForMACAddress(req *WaitForMACAddressRequest, opts ...scw.RequestOption) (*PrivateNIC, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	pn, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetPrivateNIC(&GetPrivateNICRequest{
				ServerID:     req.ServerID,
				Zone:         req.Zone,
				PrivateNicID: req.PrivateNicID,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			if len(res.PrivateNic.MacAddress) > 0 {
				return res.PrivateNic, true, err
			}

			return res.PrivateNic, false, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}
	return pn.(*PrivateNIC), nil
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *GetServerTypesAvailabilityResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// WaitForServerRDPPasswordRequest is used by WaitForServerRDPPassword method.
type WaitForServerRDPPasswordRequest struct {
	ServerID      string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForServerRDPPassword wait for an RDP password to be generated for an instance before returning.
// This function can be used to wait for a windows instance to boot up.
func (s *API) WaitForServerRDPPassword(req *WaitForServerRDPPasswordRequest, opts ...scw.RequestOption) (*Server, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	server, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetServer(&GetServerRequest{
				ServerID: req.ServerID,
				Zone:     req.Zone,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			if res.Server.AdminPasswordEncryptedValue != nil && *res.Server.AdminPasswordEncryptedValue != "" {
				return res.Server, true, err
			}

			return res.Server, false, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}
	return server.(*Server), nil
}

type WaitForServerFileSystemRequest struct {
	ServerID      string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

func (s *API) WaitForServerFileSystem(req *WaitForServerFileSystemRequest, opts ...scw.RequestOption) (*Server, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[ServerFilesystemState]struct{}{
		ServerFilesystemStateAvailable:    {},
		ServerFilesystemStateUnknownState: {},
	}

	result, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetServer(&GetServerRequest{
				ServerID: req.ServerID,
				Zone:     req.Zone,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			allFilesystemsReady := true
			for _, fs := range res.Server.Filesystems {
				if _, ok := terminalStatus[fs.State]; !ok {
					allFilesystemsReady = false
					break
				}
			}

			return res.Server, allFilesystemsReady, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server filesystems failed")
	}

	return result.(*Server), nil
}
