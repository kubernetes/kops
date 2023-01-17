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

type LBFrontend struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID          *string
	LBName      *string
	Zone        *string
	InboundPort *int32
	BackendID   *string
}

var _ fi.CloudupTask = &LBFrontend{}
var _ fi.CompareWithID = &LBFrontend{}
var _ fi.HasName = &LBFrontend{}

func (l *LBFrontend) CompareWithID() *string {
	return l.ID
}

func (l *LBFrontend) GetName() *string {
	return l.Name
}

func (l *LBFrontend) Find(context *fi.CloudupContext) (*LBFrontend, error) {
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

	frontendResponse, err := lbService.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
		Zone: scw.Zone(cloud.Zone()),
		LBID: loadBalancer.ID,
		Name: l.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("listing front-ends for load-balancer %s: %w", loadBalancer.ID, err)
	}
	if frontendResponse.TotalCount != 1 {
		return nil, nil
	}
	frontend := frontendResponse.Frontends[0]

	return &LBFrontend{
		Name:        fi.PtrTo(frontend.Name),
		Lifecycle:   l.Lifecycle,
		ID:          fi.PtrTo(frontend.ID),
		LBName:      fi.PtrTo(frontend.LB.Name),
		BackendID:   fi.PtrTo(frontend.Backend.ID),
		Zone:        fi.PtrTo(string(frontend.LB.Zone)),
		InboundPort: fi.PtrTo(frontend.InboundPort),
	}, nil
}

func (l *LBFrontend) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(l, context)
}

func (_ *LBFrontend) CheckChanges(actual, expected, changes *LBFrontend) error {
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
		if changes.BackendID != nil {
			return fi.CannotChangeField("Back-end ID")
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

func (l *LBFrontend) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *LBFrontend) error {
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

	// We fetch the ID of the back-end from the load-balancer's ID
	backendResponse, err := lbService.ListBackends(&lb.ZonedAPIListBackendsRequest{
		Zone: scw.Zone(fi.ValueOf(expected.Zone)),
		LBID: lbID,
	})
	if err != nil {
		return fmt.Errorf("listing back-ends for load-balancer %s: %w", lbID, err)
	}
	if backendResponse.TotalCount != 1 {
		return fmt.Errorf("expected 1 load-balancer back-end, got %d", backendResponse.TotalCount)
	}
	backendID := backendResponse.Backends[0].ID

	if actual != nil {

		_, err := lbService.UpdateFrontend(&lb.ZonedAPIUpdateFrontendRequest{
			Zone:        scw.Zone(fi.ValueOf(actual.Zone)),
			FrontendID:  fi.ValueOf(actual.ID),
			Name:        fi.ValueOf(actual.Name),
			InboundPort: fi.ValueOf(expected.InboundPort),
			BackendID:   backendID,
		})
		if err != nil {
			return fmt.Errorf("updating front-end for load-balancer %s: %w", fi.ValueOf(actual.LBName), err)
		}

		expected.BackendID = &backendID

	} else {

		frontendCreated, err := lbService.CreateFrontend(&lb.ZonedAPICreateFrontendRequest{
			Zone:        scw.Zone(fi.ValueOf(expected.Zone)),
			LBID:        lbID,
			Name:        fi.ValueOf(expected.Name),
			InboundPort: fi.ValueOf(expected.InboundPort),
			BackendID:   backendID,
		})
		if err != nil {
			return fmt.Errorf("creating front-end for load-balancer %s: %w", fi.ValueOf(expected.LBName), err)
		}

		expected.ID = &frontendCreated.ID
		expected.BackendID = &backendID
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
