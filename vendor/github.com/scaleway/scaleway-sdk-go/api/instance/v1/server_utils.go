package instance

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

const (
	defaultTimeout       = 5 * time.Minute
	defaultRetryInterval = 5 * time.Second
)

// CreateServer creates a server.
func (s *API) CreateServer(req *CreateServerRequest, opts ...scw.RequestOption) (*CreateServerResponse, error) {
	// If image is not a UUID we try to fetch it from marketplace.
	if req.Image != nil && !validation.IsUUID(*req.Image) {
		apiMarketplace := marketplace.NewAPI(s.client)

		getLocalImageByLabelRequest := &marketplace.GetLocalImageByLabelRequest{
			ImageLabel:     *req.Image,
			Zone:           req.Zone,
			CommercialType: req.CommercialType,
		}

		if bootVolumeType := getBootVolumeType(req.Volumes); bootVolumeType != nil {
			getLocalImageByLabelRequest.Type = *bootVolumeType
		}

		image, err := apiMarketplace.GetLocalImageByLabel(getLocalImageByLabelRequest)
		if err != nil {
			return nil, err
		}
		req.Image = scw.StringPtr(image.ID)
	}

	return s.createServer(req, opts...)
}

func getBootVolumeType(volumes map[string]*VolumeServerTemplate) *marketplace.LocalImageType {
	var bootVolumeType marketplace.LocalImageType
	foundBootVolume := false

	for _, volume := range volumes {
		if volume.Boot == nil {
			continue
		}
		if *volume.Boot {
			foundBootVolume = true
			switch volume.VolumeType {
			case VolumeVolumeTypeSbsVolume:
				bootVolumeType = marketplace.LocalImageTypeInstanceSbs
			case VolumeVolumeTypeLSSD:
				bootVolumeType = marketplace.LocalImageTypeInstanceLocal
			}
		}
	}

	if !foundBootVolume && len(volumes) > 0 {
		switch volumes["0"].VolumeType {
		case VolumeVolumeTypeSbsVolume:
			bootVolumeType = marketplace.LocalImageTypeInstanceSbs
		case VolumeVolumeTypeLSSD:
			bootVolumeType = marketplace.LocalImageTypeInstanceLocal
		}
	}
	return &bootVolumeType
}

// UpdateServer updates a server.
//
// Note: Implementation is thread-safe.
func (s *API) UpdateServer(req *UpdateServerRequest, opts ...scw.RequestOption) (*UpdateServerResponse, error) {
	defer lockServer(req.Zone, req.ServerID).Unlock()
	return s.updateServer(req, opts...)
}

