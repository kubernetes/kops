package vpcgw

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultTimeout       = 5 * time.Minute
	defaultRetryInterval = 15 * time.Second
)

// WaitForGatewayRequest is used by WaitForGateway method
type WaitForGatewayRequest struct {
	GatewayID     string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForGateway waits for the gateway to be in a "terminal state" before returning.
// This function can be used to wait for a gateway to be ready for example.
func (s *API) WaitForGateway(req *WaitForGatewayRequest, opts ...scw.RequestOption) (*Gateway, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[GatewayStatus]struct{}{
		GatewayStatusRunning: {},
		GatewayStatusDeleted: {},
		GatewayStatusUnknown: {},
		GatewayStatusFailed:  {},
	}

	gateway, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			ns, err := s.GetGateway(&GetGatewayRequest{
				Zone:      req.Zone,
				GatewayID: req.GatewayID,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalStatus[ns.Status]

			return ns, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for gateway failed")
	}

	return gateway.(*Gateway), nil
}

// WaitForGatewayNetworkRequest is used by WaitForGatewayNetwork method
type WaitForGatewayNetworkRequest struct {
	GatewayNetworkID string
	Zone             scw.Zone
	Timeout          *time.Duration
	RetryInterval    *time.Duration
}

// WaitForGatewayNetwork waits for the gateway network to be in a "terminal state" before returning.
// This function can be used to wait for a gateway network to be ready for example.
func (s *API) WaitForGatewayNetwork(req *WaitForGatewayNetworkRequest, opts ...scw.RequestOption) (*GatewayNetwork, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[GatewayNetworkStatus]struct{}{
		GatewayNetworkStatusReady:   {},
		GatewayNetworkStatusUnknown: {},
		GatewayNetworkStatusDeleted: {},
		GatewayNetworkStatusCreated: {},
	}

	gatewayNetwork, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			ns, err := s.GetGatewayNetwork(&GetGatewayNetworkRequest{
				Zone:             req.Zone,
				GatewayNetworkID: req.GatewayNetworkID,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalStatus[ns.Status]

			return ns, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for gateway network failed")
	}

	return gatewayNetwork.(*GatewayNetwork), nil
}

// WaitForDHCPEntriesRequest is used by WaitForDHCPEntries method
type WaitForDHCPEntriesRequest struct {
	GatewayNetworkID *string
	MacAddress       string

	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForDHCPEntries waits for at least one dhcp entry with the correct mac address.
// This function can be used to wait for an instance to use dhcp
func (s *API) WaitForDHCPEntries(req *WaitForDHCPEntriesRequest, opts ...scw.RequestOption) (*ListDHCPEntriesResponse, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	dhcpEntries, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			entries, err := s.ListDHCPEntries(&ListDHCPEntriesRequest{
				Zone:             req.Zone,
				GatewayNetworkID: req.GatewayNetworkID,
				MacAddress:       &req.MacAddress,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			containsMacAddress := false
			for _, entry := range entries.DHCPEntries {
				if entry.MacAddress == req.MacAddress {
					containsMacAddress = true
					break
				}
			}

			return entries, containsMacAddress, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for gateway network failed")
	}

	return dhcpEntries.(*ListDHCPEntriesResponse), nil
}
