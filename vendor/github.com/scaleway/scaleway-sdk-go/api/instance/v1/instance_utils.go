package instance

import (
	"encoding/json"
	"fmt"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"sync"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var (
	resourceLock sync.Map
)

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
	})
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
	})
	if err != nil {
		return nil, err
	}

	return &DetachIPResponse{IP: ipResponse.IP}, nil
}

// AttachVolumeRequest contains the parameters to attach a volume to a server
type AttachVolumeRequest struct {
	Zone     scw.Zone `json:"-"`
	ServerID string   `json:"-"`
	VolumeID string   `json:"-"`
}

// AttachVolumeResponse contains the updated server after attaching a volume
type AttachVolumeResponse struct {
	Server *Server `json:"-"`
}

// volumesToVolumeTemplates converts a map of *Volume to a map of *VolumeTemplate
// so it can be used in a UpdateServer request
func volumesToVolumeTemplates(volumes map[string]*VolumeServer) map[string]*VolumeServerTemplate {
	volumeTemplates := map[string]*VolumeServerTemplate{}
	for key, volume := range volumes {
		volumeTemplates[key] = &VolumeServerTemplate{
			ID:   volume.ID,
			Name: volume.Name,
		}
	}
	return volumeTemplates
}

// AttachVolume attaches a volume to a server
//
// Note: Implementation is thread-safe.
func (s *API) AttachVolume(req *AttachVolumeRequest, opts ...scw.RequestOption) (*AttachVolumeResponse, error) {
	defer lockServer(req.Zone, req.ServerID).Unlock()
	// get server with volumes
	getServerResponse, err := s.GetServer(&GetServerRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
	})
	if err != nil {
		return nil, err
	}
	volumes := getServerResponse.Server.Volumes

	newVolumes := volumesToVolumeTemplates(volumes)

	// add volume to volumes list
	// We loop through all the possible volume keys (0 to len(volumes))
	// to find a non existing key and assign it to the requested volume.
	// A key should always be found. However we return an error if no keys were found.
	found := false
	for i := 0; i <= len(volumes); i++ {
		key := fmt.Sprintf("%d", i)
		if _, ok := newVolumes[key]; !ok {
			newVolumes[key] = &VolumeServerTemplate{
				ID: req.VolumeID,
				// name is ignored on this PATCH
				Name: req.VolumeID,
			}
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("could not find key to attach volume %s", req.VolumeID)
	}

	// update server
	updateServerResponse, err := s.updateServer(&UpdateServerRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
		Volumes:  &newVolumes,
	})
	if err != nil {
		return nil, err
	}

	return &AttachVolumeResponse{Server: updateServerResponse.Server}, nil
}

// DetachVolumeRequest contains the parameters to detach a volume from a server
type DetachVolumeRequest struct {
	Zone     scw.Zone `json:"-"`
	VolumeID string   `json:"-"`
}

// DetachVolumeResponse contains the updated server after detaching a volume
type DetachVolumeResponse struct {
	Server *Server `json:"-"`
}

// DetachVolume detaches a volume from a server
//
// Note: Implementation is thread-safe.
func (s *API) DetachVolume(req *DetachVolumeRequest, opts ...scw.RequestOption) (*DetachVolumeResponse, error) {
	// get volume
	getVolumeResponse, err := s.GetVolume(&GetVolumeRequest{
		Zone:     req.Zone,
		VolumeID: req.VolumeID,
	})
	if err != nil {
		return nil, err
	}
	if getVolumeResponse.Volume == nil {
		return nil, errors.New("expected volume to have value in response")
	}
	if getVolumeResponse.Volume.Server == nil {
		return nil, errors.New("volume should be attached to a server")
	}
	serverID := getVolumeResponse.Volume.Server.ID

	defer lockServer(req.Zone, serverID).Unlock()
	// get server with volumes
	getServerResponse, err := s.GetServer(&GetServerRequest{
		Zone:     req.Zone,
		ServerID: serverID,
	})
	if err != nil {
		return nil, err
	}
	volumes := getServerResponse.Server.Volumes
	// remove volume from volumes list
	for key, volume := range volumes {
		if volume.ID == req.VolumeID {
			delete(volumes, key)
		}
	}

	newVolumes := volumesToVolumeTemplates(volumes)

	// update server
	updateServerResponse, err := s.updateServer(&UpdateServerRequest{
		Zone:     req.Zone,
		ServerID: serverID,
		Volumes:  &newVolumes,
	})
	if err != nil {
		return nil, err
	}

	return &DetachVolumeResponse{Server: updateServerResponse.Server}, nil
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeSetTotalCount(totalCount int) {
	r.TotalCount = uint32(totalCount)
}

// UnsafeSetTotalCount should not be used
// Internal usage only
func (r *ListBootscriptsResponse) UnsafeSetTotalCount(totalCount int) {
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

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListServersTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListServersTypesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListServersTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	if r.Servers == nil {
		r.Servers = make(map[string]*ServerType, len(results.Servers))
	}

	for name, serverType := range results.Servers {
		r.Servers[name] = serverType
	}

	r.TotalCount += uint32(len(results.Servers))
	return uint32(len(results.Servers)), nil
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
		Get: func() (interface{}, bool, error) {
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
		Get: func() (interface{}, bool, error) {
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
