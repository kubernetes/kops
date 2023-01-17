/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scalewaytasks

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

type LBBackend struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID                   *string
	LBName               *string
	Zone                 *string
	ForwardProtocol      *string
	ForwardPort          *int32
	ForwardPortAlgorithm *string
	StickySessions       *string
	ProxyProtocol        *string
}

var _ fi.CloudupTask = &LBBackend{}
var _ fi.CompareWithID = &LBBackend{}
var _ fi.HasName = &LBBackend{}

func (l *LBBackend) CompareWithID() *string {
	return l.ID
}

func (l *LBBackend) GetName() *string {
	return l.Name
}

func (l *LBBackend) Find(context *fi.CloudupContext) (*LBBackend, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	lbResponse, err := lbService.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone: scw.Zone(fi.ValueOf(l.Zone)),
		Name: l.LBName,
	})
	if err != nil {
		return nil, fmt.Errorf("listing load-balancers: %w", err)
	}
	if lbResponse.TotalCount != 1 {
		return nil, nil
	}
	loadBalancer := lbResponse.LBs[0]

	backendResponse, err := lbService.ListBackends(&lb.ZonedAPIListBackendsRequest{
		Zone: scw.Zone(cloud.Zone()),
		LBID: loadBalancer.ID,
		Name: l.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("listing back-ends for load-balancer %s: %w", loadBalancer.ID, err)
	}
	if backendResponse.TotalCount != 1 {
		return nil, nil
	}
	backend := backendResponse.Backends[0]

	return &LBBackend{
		Name:                 fi.PtrTo(backend.Name),
		Lifecycle:            l.Lifecycle,
		ID:                   fi.PtrTo(backend.ID),
		LBName:               fi.PtrTo(backend.LB.Name),
		Zone:                 fi.PtrTo(string(backend.LB.Zone)),
		ForwardProtocol:      fi.PtrTo(string(backend.ForwardProtocol)),
		ForwardPort:          fi.PtrTo(backend.ForwardPort),
		ForwardPortAlgorithm: fi.PtrTo(string(backend.ForwardPortAlgorithm)),
		StickySessions:       fi.PtrTo(string(backend.StickySessions)),
		ProxyProtocol:        fi.PtrTo(string(backend.ProxyProtocol)),
	}, nil
}

func (l *LBBackend) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(l, context)
}

func (_ *LBBackend) CheckChanges(actual, expected, changes *LBBackend) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.LBName != nil {
			return fi.CannotChangeField("Load-balancer name")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.LBName == nil {
			return fi.RequiredField("Load-Balancer name")
		}
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (l *LBBackend) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *LBBackend) error {
	lbService := t.Cloud.LBService()

	// We fetch the ID of the LB from its name
	lbResponse, err := lbService.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone: scw.Zone(fi.ValueOf(expected.Zone)),
		Name: l.LBName,
	}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("getting load-balancer %s: %w", fi.ValueOf(l.LBName), err)
	}
	if lbResponse.TotalCount != 1 {
		return fmt.Errorf("expected 1 load-balancer, got %d", lbResponse.TotalCount)
	}
	lbID := lbResponse.LBs[0].ID

	if actual != nil {

		_, err := lbService.UpdateBackend(&lb.ZonedAPIUpdateBackendRequest{
			Zone:                 scw.Zone(fi.ValueOf(actual.Zone)),
			BackendID:            fi.ValueOf(actual.ID),
			Name:                 fi.ValueOf(actual.Name),
			ForwardProtocol:      lb.Protocol(fi.ValueOf(expected.ForwardProtocol)),
			ForwardPort:          fi.ValueOf(expected.ForwardPort),
			ForwardPortAlgorithm: lb.ForwardPortAlgorithm(fi.ValueOf(expected.ForwardPortAlgorithm)),
			StickySessions:       lb.StickySessionsType(fi.ValueOf(expected.StickySessions)),
			ProxyProtocol:        lb.ProxyProtocol(fi.ValueOf(expected.ProxyProtocol)),
		})
		if err != nil {
			return fmt.Errorf("updating back-end for load-balancer %s: %w", fi.ValueOf(actual.LBName), err)
		}

		expected.ID = actual.ID

	} else {

		req := &lb.ZonedAPICreateBackendRequest{
			Zone:                 scw.Zone(fi.ValueOf(expected.Zone)),
			LBID:                 lbID,
			Name:                 fi.ValueOf(expected.Name),
			ForwardProtocol:      lb.Protocol(fi.ValueOf(expected.ForwardProtocol)),
			ForwardPort:          fi.ValueOf(expected.ForwardPort),
			ForwardPortAlgorithm: lb.ForwardPortAlgorithm(fi.ValueOf(expected.ForwardPortAlgorithm)),
			StickySessions:       lb.StickySessionsType(fi.ValueOf(expected.StickySessions)),
			HealthCheck: &lb.HealthCheck{
				CheckMaxRetries: 5,
				TCPConfig:       &lb.HealthCheckTCPConfig{},
				Port:            fi.ValueOf(expected.ForwardPort),
				CheckTimeout:    scw.TimeDurationPtr(3000),
				CheckDelay:      scw.TimeDurationPtr(1001),
			},
			ProxyProtocol: lb.ProxyProtocol(fi.ValueOf(expected.ProxyProtocol)),
		}

		backendCreated, err := lbService.CreateBackend(req)
		if err != nil {
			return fmt.Errorf("creating back-end for load-balancer %s: %w", fi.ValueOf(expected.LBName), err)
		}

		expected.ID = &backendCreated.ID

	}

	_, err = lbService.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID: lbID,
		Zone: scw.Zone(fi.ValueOf(expected.Zone)),
	})
	if err != nil {
		return fmt.Errorf("waiting for load-balancer %s: %w", fi.ValueOf(expected.LBName), err)
	}

	return nil
}
