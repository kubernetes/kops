package domain

import (
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
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

// WaitForOrderDomainRequest is used by WaitForOrderDomain method.
type WaitForOrderDomainRequest struct {
	Domain        string
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForOrderDomain waits until the domain reaches a terminal status.
func (s *RegistrarAPI) WaitForOrderDomain(
	req *WaitForOrderDomainRequest,
	opts ...scw.RequestOption,
) (*Domain, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	// Terminal statuses indicating success or error.
	terminalStatuses := map[DomainStatus]struct{}{
		DomainStatusActive:      {},
		DomainStatusExpired:     {},
		DomainStatusLocked:      {},
		DomainStatusCreateError: {},
		DomainStatusRenewError:  {},
		DomainStatusXferError:   {},
	}

	var lastStatus DomainStatus

	domain, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			resp, err := s.GetDomain(&RegistrarAPIGetDomainRequest{
				Domain: req.Domain,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			lastStatus = resp.Status

			if _, isTerminal := terminalStatuses[resp.Status]; isTerminal {
				return resp, true, nil
			}
			return resp, false, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("waiting for domain %s failed, last known status: %s", req.Domain, lastStatus))
	}

	return domain.(*Domain), nil
}

// WaitForAutoRenewStatusRequest defines the parameters for waiting on the auto‑renew feature.
type WaitForAutoRenewStatusRequest struct {
	Domain        string         // The domain to wait for.
	Timeout       *time.Duration // Optional timeout.
	RetryInterval *time.Duration // Optional retry interval.
}

// WaitForAutoRenewStatus polls the domain until its auto‑renew feature reaches a terminal state
// (either "enabled" or "disabled"). It uses GetDomain() to fetch the current status.
func (s *RegistrarAPI) WaitForAutoRenewStatus(req *WaitForAutoRenewStatusRequest, opts ...scw.RequestOption) (*Domain, error) {
	// Use default timeout and retry interval if not provided.
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	// Terminal statuses for auto_renew: enabled or disabled.
	terminalStatuses := map[DomainFeatureStatus]struct{}{
		DomainFeatureStatusEnabled:  {},
		DomainFeatureStatusDisabled: {},
	}

	var lastStatus DomainFeatureStatus

	domainResult, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			resp, err := s.GetDomain(&RegistrarAPIGetDomainRequest{
				Domain: req.Domain,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			lastStatus = resp.AutoRenewStatus
			if _, isTerminal := terminalStatuses[resp.AutoRenewStatus]; isTerminal {
				return resp, true, nil
			}
			return resp, false, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("waiting for auto_renew to reach a terminal state for domain %s failed, last known status: %s", req.Domain, lastStatus))
	}
	return domainResult.(*Domain), nil
}

// WaitForDNSSECStatusRequest defines the parameters for waiting on the DNSSEC feature.
type WaitForDNSSECStatusRequest struct {
	Domain        string         // The domain to wait for.
	Timeout       *time.Duration // Optional timeout.
	RetryInterval *time.Duration // Optional retry interval.
}

// WaitForDNSSECStatus polls the domain until its DNSSEC feature reaches a terminal state
// (either "enabled" or "disabled"). It uses GetDomain() to fetch the current status.
func (s *RegistrarAPI) WaitForDNSSECStatus(req *WaitForDNSSECStatusRequest, opts ...scw.RequestOption) (*Domain, error) {
	// Use default timeout and retry interval if not provided.
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	// Terminal statuses for DNSSEC: enabled or disabled.
	terminalStatuses := map[DomainFeatureStatus]struct{}{
		DomainFeatureStatusEnabled:  {},
		DomainFeatureStatusDisabled: {},
	}

	var lastStatus DomainFeatureStatus

	domainResult, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			// Retrieve the domain.
			resp, err := s.GetDomain(&RegistrarAPIGetDomainRequest{
				Domain: req.Domain,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			// Check the current DNSSEC status.
			lastStatus = resp.Dnssec.Status
			if _, isTerminal := terminalStatuses[resp.Dnssec.Status]; isTerminal {
				return resp, true, nil
			}
			return resp, false, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("waiting for dnssec to reach a terminal state for domain %s failed, last known status: %s", req.Domain, lastStatus))
	}
	return domainResult.(*Domain), nil
}