// WaitForServerRequest is used by WaitForServer method.
type WaitForServerRequest struct {
	ServerID      string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForServer wait for the server to be in a "terminal state" before returning.
// This function can be used to wait for a server to be started for example.
func (s *API) WaitForServer(req *WaitForServerRequest, opts ...scw.RequestOption) (*Server, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[ServerState]struct{}{
		ServerStateStopped:        {},
		ServerStateStoppedInPlace: {},
		ServerStateLocked:         {},
		ServerStateRunning:        {},
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
			_, isTerminal := terminalStatus[res.Server.State]

			return res.Server, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}
	return server.(*Server), nil
}

// ServerActionAndWaitRequest is used by ServerActionAndWait method.
type ServerActionAndWaitRequest struct {
	ServerID string
	Zone     scw.Zone
	Action   ServerAction

	// Timeout: maximum time to wait before (default: 5 minutes)
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// ServerActionAndWait start an action and wait for the server to be in the correct "terminal state"
// expected by this action.
func (s *API) ServerActionAndWait(req *ServerActionAndWaitRequest, opts ...scw.RequestOption) error {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	_, err := s.ServerAction(&ServerActionRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
		Action:   req.Action,
	}, opts...)
	if err != nil {
		return err
	}

	finalServer, err := s.WaitForServer(&WaitForServerRequest{
		Zone:          req.Zone,
		ServerID:      req.ServerID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, opts...)
	if err != nil {
		return err
	}

	// check the action was properly executed
	expectedState := ServerState("unknown")
	switch req.Action {
	case ServerActionPoweron, ServerActionReboot:
		expectedState = ServerStateRunning
	case ServerActionPoweroff:
		expectedState = ServerStateStopped
	case ServerActionStopInPlace:
		expectedState = ServerStateStoppedInPlace
	}

	// backup can be performed from any state
	if expectedState != ServerState("unknown") && finalServer.State != expectedState {
		return errors.New("expected state %s but found %s: %s", expectedState, finalServer.State, finalServer.StateDetail)
	}

	return nil
}

// GetServerTypeRequest is used by GetServerType.
type GetServerTypeRequest struct {
	Zone scw.Zone
	Name string
}

// GetServerType get server type info by it's name.
func (s *API) GetServerType(req *GetServerTypeRequest) (*ServerType, error) {
	res, err := s.ListServersTypes(&ListServersTypesRequest{
		Zone: req.Zone,
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	if serverType, exist := res.Servers[req.Name]; exist {
		return serverType, nil
	}

	return nil, errors.New("could not find server type %q", req.Name)
}

// GetServerUserDataRequest is used by GetServerUserData method.
type GetServerUserDataRequest struct {
	Zone     scw.Zone `json:"-"`
	ServerID string   `json:"-"`

	// Key defines the user data key to get.
	Key string `json:"-"`
}

// GetServerUserData gets the content of a user data on a server for the given key.
func (s *API) GetServerUserData(req *GetServerUserDataRequest, opts ...scw.RequestOption) (io.Reader, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Zone == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if req.ServerID == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	if req.Key == "" {
		return nil, errors.New("field Key cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + req.ServerID + "/user_data/" + req.Key,
		Headers: http.Header{},
	}

	res := &bytes.Buffer{}

	err = s.client.Do(scwReq, res, opts...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SetServerUserDataRequest is used by SetServerUserData method.
type SetServerUserDataRequest struct {
	Zone     scw.Zone `json:"-"`
	ServerID string   `json:"-"`

	// Key defines the user data key to set.
	Key string `json:"-"`

	// Content defines the data to set.
	Content io.Reader
}

// SetServerUserData sets the content of a user data on a server for the given key.
func (s *API) SetServerUserData(req *SetServerUserDataRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Zone == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if req.ServerID == "" {
		return errors.New("field ServerID cannot be empty in request")
	}

	if req.Key == "" {
		return errors.New("field Key cannot be empty in request")
	}

	if req.Content == nil {
		return errors.New("field Content cannot be nil in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + req.ServerID + "/user_data/" + req.Key,
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req.Content)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}

	return nil
}

// GetAllServerUserDataRequest is used by GetAllServerUserData method.
type GetAllServerUserDataRequest struct {
	Zone     scw.Zone `json:"-"`
	ServerID string   `json:"-"`
}

// GetAllServerUserDataResponse is used by GetAllServerUserData method.
type GetAllServerUserDataResponse struct {
	UserData map[string]io.Reader `json:"-"`
}

// GetAllServerUserData gets all user data on a server.
func (s *API) GetAllServerUserData(req *GetAllServerUserDataRequest, opts ...scw.RequestOption) (*GetAllServerUserDataResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if req.ServerID == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	// get all user data keys
	allUserDataRes, err := s.ListServerUserData(&ListServerUserDataRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	res := &GetAllServerUserDataResponse{
		UserData: make(map[string]io.Reader, len(allUserDataRes.UserData)),
	}

	// build a map with all user data
	for _, key := range allUserDataRes.UserData {
		value, err := s.GetServerUserData(&GetServerUserDataRequest{
			Zone:     req.Zone,
			ServerID: req.ServerID,
			Key:      key,
		}, opts...)
		if err != nil {
			return nil, err
		}
		res.UserData[key] = value
	}

	return res, nil
}

// SetAllServerUserDataRequest is used by SetAllServerUserData method.
type SetAllServerUserDataRequest struct {
	Zone     scw.Zone `json:"-"`
	ServerID string   `json:"-"`

	// UserData defines all user data that will be set to the server.
	// This map is idempotent, it means that all the current data will be overwritten and
	// all keys not present in this map will be deleted.. All data will be removed if this map is nil.
	UserData map[string]io.Reader `json:"-"`
}

// SetAllServerUserData sets all user data on a server, it deletes every keys previously set.
func (s *API) SetAllServerUserData(req *SetAllServerUserDataRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Zone == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if req.ServerID == "" {
		return errors.New("field ServerID cannot be empty in request")
	}

	// get all current user data keys
	allUserDataRes, err := s.ListServerUserData(&ListServerUserDataRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
	}, opts...)
	if err != nil {
		return err
	}

	// delete all current user data
	for _, key := range allUserDataRes.UserData {
		_, exist := req.UserData[key]
		if exist {
			continue
		}
		err := s.DeleteServerUserData(&DeleteServerUserDataRequest{
			Zone:     req.Zone,
			ServerID: req.ServerID,
			Key:      key,
		}, opts...)
		if err != nil {
			return err
		}
	}

	// set all new user data
	for key, value := range req.UserData {
		err := s.SetServerUserData(&SetServerUserDataRequest{
			Zone:     req.Zone,
			ServerID: req.ServerID,
			Key:      key,
			Content:  value,
		}, opts...)
		if err != nil {
			return err
		}
	}

	return nil
}
