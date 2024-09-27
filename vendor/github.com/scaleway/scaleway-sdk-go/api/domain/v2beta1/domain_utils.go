package domain

import (
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
	ErrCodeNoSuchDNSZone   = "NoSuchDNSZone"
	ErrCodeNoSuchDNSRecord = "NoSuchDNSRecord"
)

// WaitForDNSZoneRequest is used by WaitForDNSZone method.
type WaitForDNSZoneRequest struct {
	DNSZone       string
	DNSZones      []string
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

	dnsZone, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			listReq := &ListDNSZonesRequest{
				DNSZones: req.DNSZones,
			}

			if req.DNSZone != "" {
				listReq.DNSZone = &req.DNSZone
			}

			// listing dnsZone zones and take the first one
			DNSZones, err := s.ListDNSZones(listReq, opts...)
			if err != nil {
				return nil, false, err
			}

			if len(DNSZones.DNSZones) == 0 {
				return nil, true, errors.New(ErrCodeNoSuchDNSZone)
			}

			zone := DNSZones.DNSZones[0]

			_, isTerminal := terminalStatus[zone.Status]

			return zone, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for DNS failed")
	}

	return dnsZone.(*DNSZone), nil
}

// WaitForDNSRecordExistRequest is used by WaitForDNSRecordExist method.
type WaitForDNSRecordExistRequest struct {
	DNSZone       string
	RecordName    string
	RecordType    RecordType
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

func (s *API) WaitForDNSRecordExist(
	req *WaitForDNSRecordExistRequest,
	opts ...scw.RequestOption,
) (*Record, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	dns, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			// listing dns zone records and take the first one
			DNSRecords, err := s.ListDNSZoneRecords(&ListDNSZoneRecordsRequest{
				Name:    req.RecordName,
				Type:    req.RecordType,
				DNSZone: req.DNSZone,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			if DNSRecords.TotalCount == 0 {
				return nil, false, errors.New(ErrCodeNoSuchDNSRecord)
			}

			record := DNSRecords.Records[0]

			return record, true, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "check for DNS Record exist failed")
	}

	return dns.(*Record), nil
}
