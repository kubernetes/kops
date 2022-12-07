package domain

import (
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRetryInterval = 15 * time.Second
	defaultTimeout       = 5 * time.Minute
)

const (
	// ErrCodeNoSuchDNSZone for service response error code
	//
	// The specified dns zone does not exist.
	ErrCodeNoSuchDNSZone = "NoSuchDNSZone"
)

// WaitForDNSZoneRequest is used by WaitForDNSZone method.
type WaitForDNSZoneRequest struct {
	DNSZone       string
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

func (s *API) WaitForDNSZone(
	req *WaitForDNSZoneRequest,
	opts ...scw.RequestOption,
) (*DNSZone, error) {

	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[DNSZoneStatus]struct{}{
		DNSZoneStatusActive: {},
		DNSZoneStatusLocked: {},
		DNSZoneStatusError:  {},
	}

	dns, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			// listing dns zones and take the first one
			DNSZones, err := s.ListDNSZones(&ListDNSZonesRequest{
				DNSZone: req.DNSZone,
			}, opts...)

			if err != nil {
				return nil, false, err
			}

			if len(DNSZones.DNSZones) == 0 {
				return nil, true, fmt.Errorf(ErrCodeNoSuchDNSZone)
			}

			Dns := DNSZones.DNSZones[0]

			_, isTerminal := terminalStatus[Dns.Status]

			return Dns, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})

	if err != nil {
		return nil, errors.Wrap(err, "waiting for DNS failed")
	}

	return dns.(*DNSZone), nil
}
