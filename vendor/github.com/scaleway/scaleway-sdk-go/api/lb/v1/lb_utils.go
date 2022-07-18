package lb

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRetryInterval = 2 * time.Second
	defaultTimeout       = 5 * time.Minute
)

// WaitForLBRequest is used by WaitForLb method.
type WaitForLBRequest struct {
	LBID          string
	Region        scw.Region
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForLb waits for the lb to be in a "terminal state" before returning.
// This function can be used to wait for a lb to be ready for example.
func (s *API) WaitForLb(req *WaitForLBRequest, opts ...scw.RequestOption) (*LB, error) {
	return waitForLb(req.Timeout, req.RetryInterval, func() (*LB, error) {
		return s.GetLB(&GetLBRequest{
			Region: req.Region,
			LBID:   req.LBID,
		}, opts...)
	})
}

// ZonedAPIWaitForLBRequest is used by WaitForLb method.
type ZonedAPIWaitForLBRequest struct {
	LBID          string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForLb waits for the lb to be in a "terminal state" before returning.
// This function can be used to wait for a lb to be ready for example.
func (s *ZonedAPI) WaitForLb(req *ZonedAPIWaitForLBRequest, opts ...scw.RequestOption) (*LB, error) {
	return waitForLb(req.Timeout, req.RetryInterval, func() (*LB, error) {
		return s.GetLB(&ZonedAPIGetLBRequest{
			Zone: req.Zone,
			LBID: req.LBID,
		}, opts...)
	})
}

func waitForLb(timeout *time.Duration, retryInterval *time.Duration, getLB func() (*LB, error)) (*LB, error) {
	t := defaultTimeout
	if timeout != nil {
		t = *timeout
	}
	r := defaultRetryInterval
	if retryInterval != nil {
		r = *retryInterval
	}

	terminalStatus := map[LBStatus]struct{}{
		LBStatusReady:   {},
		LBStatusStopped: {},
		LBStatusError:   {},
		LBStatusLocked:  {},
	}

	lb, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := getLB()

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, isTerminal, nil
		},
		Timeout:          t,
		IntervalStrategy: async.LinearIntervalStrategy(r),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for lb failed")
	}
	return lb.(*LB), nil
}

// ZonedAPIWaitForLBInstancesRequest is used by WaitForLb method.
type ZonedAPIWaitForLBInstancesRequest struct {
	LBID          string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForLbInstances waits for the lb to be in a "terminal state" and the attached instances before returning.
func (s *ZonedAPI) WaitForLbInstances(req *ZonedAPIWaitForLBInstancesRequest, opts ...scw.RequestOption) (*LB, error) {
	return waitForLbInstances(req.Timeout, req.RetryInterval, func() (*LB, error) {
		return s.GetLB(&ZonedAPIGetLBRequest{
			Zone: req.Zone,
			LBID: req.LBID,
		}, opts...)
	})
}

func waitForLbInstances(timeout *time.Duration, retryInterval *time.Duration, getLB func() (*LB, error)) (*LB, error) {
	t := defaultTimeout
	if timeout != nil {
		t = *timeout
	}
	r := defaultRetryInterval
	if retryInterval != nil {
		r = *retryInterval
	}

	terminalLBStatus := map[LBStatus]struct{}{
		LBStatusReady:   {},
		LBStatusStopped: {},
		LBStatusError:   {},
		LBStatusLocked:  {},
	}

	terminalInstanceStatus := map[InstanceStatus]struct{}{
		InstanceStatusReady:   {},
		InstanceStatusStopped: {},
		InstanceStatusError:   {},
		InstanceStatusLocked:  {},
	}

	lb, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := getLB()
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalLBStatus[res.Status]

			for _, i := range res.Instances {
				_, isInstanceTerminal := terminalInstanceStatus[i.Status]
				if !isInstanceTerminal {
					return res, isTerminal, nil
				}
			}

			return res, isTerminal, nil
		},
		Timeout:          t,
		IntervalStrategy: async.LinearIntervalStrategy(r),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for lb failed")
	}

	return lb.(*LB), nil
}

// ZonedAPIWaitForLBPNRequest is used by WaitForLBPN method.
type ZonedAPIWaitForLBPNRequest struct {
	LBID          string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

func waitForPNLb(timeout *time.Duration, retryInterval *time.Duration, getPNs func() ([]*PrivateNetwork, error)) ([]*PrivateNetwork, error) {
	t := defaultTimeout
	if timeout != nil {
		t = *timeout
	}
	r := defaultRetryInterval
	if retryInterval != nil {
		r = *retryInterval
	}

	terminalStatus := map[PrivateNetworkStatus]struct{}{
		PrivateNetworkStatusReady: {},
		PrivateNetworkStatusError: {},
	}

	pn, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			pns, err := getPNs()

			for _, pn := range pns {
				if err != nil {
					return nil, false, err
				}
				//wait at the first not terminal state
				_, isTerminal := terminalStatus[pn.Status]
				if !isTerminal {
					return pns, isTerminal, nil
				}
			}
			return pns, true, nil
		},
		Timeout:          t,
		IntervalStrategy: async.LinearIntervalStrategy(r),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for attachment failed")
	}
	return pn.([]*PrivateNetwork), nil
}

// WaitForLBPN waits for the private_network attached status on a load balancer
// to be in a "terminal state" before returning.
// This function can be used to wait for an attached private_network to be ready for example.
func (s *ZonedAPI) WaitForLBPN(req *ZonedAPIWaitForLBPNRequest, opts ...scw.RequestOption) ([]*PrivateNetwork, error) {
	return waitForPNLb(req.Timeout, req.RetryInterval, func() ([]*PrivateNetwork, error) {
		lbPNs, err := s.ListLBPrivateNetworks(&ZonedAPIListLBPrivateNetworksRequest{
			Zone: req.Zone,
			LBID: req.LBID,
		}, opts...)
		if err != nil {
			return nil, err
		}

		return lbPNs.PrivateNetwork, nil
	})
}

// ZonedAPIWaitForLBCertificateRequest is used by WaitForLbCertificate method.
type ZonedAPIWaitForLBCertificateRequest struct {
	CertID        string
	Zone          scw.Zone
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForLBCertificate waits for the certificate to be in a "terminal state"
func (s *ZonedAPI) WaitForLBCertificate(req *ZonedAPIWaitForLBCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	return waitForLBCertificate(req.Timeout, req.RetryInterval, func() (*Certificate, error) {
		return s.GetCertificate(&ZonedAPIGetCertificateRequest{
			Zone:          req.Zone,
			CertificateID: req.CertID,
		}, opts...)
	})
}

func waitForLBCertificate(timeout *time.Duration, retryInterval *time.Duration, getCertificate func() (*Certificate, error)) (*Certificate, error) {
	t := defaultTimeout
	if timeout != nil {
		t = *timeout
	}
	r := defaultRetryInterval
	if retryInterval != nil {
		r = *retryInterval
	}

	terminalStatus := map[CertificateStatus]struct{}{
		CertificateStatusError: {},
		CertificateStatusReady: {},
	}

	crt, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := getCertificate()

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, isTerminal, nil
		},
		Timeout:          t,
		IntervalStrategy: async.LinearIntervalStrategy(r),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for lb failed")
	}
	return crt.(*Certificate), nil
}
